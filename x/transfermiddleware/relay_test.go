package transfermiddleware_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	customibctesting "github.com/notional-labs/banksy/v2/app/ibctesting"
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
