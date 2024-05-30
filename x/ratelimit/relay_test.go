package ratelimit_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/stretchr/testify/suite"

	customibctesting "github.com/notional-labs/composable/v6/app/ibctesting"
	ratelimittypes "github.com/notional-labs/composable/v6/x/ratelimit/types"
)

type RateLimitTestSuite struct {
	suite.Suite

	coordinator *customibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *customibctesting.TestChain
	chainB *customibctesting.TestChain
	chainC *customibctesting.TestChain
}

func (suite *RateLimitTestSuite) SetupTest() {
	suite.coordinator = customibctesting.NewCoordinator(suite.T(), 3)
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
	suite.Run(t, new(RateLimitTestSuite))
}

func (suite *RateLimitTestSuite) TestReceiveIBCToken() {
	var (
		transferAmount = sdkmath.NewInt(1_000_000_000_000)
		// when transfer via sdk transfer from A (module) -> B (contract)
		ibcDenom                   = "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"
		nativeDenom                = "ppica"
		nativeTokenSendOnChainA    = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
		nativeTokenReceiveOnChainB = sdk.NewCoin(nativeDenom, transferAmount)
		timeoutHeight              = clienttypes.NewHeight(1, 110)
		expChainABalanceDiff       = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
	)

	suite.SetupTest() // reset

	path := NewTransferPath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)

	// Add parachain token info
	chainBtransMiddlewareKeeper := suite.chainB.TransferMiddleware()
	err := chainBtransMiddlewareKeeper.AddParachainIBCInfo(suite.chainB.GetContext(), ibcDenom, path.EndpointB.ChannelID, nativeDenom, sdk.DefaultBondDenom)
	suite.Require().NoError(err)

	originalChainABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
	originalChainBBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())

	msg := transfertypes.NewMsgTransfer(
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		nativeTokenSendOnChainA,
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

	// and source chain balance was decreased
	newChainABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
	suite.Require().Equal(originalChainABalance.Sub(expChainABalanceDiff), newChainABalance)

	// and dest chain balance contains voucher
	expBalance := originalChainBBalance.Add(nativeTokenReceiveOnChainB)
	gotBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())

	suite.Require().Equal(expBalance, gotBalance)

	// add rate limit
	chainBRateLimitKeeper := suite.chainB.RateLimit()
	msgAddRateLimit := ratelimittypes.MsgAddRateLimit{
		Denom:              nativeDenom,
		ChannelID:          path.EndpointB.ChannelID,
		MaxPercentSend:     sdkmath.NewInt(5), // 50_000_000_000 > minRateLimitAmount(10_000_000_000) => RateLimit = 50_000_000_000
		MaxPercentRecv:     sdkmath.NewInt(5), // 50_000_000_000 > minRateLimitAmount(10_000_000_000) => RateLimit = 50_000_000_000
		MinRateLimitAmount: sdkmath.NewInt(10_000_000_000),
		DurationHours:      1,
	}
	err = chainBRateLimitKeeper.AddRateLimit(suite.chainB.GetContext(), &msgAddRateLimit)
	suite.Require().NoError(err)

	// send from A to B
	transferAmount = transferAmount.Mul(sdkmath.NewInt(5)).Quo(sdkmath.NewInt(100))
	nativeTokenSendOnChainA = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
	msg = transfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, nativeTokenSendOnChainA, suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
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

	expBalance = expBalance.Add(sdk.NewCoin(nativeDenom, transferAmount))
	gotBalance = suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
	suite.Require().Equal(expBalance, gotBalance)

	// send 1 more time
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

	// not receive token because catch the threshold => balances have no change
	gotBalance = suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
	suite.Require().Equal(expBalance, gotBalance)
}

