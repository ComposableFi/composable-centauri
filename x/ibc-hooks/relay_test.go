package ibchooks_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/stretchr/testify/suite"

	customibctesting "github.com/ComposableFi/composable-cosmos/v6/app/ibctesting"
	ibchookskeeper "github.com/ComposableFi/composable-cosmos/v6/x/ibc-hooks/keeper"
)

// TODO: use testsuite here.
type IBCHooksTestSuite struct {
	suite.Suite

	coordinator *customibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *customibctesting.TestChain
	chainB *customibctesting.TestChain
	chainC *customibctesting.TestChain
}

func (suite *IBCHooksTestSuite) SetupTest() {
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
	suite.Run(t, new(IBCHooksTestSuite))
}

func (suite *IBCHooksTestSuite) TestRecvHooks() {
	var (
		transferAmount = sdk.NewInt(1000000000)
		timeoutHeight  = clienttypes.NewHeight(1, 110)
	// when transfer via sdk transfer from A (module) -> B (contract)
	// nativeTokenSendOnChainA = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
	)

	suite.SetupTest() // reset

	path := NewTransferPath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)

	// store code
	suite.chainB.StoreContractCode(&suite.Suite, "../../tests/ibc-hooks/bytecode/counter.wasm")
	// instancetiate contract
	addr := suite.chainB.InstantiateContract(&suite.Suite, `{"count": 0}`, 1)
	suite.Require().NotEmpty(addr)

	memo := fmt.Sprintf(`{"wasm": {"contract": "%s", "msg": {"increment": {} } } }`, addr)

	msg := transfertypes.NewMsgTransfer(
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		sdk.NewCoin(sdk.DefaultBondDenom, transferAmount),
		suite.chainA.SenderAccount.GetAddress().String(),
		addr.String(),
		timeoutHeight,
		0,
		memo,
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

	senderLocalAcc, err := ibchookskeeper.DeriveIntermediateSender("channel-0", suite.chainA.SenderAccount.GetAddress().String(), "cosmos")
	suite.Require().NoError(err)

	state := suite.chainB.QueryContract(&suite.Suite, addr, []byte(fmt.Sprintf(`{"get_count": {"addr": "%s"}}`, senderLocalAcc)))
	suite.Require().Equal(`{"count":0}`, state)

	state = suite.chainB.QueryContract(&suite.Suite, addr, []byte(fmt.Sprintf(`{"get_total_funds": {"addr": "%s"}}`, senderLocalAcc)))
	suite.Require().Equal(`{"total_funds":[{"denom":"ibc/C053D637CCA2A2BA030E2C5EE1B28A16F71CCB0E45E8BE52766DC1B241B77878","amount":"1000000000"}]}`, state)
}

func (suite *IBCHooksTestSuite) TestAckHooks() {
	var (
		transferAmount = sdk.NewInt(1000000000)
		timeoutHeight  = clienttypes.NewHeight(0, 110)
	// when transfer via sdk transfer from A (module) -> B (contract)
	// nativeTokenSendOnChainA = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
	)

	suite.SetupTest() // reset

	path := NewTransferPath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)

	// store code
	suite.chainA.StoreContractCode(&suite.Suite, "../../tests/ibc-hooks/bytecode/counter.wasm")
	// instancetiate contract
	addr := suite.chainA.InstantiateContract(&suite.Suite, `{"count": 0}`, 1)
	suite.Require().NotEmpty(addr)

	fmt.Println(addr.String())

	// Generate swap instructions for the contract
	callbackMemo := fmt.Sprintf(`{"ibc_callback":"%s"}`, addr)

	msg := transfertypes.NewMsgTransfer(
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		sdk.NewCoin(sdk.DefaultBondDenom, transferAmount),
		suite.chainA.SenderAccount.GetAddress().String(),
		addr.String(),
		timeoutHeight,
		0,
		callbackMemo,
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

	state := suite.chainA.QueryContract(
		&suite.Suite, addr,
		[]byte(fmt.Sprintf(`{"get_count": {"addr": "%s"}}`, addr)))
	suite.Require().Equal(`{"count":1}`, state)

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

	state = suite.chainA.QueryContract(
		&suite.Suite, addr,
		[]byte(fmt.Sprintf(`{"get_count": {"addr": "%s"}}`, addr)))
	suite.Require().Equal(`{"count":2}`, state)
}

func (suite *IBCHooksTestSuite) TestTimeoutHooks() {
	var (
		transferAmount = sdk.NewInt(1000000000)
		timeoutHeight  = clienttypes.NewHeight(0, 500)
	// when transfer via sdk transfer from A (module) -> B (contract)
	// nativeTokenSendOnChainA = sdk.NewCoin(sdk.DefaultBondDenom, transferAmount)
	)

	suite.SetupTest() // reset

	path := NewTransferPath(suite.chainA, suite.chainB)
	suite.coordinator.Setup(path)

	// store code
	suite.chainA.StoreContractCode(&suite.Suite, "../../tests/ibc-hooks/bytecode/counter.wasm")
	// instancetiate contract
	addr := suite.chainA.InstantiateContract(&suite.Suite, `{"count": 0}`, 1)
	suite.Require().NotEmpty(addr)

	fmt.Println(addr.String())

	// Generate swap instructions for the contract
	callbackMemo := fmt.Sprintf(`{"ibc_callback":"%s"}`, addr)

	msg := transfertypes.NewMsgTransfer(
		path.EndpointA.ChannelConfig.PortID,
		path.EndpointA.ChannelID,
		sdk.NewCoin(sdk.DefaultBondDenom, transferAmount),
		suite.chainA.SenderAccount.GetAddress().String(),
		addr.String(),
		timeoutHeight,
		uint64(suite.coordinator.CurrentTime.Add(time.Minute).UnixNano()),
		callbackMemo,
	)
	sdkResult, err := suite.chainA.SendMsgs(msg)
	suite.Require().NoError(err)

	packet, err := customibctesting.ParsePacketFromEvents(sdkResult.GetEvents())
	suite.Require().NoError(err)

	// Move chainB forward one block
	suite.chainB.NextBlock()
	// One month later
	suite.coordinator.IncrementTimeBy(time.Hour)
	err = path.EndpointA.UpdateClient()
	suite.Require().NoError(err)

	err = path.EndpointA.TimeoutPacket(packet)
	suite.Require().NoError(err)

	// The test contract will increment the counter for itself by 10 when a packet times out
	state := suite.chainA.QueryContract(&suite.Suite, addr, []byte(fmt.Sprintf(`{"get_count": {"addr": "%s"}}`, addr)))
	suite.Require().Equal(`{"count":10}`, state)
}
