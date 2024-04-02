package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ComposableFi/composable-cosmos/v6/x/mint/types"
)

// InitGenesis new mint genesis
func (keeper Keeper) InitGenesis(ctx sdk.Context, ak types.AccountKeeper, data *types.GenesisState) {
	keeper.SetMinter(ctx, data.Minter)

	if err := keeper.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	newCoins := sdk.NewCoins(data.IncentivesSupply)
	if err := keeper.MintCoins(ctx, newCoins); err != nil {
		panic(err)
	}

	ak.GetModuleAccount(ctx, types.ModuleName)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func (keeper Keeper) ExportGenesis(ctx sdk.Context, authKeeper types.AccountKeeper) *types.GenesisState {
	minter := keeper.GetMinter(ctx)
	params := keeper.GetParams(ctx)

	remIncentives := keeper.bankKeeper.GetBalance(ctx, authKeeper.GetModuleAddress(types.ModuleName), params.MintDenom)
	return types.NewGenesisState(minter, params, remIncentives)
}
