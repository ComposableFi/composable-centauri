package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ComposableFi/composable-cosmos/v6/x/transfermiddleware/types"
)

// BeginBlocker of epochs module.
func (k Keeper) BeginBlocker(ctx sdk.Context) {
	// Iterate over remove list
	k.IterateRemoveListInfo(ctx, func(removeList types.RemoveParachainIBCTokenInfo) (stop bool) {
		// If pass the duration, remove parachain token info
		if ctx.BlockTime().After(removeList.RemoveTime) {
			k.RemoveParachainIBCInfo(ctx, removeList.NativeDenom)
		}
		return false
	})
}
