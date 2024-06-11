package transfermiddleware_test

import (
	"encoding/json"
	"os"
	"testing"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	wasmkeeper "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/keeper"
	wasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	"github.com/stretchr/testify/suite"

	customibctesting "github.com/notional-labs/composable/v6/app/ibctesting"
)

var govAuthorityAddress = "pica10556m38z4x6pqalr9rl5ytf3cff8q46nf36090" // convert from: centauri10556m38z4x6pqalr9rl5ytf3cff8q46nk85k9m

type TransferTestSuite struct {
	suite.Suite

	coordinator *customibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *customibctesting.TestChain
	chainB *customibctesting.TestChain

	ctx      sdk.Context
	store    storetypes.KVStore
	testData map[string]string

	wasmKeeper wasmkeeper.Keeper
}

func (suite *TransferTestSuite) SetupTest() {
	suite.coordinator = customibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(customibctesting.GetChainID(1))
	suite.chainB = suite.coordinator.GetChain(customibctesting.GetChainID(2))

	suite.chainB.SetWasm(true)
	suite.coordinator.CommitNBlocks(suite.chainA, 2)
	suite.coordinator.CommitNBlocks(suite.chainB, 2)

	data, err := os.ReadFile("../../app/ibctesting/test_data/raw.json")
	suite.Require().NoError(err)
	err = json.Unmarshal(data, &suite.testData)
	suite.Require().NoError(err)

	suite.ctx = suite.chainB.GetContext().WithBlockGasMeter(storetypes.NewInfiniteGasMeter())
	suite.store = suite.chainB.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.ctx, "08-wasm-0")

	wasmContract, err := os.ReadFile("../../contracts/ics10_grandpa_cw.wasm")
	suite.Require().NoError(err)

	suite.wasmKeeper = suite.chainB.Wasm08Keeper()

	msg := wasmtypes.NewMsgStoreCode(govAuthorityAddress, wasmContract)

	response, err := suite.wasmKeeper.StoreCode(suite.ctx, msg)

	suite.Require().NoError(err)
	suite.Require().NotNil(response.Checksum)
	suite.coordinator.CodeID = response.Checksum
}

func TestTransferTestSuite(t *testing.T) {
	suite.Run(t, new(TransferTestSuite))
}

func (suite *TransferTestSuite) TestIbcAnteWithWasmUpdateClient() {
	suite.SetupTest()
	path := customibctesting.NewPath(suite.chainA, suite.chainB)
	suite.coordinator.SetupClients(path)

	// ensure counterparty has committed state
	suite.chainA.Coordinator.CommitBlock(suite.chainA)

	var header exported.ClientMessage
	header, err := suite.chainB.ConstructUpdateWasmClientHeader(suite.chainA, path.EndpointB.ClientID)
	suite.Require().NoError(err)

	msg, err := clienttypes.NewMsgUpdateClient(
		path.EndpointB.ClientID, header,
		suite.chainB.SenderAccount.GetAddress().String(),
	)
	suite.Require().NoError(err)

	_, err = suite.chainB.SendMsgsWithExpPass(false, msg)
	suite.Require().Error(err)
}

func (suite *TransferTestSuite) TestIbcAnteWithTenderMintUpdateClient() {
	suite.SetupTest()
	path := customibctesting.NewPath(suite.chainA, suite.chainB)
	suite.coordinator.SetupClients(path)

	// ensure counterparty has committed state
	suite.chainA.Coordinator.CommitBlock(suite.chainA)

	header, _ := suite.chainA.ConstructUpdateTMClientHeader(path.EndpointA.Chain, path.EndpointB.ClientID)

	msg, err := clienttypes.NewMsgUpdateClient(
		path.EndpointB.ClientID, header,
		suite.chainB.SenderAccount.GetAddress().String(),
	)
	suite.Require().NoError(err)

	_, err = suite.chainB.SendMsgsWithExpPass(false, msg)
	suite.Require().Error(err)
}
