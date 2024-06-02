package interchaintest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/chain/polkadot"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/relayer"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
)

// TestHyperspace features
// * sets up a Polkadot parachain
// * sets up a Cosmos chain
// * sets up the Hyperspace relayer
// * Funds a user wallet on both chains
// * Pushes a wasm client contract to the Cosmos chain
// * create client, connection, and channel in relayer
// * start relayer
// * send transfer over ibc
func TestCentauriPicassoIBCTransfer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	client, network := interchaintest.DockerSetup(t)

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	ctx := context.Background()

	nv := 5 // Number of validators
	nf := 3 // Number of full nodes

	consensusOverrides := make(testutil.Toml)
	blockTime := 5 // seconds, parachain is 12 second blocks, don't make relayer work harder than needed
	blockT := (time.Duration(blockTime) * time.Second).String()
	consensusOverrides["timeout_commit"] = blockT
	consensusOverrides["timeout_propose"] = blockT

	configTomlOverrides := make(testutil.Toml)
	configTomlOverrides["consensus"] = consensusOverrides

	configFileOverrides := make(map[string]any)
	configFileOverrides["config/config.toml"] = configTomlOverrides

	// Get both chains
	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			// Name:    "composable",
			// Version: "seunlanlege/centauri-polkadot:v0.9.27,seunlanlege/centauri-parachain:v0.9.27",
			ChainName: "composable", // Set ChainName so that a suffix with a "dash" is not appended (required for hyperspace)
			ChainConfig: ibc.ChainConfig{
				Type:    "polkadot",
				Name:    "composable",
				ChainID: "rococo-local",
				Images: []ibc.DockerImage{
					{
						Repository: "seunlanlege/centauri-polkadot",
						Version:    "v0.9.27",
						UidGid:     "1000:1000",
					},
					{
						Repository: "seunlanlege/centauri-parachain",
						Version:    "v0.9.27",
						// UidGid: "1025:1025",
					},
				},
				Bin:            "polkadot",
				Bech32Prefix:   "composable",
				Denom:          "uDOT",
				GasPrices:      "",
				GasAdjustment:  0,
				TrustingPeriod: "",
				CoinType:       "354",
			},
			NumValidators: &nv,
			NumFullNodes:  &nf,
		},
		{
			ChainConfig: ibc.ChainConfig{
				Type:           "cosmos",
				Name:           "centauri",
				ChainID:        "centaurid",
				Images:         []ibc.DockerImage{CentauriImage},
				Bin:            "centaurid",
				Bech32Prefix:   "centauri",
				Denom:          "stake",
				GasPrices:      "0.00stake",
				GasAdjustment:  1.3,
				TrustingPeriod: "504h",
				CoinType:       "118",
				// EncodingConfig: WasmClientEncoding(),
				NoHostMount:         true,
				ConfigFileOverrides: configFileOverrides,
				ModifyGenesis:       modifyGenesisShortProposals(votingPeriod, maxDepositPeriod),
			},
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	composable := chains[0].(*polkadot.PolkadotChain)
	centaurid := chains[1].(*cosmos.CosmosChain)

	// Get a relayer instance
	r := interchaintest.NewBuiltinRelayerFactory(
		ibc.Hyperspace,
		zaptest.NewLogger(t),
		// These two fields are used to pass in a custom Docker image built locally
		// relayer.ImagePull(false),
		relayer.CustomDockerImage("composablefi/hyperspace", "latest", "1000:1000"),
	).Build(t, client, network)

	// Build the network; spin up the chains and configure the relayer
	const pathName = "composable-centaurid"
	const relayerName = "hyperspace"

	ic := interchaintest.NewInterchain().
		AddChain(composable).
		AddChain(centaurid).
		AddRelayer(r, relayerName).
		AddLink(interchaintest.InterchainLink{
			Chain1:  composable,
			Chain2:  centaurid,
			Relayer: r,
			Path:    pathName,
		})

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:          t.Name(),
		Client:            client,
		NetworkID:         network,
		BlockDatabaseFile: interchaintest.DefaultBlockDatabaseFilepath(),
		SkipPathCreation:  true, // Skip path creation, so we can have granular control over the process
	}))

	fmt.Println("Interchain built")

	t.Cleanup(func() {
		_ = ic.Close()
	})
	// Create a proposal, vote, and wait for it to pass. Return code hash for relayer.
	codeHash := pushWasmContractViaGov(t, ctx, centaurid)

	// Set client contract hash in cosmos chain config
	err = r.SetClientContractHash(ctx, eRep, centaurid.Config(), codeHash)
	require.NoError(t, err)

	// Ensure parachain has started (starts 1 session/epoch after relay chain)
	err = testutil.WaitForBlocks(ctx, 1, composable)
	require.NoError(t, err, "polkadot chain failed to make blocks")

	// Fund users on both cosmos and parachain, mints Asset 1 for Alice
	fundAmount := int64(12_333_000_000_000)
	polkadotUser, cosmosUser := fundUsers(t, ctx, fundAmount, composable, centaurid)

	err = r.GeneratePath(ctx, eRep, centaurid.Config().ChainID, composable.Config().ChainID, pathName)
	require.NoError(t, err)

	// Create new clients
	err = r.CreateClients(ctx, eRep, pathName, ibc.DefaultClientOpts())
	require.NoError(t, err)
	err = testutil.WaitForBlocks(ctx, 1, centaurid, composable) // these 1 block waits may be needed, not sure
	require.NoError(t, err)

	// Create a new connection
	err = r.CreateConnections(ctx, eRep, pathName)
	require.NoError(t, err)
	err = testutil.WaitForBlocks(ctx, 1, centaurid, composable)
	require.NoError(t, err)

	// Create a new channel & get channels from each chain
	err = r.CreateChannel(ctx, eRep, pathName, ibc.DefaultChannelOpts())
	require.NoError(t, err)
	err = testutil.WaitForBlocks(ctx, 1, centaurid, composable)
	require.NoError(t, err)

	// Get channels - Query channels was removed
	/*cosmosChannelOutput, err := r.GetChannels(ctx, eRep, centaurid.Config().ChainID)
	require.NoError(t, err)
	require.Equal(t, len(cosmosChannelOutput), 1)
	require.Equal(t, cosmosChannelOutput[0].ChannelID, "channel-0")
	require.Equal(t, cosmosChannelOutput[0].PortID, "transfer")
	polkadotChannelOutput, err := r.GetChannels(ctx, eRep, composable.Config().ChainID)
	require.NoError(t, err)
	require.Equal(t, len(polkadotChannelOutput), 1)
	require.Equal(t, polkadotChannelOutput[0].ChannelID, "channel-0")
	require.Equal(t, polkadotChannelOutput[0].PortID, "transfer")*/

	// Start relayer
	err = r.StartRelayer(ctx, eRep, pathName)
	require.NoError(t, err)
	t.Cleanup(func() {
		err = r.StopRelayer(ctx, eRep)
		if err != nil {
			panic(err)
		}
	})

	// Send 1.77 stake from cosmosUser to parachainUser
	amountToSend := int64(1_770_000)
	transfer := ibc.WalletAmount{
		Address: polkadotUser.FormattedAddress(),
		Denom:   centaurid.Config().Denom,
		Amount:  amountToSend,
	}
	tx, err := centaurid.SendIBCTransfer(ctx, "channel-0", cosmosUser.KeyName(), transfer, ibc.TransferOptions{})
	require.NoError(t, err)
	require.NoError(t, tx.Validate()) // test source wallet has decreased funds
	err = testutil.WaitForBlocks(ctx, 5, centaurid, composable)
	require.NoError(t, err)

	/*// Trace IBC Denom of stake on parachain
	srcDenomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(cosmosChannelOutput[0].PortID, cosmosChannelOutput[0].ChannelID, centaurid.Config().Denom))
	dstIbcDenom := srcDenomTrace.IBCDenom()
	fmt.Println("Dst Ibc denom: ", dstIbcDenom)
	// Test destination wallet has increased funds, this is not working, want to verify IBC balance on parachain
	polkadotUserIbcCoins, err := composable.GetIbcBalance(ctx, string(polkadotUser.Address()))
	fmt.Println("UserIbcCoins: ", polkadotUserIbcCoins.String())
	aliceIbcCoins, err := composable.GetIbcBalance(ctx, "5yNZjX24n2eg7W6EVamaTXNQbWCwchhThEaSWB7V3GRjtHeL")
	fmt.Println("AliceIbcCoins: ", aliceIbcCoins.String())*/

	// Send 1.16 stake from parachainUser to cosmosUser
	amountToReflect := int64(1_160_000)
	reflectTransfer := ibc.WalletAmount{
		Address: cosmosUser.FormattedAddress(),
		Denom:   "2", // stake
		Amount:  amountToReflect,
	}
	_, err = composable.SendIBCTransfer(ctx, "channel-0", polkadotUser.KeyName(), reflectTransfer, ibc.TransferOptions{})
	require.NoError(t, err)

	// Send 1.88 "UNIT" from Alice to cosmosUser
	amountUnits := int64(1_880_000_000_000)
	unitTransfer := ibc.WalletAmount{
		Address: cosmosUser.FormattedAddress(),
		Denom:   "1", // UNIT
		Amount:  amountUnits,
	}
	_, err = composable.SendIBCTransfer(ctx, "channel-0", "alice", unitTransfer, ibc.TransferOptions{})
	require.NoError(t, err)

	// Wait for MsgRecvPacket on cosmos chain
	finalStakeBal := fundAmount - amountToSend + amountToReflect
	err = cosmos.PollForBalance(ctx, centaurid, 20, ibc.WalletAmount{
		Address: cosmosUser.FormattedAddress(),
		Denom:   centaurid.Config().Denom,
		Amount:  finalStakeBal,
	})
	require.NoError(t, err)

	// Verify final cosmos user "stake" balance
	cosmosUserStakeBal, err := centaurid.GetBalance(ctx, cosmosUser.FormattedAddress(), centaurid.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, finalStakeBal, cosmosUserStakeBal)
	// Verify final cosmos user "unit" balance
	unitDenomTrace := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom("transfer", "channel-0", "UNIT"))
	cosmosUserUnitBal, err := centaurid.GetBalance(ctx, cosmosUser.FormattedAddress(), unitDenomTrace.IBCDenom())
	require.NoError(t, err)
	require.Equal(t, amountUnits, cosmosUserUnitBal)
	/*polkadotUserIbcCoins, err = composable.GetIbcBalance(ctx, string(polkadotUser.Address()))
	fmt.Println("UserIbcCoins: ", polkadotUserIbcCoins.String())
	aliceIbcCoins, err = composable.GetIbcBalance(ctx, "5yNZjX24n2eg7W6EVamaTXNQbWCwchhThEaSWB7V3GRjtHeL")
	fmt.Println("AliceIbcCoins: ", aliceIbcCoins.String())*/

	fmt.Println("********************************")
	fmt.Println("********* Test passed **********")
	fmt.Println("********************************")

	// err = testutil.WaitForBlocks(ctx, 50, centaurid, composable)
	// require.NoError(t, err)
}

