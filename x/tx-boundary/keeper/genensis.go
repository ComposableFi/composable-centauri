package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ComposableFi/composable-cosmos/v6/x/tx-boundary/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetDelegateBoundary(ctx, genState.DelegateBoundary)
	k.SetRedelegateBoundary(ctx, genState.RedelegateBoundary)
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesisState()

	genesis.DelegateBoundary = k.GetDelegateBoundary(ctx)
	genesis.RedelegateBoundary = k.GetRedelegateBoundary(ctx)

	return genesis
}
