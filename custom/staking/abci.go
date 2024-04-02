package bank

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	customstakingkeeper "github.com/ComposableFi/composable-cosmos/v6/custom/staking/keeper"
)

// Called every block, update validator set
func EndBlocker(ctx sdk.Context, k *customstakingkeeper.Keeper) []abci.ValidatorUpdate {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	return k.BlockValidatorUpdates(ctx, ctx.BlockHeight())
}
