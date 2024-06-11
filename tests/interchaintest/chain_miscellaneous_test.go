package interchaintest

import (
	"context"
	"fmt"
	"testing"

	"cosmossdk.io/math"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

var (
	numVals      = 1
	numFullNodes = 0
)

func TestICTestMiscellaneous(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	cosmos.SetSDKConfig("pica")

	sdk47Genesis := []cosmos.GenesisKV{
		cosmos.NewGenesisKV("app_state.gov.params.voting_period", "15s"),
		cosmos.NewGenesisKV("app_state.gov.params.max_deposit_period", "10s"),
		cosmos.NewGenesisKV("app_state.gov.params.min_deposit.0.denom", "ppica"),
		cosmos.NewGenesisKV("app_state.gov.params.min_deposit.0.amount", "1"),
	}

	config := CentauriConfig
	config.ModifyGenesis = cosmos.ModifyGenesis(sdk47Genesis)

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "centauri",
			ChainConfig:   config,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	chain := chains[0].(*cosmos.CosmosChain)

	ic := interchaintest.NewInterchain().
		AddChain(chain)

	ctx := context.Background()
	client, network := interchaintest.DockerSetup(t)

	require.NoError(t, ic.Build(ctx, nil, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: true,
	}))
	t.Cleanup(func() {
		_ = ic.Close()
	})

	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", math.NewInt(10_000_000_000), chain, chain)

	testWalletKeys(ctx, t, chain)
	testSendingTokens(ctx, t, chain, users)
	testFindTxs(ctx, t, chain, users) // not supported with CometMock
	testPollForBalance(ctx, t, chain, users)
	testRangeBlockMessages(ctx, t, chain, users)
	testBroadcaster(ctx, t, chain, users)
	testQueryCmd(ctx, t, chain)
	testHasCommand(ctx, t, chain)
	testFailedCWExecute(ctx, t, chain, users)
	testAddingNode(ctx, t, chain)
	testGetGovernanceAddress(ctx, t, chain)
	testTXFailsOnBlockInclusion(ctx, t, chain, users)
}

func wasmEncoding() *testutil.TestEncodingConfig {
	cfg := cosmos.DefaultEncoding()
	wasmtypes.RegisterInterfaces(cfg.InterfaceRegistry)
	return &cfg
}

func testFailedCWExecute(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain, users []ibc.Wallet) {
	user := users[0]
	keyName := user.KeyName()

	codeId, err := chain.StoreContract(ctx, keyName, "contracts/cw_template.wasm")
	if err != nil {
		t.Fatal(err)
	}

	contractAddr, err := chain.InstantiateContract(ctx, keyName, codeId, `{"count":0}`, true)
	if err != nil {
		t.Fatal(err)
	}

	// execute on the contract with the wrong message (err)
	txResp, err := chain.ExecuteContract(ctx, keyName, contractAddr, `{"not_a_func":{}}`)
	require.Error(t, err)
	fmt.Printf("txResp.RawLog: %+v\n", txResp.RawLog)
	fmt.Printf("err: %+v\n", err)
	require.Contains(t, err.Error(), "failed to execute message")
}

func testWalletKeys(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain) {
	// create a general key
	randKey := "randkey123"
	err := chain.CreateKey(ctx, randKey)
	require.NoError(t, err)

	// verify key was created properly
	_, err = chain.GetAddress(ctx, randKey)
	require.NoError(t, err)

	// recover a key
	// juno1hj5fveer5cjtn4wd6wstzugjfdxzl0xps73ftl
	keyName := "key-abc"
	testMnemonic := "decorate bright ozone fork gallery riot bus exhaust worth way bone indoor calm squirrel merry zero scheme cotton until shop any excess stage laundry"
	wallet, err := chain.BuildWallet(ctx, keyName, testMnemonic)
	require.NoError(t, err)

	// verify
	addr, err := chain.GetAddress(ctx, keyName)
	require.NoError(t, err)
	require.Equal(t, wallet.Address(), addr)

	tn := chain.Validators[0]
	a, err := tn.KeyBech32(ctx, "key-abc", "val")
	require.NoError(t, err)
	require.Equal(t, a, "picavaloper1hj5fveer5cjtn4wd6wstzugjfdxzl0xpaf78xu")

	a, err = tn.KeyBech32(ctx, "key-abc", "acc")
	require.NoError(t, err)
	require.Equal(t, a, wallet.FormattedAddress())

	a, err = tn.AccountKeyBech32(ctx, "key-abc")
	require.NoError(t, err)
	require.Equal(t, a, wallet.FormattedAddress())
}

