package keeper

import (
	"github.com/ComposableFi/composable-cosmos/v6/x/stakingmiddleware/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis new stake middleware genesis
func (keeper Keeper) InitGenesis(ctx sdk.Context, data *types.GenesisState) {
	if err := keeper.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (keeper Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := keeper.GetParams(ctx)
	return types.NewGenesisState(params)
}
