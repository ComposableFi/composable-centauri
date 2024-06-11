package transfermiddleware_test

import (
	"encoding/binary"
	"encoding/json"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	customibctesting "github.com/notional-labs/composable/v6/app/ibctesting"
	"github.com/stretchr/testify/suite"
)

type PacketMetadata struct {
	Forward *ForwardMetadata `json:"forward"`
}

type ForwardMetadata struct {
	Receiver string        `json:"receiver,omitempty"`
	Port     string        `json:"port,omitempty"`
	Channel  string        `json:"channel,omitempty"`
	Timeout  time.Duration `json:"timeout,omitempty"`
	Retries  *uint8        `json:"retries,omitempty"`
	Next     string        `json:"next,omitempty"`
}

// TODO: use testsuite here.
type TransferMiddlewareTestSuite struct {
	suite.Suite

	coordinator *customibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *customibctesting.TestChain
	chainB *customibctesting.TestChain
	chainC *customibctesting.TestChain
}

func (suite *TransferMiddlewareTestSuite) SetupTest() {
	suite.coordinator = customibctesting.NewCoordinator(suite.T(), 4)
	suite.chainA = suite.coordinator.GetChain(customibctesting.GetChainID(1))
	suite.chainB = suite.coordinator.GetChain(customibctesting.GetChainID(2))
	suite.chainC = suite.coordinator.GetChain(customibctesting.GetChainID(3))
}

func NewTransferPath(chainA, chainB *customibctesting.TestChain) *customibctesting.Path {
	path := customibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = customibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = customibctesting.TransferPort
	path.EndpointA.ChannelConfig.Version = transfertypes.Version
	path.EndpointB.ChannelConfig.Version = transfertypes.Version

	return path
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TransferMiddlewareTestSuite))
}

var keyCounter uint64

// we need to make this deterministic (same every test run), as encoded address size and thus gas cost,
// depends on the actual bytes (due to ugly CanonicalAddress encoding)
func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.AccAddress) {
	keyCounter++
	seed := make([]byte, 8)
	binary.BigEndian.PutUint64(seed, keyCounter)

	key := ed25519.GenPrivKeyFromSecret(seed)
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}

func RandomAccountAddress(_ testing.TB) sdk.AccAddress {
	_, _, addr := keyPubAddr()
	return addr
}

func RandomBech32AccountAddress(tb testing.TB) string {
	tb.Helper()
	return RandomAccountAddress(tb).String()
}

