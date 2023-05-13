package transfermiddleware_test

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	customibctesting "github.com/notional-labs/banksy/v2/app/ibctesting"
	routertypes "github.com/strangelove-ventures/packet-forward-middleware/v7/router/types"
	"github.com/stretchr/testify/suite"
)

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
	path.EndpointA.ChannelConfig.Version = ibctransfertypes.Version
	path.EndpointB.ChannelConfig.Version = ibctransfertypes.Version

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

func RandomBech32AccountAddress(t testing.TB) string {
	return RandomAccountAddress(t).String()
}

func (suite *TransferMiddlewareTestSuite) TestSendTransfer() {
	var (
		transferAmount = sdk.NewInt(1000000000)
		// when transfer via sdk transfer from A (module) -> B (contract)
		timeoutHeight = clienttypes.NewHeight(1, 110)
		pathAtoB      *customibctesting.Path
		pathCtoB      *customibctesting.Path
		path          *customibctesting.Path
		srcPort       string
		srcChannel    string
		chain         *customibctesting.TestChain
		expDenom      string
		// pathBtoC      = NewTransferPath(suite.chainB, suite.chainC)
	)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"Receiver is Parachain chain",
			func() {
				path = pathAtoB
				srcPort = pathAtoB.EndpointB.ChannelConfig.PortID
				srcChannel = pathAtoB.EndpointB.ChannelID
				chain = suite.chainA
				expDenom = sdk.DefaultBondDenom
			},
		},
		{
			"Receiver is cosmos chain chain",
			func() {
				path = pathCtoB
				srcPort = pathCtoB.EndpointB.ChannelConfig.PortID
				srcChannel = pathCtoB.EndpointB.ChannelID
				chain = suite.chainC
				expDenom = "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			pathAtoB = NewTransferPath(suite.chainA, suite.chainB)
			suite.coordinator.Setup(pathAtoB)
			pathCtoB = NewTransferPath(suite.chainC, suite.chainB)
			suite.coordinator.Setup(pathCtoB)
			// Add parachain token info
			chainBtransMiddleware := suite.chainB.TransferMiddleware()
			err := chainBtransMiddleware.AddParachainIBCInfo(suite.chainB.GetContext(), "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878", pathAtoB.EndpointB.ChannelID, sdk.DefaultBondDenom)
			suite.Require().NoError(err)
			// send coin from A to B

			msg := ibctransfertypes.NewMsgTransfer(
				pathAtoB.EndpointA.ChannelConfig.PortID,
				pathAtoB.EndpointA.ChannelID,
				sdk.NewCoin(sdk.DefaultBondDenom, transferAmount),
				suite.chainA.SenderAccount.GetAddress().String(),
				suite.chainB.SenderAccount.GetAddress().String(),
				timeoutHeight,
				0,
				"",
			)
			_, err = suite.chainA.SendMsgs(msg)
			suite.Require().NoError(err)
			suite.Require().NoError(err, pathAtoB.EndpointB.UpdateClient())

			// then
			suite.Require().Equal(1, len(suite.chainA.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))

			// and when relay to chain A and handle Ack on chain B
			err = suite.coordinator.RelayAndAckPendingPackets(pathAtoB)
			suite.Require().NoError(err)

			// then
			suite.Require().Equal(0, len(suite.chainA.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))

			tc.malleate()

			testAcc2 := RandomAccountAddress(suite.T())
			msg = ibctransfertypes.NewMsgTransfer(
				srcPort,
				srcChannel,
				sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(500000)),
				suite.chainB.SenderAccount.GetAddress().String(),
				testAcc2.String(),
				timeoutHeight,
				0,
				"",
			)
			_, err = suite.chainB.SendMsgs(msg)
			suite.Require().NoError(err)
			suite.Require().NoError(err, path.EndpointB.UpdateClient())

			suite.Require().Equal(1, len(suite.chainB.PendingSendPackets))
			suite.Require().Equal(0, len(chain.PendingSendPackets))

			// and when relay to chain B and handle Ack on chain A
			err = suite.coordinator.RelayAndAckPendingPacketsReverse(path)
			suite.Require().NoError(err)

			suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))
			suite.Require().Equal(0, len(chain.PendingSendPackets))

			balance := chain.AllBalances(testAcc2)
			expBalance := sdk.NewCoins(sdk.NewCoin(expDenom, sdk.NewInt(500000)))
			suite.Require().Equal(expBalance, balance)
		})
	}
}

