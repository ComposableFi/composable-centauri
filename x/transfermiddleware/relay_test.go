package transfermiddleware_test

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/stretchr/testify/suite"

	customibctesting "github.com/notional-labs/composable/v6/app/ibctesting"
)

// TODO: use testsuite here.
func (suite *TransferMiddlewareTestSuite) TestOnrecvPacket() {
	var (
		transferAmount = sdkmath.NewInt(1000000000)
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
				err := chainBtransMiddleware.AddParachainIBCInfo(suite.chainB.GetContext(), "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878", "channel-0", sdk.DefaultBondDenom, sdk.DefaultBondDenom)
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
			suite.Require().Equal(expBalance, gotBalance)
			suite.Require().NoError(err)
		})
	}
}

// TODO: use testsuite here.
func (suite *TransferMiddlewareTestSuite) TestSendPacket() {
	var (
		transferAmount = sdkmath.NewInt(1000000000)
		// when transfer via sdk transfer from A (module) -> B (contract)
		nativeTokenSendOnChainA    = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
		nativeTokenReceiveOnChainB = sdk.NewCoin("ppica", transferAmount)
		timeoutHeight              = clienttypes.NewHeight(1, 110)
		expChainABalanceDiff       = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
	)

	suite.SetupTest() // reset

	path := NewTransferPath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)

	// Add parachain token info
	chainBtransMiddlewareKeeper := suite.chainB.TransferMiddleware()
	err := chainBtransMiddlewareKeeper.AddParachainIBCInfo(suite.chainB.GetContext(), "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878", path.EndpointB.ChannelID, "ppica", sdk.DefaultBondDenom)
	suite.Require().NoError(err)

	originalChainABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
	originalChainBBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())

	msg := ibctransfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, nativeTokenSendOnChainA, suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
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
	expBalance := originalChainBBalance.Add(nativeTokenReceiveOnChainB)
	gotBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
	suite.Require().Equal(expBalance, gotBalance)

	// send token back
	msg = ibctransfertypes.NewMsgTransfer(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, nativeTokenReceiveOnChainB, suite.chainB.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
	_, err = suite.chainB.SendMsgs(msg)
	suite.Require().NoError(err)
	suite.Require().NoError(err, path.EndpointA.UpdateClient())
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

	// equal chain A sender address balances
	chainBReceiverBalances := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
	suite.Require().Equal(originalChainBBalance, chainBReceiverBalances)
}

// TODO: use testsuite here.
func (suite *TransferMiddlewareTestSuite) TestTimeOutPacket() {
	var (
		transferAmount = sdkmath.NewInt(1000000000)
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
	err := chainBtransMiddlewareKeeper.AddParachainIBCInfo(suite.chainB.GetContext(), "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878", "channel-0", sdk.DefaultBondDenom, sdk.DefaultBondDenom)
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
	suite.Require().NoError(err)
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

func (suite *TransferMiddlewareTestSuite) TestMintAndEscrowProcessWhenLaunchChain() {
	var (
		// when transfer via sdk transfer from A (module) -> B (contract)
		timeoutHeight                    = clienttypes.NewHeight(1, 110)
		path                             *customibctesting.Path
		expDenom                         = "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"
		transferAmountFromChainBToChainA = sdkmath.NewInt(100000000000000)
		transferAmountFromChainAToChainB = sdkmath.NewInt(1000000000000)

		// pathBtoC      = NewTransferPath(suite.chainB, suite.chainC)
	)

	suite.Run("Test mint and escrow process", func() {
		suite.SetupTest()
		// When setup chainB(Composable already have 10^19 stake in test account (genesis))
		path = NewTransferPath(suite.chainA, suite.chainB)
		suite.coordinator.Setup(path)

		chainBSupply := suite.chainB.Balance(suite.chainB.SenderAccount.GetAddress(), "stake")
		// Send coin from (chainA) to escrow address in chain B
		escrowAddress := ibctransfertypes.GetEscrowAddress(ibctransfertypes.PortID, path.EndpointB.ChannelID)
		msg := ibctransfertypes.NewMsgTransfer(
			path.EndpointA.ChannelConfig.PortID,
			path.EndpointA.ChannelID,
			chainBSupply,
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

		// Check escrow address have ibcPICA tokens
		balance := suite.chainB.AllBalances(escrowAddress)
		expBalance := sdk.NewCoins(sdk.NewCoin(expDenom, chainBSupply.Amount))
		suite.Require().Equal(expBalance, balance)

		// Add parachain token info
		chainBtransMiddleware := suite.chainB.TransferMiddleware()
		err = chainBtransMiddleware.AddParachainIBCInfo(suite.chainB.GetContext(), expDenom, path.EndpointB.ChannelID, sdk.DefaultBondDenom, sdk.DefaultBondDenom)
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

		// check balances in sender address in chain B
		balance = suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
		expBalance = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, chainBSupply.Amount.Sub(transferAmountFromChainBToChainA)))
		suite.Require().Equal(expBalance, balance)

		// check balances in escrow address in chain B
		balance = suite.chainB.AllBalances(escrowAddress)
		expBalance = sdk.NewCoins(sdk.NewCoin(expDenom, chainBSupply.Amount.Sub(transferAmountFromChainBToChainA)))
		suite.Require().Equal(expBalance, balance)

		//  receiver in chain A receive exactly native token that transferred from chain B
		balance = suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
		expBalance = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, transferAmountFromChainBToChainA))
		suite.Require().Equal(expBalance, balance)

		// Continue send coin from (chainA) to sender account in chain B
		msg = ibctransfertypes.NewMsgTransfer(
			path.EndpointA.ChannelConfig.PortID,
			path.EndpointA.ChannelID,
			sdk.NewCoin("stake", transferAmountFromChainAToChainB),
			suite.chainA.SenderAccount.GetAddress().String(),
			suite.chainB.SenderAccount.GetAddress().String(),
			timeoutHeight,
			0,
			"",
		)

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

		// check new escrow address: newbalances := chainBSupply - transferAmountFromChainBToChainA (first IBC transfer from chain B -> A)  + transferAmountFromChainAToChainB (second IBC transfer from chain A -> B)
		balance = suite.chainB.AllBalances(escrowAddress)
		expBalance = sdk.NewCoins(sdk.NewCoin(expDenom, chainBSupply.Amount.Add(transferAmountFromChainAToChainB).Sub(transferAmountFromChainBToChainA)))
		suite.Require().Equal(expBalance, balance)

		// check new chain B supply: newbalances := chainBSupply - transferAmountFromChainBToChainA (first IBC transfer from chain B -> A)  + transferAmountFromChainAToChainB (second IBC transfer from chain A -> B)
		balance = suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
		expBalance = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, chainBSupply.Amount.Sub(transferAmountFromChainBToChainA).Add(transferAmountFromChainAToChainB)))
		suite.Require().Equal(expBalance, balance)
	})
}