func (suite *TransferMiddlewareTestSuite) TestTransferWithPFM_ErrorAck() {
	var (
		transferAmount = sdkmath.NewInt(1000000000)
		timeoutHeight  = clienttypes.NewHeight(1, 110)
		pathAtoB       *customibctesting.Path
		pathBtoC       *customibctesting.Path
		nativeDenom    = "ppica"
		ibcDenom       = "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"
		assetID        = sdk.DefaultBondDenom
		// expDenom       = "ibc/3262D378E1636BE287EC355990D229DCEB828F0C60ED5049729575E235C60E8B"
	)

	suite.SetupTest()
	pathAtoB = NewTransferPath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(pathAtoB)
	pathBtoC = NewTransferPath(suite.chainB, suite.chainC)
	suite.coordinator.Setup(pathBtoC)
	// Add parachain token info
	chainBtransMiddleware := suite.chainB.TransferMiddleware()
	err := chainBtransMiddleware.AddParachainIBCInfo(suite.chainB.GetContext(), ibcDenom, pathAtoB.EndpointB.ChannelID, nativeDenom, assetID)
	suite.Require().NoError(err)

	params := transfertypes.Params{
		SendEnabled:    true,
		ReceiveEnabled: false,
	}

	// set send params
	suite.chainC.App.GetTransferKeeper().SetParams(suite.chainC.GetContext(), params)
	senderAOriginalBalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())

	testAcc := RandomAccountAddress(suite.T())
	timeOut := 10 * time.Minute
	retries := uint8(0)
	// Build MEMOtransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	memo := PacketMetadata{
		Forward: &ForwardMetadata{
			Receiver: testAcc.String(),
			Port:     pathBtoC.EndpointA.ChannelConfig.PortID,
			Channel:  pathBtoC.EndpointA.ChannelID,
			Timeout:  timeOut,
			Retries:  &retries,
			Next:     "",
		},
	}
	memoMarshalled, err := json.Marshal(&memo)
	suite.Require().NoError(err)

	msg := transfertypes.NewMsgTransfer(
		pathAtoB.EndpointA.ChannelConfig.PortID,
		pathAtoB.EndpointA.ChannelID,
		sdk.NewCoin(assetID, transferAmount),
		suite.chainA.SenderAccount.GetAddress().String(),
		testAcc.String(),
		timeoutHeight,
		0,
		string(memoMarshalled),
	)
	_, err = suite.chainA.SendMsgs(msg)
	suite.Require().NoError(err)
	suite.Require().NoError(err, pathAtoB.EndpointB.UpdateClient())

	// then
	suite.Require().Equal(1, len(suite.chainA.PendingSendPackets))
	suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))
	// relay packet
	sendingPacket := suite.chainA.PendingSendPackets[0]
	suite.coordinator.IncrementTime()
	suite.coordinator.CommitBlock(suite.chainA)
	err = pathAtoB.EndpointB.UpdateClient()
	suite.Require().NoError(err)

	err = pathAtoB.EndpointB.RecvPacket(sendingPacket)
	suite.Require().NoError(err)
	suite.chainA.PendingSendPackets = nil

	// Check after first Hop
	senderABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
	suite.Require().Equal(senderAOriginalBalance.Sub(sdk.NewCoin(assetID, transferAmount)), senderABalance)

	escrowIbcDenomAddress := transfertypes.GetEscrowAddress(pathAtoB.EndpointB.ChannelConfig.PortID, pathAtoB.EndpointB.ChannelID)
	escrowIbcDenomAddressBalance := suite.chainB.AllBalances(escrowIbcDenomAddress)
	suite.Require().Equal(sdk.NewCoins(sdk.NewCoin(ibcDenom, transferAmount)), escrowIbcDenomAddressBalance)

	escrowNativeDenomAddress := transfertypes.GetEscrowAddress(pathBtoC.EndpointA.ChannelConfig.PortID, pathBtoC.EndpointA.ChannelID)
	escrowNativeDenomAddressBalance := suite.chainB.AllBalances(escrowNativeDenomAddress)
	suite.Require().Equal(sdk.NewCoins(sdk.NewCoin(nativeDenom, transferAmount)), escrowNativeDenomAddressBalance)

	// then should have a packet from B to C
	suite.Require().Equal(1, len(suite.chainB.PendingSendPackets))
	suite.Require().Equal(0, len(suite.chainC.PendingSendPackets))

	// relay packet
	sendingPacket = suite.chainB.PendingSendPackets[0]
	suite.coordinator.IncrementTime()
	suite.coordinator.CommitBlock(suite.chainB)
	err = pathBtoC.EndpointB.UpdateClient()
	suite.Require().NoError(err)

	err = pathBtoC.EndpointB.RecvPacket(sendingPacket)
	suite.Require().NoError(err)
	suite.chainB.PendingSendPackets = nil

	// relay ack C to B
	suite.Require().Equal(1, len(suite.chainC.PendingAckPackets))
	ack := suite.chainC.PendingAckPackets[0]
	suite.coordinator.IncrementTime()
	suite.coordinator.CommitBlock(suite.chainC)
	err = pathBtoC.EndpointA.UpdateClient()
	suite.Require().NoError(err)
	// relay failed ack
	err = pathBtoC.EndpointA.AcknowledgePacket(ack.Packet, ack.Ack)
	suite.Require().NoError(err)
	suite.chainC.PendingAckPackets = nil

	// relay ack B to A
	suite.Require().Equal(1, len(suite.chainB.PendingAckPackets))
	ack = suite.chainB.PendingAckPackets[0]
	suite.coordinator.IncrementTime()
	suite.coordinator.CommitBlock(suite.chainB)
	err = pathAtoB.EndpointA.UpdateClient()
	suite.Require().NoError(err)

	err = pathAtoB.EndpointA.AcknowledgePacket(ack.Packet, ack.Ack)
	suite.Require().NoError(err)
	suite.chainB.PendingAckPackets = nil

	escrowIbcDenomAddress = transfertypes.GetEscrowAddress(pathAtoB.EndpointB.ChannelConfig.PortID, pathAtoB.EndpointB.ChannelID)
	escrowIbcDenomAddressBalance = suite.chainB.AllBalances(escrowIbcDenomAddress)
	suite.Require().Empty(escrowIbcDenomAddressBalance)

	escrowNativeDenomAddress = transfertypes.GetEscrowAddress(pathBtoC.EndpointA.ChannelConfig.PortID, pathBtoC.EndpointA.ChannelID)
	escrowNativeDenomAddressBalance = suite.chainB.AllBalances(escrowNativeDenomAddress)
	suite.Require().Empty(escrowNativeDenomAddressBalance)

	balance := suite.chainB.AllBalances(testAcc)
	suite.Require().Empty(balance)

	balance = suite.chainC.AllBalances(testAcc)
	suite.Require().Empty(balance)

	balance = suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
	suite.Require().Equal(senderAOriginalBalance, balance)
}

