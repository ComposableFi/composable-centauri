package simulation_test

import (
	"math/rand"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/ComposableFi/composable-cosmos/v6/x/mint/simulation"
	"github.com/ComposableFi/composable-cosmos/v6/x/mint/types"
)

func TestProposalMsgs(t *testing.T) {
	// initialize parameters
	s := rand.NewSource(1)
	r := rand.New(s)

	ctx := sdk.NewContext(nil, tmproto.Header{}, true, nil)
	accounts := simtypes.RandomAccounts(r, 3)

	// execute ProposalMsgs function
	weightedProposalMsgs := simulation.ProposalMsgs()
	assert.Assert(t, len(weightedProposalMsgs) == 1)

	w0 := weightedProposalMsgs[0]

	// tests w0 interface:
	assert.Equal(t, simulation.OpWeightMsgUpdateParams, w0.AppParamsKey())
	assert.Equal(t, simulation.DefaultWeightMsgUpdateParams, w0.DefaultWeight())

	msg := w0.MsgSimulatorFn()(r, ctx, accounts)
	msgUpdateParams, ok := msg.(*types.MsgUpdateParams)
	assert.Assert(t, ok)

	assert.Equal(t, sdk.AccAddress(address.Module("gov")).String(), msgUpdateParams.Authority)
	assert.Equal(t, uint64(20546551), msgUpdateParams.Params.BlocksPerYear)
	assert.DeepEqual(t, sdk.NewDecWithPrec(56, 2), msgUpdateParams.Params.GoalBonded)
	assert.DeepEqual(t, sdk.NewDecWithPrec(1, 2), msgUpdateParams.Params.InflationRateChange)
	assert.DeepEqual(t, sdk.NewInt(99997750760398084), msgUpdateParams.Params.MaxTokenPerYear)
	assert.DeepEqual(t, sdk.NewInt(504064263676792), msgUpdateParams.Params.MinTokenPerYear)
	assert.Equal(t, "XhhuTSkuxK", msgUpdateParams.Params.MintDenom)
}