// TODO: use testsuite here.
func (suite *TransferMiddlewareTestSuite) TestOnrecvPacket() {
	var (
		transferAmount = sdk.NewInt(1000000000)
		// when transfer via sdk transfer from A (module) -> B (contract)
		coinToSendToB = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
		timeoutHeight = clienttypes.NewHeight(1, 110)
	)
	var (
		expChainBBalanceDiff sdk.Coin
		path                 = NewTransferPath(suite.chainA, suite.chainB)
	)

	testCases := []struct {
		name                 string
		expChainABalanceDiff sdk.Coin
		malleate             func()
	}{
		{
			"Transfer with no pre-set ParachainIBCTokenInfo",
			sdk.NewCoin(sdk.DefaultBondDenom, transferAmount),
			func() {
				expChainBBalanceDiff = ibctransfertypes.GetTransferCoin(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, coinToSendToB.Denom, transferAmount)

			},
		},
		{
			"Transfer with pre-set ParachainIBCTokenInfo",
			sdk.NewCoin(sdk.DefaultBondDenom, transferAmount),
			func() {
				// Add parachain token info
				chainBtransMiddleware := suite.chainB.TransferMiddleware()
				expChainBBalanceDiff = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
				err := chainBtransMiddleware.AddParachainIBCInfo(suite.chainB.GetContext(), "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878", "channel-0", sdk.DefaultBondDenom)
				suite.Require().NoError(err)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			path = NewTransferPath(suite.chainA, suite.chainB)
			suite.coordinator.Setup(path)

			tc.malleate()
			originalChainABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
			// chainB.SenderAccount: 10000000000000000000stake
			originalChainBBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())

			fmt.Println("chainB.AllBalances(chainB.SenderAccount.GetAddress())", suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress()))
			msg := ibctransfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, coinToSendToB, suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
			_, err := suite.chainA.SendMsgs(msg)
			suite.Require().NoError(err)
			suite.Require().NoError(err, path.EndpointB.UpdateClient())

			// then
			suite.Require().Equal(1, len(suite.chainA.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))

			// and when relay to chain B and handle Ack on chain A
			err = suite.coordinator.RelayAndAckPendingPackets(path)
			suite.Require().NoError(err)

			// then
			suite.Require().Equal(0, len(suite.chainA.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))

			// and source chain balance was decreased
			newChainABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
			suite.Require().Equal(originalChainABalance.Sub(tc.expChainABalanceDiff), newChainABalance)

			// and dest chain balance contains voucher
			expBalance := originalChainBBalance.Add(expChainBBalanceDiff)
			gotBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
			fmt.Println("expBalance", expBalance)
			fmt.Println("gotBalance", gotBalance)
			suite.Require().Equal(expBalance, gotBalance)
		})
	}
}
func (suite *TransferMiddlewareTestSuite) TestOnrecvPacketBetween3Chain() {
	var (
		transferAmount = sdk.NewInt(1000000000)
		// when transfer via sdk transfer from A -> B -> C
		coinASendToB  = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
		timeoutHeight = clienttypes.NewHeight(1, 110)
	)
	var (
		pathAB       = NewTransferPath(suite.chainA, suite.chainB)
		pathBC       = NewTransferPath(suite.chainB, suite.chainC)
		ibcDenomAtoB = ibctransfertypes.GetPrefixedDenom(pathAB.EndpointB.ChannelConfig.PortID, pathAB.EndpointB.ChannelID, sdk.DefaultBondDenom)
	)
	testCases := []struct {
		name                 string
		expChainABalanceDiff sdk.Coin
		expChainBBalanceDiff sdk.Coin
		expChainCBalanceDiff sdk.Coin
		malleate             func()
	}{
		{
			name:                 "Transfer with no pre-set ParachainIBCTokenInfo",
			expChainABalanceDiff: sdk.NewCoin(sdk.DefaultBondDenom, transferAmount),
			expChainBBalanceDiff: ibctransfertypes.GetTransferCoin(pathAB.EndpointB.ChannelConfig.PortID, pathAB.EndpointB.ChannelID, ibcDenomAtoB, transferAmount),
			expChainCBalanceDiff: ibctransfertypes.GetTransferCoin(pathBC.EndpointB.ChannelConfig.PortID, pathBC.EndpointB.ChannelID, ibcDenomAtoB, transferAmount),
			malleate:             func() {},
		},
		// {
		// 	"Transfer with pre-set ParachainIBCTokenInfo",
		// 	sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(2000000000)),
		// 	sdk.NewCoin(sdk.DefaultBondDenom, transferAmount),
		// 	sdk.NewCoin(sdk.DefaultBondDenom, transferAmount),
		// 	func() {
		// 		// Add parachain token info
		// 		chainBtransMiddleware := chainB.TransferMiddleware()
		// 		err := chainBtransMiddleware.AddParachainIBCInfo(chainB.GetContext(), "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878", "channel-0", sdk.DefaultBondDenom)
		// 		require.NoError(t, err)

		// 		chainCtransMiddleware := chainC.TransferMiddleware()
		// 		err = chainCtransMiddleware.AddParachainIBCInfo(chainC.GetContext(), "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878", "channel-0", sdk.DefaultBondDenom)
		// 		require.NoError(t, err)
		// 	},
		// },
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset
			pathAB = NewTransferPath(suite.chainA, suite.chainB)
			suite.coordinator.Setup(pathAB)
			pathBC = NewTransferPath(suite.chainB, suite.chainC)
			suite.coordinator.Setup(pathBC)

			tc.malleate()

			originalChainABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
			// chainB.SenderAccount: 10000000000000000000stake
			// originalChainBBalance := chainB.AllBalances(chainB.SenderAccount.GetAddress())

			originalChainCBalance := suite.chainC.AllBalances(suite.chainC.SenderAccount.GetAddress())

			fmt.Println("Begin")
			fmt.Println("chainA.AllBalances(chainA.SenderAccount.GetAddress())", suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress()))
			fmt.Println("chainB.AllBalances(chainB.SenderAccount.GetAddress())", suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress()))
			fmt.Println("chainC.AllBalances(chainC.SenderAccount.GetAddress())", suite.chainC.AllBalances(suite.chainC.SenderAccount.GetAddress()))
			forwardMetadata := routertypes.PacketMetadata{
				Forward: &routertypes.ForwardMetadata{
					Receiver: suite.chainC.SenderAccount.GetAddress().String(),
					Port:     "transfer",
					Channel:  pathBC.EndpointA.ChannelID,
				},
			}
			memo, err := json.Marshal(forwardMetadata)

			msg := ibctransfertypes.NewMsgTransfer(pathAB.EndpointA.ChannelConfig.PortID, pathAB.EndpointA.ChannelID, coinASendToB, suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), timeoutHeight, 0, string(memo))
			_, err = suite.chainA.SendMsgs(msg)
			suite.Require().NoError(err)
			suite.Require().NoError(pathAB.EndpointB.UpdateClient())

			// then
			suite.Require().Equal(1, len(suite.chainA.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))

			// and when relay to chain B and handle Ack on chain A
			err = suite.coordinator.RelayAndAckPendingPackets(pathAB)
			suite.Require().NoError(err)

			err = suite.coordinator.RelayAndAckPendingPackets(pathBC)
			suite.Require().NoError(err)
			// then
			suite.Require().Equal(0, len(suite.chainA.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))
			fmt.Println("After A -> B")
			fmt.Println("chainA.AllBalances(chainA.SenderAccount.GetAddress())", suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress()))
			fmt.Println("chainB.AllBalances(chainB.SenderAccount.GetAddress())", suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress()))
			fmt.Println("chainC.AllBalances(chainC.SenderAccount.GetAddress())", suite.chainC.AllBalances(suite.chainC.SenderAccount.GetAddress()))
			// and source chain balance was decreased
			newChainABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
			suite.Require().Equal(originalChainABalance.Sub(tc.expChainABalanceDiff), newChainABalance)

			// and dest chain balance contains voucher
			expBalance := originalChainCBalance.Add(tc.expChainCBalanceDiff)
			gotBalance := suite.chainC.AllBalances(suite.chainC.SenderAccount.GetAddress())
			fmt.Println("expBalance", expBalance)
			fmt.Println("gotBalance", gotBalance)
			suite.Require().Equal(expBalance, gotBalance)
		})
	}
}

