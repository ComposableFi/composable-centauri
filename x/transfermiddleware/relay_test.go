package transfermiddleware_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	customibctesting "github.com/notional-labs/banksy/v2/app/ibctesting"
	"github.com/stretchr/testify/suite"
)

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
		expChainABalanceDiff = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
	)

	suite.SetupTest() // reset

	path := NewTransferPath(suite.chainA, suite.chainB)
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
		expChainABalanceDiff = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
	)

	suite.SetupTest() // reset

	path := NewTransferPath(suite.chainA, suite.chainB)
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

func TestTransferMiddlewareTestSuiteTestSuite(t *testing.T) {
	suite.Run(t, new(TransferMiddlewareTestSuite))
}

func (suite *TransferMiddlewareTestSuite) TestMintAndBurnProcessWhenLaunchChain() {
	var (
		transferAmount, _ = sdk.NewIntFromString("10000000000000000000")
		// when transfer via sdk transfer from A (module) -> B (contract)
		timeoutHeight                    = clienttypes.NewHeight(1, 110)
		path                             *customibctesting.Path
		expDenom                         = "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"
		transferAmountFromChainBToChainA = sdk.NewInt(1000000000)

		// pathBtoC      = NewTransferPath(suite.chainB, suite.chainC)
	)

	testCases := []struct {
		name string
	}{
		{
			"Test Mint",
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			// When setup chainB(Composable already have 10^19 stake in test account (genesis))
			path = NewTransferPath(suite.chainA, suite.chainB)
			suite.coordinator.Setup(path)

			senderABalance := suite.chainB.Balance(suite.chainB.SenderAccount.GetAddress(), "stake")

			// Send coin from picasso (chainA) to escrow address
			escrowAddress := ibctransfertypes.GetEscrowAddress(ibctransfertypes.PortID, path.EndpointB.ChannelID)
			msg := ibctransfertypes.NewMsgTransfer(
				path.EndpointA.ChannelConfig.PortID,
				path.EndpointA.ChannelID,
				senderABalance,
				suite.chainA.SenderAccount.GetAddress().String(),
				escrowAddress.String(),
				timeoutHeight,
				0,
				"",
			)
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

			balance := suite.chainB.AllBalances(escrowAddress)
			expBalance := sdk.NewCoins(sdk.NewCoin(expDenom, transferAmount))
			suite.Require().Equal(expBalance, balance)

			// Add parachain token info
			chainBtransMiddleware := suite.chainB.TransferMiddleware()
			err = chainBtransMiddleware.AddParachainIBCInfo(suite.chainB.GetContext(), expDenom, path.EndpointB.ChannelID, sdk.DefaultBondDenom)
			suite.Require().NoError(err)

			// send coin from B to A
			msg = ibctransfertypes.NewMsgTransfer(
				path.EndpointB.ChannelConfig.PortID,
				path.EndpointB.ChannelID,
				sdk.NewCoin("stake", transferAmountFromChainBToChainA),
				suite.chainB.SenderAccount.GetAddress().String(),
				suite.chainA.SenderAccount.GetAddress().String(),
				timeoutHeight,
				0,
				"",
			)
			_, err = suite.chainB.SendMsgs(msg)
			suite.Require().NoError(err)
			suite.Require().NoError(err, path.EndpointA.UpdateClient())

			// then
			suite.Require().Equal(1, len(suite.chainB.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainA.PendingSendPackets))

			// and when relay to chain A and handle Ack on chain B
			err = suite.coordinator.RelayAndAckPendingPacketsReverse(path)
			suite.Require().NoError(err)

			// then
			suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainA.PendingSendPackets))

			balance = suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
			expBalance = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, transferAmount.Sub(transferAmountFromChainBToChainA)))
			suite.Require().Equal(expBalance, balance)

			balance = suite.chainB.AllBalances(escrowAddress)
			expBalance = sdk.NewCoins(sdk.NewCoin(expDenom, transferAmount.Sub(transferAmountFromChainBToChainA)))
			suite.Require().Equal(expBalance, balance)

			balance = suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
			expBalance = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, transferAmountFromChainBToChainA))
			suite.Require().Equal(expBalance, balance)
		})
	}
}
