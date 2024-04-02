package bank_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/stretchr/testify/suite"

	customibctesting "github.com/ComposableFi/composable-cosmos/v6/app/ibctesting"
)

type CustomBankTestSuite struct {
	suite.Suite

	coordinator *customibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *customibctesting.TestChain
	chainB *customibctesting.TestChain
	chainC *customibctesting.TestChain
}

func NewTransferPath(chainA, chainB *customibctesting.TestChain) *customibctesting.Path {
	path := customibctesting.NewPath(chainA, chainB)
	path.EndpointA.ChannelConfig.PortID = customibctesting.TransferPort
	path.EndpointB.ChannelConfig.PortID = customibctesting.TransferPort
	path.EndpointA.ChannelConfig.Version = ibctransfertypes.Version
	path.EndpointB.ChannelConfig.Version = ibctransfertypes.Version

	return path
}

func (suite *CustomBankTestSuite) SetupTest() {
	suite.coordinator = customibctesting.NewCoordinator(suite.T(), 4)
	suite.chainA = suite.coordinator.GetChain(customibctesting.GetChainID(1))
	suite.chainB = suite.coordinator.GetChain(customibctesting.GetChainID(2))
	suite.chainC = suite.coordinator.GetChain(customibctesting.GetChainID(3))
}

func TestBankTestSuite(t *testing.T) {
	suite.Run(t, new(CustomBankTestSuite))
}

// TODO: use testsuite here.
func (suite *CustomBankTestSuite) TestTotalSupply() {
	var (
		transferAmount = sdk.NewInt(1000000000)
		// when transfer via sdk transfer from A (module) -> B (contract)
		coinToSendToB     = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
		timeoutHeight     = clienttypes.NewHeight(1, 110)
		originAmt, err    = sdk.NewIntFromString("10000000001100000000000")
		chainBOriginSuply = sdk.NewCoin("stake", originAmt)
	)
	suite.Require().True(err)
	var (
		expChainBBalanceDiff sdk.Coin
		path                 = NewTransferPath(suite.chainA, suite.chainB)
		escrowAddr           = ibctransfertypes.GetEscrowAddress(ibctransfertypes.PortID, "channel-0")
	)

	testCases := []struct {
		name                 string
		expChainABalanceDiff sdk.Coin
		expTotalSupplyDiff   sdk.Coins
		expChainBTotalSuppy  sdk.Coins
		malleate             func()
	}{
		{
			"Total supply with no transfermiddleware setup",
			sdk.NewCoin(sdk.DefaultBondDenom, transferAmount),
			sdk.Coins{sdk.NewCoin("ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878", transferAmount)},
			sdk.Coins{sdk.NewCoin("ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878", transferAmount), chainBOriginSuply},
			func() {
				expChainBBalanceDiff = ibctransfertypes.GetTransferCoin(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, coinToSendToB.Denom, transferAmount)
			},
		},
		{
			"Total supply with transfermiddleware setup",
			sdk.NewCoin(sdk.DefaultBondDenom, transferAmount),
			sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)),
			sdk.Coins{chainBOriginSuply.Add(sdk.NewCoin("stake", transferAmount))},
			func() {
				// Add parachain token info
				chainBtransMiddleware := suite.chainB.TransferMiddleware()
				expChainBBalanceDiff = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
				err := chainBtransMiddleware.AddParachainIBCInfo(suite.chainB.GetContext(), "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878", "channel-0", sdk.DefaultBondDenom, sdk.DefaultBondDenom)
				suite.Require().NoError(err)
			},
		},
		{
			"Total supply with transfermiddleware setup and pre mint",
			sdk.NewCoin(sdk.DefaultBondDenom, transferAmount),
			sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)),
			sdk.Coins{chainBOriginSuply.Add(sdk.NewCoin("stake", transferAmount))},
			func() {
				// Premint for escrow
				err := suite.chainB.GetBankKeeper().MintCoins(suite.chainB.GetContext(), "mint", sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000000))))
				suite.Require().NoError(err)
				err = suite.chainB.GetBankKeeper().SendCoinsFromModuleToAccount(suite.chainB.GetContext(), "mint", escrowAddr, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000000))))
				suite.Require().NoError(err)

				// Add parachain token info
				chainBtransMiddleware := suite.chainB.TransferMiddleware()
				expChainBBalanceDiff = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
				err = chainBtransMiddleware.AddParachainIBCInfo(suite.chainB.GetContext(), "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878", "channel-0", sdk.DefaultBondDenom, sdk.DefaultBondDenom)
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
			originalChainBTotalSupply, err := suite.chainB.GetBankKeeper().TotalSupply(suite.chainB.GetContext(), &banktypes.QueryTotalSupplyRequest{})
			suite.Require().NoError(err)

			msg := ibctransfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, coinToSendToB, suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
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
			suite.Require().Equal(originalChainABalance.Sub(tc.expChainABalanceDiff), newChainABalance)

			// and dest chain balance contains voucher
			expBalance := originalChainBBalance.Add(expChainBBalanceDiff)
			gotBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
			suite.Require().Equal(expBalance, gotBalance)
			suite.Require().NoError(err)

			totalSupply, err := suite.chainB.GetBankKeeper().TotalSupply(suite.chainB.GetContext(), &banktypes.QueryTotalSupplyRequest{})
			suite.Require().NoError(err)
			suite.Require().Equal(totalSupply.Supply, originalChainBTotalSupply.Supply.Add(tc.expTotalSupplyDiff...))
			suite.Require().Equal(totalSupply.Supply, tc.expChainBTotalSuppy)
		})
	}
}

