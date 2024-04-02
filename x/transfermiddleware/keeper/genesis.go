package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ComposableFi/composable-cosmos/v6/x/transfermiddleware/types"
)

// TODO: add init genesis logic
// InitGenesis initializes the transfermiddleware module's state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	for _, tokenInfo := range genState.TokenInfos {
		k.AddParachainIBCInfo(ctx, tokenInfo.IbcDenom, tokenInfo.ChannelID, tokenInfo.NativeDenom, tokenInfo.AssetId)
	}
	k.SetParams(ctx, genState.Params)
}

// IterateParaTokenInfos iterate through all parachain token info.
func (k Keeper) IterateParaTokenInfos(ctx sdk.Context, fn func(index int64, info types.ParachainIBCTokenInfo) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.KeyParachainIBCTokenInfoByAssetID)
	defer iterator.Close()

	i := int64(0)

	for ; iterator.Valid(); iterator.Next() {
		info := types.ParachainIBCTokenInfo{}
		err := k.cdc.Unmarshal(iterator.Value(), &info)
		if err != nil {
			panic(err)
		}
		stop := fn(i, info)

		if stop {
			break
		}
		i++
	}
}

// ExportGenesis returns the x/transfermiddleware module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	infos := []types.ParachainIBCTokenInfo{}
	k.IterateParaTokenInfos(ctx, func(index int64, info types.ParachainIBCTokenInfo) (stop bool) {
		infos = append(infos, info)
		return false
	})

	return &types.GenesisState{
		TokenInfos: infos,
		Params:     k.GetParams(ctx),
	}
}
