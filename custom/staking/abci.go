package bank

import (
	"context"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	// "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	customstakingkeeper "github.com/notional-labs/composable/v6/custom/staking/keeper"
)

// EndBlocker returns the end blocker for the staking module.
func EndBlocker(ctx context.Context, k *customstakingkeeper.Keeper) ([]abci.ValidatorUpdate, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	return k.BlockValidatorUpdates(ctx)
}
