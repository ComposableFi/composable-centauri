package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/notional-labs/composable/v6/x/transfermiddleware/types"
)

// BeginBlocker of epochs module.
func (k Keeper) BeginBlocker(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Iterate over remove list
	k.IterateRemoveListInfo(sdkCtx, func(removeList types.RemoveParachainIBCTokenInfo) (stop bool) {
		// If pass the duration, remove parachain token info
		if sdkCtx.BlockTime().After(removeList.RemoveTime) {
			k.RemoveParachainIBCInfo(sdkCtx, removeList.NativeDenom)
		}
		return false
	})
	return nil
}
