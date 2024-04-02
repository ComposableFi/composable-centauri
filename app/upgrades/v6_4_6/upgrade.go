package v6_4_6

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/ComposableFi/composable-cosmos/v6/app/keepers"
	"github.com/ComposableFi/composable-cosmos/v6/app/upgrades"
	ibctransfermiddleware "github.com/ComposableFi/composable-cosmos/v6/x/ibctransfermiddleware/types"
	"github.com/cosmos/cosmos-sdk/codec"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	_ upgrades.BaseAppParamManager,
	_ codec.Codec,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Add params for custom middleware
		custommiddlewareparams := ibctransfermiddleware.DefaultGenesisState()
		keepers.IbcTransferMiddlewareKeeper.SetParams(ctx, custommiddlewareparams.Params)

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