func (suite *TransferMiddlewareTestSuite) TestTransferWithPFM() {
	var (
		transferAmount = sdkmath.NewInt(1000000000)
		// when transfer via sdk transfer from A (module) -> B (contract)
		timeoutHeight = clienttypes.NewHeight(1, 110)
		pathAtoB      *customibctesting.Path
		pathBtoC      *customibctesting.Path
		nativeDenom   = "ppica"
		ibcDenom      = "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"
		assetID       = sdk.DefaultBondDenom
		expDenom      = "ibc/3262D378E1636BE287EC355990D229DCEB828F0C60ED5049729575E235C60E8B"
	)

	testCases := []struct {
		name string
	}{
		{
			"Success case A -> B -> C",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			pathAtoB = NewTransferPath(suite.chainA, suite.chainB)
			suite.coordinator.Setup(pathAtoB)
			pathBtoC = NewTransferPath(suite.chainB, suite.chainC)
			suite.coordinator.Setup(pathBtoC)

			// Add parachain token info
			senderAOriginalBalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
			chainBtransMiddleware := suite.chainB.TransferMiddleware()
			err := chainBtransMiddleware.AddParachainIBCInfo(suite.chainB.GetContext(), ibcDenom, pathAtoB.EndpointB.ChannelID, nativeDenom, assetID)
			suite.Require().NoError(err)

			testAcc := RandomAccountAddress(suite.T())
			timeOut := 10 * time.Minute
			retries := uint8(0)
			// Build MEMO
			memo := PacketMetadata{
				Forward: &ForwardMetadata{
					Receiver: testAcc.String(),
					Port:     pathBtoC.EndpointA.ChannelConfig.PortID,
					Channel:  pathBtoC.EndpointA.ChannelID,
					Timeout:  timeOut,
					Retries:  &retries,
					Next:     "",
				},
			}
			memoMarshalled, err := json.Marshal(&memo)
			suite.Require().NoError(err)

			intermediaryOriginalBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())

			msg := transfertypes.NewMsgTransfer(
				pathAtoB.EndpointA.ChannelConfig.PortID,
				pathAtoB.EndpointA.ChannelID,
				sdk.NewCoin(assetID, transferAmount),
				suite.chainA.SenderAccount.GetAddress().String(),
				suite.chainB.SenderAccount.GetAddress().String(),
				timeoutHeight,
				0,
				string(memoMarshalled),
			)
			_, err = suite.chainA.SendMsgs(msg)
			suite.Require().NoError(err)
			suite.Require().NoError(err, pathAtoB.EndpointB.UpdateClient())

			// then
			suite.Require().Equal(1, len(suite.chainA.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))
			// relay packet
			sendingPacket := suite.chainA.PendingSendPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainA)
			err = pathAtoB.EndpointB.UpdateClient()
			suite.Require().NoError(err)

			err = pathAtoB.EndpointB.RecvPacket(sendingPacket)
			suite.Require().NoError(err)
			suite.chainA.PendingSendPackets = nil
			// then should have a packet from B to C
			suite.Require().Equal(1, len(suite.chainB.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainC.PendingSendPackets))

			// relay packet
			sendingPacket = suite.chainB.PendingSendPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainB)
			err = pathBtoC.EndpointB.UpdateClient()
			suite.Require().NoError(err)

			err = pathBtoC.EndpointB.RecvPacket(sendingPacket)
			suite.Require().NoError(err)
			suite.chainB.PendingSendPackets = nil

			// Check escrow address
			escrowIbcDenomAddress := transfertypes.GetEscrowAddress(pathAtoB.EndpointB.ChannelConfig.PortID, pathAtoB.EndpointB.ChannelID)
			escrowIbcDenomAddressBalance := suite.chainB.AllBalances(escrowIbcDenomAddress)
			expectBalance := sdk.NewCoins(sdk.NewCoin(ibcDenom, transferAmount))
			suite.Require().Equal(expectBalance, escrowIbcDenomAddressBalance)

			escrowNativeDenomAddress := transfertypes.GetEscrowAddress(pathBtoC.EndpointA.ChannelConfig.PortID, pathBtoC.EndpointA.ChannelID)
			escrowNativeDenomAddressBalance := suite.chainB.AllBalances(escrowNativeDenomAddress)
			expectBalance = sdk.NewCoins(sdk.NewCoin(nativeDenom, transferAmount))
			suite.Require().Equal(expectBalance, escrowNativeDenomAddressBalance)

			// relay ack C to B
			suite.Require().Equal(1, len(suite.chainC.PendingAckPackets))
			ack := suite.chainC.PendingAckPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainC)
			err = pathBtoC.EndpointA.UpdateClient()
			suite.Require().NoError(err)

			err = pathBtoC.EndpointA.AcknowledgePacket(ack.Packet, ack.Ack)
			suite.Require().NoError(err)
			suite.chainC.PendingAckPackets = nil

			// relay ack B to A
			suite.Require().Equal(1, len(suite.chainB.PendingAckPackets))
			ack = suite.chainB.PendingAckPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainB)
			err = pathAtoB.EndpointA.UpdateClient()
			suite.Require().NoError(err)

			err = pathAtoB.EndpointA.AcknowledgePacket(ack.Packet, ack.Ack)
			suite.Require().NoError(err)
			suite.chainB.PendingAckPackets = nil

			// Check escrow address
			escrowIbcDenomAddress = transfertypes.GetEscrowAddress(pathAtoB.EndpointB.ChannelConfig.PortID, pathAtoB.EndpointB.ChannelID)
			escrowIbcDenomAddressBalance = suite.chainB.AllBalances(escrowIbcDenomAddress)
			expectBalance = sdk.NewCoins(sdk.NewCoin(ibcDenom, transferAmount))
			suite.Require().Equal(expectBalance, escrowIbcDenomAddressBalance)

			escrowNativeDenomAddress = transfertypes.GetEscrowAddress(pathBtoC.EndpointA.ChannelConfig.PortID, pathBtoC.EndpointA.ChannelID)
			escrowNativeDenomAddressBalance = suite.chainB.AllBalances(escrowNativeDenomAddress)
			expectBalance = sdk.NewCoins(sdk.NewCoin(nativeDenom, transferAmount))
			suite.Require().Equal(expectBalance, escrowNativeDenomAddressBalance)

			intermediaryBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
			suite.Require().Equal(intermediaryOriginalBalance, intermediaryBalance)

			senderABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
			suite.Require().Equal(senderAOriginalBalance.Sub(sdk.NewCoin(assetID, transferAmount)), senderABalance)

			expBalance := sdk.NewCoins(sdk.NewCoin(expDenom, transferAmount))
			balance := suite.chainC.AllBalances(testAcc)
			suite.Require().Equal(expBalance, balance)
		})
	}
}