// TODO: use testsuite here.
func (suite *TransferMiddlewareTestSuite) TestSendPacket() {
	var (
		transferAmount = sdk.NewInt(1000000000)
		// when transfer via sdk transfer from A (module) -> B (contract)
		nativeToken   = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
		timeoutHeight = clienttypes.NewHeight(1, 110)
	)
	var (
		expChainBBalanceDiff sdk.Coin
		path                 = NewTransferPath(suite.chainA, suite.chainB)
		expChainABalanceDiff = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
	)

	suite.SetupTest() // reset

	path = NewTransferPath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)

	// Add parachain token info
	chainBtransMiddlewareKeeper := suite.chainB.TransferMiddleware()
	expChainBBalanceDiff = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
	err := chainBtransMiddlewareKeeper.AddParachainIBCInfo(suite.chainB.GetContext(), "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878", "channel-0", sdk.DefaultBondDenom)
	suite.Require().NoError(err)

	originalChainABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
	originalChainBBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())

	msg := ibctransfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, nativeToken, suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
	_, err = suite.chainA.SendMsgs(msg)
	suite.Require().NoError(err)
	suite.Require().NoError(err, path.EndpointB.UpdateClient())

	// then
	suite.Require().Equal(1, len(suite.chainA.PendingSendPackets))
	suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))

	// and when relay to chain B and handle Ack on chain A
	err = suite.coordinator.RelayAndAckPendingPackets(path)
	suite.Require().NoError(err)

	// then
	suite.Require().Equal(0, len(suite.chainA.PendingSendPackets))
	suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))

	// and source chain balance was decreased
	newChainABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
	suite.Require().Equal(originalChainABalance.Sub(expChainABalanceDiff), newChainABalance)

	// and dest chain balance contains voucher
	expBalance := originalChainBBalance.Add(expChainBBalanceDiff)
	gotBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
	suite.Require().Equal(expBalance, gotBalance)

	// send token back
	msg = ibctransfertypes.NewMsgTransfer(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, nativeToken, suite.chainB.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
	_, err = suite.chainB.SendMsgs(msg)
	suite.Require().NoError(err)
	suite.Require().NoError(err, path.EndpointA.UpdateClient())

	// then
	suite.Require().Equal(1, len(suite.chainB.PendingSendPackets))

	// and when relay to chain B and handle Ack on chain A
	err = suite.coordinator.RelayAndAckPendingPacketsReverse(path)
	suite.Require().NoError(err)

	// then
	suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))

	// check escrow address don't have any token in chain B
	escrowAddressChainB := ibctransfertypes.GetEscrowAddress(ibctransfertypes.PortID, path.EndpointB.ChannelID)
	escrowTokenChainB := suite.chainB.AllBalances(escrowAddressChainB)
	suite.Require().Equal(sdk.Coins{}, escrowTokenChainB)

	// check escrow address don't have any token in chain A
	escrowAddressChainA := ibctransfertypes.GetEscrowAddress(ibctransfertypes.PortID, path.EndpointA.ChannelID)
	escrowTokenChainA := suite.chainA.AllBalances(escrowAddressChainA)
	suite.Require().Equal(sdk.Coins{}, escrowTokenChainA)

	// equal chain A sender address balances
	chainASenderBalances := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
	suite.Require().Equal(originalChainABalance, chainASenderBalances)
}

