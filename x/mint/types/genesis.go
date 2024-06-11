package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// InflationCalculationFn defines the function required to calculate inflation rate during
// BeginBlock. It receives the minter and params stored in the keeper, along with the current
// bondedRatio and returns the newly calculated inflation rate.
// It can be used to specify a custom inflation calculation logic, instead of relying on the
// default logic provided by the sdk.
type InflationCalculationFn func(ctx sdk.Context, minter Minter, params Params, bondedRatio math.LegacyDec, totalStakingSupply math.Int) math.LegacyDec

// DefaultInflationCalculationFn is the default function used to calculate inflation.
func DefaultInflationCalculationFn(_ sdk.Context, minter Minter, params Params, bondedRatio math.LegacyDec, totalStakingSupply math.Int) math.LegacyDec {
	return minter.NextInflationRate(params, bondedRatio, totalStakingSupply)
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(minter Minter, params Params, incentives sdk.Coin) *GenesisState {
	return &GenesisState{
		Minter:           minter,
		Params:           params,
		IncentivesSupply: incentives,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Minter:           DefaultInitialMinter(),
		Params:           DefaultParams(),
		IncentivesSupply: sdk.NewCoin(stakingtypes.DefaultParams().BondDenom, math.NewInt(IncentivesSupply)),
	}
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants holds.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	return ValidateMinter(data.Minter)
}