func (suite *TransferMiddlewareTestSuite) TestTransferWithPFMReverse_ErrorAck() {
	var (
		transferAmount = sdkmath.NewInt(1000000000)
		// when transfer via sdk transfer from A (module) -> B (contract)
		timeoutHeight = clienttypes.NewHeight(1, 110)
		pathAtoB      *customibctesting.Path
		pathBtoC      *customibctesting.Path
		nativeDenom   = "ppica"
		ibcDenom      = "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"
		assetID       = sdk.DefaultBondDenom
		expDenom      = "ibc/3262D378E1636BE287EC355990D229DCEB828F0C60ED5049729575E235C60E8B"
	)

	testCases := []struct {
		name string
	}{
		{
			"Success transfer from Picasso -> Composable -> Osmosis and error reverse from Osmosis -> Composable -> Picasso",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			pathAtoB = NewTransferPath(suite.chainA, suite.chainB)
			suite.coordinator.Setup(pathAtoB)
			pathBtoC = NewTransferPath(suite.chainB, suite.chainC)
			suite.coordinator.Setup(pathBtoC)
			senderAOriginalBalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
			senderBOriginalBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
			senderCOriginalBalance := suite.chainC.AllBalances(suite.chainC.SenderAccount.GetAddress())
			_ = senderAOriginalBalance
			_ = senderBOriginalBalance
			_ = senderCOriginalBalance
			// Add parachain token info
			chainBtransMiddleware := suite.chainB.TransferMiddleware()
			err := chainBtransMiddleware.AddParachainIBCInfo(suite.chainB.GetContext(), ibcDenom, pathAtoB.EndpointB.ChannelID, nativeDenom, assetID)
			suite.Require().NoError(err)

			// Disable receiveEnabled on chain A so it will return error ack
			params := transfertypes.Params{
				SendEnabled:    true,
				ReceiveEnabled: false,
			}
			// set send params
			suite.chainA.App.GetTransferKeeper().SetParams(suite.chainA.GetContext(), params)

			timeOut := 10 * time.Minute
			retries := uint8(0)
			// Build MEMO
			memo := PacketMetadata{
				Forward: &ForwardMetadata{
					Receiver: suite.chainC.SenderAccount.GetAddress().String(),
					Port:     pathBtoC.EndpointA.ChannelConfig.PortID,
					Channel:  pathBtoC.EndpointA.ChannelID,
					Timeout:  timeOut,
					Retries:  &retries,
					Next:     "",
				},
			}
			memoMarshalled, err := json.Marshal(&memo)
			suite.Require().NoError(err)

			msg := transfertypes.NewMsgTransfer(
				pathAtoB.EndpointA.ChannelConfig.PortID,
				pathAtoB.EndpointA.ChannelID,
				sdk.NewCoin(assetID, transferAmount),
				suite.chainA.SenderAccount.GetAddress().String(),
				suite.chainB.SenderAccount.GetAddress().String(),
				timeoutHeight,
				0,
				string(memoMarshalled),
			)
			_, err = suite.chainA.SendMsgs(msg)
			suite.Require().NoError(err)
			suite.Require().NoError(err, pathAtoB.EndpointB.UpdateClient())

			// then
			suite.Require().Equal(1, len(suite.chainA.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))
			// relay packet
			sendingPacket := suite.chainA.PendingSendPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainA)
			err = pathAtoB.EndpointB.UpdateClient()
			suite.Require().NoError(err)

			err = pathAtoB.EndpointB.RecvPacket(sendingPacket)
			suite.Require().NoError(err)
			suite.chainA.PendingSendPackets = nil
			// Check escrow address
			escrowIbcDenomAddress := transfertypes.GetEscrowAddress(pathAtoB.EndpointB.ChannelConfig.PortID, pathAtoB.EndpointB.ChannelID)
			escrowIbcDenomAddressBalance := suite.chainB.AllBalances(escrowIbcDenomAddress)
			expBalance := sdk.NewCoins(sdk.NewCoin(ibcDenom, transferAmount))
			suite.Require().Equal(expBalance, escrowIbcDenomAddressBalance)

			escrowNativeDenomAddress := transfertypes.GetEscrowAddress(pathBtoC.EndpointA.ChannelConfig.PortID, pathBtoC.EndpointA.ChannelID)
			escrowNativeDenomAddressBalance := suite.chainB.AllBalances(escrowNativeDenomAddress)
			expBalance = sdk.NewCoins(sdk.NewCoin(nativeDenom, transferAmount))
			suite.Require().Equal(expBalance, escrowNativeDenomAddressBalance)

			// then should have a packet from B to C
			suite.Require().Equal(1, len(suite.chainB.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainC.PendingSendPackets))

			// relay packet
			sendingPacket = suite.chainB.PendingSendPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainB)
			err = pathBtoC.EndpointB.UpdateClient()
			suite.Require().NoError(err)

			err = pathBtoC.EndpointB.RecvPacket(sendingPacket)
			suite.Require().NoError(err)
			suite.chainB.PendingSendPackets = nil

			// relay ack C to B
			suite.Require().Equal(1, len(suite.chainC.PendingAckPackets))
			ack := suite.chainC.PendingAckPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainC)
			err = pathBtoC.EndpointA.UpdateClient()
			suite.Require().NoError(err)

			err = pathBtoC.EndpointA.AcknowledgePacket(ack.Packet, ack.Ack)
			suite.Require().NoError(err)
			suite.chainC.PendingAckPackets = nil

			// relay ack B to A
			suite.Require().Equal(1, len(suite.chainB.PendingAckPackets))
			ack = suite.chainB.PendingAckPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainB)
			err = pathAtoB.EndpointA.UpdateClient()
			suite.Require().NoError(err)

			err = pathAtoB.EndpointA.AcknowledgePacket(ack.Packet, ack.Ack)
			suite.Require().NoError(err)
			suite.chainB.PendingAckPackets = nil

			senderBCurrentBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
			suite.Require().Equal(senderBOriginalBalance, senderBCurrentBalance)

			escrowIbcDenomAddress = transfertypes.GetEscrowAddress(pathAtoB.EndpointB.ChannelConfig.PortID, pathAtoB.EndpointB.ChannelID)
			escrowIbcDenomAddressBalance = suite.chainB.AllBalances(escrowIbcDenomAddress)
			expBalance = sdk.NewCoins(sdk.NewCoin(ibcDenom, transferAmount))
			suite.Require().Equal(expBalance, escrowIbcDenomAddressBalance)

			escrowNativeDenomAddress = transfertypes.GetEscrowAddress(pathBtoC.EndpointA.ChannelConfig.PortID, pathBtoC.EndpointA.ChannelID)
			escrowNativeDenomAddressBalance = suite.chainB.AllBalances(escrowNativeDenomAddress)
			expBalance = sdk.NewCoins(sdk.NewCoin(nativeDenom, transferAmount))
			suite.Require().Equal(expBalance, escrowNativeDenomAddressBalance)

			balance := suite.chainC.AllBalances(suite.chainC.SenderAccount.GetAddress())
			receiveBalance := balance.AmountOf(expDenom)
			suite.Require().Equal(transferAmount, receiveBalance)

			senderAOriginalBalance = suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
			senderBOriginalBalance = suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
			senderCOriginalBalance = suite.chainC.AllBalances(suite.chainC.SenderAccount.GetAddress())
			// transfer back from osmosis to picasso
			memo = PacketMetadata{
				Forward: &ForwardMetadata{
					Receiver: suite.chainA.SenderAccount.GetAddress().String(),
					Port:     pathAtoB.EndpointB.ChannelConfig.PortID,
					Channel:  pathAtoB.EndpointB.ChannelID,
					Timeout:  timeOut,
					Retries:  &retries,
					Next:     "",
				},
			}

			memoMarshalled, err = json.Marshal(&memo)
			suite.Require().NoError(err)

			msg = transfertypes.NewMsgTransfer(
				pathBtoC.EndpointB.ChannelConfig.PortID,
				pathBtoC.EndpointB.ChannelID,
				sdk.NewCoin(expDenom, transferAmount),
				suite.chainC.SenderAccount.GetAddress().String(),
				suite.chainB.SenderAccount.GetAddress().String(),
				timeoutHeight,
				0,
				string(memoMarshalled),
			)

			_, err = suite.chainC.SendMsgs(msg)
			suite.Require().NoError(err)
			suite.Require().NoError(err, pathBtoC.EndpointA.UpdateClient())

			// then
			suite.Require().Equal(1, len(suite.chainC.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))
			// relay packet
			sendingPacket = suite.chainC.PendingSendPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainC)
			err = pathBtoC.EndpointA.UpdateClient()
			suite.Require().NoError(err)

			err = pathBtoC.EndpointA.RecvPacket(sendingPacket)
			suite.Require().NoError(err)
			suite.chainC.PendingSendPackets = nil
			// Check escrow address
			escrowIbcDenomAddress = transfertypes.GetEscrowAddress(pathAtoB.EndpointB.ChannelConfig.PortID, pathAtoB.EndpointB.ChannelID)
			escrowIbcDenomAddressBalance = suite.chainB.AllBalances(escrowIbcDenomAddress)
			suite.Require().Empty(escrowIbcDenomAddressBalance)

			escrowNativeDenomAddress = transfertypes.GetEscrowAddress(pathBtoC.EndpointA.ChannelConfig.PortID, pathBtoC.EndpointA.ChannelID)
			escrowNativeDenomAddressBalance = suite.chainB.AllBalances(escrowNativeDenomAddress)
			suite.Require().Empty(escrowNativeDenomAddressBalance)

			// then should have a packet from B to A
			suite.Require().Equal(1, len(suite.chainB.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainA.PendingSendPackets))

			// relay packet
			sendingPacket = suite.chainB.PendingSendPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainB)
			err = pathAtoB.EndpointA.UpdateClient()
			suite.Require().NoError(err)

			err = pathAtoB.EndpointA.RecvPacket(sendingPacket)
			suite.Require().NoError(err)
			suite.chainB.PendingSendPackets = nil

			// relay error ack A to B
			suite.Require().Equal(1, len(suite.chainA.PendingAckPackets))
			ack = suite.chainA.PendingAckPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainA)
			err = pathAtoB.EndpointB.UpdateClient()
			suite.Require().NoError(err)

			err = pathAtoB.EndpointB.AcknowledgePacket(ack.Packet, ack.Ack)
			suite.Require().NoError(err)
			suite.chainA.PendingAckPackets = nil

			// relay error ack B to C
			suite.Require().Equal(1, len(suite.chainB.PendingAckPackets))
			ack = suite.chainB.PendingAckPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainB)
			err = pathBtoC.EndpointB.UpdateClient()
			suite.Require().NoError(err)

			err = pathBtoC.EndpointB.AcknowledgePacket(ack.Packet, ack.Ack)
			suite.Require().NoError(err)
			suite.chainB.PendingAckPackets = nil

			senderACurrentBalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
			suite.Require().Equal(senderAOriginalBalance, senderACurrentBalance)

			senderBCurrentBalance = suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
			suite.Require().Equal(senderBOriginalBalance, senderBCurrentBalance)

			senderCCurrentBalance := suite.chainC.AllBalances(suite.chainC.SenderAccount.GetAddress())
			suite.Require().Equal(senderCOriginalBalance, senderCCurrentBalance)

			escrowIbcDenomAddress = transfertypes.GetEscrowAddress(pathAtoB.EndpointB.ChannelConfig.PortID, pathAtoB.EndpointB.ChannelID)
			escrowIbcDenomAddressBalance = suite.chainB.AllBalances(escrowIbcDenomAddress)
			expBalance = sdk.NewCoins(sdk.NewCoin(ibcDenom, transferAmount))
			suite.Require().Equal(expBalance, escrowIbcDenomAddressBalance)

			escrowNativeDenomAddress = transfertypes.GetEscrowAddress(pathBtoC.EndpointA.ChannelConfig.PortID, pathBtoC.EndpointA.ChannelID)
			escrowNativeDenomAddressBalance = suite.chainB.AllBalances(escrowNativeDenomAddress)
			expBalance = sdk.NewCoins(sdk.NewCoin(nativeDenom, transferAmount))
			suite.Require().Equal(expBalance, escrowNativeDenomAddressBalance)
		})
	}
}

