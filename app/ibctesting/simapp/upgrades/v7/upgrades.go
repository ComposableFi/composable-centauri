package v7

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	consensusparamskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	clientkeeper "github.com/cosmos/ibc-go/v8/modules/core/02-client/keeper"
	ibctmmigrations "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint/migrations"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// UpgradeName defines the on-chain upgrade name for the SimApp v7 upgrade.
	UpgradeName = "v7"
)

// CreateUpgradeHandler creates an upgrade handler for the v7 SimApp upgrade.
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	cdc codec.BinaryCodec,
	clientKeeper clientkeeper.Keeper,
	consensusParamsKeeper consensusparamskeeper.Keeper,
	paramsKeeper paramskeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// OPTIONAL: prune expired tendermint consensus states to save storage space
		sdkctx := sdk.UnwrapSDKContext(ctx)
		if _, err := ibctmmigrations.PruneExpiredConsensusStates(sdkctx, cdc, clientKeeper); err != nil {
			return nil, err
		}

		legacyBaseAppSubspace := paramsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
		baseapp.MigrateParams(sdkctx, legacyBaseAppSubspace, &consensusParamsKeeper.ParamsStore)

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
