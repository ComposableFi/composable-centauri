package simulation

import (
	"math/rand"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/notional-labs/composable/v6/x/mint/types"
)

// Simulation operation weights constants
const (
	DefaultWeightMsgUpdateParams int = 100

	OpWeightMsgUpdateParams = "op_weight_msg_update_params" //nolint:gosec
)

// ProposalMsgs defines the module weighted proposals' contents
func ProposalMsgs() []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			OpWeightMsgUpdateParams,
			DefaultWeightMsgUpdateParams,
			SimulateMsgUpdateParams,
		),
	}
}

// SimulateMsgUpdateParams returns a random MsgUpdateParams
func SimulateMsgUpdateParams(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) sdk.Msg {
	// use the default gov module account address as authority
	var authority sdk.AccAddress = address.Module("gov")

	params := types.DefaultParams()
	params.BlocksPerYear = uint64(simtypes.RandIntBetween(r, 1, 60*60*8766))
	params.GoalBonded = sdkmath.LegacyNewDecWithPrec(int64(simtypes.RandIntBetween(r, 0, 100)), 2)
	params.InflationRateChange = sdkmath.LegacyNewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 20)), 2)
	params.MaxTokenPerYear = sdkmath.NewIntFromUint64(uint64(simtypes.RandIntBetween(r, 1000000000000000, 100000000000000000)))
	params.MinTokenPerYear = sdkmath.NewIntFromUint64(uint64(simtypes.RandIntBetween(r, 1, 1000000000000000)))
	params.MintDenom = simtypes.RandStringOfLength(r, 10)

	return &types.MsgUpdateParams{
		Authority: authority.String(),
		Params:    params,
	}
}