func (suite *CustomBankTestSuite) TestTotalSupply2() {
	var (
		transferAmount  = sdk.NewInt(1000000000)
		transferAmount2 = sdk.NewInt(3500000000)
		// when transfer via sdk transfer from A (module) -> B (contract)
		coinChainASendToB = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
		coinChainCSentToB = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount2)
		timeoutHeight     = clienttypes.NewHeight(1, 110)
	)
	var (
		expChainBBalanceDiff sdk.Coins
		pathAB               = NewTransferPath(suite.chainA, suite.chainB)
		pathCB               = NewTransferPath(suite.chainC, suite.chainB)
	)

	testCases := []struct {
		name                 string
		expChainABalanceDiff sdk.Coin
		expChainCBalanceDiff sdk.Coin
		expTotalSupplyDiff   sdk.Coins
		malleate             func()
	}{
		{
			"Total supply with no transfermiddleware setup",
			sdk.NewCoin(sdk.DefaultBondDenom, transferAmount),
			sdk.NewCoin(sdk.DefaultBondDenom, transferAmount2),
			sdk.Coins{sdk.NewCoin("ibc/3C3D7B3BE4ECC85A0E5B52A3AEC3B7DFC2AA9CA47C37821E57020D6807043BE9", transferAmount2), sdk.NewCoin("ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878", transferAmount)},
			func() {
				transferCoinFromA := ibctransfertypes.GetTransferCoin(pathAB.EndpointB.ChannelConfig.PortID, pathAB.EndpointB.ChannelID, coinChainASendToB.Denom, transferAmount)
				transferCoinFromC := ibctransfertypes.GetTransferCoin(pathCB.EndpointB.ChannelConfig.PortID, pathCB.EndpointB.ChannelID, coinChainCSentToB.Denom, transferAmount2)
				expChainBBalanceDiff = sdk.NewCoins(transferCoinFromA, transferCoinFromC)
			},
		},
		{
			"Total supply with transfermiddleware setup",
			sdk.NewCoin(sdk.DefaultBondDenom, transferAmount),
			sdk.NewCoin(sdk.DefaultBondDenom, transferAmount2),
			sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, transferAmount), sdk.NewCoin("ibc/3C3D7B3BE4ECC85A0E5B52A3AEC3B7DFC2AA9CA47C37821E57020D6807043BE9", transferAmount2)),
			func() {
				// Add parachain token info
				chainBtransMiddleware := suite.chainB.TransferMiddleware()

				transferCoinFromA := sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
				transferCoinFromC := ibctransfertypes.GetTransferCoin(pathCB.EndpointB.ChannelConfig.PortID, pathCB.EndpointB.ChannelID, coinChainASendToB.Denom, transferAmount2)
				expChainBBalanceDiff = sdk.NewCoins(transferCoinFromA, transferCoinFromC)

				err := chainBtransMiddleware.AddParachainIBCInfo(suite.chainB.GetContext(), "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878", "channel-0", sdk.DefaultBondDenom, sdk.DefaultBondDenom)
				suite.Require().NoError(err)
			},
		},
	}
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest() // reset

			pathAB = NewTransferPath(suite.chainA, suite.chainB)
			pathCB = NewTransferPath(suite.chainC, suite.chainB)
			suite.coordinator.Setup(pathAB)
			suite.coordinator.Setup(pathCB)

			tc.malleate()

			originalChainABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
			// chainB.SenderAccount: 10000000000000000000stake
			originalChainBBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
			originalChainCBalance := suite.chainC.AllBalances(suite.chainC.SenderAccount.GetAddress())
			originalChainBTotalSupply, err := suite.chainB.GetBankKeeper().TotalSupply(suite.chainB.GetContext(), &banktypes.QueryTotalSupplyRequest{})
			suite.Require().NoError(err)

			// Send from A to B
			msg := ibctransfertypes.NewMsgTransfer(pathAB.EndpointA.ChannelConfig.PortID, pathAB.EndpointA.ChannelID, coinChainASendToB, suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
			_, err = suite.chainA.SendMsgs(msg)
			suite.Require().NoError(err)
			suite.Require().NoError(err, pathAB.EndpointB.UpdateClient())

			// then
			suite.Require().Equal(1, len(suite.chainA.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))

			// and when relay to chain B and handle Ack on chain A
			err = suite.coordinator.RelayAndAckPendingPackets(pathAB)
			suite.Require().NoError(err)

			// then
			suite.Require().Equal(0, len(suite.chainA.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))

			// Send from C to B
			msg = ibctransfertypes.NewMsgTransfer(pathCB.EndpointA.ChannelConfig.PortID, pathCB.EndpointA.ChannelID, coinChainCSentToB, suite.chainC.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
			_, err = suite.chainC.SendMsgs(msg)
			suite.Require().NoError(err)
			suite.Require().NoError(err, pathCB.EndpointB.UpdateClient())

			// then
			suite.Require().Equal(1, len(suite.chainC.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))

			// and when relay to chain B and handle Ack on chain C
			err = suite.coordinator.RelayAndAckPendingPackets(pathCB)
			suite.Require().NoError(err)

			// then
			suite.Require().Equal(0, len(suite.chainC.PendingSendPackets))
			suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))

			// and source chain balance was decreased
			newChainABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
			suite.Require().Equal(originalChainABalance.Sub(tc.expChainABalanceDiff), newChainABalance)

			newChainCBalance := suite.chainC.AllBalances(suite.chainC.SenderAccount.GetAddress())
			suite.Require().Equal(originalChainCBalance.Sub(tc.expChainCBalanceDiff), newChainCBalance)

			// and dest chain balance contains voucher
			expBalance := originalChainBBalance.Add(expChainBBalanceDiff...)
			gotBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
			suite.Require().Equal(expBalance, gotBalance)
			suite.Require().NoError(err)

			totalSupply, err := suite.chainB.GetBankKeeper().TotalSupply(suite.chainB.GetContext(), &banktypes.QueryTotalSupplyRequest{})
			suite.Require().NoError(err)
			suite.Require().Equal(totalSupply.Supply, originalChainBTotalSupply.Supply.Add(tc.expTotalSupplyDiff...))
		})
	}
}