func pushWasmContractViaGov(t *testing.T, ctx context.Context, centaurid *cosmos.CosmosChain) string {
	// Set up cosmos user for pushing new wasm code msg via governance
	fundAmountForGov := int64(10_000_000_000)
	contractUsers := interchaintest.GetAndFundTestUsers(t, ctx, "default", fundAmountForGov, centaurid)
	contractUser := contractUsers[0]

	contractUserBalInitial, err := centaurid.GetBalance(ctx, contractUser.FormattedAddress(), centaurid.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, fundAmountForGov, contractUserBalInitial)

	proposal := cosmos.TxProposalv1{
		Metadata: "none",
		Deposit:  "500000000" + centaurid.Config().Denom, // greater than min deposit
		Title:    "Grandpa Contract",
		Summary:  "new grandpa contract",
	}

	proposalTx, codeHash, err := centaurid.PushNewWasmClientProposal(ctx, contractUser.KeyName(), "ics10_grandpa_cw.wasm", proposal)
	require.NoError(t, err, "error submitting new wasm contract proposal tx")

	height, err := centaurid.Height(ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")

	err = centaurid.VoteOnProposalAllValidators(ctx, proposalTx.ProposalID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	_, err = cosmos.PollForProposalStatus(ctx, centaurid, height, height+heightDelta, proposalTx.ProposalID, cosmos.ProposalStatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	err = testutil.WaitForBlocks(ctx, 1, centaurid)
	require.NoError(t, err)

	var getCodeQueryMsgRsp GetCodeQueryMsgResponse
	err = centaurid.QueryClientContractCode(ctx, codeHash, &getCodeQueryMsgRsp)
	codeHashByte32 := sha256.Sum256(getCodeQueryMsgRsp.Code)
	codeHash2 := hex.EncodeToString(codeHashByte32[:])
	t.Logf("Contract codeHash from code: %s", codeHash2)
	require.NoError(t, err)
	require.NotEmpty(t, getCodeQueryMsgRsp.Code)
	require.Equal(t, codeHash, codeHash2)

	return codeHash
}

func fundUsers(t *testing.T, ctx context.Context, fundAmount int64, composable, centaurid ibc.Chain) (ibc.Wallet, ibc.Wallet) {
	users := interchaintest.GetAndFundTestUsers(t, ctx, "user", fundAmount, composable, centaurid)
	polkadotUser, cosmosUser := users[0], users[1]
	err := testutil.WaitForBlocks(ctx, 2, composable, centaurid) // Only waiting 1 block is flaky for parachain
	require.NoError(t, err, "cosmos or polkadot chain failed to make blocks")

	// Check balances are correct
	polkadotUserAmount, err := composable.GetBalance(ctx, polkadotUser.FormattedAddress(), composable.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, fundAmount, polkadotUserAmount, "Initial polkadot user amount not expected")
	parachainUserAmount, err := composable.GetBalance(ctx, polkadotUser.FormattedAddress(), "")
	require.NoError(t, err)
	require.Equal(t, fundAmount, parachainUserAmount, "Initial parachain user amount not expected")
	cosmosUserAmount, err := centaurid.GetBalance(ctx, cosmosUser.FormattedAddress(), centaurid.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, fundAmount, cosmosUserAmount, "Initial cosmos user amount not expected")

	return polkadotUser, cosmosUser
}