func (suite *RateLimitTestSuite) TestSendIBCToken() {
	var (
		transferAmount = sdkmath.NewInt(1_000_000_000_000)
		// when transfer via sdk transfer from A (module) -> B (contract)
		ibcDenom                   = "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"
		nativeDenom                = "ppica"
		nativeTokenSendOnChainA    = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
		nativeTokenReceiveOnChainB = sdk.NewCoin(nativeDenom, transferAmount)
		timeoutHeight              = clienttypes.NewHeight(1, 110)
		expChainABalanceDiff       = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
	)

	suite.SetupTest() // reset

	path := NewTransferPath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)

	// Add parachain token info
	chainBtransMiddlewareKeeper := suite.chainB.TransferMiddleware()
	err := chainBtransMiddlewareKeeper.AddParachainIBCInfo(suite.chainB.GetContext(), ibcDenom, path.EndpointB.ChannelID, nativeDenom, sdk.DefaultBondDenom)
	suite.Require().NoError(err)

	originalChainABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
	originalChainBBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())

	msg := transfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, nativeTokenSendOnChainA, suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
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

	originalChainBBalance = gotBalance
	// add rate limit
	chainBRateLimitKeeper := suite.chainB.RateLimit()
	msgAddRateLimit := ratelimittypes.MsgAddRateLimit{
		Denom:              nativeDenom,
		ChannelID:          path.EndpointB.ChannelID,
		MaxPercentSend:     sdkmath.NewInt(5), // 50_000_000_000 > minRateLimitAmount(10_000_000_000) => RateLimit = 50_000_000_000
		MaxPercentRecv:     sdkmath.NewInt(5), // 50_000_000_000 > minRateLimitAmount(10_000_000_000) => RateLimit = 50_000_000_000
		MinRateLimitAmount: sdkmath.NewInt(10_000_000_000),
		DurationHours:      1,
	}
	err = chainBRateLimitKeeper.AddRateLimit(suite.chainB.GetContext(), &msgAddRateLimit)
	suite.Require().NoError(err)

	// send from B to A
	transferAmount = transferAmount.Mul(sdkmath.NewInt(5)).Quo(sdkmath.NewInt(100))
	nativeTokenSendOnChainB := sdk.NewCoin(nativeDenom, transferAmount)
	msg = transfertypes.NewMsgTransfer(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, nativeTokenSendOnChainB, suite.chainB.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
	_, err = suite.chainB.SendMsgs(msg)
	suite.Require().NoError(err)
	suite.Require().NoError(err, path.EndpointA.UpdateClient())

	// then
	suite.Require().Equal(1, len(suite.chainB.PendingSendPackets))
	suite.Require().Equal(0, len(suite.chainA.PendingSendPackets))

	// and when relay to chain B and handle Ack on chain A
	err = suite.coordinator.RelayAndAckPendingPacketsReverse(path)
	suite.Require().NoError(err)

	// then
	suite.Require().Equal(0, len(suite.chainA.PendingSendPackets))
	suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))

	expBalance = originalChainBBalance.Sub(nativeTokenSendOnChainB)
	gotBalance = suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
	suite.Require().Equal(expBalance, gotBalance)

	// send 1 more time
	_, err = suite.chainB.SendMsgsWithExpPass(false, msg)
	suite.Require().Error(err) // catch the threshold so should not be sent

	// SignAndDeliver calls app.Commit()
	suite.chainB.NextBlock()

	// increment sequence for successful transaction execution
	err = suite.chainB.SenderAccount.SetSequence(suite.chainB.SenderAccount.GetSequence() + 1)
	suite.Require().NoError(err)

	suite.chainB.Coordinator.IncrementTime()

	// not receive token because catch the threshold => balances have no change
	balances := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
	suite.Require().Equal(expBalance, balances)
}

func (suite *RateLimitTestSuite) TestReceiveIBCTokenWithMinRateLimitAmount() {
	var (
		transferAmount = sdkmath.NewInt(100_000_000_000)
		// when transfer via sdk transfer from A (module) -> B (contract)
		ibcDenom                   = "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"
		nativeDenom                = "ppica"
		nativeTokenSendOnChainA    = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
		nativeTokenReceiveOnChainB = sdk.NewCoin(nativeDenom, transferAmount)
		timeoutHeight              = clienttypes.NewHeight(1, 110)
		expChainABalanceDiff       = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
	)

	suite.SetupTest() // reset

	path := NewTransferPath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)

	// Add parachain token info
	chainBtransMiddlewareKeeper := suite.chainB.TransferMiddleware()
	err := chainBtransMiddlewareKeeper.AddParachainIBCInfo(suite.chainB.GetContext(), ibcDenom, path.EndpointB.ChannelID, nativeDenom, sdk.DefaultBondDenom)
	suite.Require().NoError(err)

	originalChainABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
	originalChainBBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())

	msg := transfertypes.NewMsgTransfer(
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		nativeTokenSendOnChainA,
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

	// and source chain balance was decreased
	newChainABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
	suite.Require().Equal(originalChainABalance.Sub(expChainABalanceDiff), newChainABalance)

	// and dest chain balance contains voucher
	expBalance := originalChainBBalance.Add(nativeTokenReceiveOnChainB)
	gotBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())

	suite.Require().Equal(expBalance, gotBalance)

	// add rate limit
	chainBRateLimitKeeper := suite.chainB.RateLimit()
	msgAddRateLimit := ratelimittypes.MsgAddRateLimit{
		Denom:              nativeDenom,
		ChannelID:          path.EndpointB.ChannelID,
		MaxPercentSend:     sdkmath.NewInt(5), // 5_000_000_000 < minRateLimitAmount(10_000_000_000) => RateLimit = 10_000_000_000
		MaxPercentRecv:     sdkmath.NewInt(5), // 5_000_000_000 < minRateLimitAmount(10_000_000_000) => RateLimit = 10_000_000_000
		MinRateLimitAmount: sdkmath.NewInt(10_000_000_000),
		DurationHours:      1,
	}
	err = chainBRateLimitKeeper.AddRateLimit(suite.chainB.GetContext(), &msgAddRateLimit)
	suite.Require().NoError(err)

	// send from A to B
	transferAmount = sdkmath.NewInt(10_000_000_000)
	nativeTokenSendOnChainA = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
	msg = transfertypes.NewMsgTransfer(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, nativeTokenSendOnChainA, suite.chainA.SenderAccount.GetAddress().String(), suite.chainB.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
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

	expBalance = expBalance.Add(sdk.NewCoin(nativeDenom, transferAmount))
	gotBalance = suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
	suite.Require().Equal(expBalance, gotBalance)

	// send 1 more time
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

	// not receive token because catch the threshold => balances have no change
	gotBalance = suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
	fmt.Println(expBalance.String(), gotBalance.String())
	suite.Require().Equal(expBalance, gotBalance)
}

