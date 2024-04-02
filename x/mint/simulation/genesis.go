package simulation

import (
	"math/rand"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/ComposableFi/composable-cosmos/v6/x/mint/types"
)

// Simulation parameter constants
const (
	Inflation           = "inflation"
	InflationRateChange = "inflation_rate_change"
	InflationMax        = "inflation_max"
	InflationMin        = "inflation_min"
	AnnualProvisions    = "annual_provisions"
	GoalBonded          = "goal_bonded"
)

// GenInflation randomized Inflation
func GenInflation(r *rand.Rand) math.LegacyDec {
	return sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
}

// GenInflationRateChange randomized InflationRateChange
func GenInflationRateChange(r *rand.Rand) math.LegacyDec {
	return sdk.NewDecWithPrec(int64(r.Intn(99)), 2)
}

// GenInflationMax randomized InflationMax
func GenInflationMax(r *rand.Rand) math.LegacyDec {
	return sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 10, 30)), 2)
}

// GenAnnualProvisions randomized AnnualProvisions
func GenAnnualProvisions(r *rand.Rand) math.LegacyDec {
	return sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 10)), 2)
}

// GenInflationMin randomized InflationMin
func GenInflationMin(r *rand.Rand) math.LegacyDec {
	return sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 10)), 2)
}

// GenGoalBonded randomized GoalBonded
func GenGoalBonded(r *rand.Rand) math.LegacyDec {
	return sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 50, 100)), 2)
}

// RandomizeGenState generates a random GenesisState for wasm
func RandomizedGenState(simState *module.SimulationState) {
	// minter
	var inflation sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, Inflation, &inflation, simState.Rand,
		func(r *rand.Rand) { inflation = GenInflation(r) },
	)

	// params
	var inflationRateChange sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, InflationRateChange, &inflationRateChange, simState.Rand,
		func(r *rand.Rand) { inflationRateChange = GenInflationRateChange(r) },
	)

	var annualProvisions sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, AnnualProvisions, &annualProvisions, simState.Rand,
		func(r *rand.Rand) { annualProvisions = GenAnnualProvisions(r) },
	)

	var goalBonded sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, GoalBonded, &goalBonded, simState.Rand,
		func(r *rand.Rand) { goalBonded = GenGoalBonded(r) },
	)

	var inflationMax sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, InflationMax, &inflationMax, simState.Rand,
		func(r *rand.Rand) { inflationMax = GenInflationMax(r) },
	)

	blocksPerYear := uint64(60 * 60 * 8766 / 5)
	// params := types.DefaultParams()
	mintGenesis := types.GenesisState{
		Minter: types.Minter{
			Inflation:        inflation,
			AnnualProvisions: annualProvisions,
		},
		Params: types.Params{
			MintDenom:           sdk.DefaultBondDenom,
			InflationRateChange: inflationRateChange,
			GoalBonded:          goalBonded,
			BlocksPerYear:       blocksPerYear,
			MaxTokenPerYear:     sdk.NewIntFromUint64(1000000000000000),
			MinTokenPerYear:     sdk.NewIntFromUint64(800000000000000),
		},
		IncentivesSupply: sdk.NewCoin(stakingtypes.DefaultParams().BondDenom, sdk.NewInt(100000000000)),
	}

	_, err := simState.Cdc.MarshalJSON(&mintGenesis)
	if err != nil {
		panic(err)
	}

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&mintGenesis)
}
