package mint

import (
	"context"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/notional-labs/composable/v6/x/mint/keeper"
	"github.com/notional-labs/composable/v6/x/mint/types"
)

// BeginBlocker mints new tokens for the previous block.
func BeginBlocker(ctx context.Context, k keeper.Keeper, ic types.InflationCalculationFn) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// fetch stored minter & params
	minter := k.GetMinter(sdkCtx)
	params := k.GetParams(sdkCtx)

	// recalculate inflation rate
	totalStakingSupply, _ := k.StakingTokenSupply(sdkCtx)
	bondedRatio, _ := k.BondedRatio(sdkCtx)
	minter.Inflation = ic(sdkCtx, minter, params, bondedRatio, totalStakingSupply)
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalStakingSupply)
	k.SetMinter(sdkCtx, minter)

	// calculate how many we would mint, but we dont mint them, we take them from the prefunded account
	mintedCoin := minter.BlockProvision(params)
	mintedCoins := sdk.NewCoins(mintedCoin)
	// send the minted coins to the fee collector account
	err := k.AddCollectedFees(sdkCtx, mintedCoins)
	if err != nil {
		k.Logger(sdkCtx).Info("Not enough incentive tokens in the mint pool to distribute")
	}

	if mintedCoin.Amount.IsInt64() {
		defer telemetry.ModuleSetGauge(types.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
	}

	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeReward,
			sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
		),
	)
	return nil
}