func (suite *TransferMiddlewareTestSuite) TestTransferWithPFMReverse() {
	var (
		transferAmount = sdkmath.NewInt(1000000000)
		// when transfer via sdk transfer from A (module) -> B (contract)
		timeoutHeight = clienttypes.NewHeight(1, 110)
		pathAtoB      *customibctesting.Path
		pathBtoC      *customibctesting.Path
		nativeDenom   = "ppica"
		ibcDenom      = "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"
		assetID       = sdk.DefaultBondDenom
		expDenom      = "ibc/3262D378E1636BE287EC355990D229DCEB828F0C60ED5049729575E235C60E8B"
	)

	testCases := []struct {
		name string
	}{
		{
			"Success case Picasso -> Composable -> Osmosis and reverse from Osmosis -> Composable -> Picasso",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			pathAtoB = NewTransferPath(suite.chainA, suite.chainB)
			suite.coordinator.Setup(pathAtoB)
			pathBtoC = NewTransferPath(suite.chainB, suite.chainC)
			suite.coordinator.Setup(pathBtoC)
			senderAOriginalBalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
			senderBOriginalBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
			senderCOriginalBalance := suite.chainC.AllBalances(suite.chainC.SenderAccount.GetAddress())
			// Add parachain token info
			chainBtransMiddleware := suite.chainB.TransferMiddleware()
			err := chainBtransMiddleware.AddParachainIBCInfo(suite.chainB.GetContext(), ibcDenom, pathAtoB.EndpointB.ChannelID, nativeDenom, assetID)
			suite.Require().NoError(err)

			timeOut := 10 * time.Minute
			retries := uint8(0)
			// Build MEMO
			memo := PacketMetadata{
				Forward: &ForwardMetadata{
					Receiver: suite.chainC.SenderAccount.GetAddress().String(),
					Port:     pathBtoC.EndpointA.ChannelConfig.PortID,
					Channel:  pathBtoC.EndpointA.ChannelID,
					Timeout:  timeOut,
					Retries:  &retries,
					Next:     "",
				},
			}
			memoMarshalled, err := json.Marshal(&memo)
			suite.Require().NoError(err)

			msg := transfertypes.NewMsgTransfer(
				pathAtoB.EndpointA.ChannelConfig.PortID,
				pathAtoB.EndpointA.ChannelID,
				sdk.NewCoin(assetID, transferAmount),
				suite.chainA.SenderAccount.GetAddress().String(),
				suite.chainB.SenderAccount.GetAddress().String(),
				timeoutHeight,
				0,
				string(memoMarshalled),
			)
			_, err = suite.chainA.SendMsgs(msg)
			suite.Require().NoError(err)
			suite.Require().NoError(err, pathAtoB.EndpointB.UpdateClient())

			// then
			suite.Require().Equal(1, len(suite.chainA.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))
			// relay packet
			sendingPacket := suite.chainA.PendingSendPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainA)
			err = pathAtoB.EndpointB.UpdateClient()
			suite.Require().NoError(err)

			err = pathAtoB.EndpointB.RecvPacket(sendingPacket)
			suite.Require().NoError(err)
			suite.chainA.PendingSendPackets = nil
			// Check escrow address
			escrowIbcDenomAddress := transfertypes.GetEscrowAddress(pathAtoB.EndpointB.ChannelConfig.PortID, pathAtoB.EndpointB.ChannelID)
			escrowIbcDenomAddressBalance := suite.chainB.AllBalances(escrowIbcDenomAddress)
			expBalance := sdk.NewCoins(sdk.NewCoin(ibcDenom, transferAmount))
			suite.Require().Equal(expBalance, escrowIbcDenomAddressBalance)

			escrowNativeDenomAddress := transfertypes.GetEscrowAddress(pathBtoC.EndpointA.ChannelConfig.PortID, pathBtoC.EndpointA.ChannelID)
			escrowNativeDenomAddressBalance := suite.chainB.AllBalances(escrowNativeDenomAddress)
			expBalance = sdk.NewCoins(sdk.NewCoin(nativeDenom, transferAmount))
			suite.Require().Equal(expBalance, escrowNativeDenomAddressBalance)

			// then should have a packet from B to C
			suite.Require().Equal(1, len(suite.chainB.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainC.PendingSendPackets))

			// relay packet
			sendingPacket = suite.chainB.PendingSendPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainB)
			err = pathBtoC.EndpointB.UpdateClient()
			suite.Require().NoError(err)

			err = pathBtoC.EndpointB.RecvPacket(sendingPacket)
			suite.Require().NoError(err)
			suite.chainB.PendingSendPackets = nil

			// relay ack C to B
			suite.Require().Equal(1, len(suite.chainC.PendingAckPackets))
			ack := suite.chainC.PendingAckPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainC)
			err = pathBtoC.EndpointA.UpdateClient()
			suite.Require().NoError(err)

			err = pathBtoC.EndpointA.AcknowledgePacket(ack.Packet, ack.Ack)
			suite.Require().NoError(err)
			suite.chainC.PendingAckPackets = nil

			// relay ack B to A
			suite.Require().Equal(1, len(suite.chainB.PendingAckPackets))
			ack = suite.chainB.PendingAckPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainB)
			err = pathAtoB.EndpointA.UpdateClient()
			suite.Require().NoError(err)

			err = pathAtoB.EndpointA.AcknowledgePacket(ack.Packet, ack.Ack)
			suite.Require().NoError(err)
			suite.chainB.PendingAckPackets = nil

			senderBCurrentBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
			suite.Require().Equal(senderBOriginalBalance, senderBCurrentBalance)

			escrowIbcDenomAddress = transfertypes.GetEscrowAddress(pathAtoB.EndpointB.ChannelConfig.PortID, pathAtoB.EndpointB.ChannelID)
			escrowIbcDenomAddressBalance = suite.chainB.AllBalances(escrowIbcDenomAddress)
			expBalance = sdk.NewCoins(sdk.NewCoin(ibcDenom, transferAmount))
			suite.Require().Equal(expBalance, escrowIbcDenomAddressBalance)

			escrowNativeDenomAddress = transfertypes.GetEscrowAddress(pathBtoC.EndpointA.ChannelConfig.PortID, pathBtoC.EndpointA.ChannelID)
			escrowNativeDenomAddressBalance = suite.chainB.AllBalances(escrowNativeDenomAddress)
			expBalance = sdk.NewCoins(sdk.NewCoin(nativeDenom, transferAmount))
			suite.Require().Equal(expBalance, escrowNativeDenomAddressBalance)

			balance := suite.chainC.AllBalances(suite.chainC.SenderAccount.GetAddress())
			receiveBalance := balance.AmountOf(expDenom)
			suite.Require().Equal(transferAmount, receiveBalance)

			// transfer back from osmosis to picasso
			memo = PacketMetadata{
				Forward: &ForwardMetadata{
					Receiver: suite.chainA.SenderAccount.GetAddress().String(),
					Port:     pathAtoB.EndpointB.ChannelConfig.PortID,
					Channel:  pathAtoB.EndpointB.ChannelID,
					Timeout:  timeOut,
					Retries:  &retries,
					Next:     "",
				},
			}

			memoMarshalled, err = json.Marshal(&memo)
			suite.Require().NoError(err)

			msg = transfertypes.NewMsgTransfer(
				pathBtoC.EndpointB.ChannelConfig.PortID,
				pathBtoC.EndpointB.ChannelID,
				sdk.NewCoin(expDenom, transferAmount),
				suite.chainC.SenderAccount.GetAddress().String(),
				suite.chainB.SenderAccount.GetAddress().String(),
				timeoutHeight,
				0,
				string(memoMarshalled),
			)

			_, err = suite.chainC.SendMsgs(msg)
			suite.Require().NoError(err)
			suite.Require().NoError(err, pathBtoC.EndpointA.UpdateClient())

			// then
			suite.Require().Equal(1, len(suite.chainC.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))
			// relay packet
			sendingPacket = suite.chainC.PendingSendPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainC)
			err = pathBtoC.EndpointA.UpdateClient()
			suite.Require().NoError(err)

			err = pathBtoC.EndpointA.RecvPacket(sendingPacket)
			suite.Require().NoError(err)
			suite.chainC.PendingSendPackets = nil

			// then should have a packet from B to A
			suite.Require().Equal(1, len(suite.chainB.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainA.PendingSendPackets))

			// relay packet
			sendingPacket = suite.chainB.PendingSendPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainB)
			err = pathAtoB.EndpointA.UpdateClient()
			suite.Require().NoError(err)

			err = pathAtoB.EndpointA.RecvPacket(sendingPacket)
			suite.Require().NoError(err)
			suite.chainB.PendingSendPackets = nil

			// relay ack A to B
			suite.Require().Equal(1, len(suite.chainA.PendingAckPackets))
			ack = suite.chainA.PendingAckPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainA)
			err = pathAtoB.EndpointB.UpdateClient()
			suite.Require().NoError(err)

			err = pathAtoB.EndpointB.AcknowledgePacket(ack.Packet, ack.Ack)
			suite.Require().NoError(err)
			suite.chainA.PendingAckPackets = nil

			// relay ack B to C
			suite.Require().Equal(1, len(suite.chainB.PendingAckPackets))
			ack = suite.chainB.PendingAckPackets[0]
			suite.coordinator.IncrementTime()
			suite.coordinator.CommitBlock(suite.chainB)
			err = pathBtoC.EndpointB.UpdateClient()
			suite.Require().NoError(err)

			err = pathBtoC.EndpointB.AcknowledgePacket(ack.Packet, ack.Ack)
			suite.Require().NoError(err)
			suite.chainB.PendingAckPackets = nil

			senderACurrentBalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
			suite.Require().Equal(senderAOriginalBalance, senderACurrentBalance)

			senderBCurrentBalance = suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
			suite.Require().Equal(senderBOriginalBalance, senderBCurrentBalance)

			senderCCurrentBalance := suite.chainC.AllBalances(suite.chainC.SenderAccount.GetAddress())
			suite.Require().Equal(senderCOriginalBalance, senderCCurrentBalance)

			escrowIbcDenomAddressBalance = suite.chainB.AllBalances(escrowIbcDenomAddress)
			suite.Require().Empty(escrowIbcDenomAddressBalance)

			escrowNativeDenomAddressBalance = suite.chainB.AllBalances(escrowNativeDenomAddress)
			suite.Require().Empty(escrowNativeDenomAddressBalance)
		})
	}
}
