package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	mintkeeper "github.com/ComposableFi/composable-cosmos/v6/x/mint/keeper"
	tfmwkeeper "github.com/ComposableFi/composable-cosmos/v6/x/transfermiddleware/keeper"
)

var listAllowedRelayAddress = []string{
	"centauri1eqv3xl0vk0md74qukfghfff4z3axsp29rr9c85",
	"centauri1av6x9sll0yx4anske424jtgxejnrgqv6j6tjjt",
	"centauri1c8sxuxfgj5qj0l9gehs7any7s8mmx03qd7yd3f",
	"centauri17qv55sj9rgxs722wkkg0gewjv45msem90v6fpw",
}

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	tfmwKeeper tfmwkeeper.Keeper,
	mintKeeper mintkeeper.Keeper,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		// Add relayer address to store
		for _, allowedRelayAddress := range listAllowedRelayAddress {
			tfmwKeeper.SetAllowRlyAddress(ctx, allowedRelayAddress)
		}

		// enable staking reward
		mintParam := mintKeeper.GetParams(ctx)
		maxTokenPerYear, _ := sdk.NewIntFromString("99999999000000000000")
		minTokenPerYear, _ := sdk.NewIntFromString("99999999000000000000")

		mintParam.MaxTokenPerYear = maxTokenPerYear
		mintParam.MinTokenPerYear = minTokenPerYear

		err := mintKeeper.SetParams(ctx, mintParam)
		if err != nil {
			return vm, err
		}

		return mm.RunMigrations(ctx, configurator, vm)
	}
}