func (suite *RateLimitTestSuite) TestSendIBCTokenWithMinRateLimitAmount() {
	var (
		transferAmount = sdkmath.NewInt(100_000_000_000)
		// when transfer via sdk transfer from A (module) -> B (contract)
		ibcDenom                   = "ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878"
		nativeDenom                = "ppica"
		nativeTokenSendOnChainA    = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
		nativeTokenReceiveOnChainB = sdk.NewCoin(nativeDenom, transferAmount)
		timeoutHeight              = clienttypes.NewHeight(1, 110)
		expChainABalanceDiff       = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
	)

	suite.SetupTest() // reset

	path := NewTransferPath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)

	// Add parachain token info
	chainBtransMiddlewareKeeper := suite.chainB.TransferMiddleware()
	err := chainBtransMiddlewareKeeper.AddParachainIBCInfo(suite.chainB.GetContext(), ibcDenom, path.EndpointB.ChannelID, nativeDenom, sdk.DefaultBondDenom)
	suite.Require().NoError(err)

	originalChainABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
	originalChainBBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())

	msg := transfertypes.NewMsgTransfer(
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		nativeTokenSendOnChainA,
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

	// and source chain balance was decreased
	newChainABalance := suite.chainA.AllBalances(suite.chainA.SenderAccount.GetAddress())
	suite.Require().Equal(originalChainABalance.Sub(expChainABalanceDiff), newChainABalance)

	// and dest chain balance contains voucher
	expBalance := originalChainBBalance.Add(nativeTokenReceiveOnChainB)
	gotBalance := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())

	suite.Require().Equal(expBalance, gotBalance)

	originalChainBBalance = gotBalance
	// add rate limit 5%
	chainBRateLimitKeeper := suite.chainB.RateLimit()
	msgAddRateLimit := ratelimittypes.MsgAddRateLimit{
		Denom:              nativeDenom,
		ChannelID:          path.EndpointB.ChannelID,
		MaxPercentSend:     sdkmath.NewInt(5), // 5_000_000_000 < minRateLimitAmount(10_000_000_000) => RateLimit = 10_000_000_000
		MaxPercentRecv:     sdkmath.NewInt(5), // 5_000_000_000 < minRateLimitAmount(10_000_000_000) => RateLimit = 10_000_000_000
		MinRateLimitAmount: sdkmath.NewInt(10_000_000_000),
		DurationHours:      1,
	}
	err = chainBRateLimitKeeper.AddRateLimit(suite.chainB.GetContext(), &msgAddRateLimit)
	suite.Require().NoError(err)

	// send from B to A
	transferAmount = sdkmath.NewInt(10_000_000_000)
	nativeTokenSendOnChainB := sdk.NewCoin(nativeDenom, transferAmount)
	msg = transfertypes.NewMsgTransfer(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, nativeTokenSendOnChainB, suite.chainB.SenderAccount.GetAddress().String(), suite.chainA.SenderAccount.GetAddress().String(), timeoutHeight, 0, "")
	_, err = suite.chainB.SendMsgs(msg)
	suite.Require().NoError(err)
	suite.Require().NoError(err, path.EndpointA.UpdateClient())

	// then
	suite.Require().Equal(1, len(suite.chainB.PendingSendPackets))
	suite.Require().Equal(0, len(suite.chainA.PendingSendPackets))

	// and when relay to chain B and handle Ack on chain A
	err = suite.coordinator.RelayAndAckPendingPacketsReverse(path)
	suite.Require().NoError(err)

	// then
	suite.Require().Equal(0, len(suite.chainA.PendingSendPackets))
	suite.Require().Equal(0, len(suite.chainB.PendingSendPackets))

	expBalance = originalChainBBalance.Sub(nativeTokenSendOnChainB)
	gotBalance = suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
	suite.Require().Equal(expBalance, gotBalance)

	// send 1 more time
	_, err = suite.chainB.SendMsgsWithExpPass(false, msg)
	suite.Require().Error(err) // catch the threshold so should not be sent

	// SignAndDeliver calls app.Commit()
	suite.chainB.NextBlock()

	// increment sequence for successful transaction execution
	err = suite.chainB.SenderAccount.SetSequence(suite.chainB.SenderAccount.GetSequence() + 1)
	suite.Require().NoError(err)

	suite.chainB.Coordinator.IncrementTime()

	// not receive token because catch the threshold => balances have no change
	balances := suite.chainB.AllBalances(suite.chainB.SenderAccount.GetAddress())
	suite.Require().Equal(expBalance, balances)
}