func testSendingTokens(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain, users []ibc.Wallet) {
	_, err := chain.GetBalance(ctx, users[0].FormattedAddress(), chain.Config().Denom)
	require.NoError(t, err)
	b2, err := chain.GetBalance(ctx, users[1].FormattedAddress(), chain.Config().Denom)
	require.NoError(t, err)

	sendAmt := int64(1)
	_, err = sendTokens(ctx, chain, users[0], users[1], "", sendAmt)
	require.NoError(t, err)

	b2New, err := chain.GetBalance(ctx, users[1].FormattedAddress(), chain.Config().Denom)
	require.NoError(t, err)

	require.Equal(t, b2.Add(math.NewInt(sendAmt)), b2New)
}

func testFindTxs(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain, users []ibc.Wallet) {
	height, _ := chain.Height(ctx)

	_, err := sendTokens(ctx, chain, users[0], users[1], "", 1)
	require.NoError(t, err)

	txs, err := chain.FindTxs(ctx, height+1)
	require.NoError(t, err)
	require.NotEmpty(t, txs)
	require.Equal(t, txs[0].Events[0].Type, "tx")
}

func testPollForBalance(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain, users []ibc.Wallet) {
	bal2, err := chain.GetBalance(ctx, users[1].FormattedAddress(), chain.Config().Denom)
	require.NoError(t, err)

	amt := ibc.WalletAmount{
		Address: users[1].FormattedAddress(),
		Denom:   chain.Config().Denom,
		Amount:  math.NewInt(1),
	}

	delta := int64(3)

	ch := make(chan error)
	go func() {
		new := amt
		new.Amount = bal2.Add(math.NewInt(1))
		ch <- cosmos.PollForBalance(ctx, chain, delta, new)
	}()

	err = chain.SendFunds(ctx, users[0].KeyName(), amt)
	require.NoError(t, err)
	require.NoError(t, <-ch)
}

func testRangeBlockMessages(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain, users []ibc.Wallet) {
	height, _ := chain.Height(ctx)

	_, err := sendTokens(ctx, chain, users[0], users[1], "", 1)
	require.NoError(t, err)

	var bankMsgs []*banktypes.MsgSend
	err = cosmos.RangeBlockMessages(ctx, chain.Config().EncodingConfig.InterfaceRegistry, chain.Validators[0].Client, height+1, func(msg sdk.Msg) bool {
		found, ok := msg.(*banktypes.MsgSend)
		if ok {
			bankMsgs = append(bankMsgs, found)
		}
		return false
	})
	require.NoError(t, err)
}

func testAddingNode(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain) {
	// This should be tested last or else Txs will fail on the new full node.
	nodesAmt := len(chain.Nodes())
	chain.AddFullNodes(ctx, nil, 1)
	require.Equal(t, nodesAmt+1, len(chain.Nodes()))
}

