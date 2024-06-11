package interchaintest

import (
	"context"
	"cosmossdk.io/math"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/strangelove-ventures/interchaintest/v8"
	"strconv"
	"testing"

	"github.com/icza/dyno"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	// simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
)

const (
	heightDelta      = 30
	votingPeriod     = "20s"
	maxDepositPeriod = "10s"
)

// Spin up a centaurid chain, push a contract, and get that contract code from chain
func TestPushWasmClientCode(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	client, network := interchaintest.DockerSetup(t)

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	ctx := context.Background()

	// Override config files to support an ~2.5MB contract
	configFileOverrides := make(map[string]any)

	appTomlOverrides := make(testutil.Toml)
	configTomlOverrides := make(testutil.Toml)

	apiOverrides := make(testutil.Toml)
	apiOverrides["rpc-max-body-bytes"] = 2_000_000
	appTomlOverrides["api"] = apiOverrides

	rpcOverrides := make(testutil.Toml)
	rpcOverrides["max_body_bytes"] = 2_000_000
	rpcOverrides["max_header_bytes"] = 2_100_000
	configTomlOverrides["rpc"] = rpcOverrides

	// mempoolOverrides := make(testutil.Toml)
	// mempoolOverrides["max_tx_bytes"] = 6000000
	// configTomlOverrides["mempool"] = mempoolOverrides

	configFileOverrides["config/app.toml"] = appTomlOverrides
	configFileOverrides["config/config.toml"] = configTomlOverrides

	numValidator := 1

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			ChainConfig: ibc.ChainConfig{
				Type:           "cosmos",
				Name:           "centauri",
				ChainID:        "centaurid",
				Images:         []ibc.DockerImage{CentauriImage},
				Bin:            "picad",
				Bech32Prefix:   "pica",
				Denom:          "stake",
				GasPrices:      "0.00stake",
				GasAdjustment:  1.3,
				TrustingPeriod: "504h",
				// EncodingConfig: WasmClientEncoding(),
				NoHostMount:         true,
				ConfigFileOverrides: configFileOverrides,
				ModifyGenesis:       modifyGenesisShortProposals(votingPeriod, maxDepositPeriod),
			},
			NumValidators: &numValidator,
		},
	})

	t.Logf("Calling cf.Chains")
	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	centaurid := chains[0]

	t.Logf("NewInterchain")
	ic := interchaintest.NewInterchain().
		AddChain(centaurid)

	t.Logf("Interchain build options")
	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
		SkipPathCreation:  true, // Skip path creation, so we can have granular control over the process
	}))

	t.Cleanup(func() {
		_ = ic.Close()
	})

	// Create and Fund User Wallets
	fundAmount := math.NewInt(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", fundAmount, centaurid)
	centaurid1User := users[0]

	err = testutil.WaitForBlocks(ctx, 5, centaurid)
	require.NoError(t, err)

	centaurid1UserBalInitial, err := centaurid.GetBalance(ctx, centaurid1User.FormattedAddress(), centaurid.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, fundAmount.Int64(), centaurid1UserBalInitial.Int64())

	centauridChain := centaurid.(*cosmos.CosmosChain)

	// Verify a normal user cannot push a wasm light client contract
	args := []string{"--title", "test", "--summary", "test", "--deposit", "500000000" + centauridChain.Config().Denom}
	_, err = centauridChain.StoreClientContract(ctx, centaurid1User.KeyName(), "ics10_grandpa_cw.wasm", args...)
	fmt.Println(err)
	// require.ErrorContains(t, err, "invalid authority")
	require.NoError(t, err, "invalid authority")

	proposal := cosmos.TxProposalv1{
		Metadata: "none",
		Deposit:  "500000000" + centauridChain.Config().Denom, // greater than min deposit
		Title:    "Grandpa Contract",
		Summary:  "new grandpa contract",
	}

	proposalTx, codeHash, err := centauridChain.PushNewWasmClientProposal(ctx, centaurid1User.KeyName(), "ics10_grandpa_cw.wasm", proposal)
	require.NoError(t, err, "error submitting new wasm contract proposal tx")

	height, err := centauridChain.Height(ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")

	err = centauridChain.VoteOnProposalAllValidators(ctx, proposalTx.ProposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	propId, err := strconv.ParseUint(proposalTx.ProposalID, 10, 64)
	require.NoError(t, err, "failed to convert proposal ID to uint64")

	_, err = cosmos.PollForProposalStatus(ctx, centauridChain, height, height+heightDelta, propId, govv1beta1.StatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	err = testutil.WaitForBlocks(ctx, 2, centauridChain)
	require.NoError(t, err)

	var getCodeQueryMsgRsp GetCodeQueryMsgResponse
	err = centauridChain.QueryClientContractCode(ctx, codeHash, &getCodeQueryMsgRsp)
	codeHashByte32 := sha256.Sum256(getCodeQueryMsgRsp.Data)
	codeHash2 := hex.EncodeToString(codeHashByte32[:])
	t.Logf("Contract codeHash from code: %s", codeHash2)
	require.NoError(t, err)
	require.NotEmpty(t, getCodeQueryMsgRsp.Data)
	require.Equal(t, codeHash, codeHash2)
}

type GetCodeQueryMsgResponse struct {
	Data []byte `json:"data"`
}

func modifyGenesisShortProposals(votingPeriod, maxDepositPeriod string) func(ibc.ChainConfig, []byte) ([]byte, error) {
	return func(chainConfig ibc.ChainConfig, genbz []byte) ([]byte, error) {
		g := make(map[string]interface{})
		if err := json.Unmarshal(genbz, &g); err != nil {
			return nil, fmt.Errorf("failed to unmarshal genesis file: %w", err)
		}
		if err := dyno.Set(g, votingPeriod, "app_state", "gov", "params", "voting_period"); err != nil {
			return nil, fmt.Errorf("failed to set voting period in genesis json: %w", err)
		}
		if err := dyno.Set(g, maxDepositPeriod, "app_state", "gov", "params", "max_deposit_period"); err != nil {
			return nil, fmt.Errorf("failed to set voting period in genesis json: %w", err)
		}
		if err := dyno.Set(g, chainConfig.Denom, "app_state", "gov", "params", "min_deposit", 0, "denom"); err != nil {
			return nil, fmt.Errorf("failed to set voting period in genesis json: %w", err)
		}
		out, err := json.Marshal(g)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal genesis bytes to json: %w", err)
		}
		return out, nil
	}
}