// TODO: use testsuite here.
func (suite *TransferMiddlewareTestSuite) TestTimeOutPacket() {
	var (
		transferAmount = sdk.NewInt(1000000000)
		// when transfer via sdk transfer from A (module) -> B (contract)
		nativeToken   = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
		timeoutHeight = clienttypes.NewHeight(1, 110)
	)
	var (
		expChainBBalanceDiff sdk.Coin
		path                 = NewTransferPath(suite.chainA, suite.chainB)
		expChainABalanceDiff = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
	)

	suite.SetupTest() // reset

	path = NewTransferPath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)

	// Add parachain token info
	chainBtransMiddlewareKeeper := suite.chainB.TransferMiddleware()
	expChainBBalanceDiff = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
	err := chainBtransMiddlewareKeeper.AddParachainIBCInfo(suite.chainB.GetContext(), "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878", "channel-0", sdk.DefaultBondDenom)
	suite.Require().NoError(err)

	originalChainABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
	originalChainBBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())

	msg := ibctransfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, nativeToken, suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
	_, err = suite.chainA.SendMsgs(msg)
	suite.Require().NoError(err)
	suite.Require().NoError(err, path.EndpointB.UpdateClient())

	// then
	suite.Require().Equal(1, len(suite.chainA.PendingSendPackets))
	suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))

	// and when relay to chain B and handle Ack on chain A
	err = suite.coordinator.RelayAndAckPendingPackets(path)
	suite.Require().NoError(err)

	// then
	suite.Require().Equal(0, len(suite.chainA.PendingSendPackets))
	suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))

	// and source chain balance was decreased
	newChainABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
	suite.Require().Equal(originalChainABalance.Sub(expChainABalanceDiff), newChainABalance)

	// and dest chain balance contains voucher
	expBalance := originalChainBBalance.Add(expChainBBalanceDiff)
	gotBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
	suite.Require().Equal(expBalance, gotBalance)

	// send token back
	timeout := uint64(suite.chainB.LastHeader.Header.Time.Add(time.Nanosecond).UnixNano()) // will timeout
	msg = ibctransfertypes.NewMsgTransfer(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, nativeToken, suite.chainB.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String(), clienttypes.NewHeight(1, 20), timeout, "")
	_, err = suite.chainB.SendMsgs(msg)
	suite.Require().NoError(err)
	suite.Require().NoError(err, path.EndpointA.UpdateClient())

	// then
	suite.Require().Equal(1, len(suite.chainB.PendingSendPackets))
	// and when relay to chain B and handle Ack on chain A
	err = suite.coordinator.TimeoutPendingPacketsReverse(path)
	suite.Require().NoError(err)

	// then
	suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))

	// equal chain A sender address balances
	chainBSenderBalances := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
	suite.Equal(expBalance, chainBSenderBalances)
}