func testBroadcaster(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain, users []ibc.Wallet) {
	from := users[0].FormattedAddress()
	addr1 := "pica190g5j8aszqhvtg7cprmev8xcxs6csra7tak0s2"
	addr2 := "pica1a53udazy8ayufvy0s434pfwjcedzqv34dspq0f"

	c1 := sdk.NewCoins(sdk.NewCoin(chain.Config().Denom, math.NewInt(1)))
	c2 := sdk.NewCoins(sdk.NewCoin(chain.Config().Denom, math.NewInt(2)))

	b := cosmos.NewBroadcaster(t, chain)

	in := banktypes.Input{
		Address: from,
		Coins:   c1.Add(c2[0]),
	}
	out := []banktypes.Output{
		{
			Address: addr1,
			Coins:   c1,
		},
		{
			Address: addr2,
			Coins:   c2,
		},
	}

	txResp, err := cosmos.BroadcastTx(
		ctx,
		b,
		users[0],
		banktypes.NewMsgMultiSend(in, out),
	)
	require.NoError(t, err)
	require.NotEmpty(t, txResp.TxHash)
	fmt.Printf("txResp: %+v\n", txResp)

	updatedBal1, err := chain.GetBalance(ctx, addr1, chain.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(1), updatedBal1)

	updatedBal2, err := chain.GetBalance(ctx, addr2, chain.Config().Denom)
	require.NoError(t, err)
	require.Equal(t, math.NewInt(2), updatedBal2)

	txResp, err = cosmos.BroadcastTx(
		ctx,
		b,
		users[0],
		banktypes.NewMsgMultiSend(banktypes.Input{
			Address: addr1,
			Coins:   c1.Add(c2[0]),
		}, out),
	)
	require.Error(t, err)
}

func testQueryCmd(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain) {
	tn := chain.Validators[0]
	stdout, stderr, err := tn.ExecQuery(ctx, "slashing", "params")
	require.NoError(t, err)
	require.NotEmpty(t, stdout)
	require.Empty(t, stderr)
}

func testHasCommand(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain) {
	tn := chain.Validators[0]
	res := tn.HasCommand(ctx, "query")
	require.True(t, res)

	if tn.IsAboveSDK47(ctx) {
		require.True(t, tn.HasCommand(ctx, "genesis"))
	} else {
		// 45 does not have this
		require.False(t, tn.HasCommand(ctx, "genesis"))
	}

	require.True(t, tn.HasCommand(ctx, "tx", "ibc"))
	require.True(t, tn.HasCommand(ctx, "q", "ibc"))
	require.True(t, tn.HasCommand(ctx, "keys"))
	require.True(t, tn.HasCommand(ctx, "help"))
	require.True(t, tn.HasCommand(ctx, "tx", "bank", "send"))

	require.False(t, tn.HasCommand(ctx, "tx", "bank", "send2notrealcmd"))
	require.False(t, tn.HasCommand(ctx, "incorrectcmd"))
}

func testGetGovernanceAddress(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain) {
	govAddr, err := chain.GetGovernanceAddress(ctx)
	require.NoError(t, err)
	_, err = chain.AccAddressFromBech32(govAddr)
	require.NoError(t, err)
}

func testTXFailsOnBlockInclusion(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain, users []ibc.Wallet) {
	// this isn't a real validator, but is well formed, so it will only fail once a validator checks the staking transaction
	fakeValoper, err := chain.GetNode().KeyBech32(ctx, users[0].KeyName(), "val")
	require.NoError(t, err)

	txHash, err := chain.GetNode().ExecTx(ctx, users[0].FormattedAddress(),
		"staking", "delegate", fakeValoper, "100"+chain.Config().Denom)
	transaction, err := chain.GetTransaction(txHash)
	require.NoError(t, err)
	require.Equal(t, "failed to execute message; message index: 0: validator does not exist", transaction.RawLog)
}

// helpers
func sendTokens(ctx context.Context, chain *cosmos.CosmosChain, from, to ibc.Wallet, token string, amount int64) (ibc.WalletAmount, error) {
	if token == "" {
		token = chain.Config().Denom
	}

	sendAmt := ibc.WalletAmount{
		Address: to.FormattedAddress(),
		Denom:   token,
		Amount:  math.NewInt(amount),
	}
	err := chain.SendFunds(ctx, from.KeyName(), sendAmt)
	return sendAmt, err
}

func validateBalance(ctx context.Context, t *testing.T, chain *cosmos.CosmosChain, user ibc.Wallet, tfDenom string, expected int64) {
	balance, err := chain.GetBalance(ctx, user.FormattedAddress(), tfDenom)
	require.NoError(t, err)
	require.Equal(t, balance, math.NewInt(expected))
}
