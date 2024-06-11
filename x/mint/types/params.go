package types

import (
	"errors"
	"fmt"
	"strings"

	"sigs.k8s.io/yaml"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	Precision        = 2
	InflationRate    = 13 // inflation rate per year
	DesiredRatio     = 67 // the distance from the desired ratio (67%)
	MaxTokenPerYear  = 1000000000000000
	MinTokenPerYear  = 800000000000000
	BlockTime        = 5 // assuming 5 second block times
	IncentivesSupply = 100000000000
)

// Parameter store keys
var (
	KeyMintDenom           = []byte("MintDenom")
	KeyInflationRateChange = []byte("InflationRateChange")
	KeyInflationMax        = []byte("InflationMax")
	KeyInflationMin        = []byte("InflationMin")
	KeyGoalBonded          = []byte("GoalBonded")
	KeyBlocksPerYear       = []byte("BlocksPerYear")
	ParamsKey              = []byte("ParamsKey")
)

// ParamTable for minting module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string, inflationRateChange, _, _, goalBonded math.LegacyDec, blocksPerYear uint64, tokenPerYear math.Int,
) Params {
	return Params{
		MintDenom:           mintDenom,
		InflationRateChange: inflationRateChange,
		GoalBonded:          goalBonded,
		BlocksPerYear:       blocksPerYear,
		MaxTokenPerYear:     tokenPerYear,
		MinTokenPerYear:     tokenPerYear,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:           sdk.DefaultBondDenom,
		InflationRateChange: math.LegacyNewDecWithPrec(InflationRate, Precision),
		GoalBonded:          math.LegacyNewDecWithPrec(DesiredRatio, Precision),
		BlocksPerYear:       uint64(60 * 60 * 8766 / BlockTime),
		MaxTokenPerYear:     math.NewIntFromUint64(MaxTokenPerYear),
		MinTokenPerYear:     math.NewIntFromUint64(MinTokenPerYear),
	}
}

// validate params
func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateInflationRateChange(p.InflationRateChange); err != nil {
		return err
	}
	if err := validateGoalBonded(p.GoalBonded); err != nil {
		return err
	}
	if err := validateBlocksPerYear(p.BlocksPerYear); err != nil {
		return err
	}

	if p.MaxTokenPerYear.LT(p.MinTokenPerYear) {
		return fmt.Errorf(
			"MaxTokenPerYear (%s) must be greater than or equal to MinTokenPerYear (%s)",
			p.MinTokenPerYear, p.MaxTokenPerYear,
		)
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		paramtypes.NewParamSetPair(KeyInflationRateChange, &p.InflationRateChange, validateInflationRateChange),
		paramtypes.NewParamSetPair(KeyGoalBonded, &p.GoalBonded, validateGoalBonded),
		paramtypes.NewParamSetPair(KeyBlocksPerYear, &p.BlocksPerYear, validateBlocksPerYear),
	}
}

func validateMintDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("mint denom cannot be blank")
	}
	err := sdk.ValidateDenom(v)
	if err != nil {
		return err
	}

	return nil
}

func validateInflationRateChange(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("inflation rate change cannot be negative: %s", v)
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("inflation rate change too large: %s", v)
	}

	return nil
}

// func validateInflationMax(i interface{}) error {
// 	v, ok := i.(sdk.Dec)
// 	if !ok {
// 		return fmt.Errorf("invalid parameter type: %T", i)
// 	}

// 	if v.IsNegative() {
// 		return fmt.Errorf("max inflation cannot be negative: %s", v)
// 	}
// 	if v.GT(sdk.OneDec()) {
// 		return fmt.Errorf("max inflation too large: %s", v)
// 	}

// 	return nil
// }

// func validateInflationMin(i interface{}) error {
// 	v, ok := i.(sdk.Dec)
// 	if !ok {
// 		return fmt.Errorf("invalid parameter type: %T", i)
// 	}

// 	if v.IsNegative() {
// 		return fmt.Errorf("min inflation cannot be negative: %s", v)
// 	}
// 	if v.GT(sdk.OneDec()) {
// 		return fmt.Errorf("min inflation too large: %s", v)
// 	}

// 	return nil
// }

func validateGoalBonded(i interface{}) error {
	v, ok := i.(math.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() || v.IsZero() {
		return fmt.Errorf("goal bonded must be positive: %s", v)
	}
	if v.GT(math.LegacyOneDec()) {
		return fmt.Errorf("goal bonded too large: %s", v)
	}

	return nil
}

func validateBlocksPerYear(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("blocks per year must be positive: %d", v)
	}

	return nil
}
