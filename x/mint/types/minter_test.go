package types

import (
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// func TestNextInflation(t *testing.T) {
// 	minter := DefaultInitialMinter()
// 	params := DefaultParams()
// 	blocksPerYr := sdkmath.LegacyNewDec(int64(params.BlocksPerYear))

// 	// Governing Mechanism:
// 	//    inflationRateChangePerYear = (1- BondedRatio/ GoalBonded) * MaxInflationRateChange

// 	tests := []struct {
// 		bondedRatio, setInflation, expChange sdk.Dec
// 	}{
// 		// with 0% bonded atom supply the inflation should increase by InflationRateChange
// 		{sdk.ZeroDec(), sdkmath.LegacyNewDecWithPrec(7, 2), params.InflationRateChange.Quo(blocksPerYr)},

// 		// 100% bonded, starting at 20% inflation and being reduced
// 		// (1 - (1/0.67))*(0.13/8667)
// 		{
// 			sdk.OneDec(), sdkmath.LegacyNewDecWithPrec(20, 2),
// 			sdk.OneDec().Sub(sdk.OneDec().Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(blocksPerYr),
// 		},

// 		// 50% bonded, starting at 10% inflation and being increased
// 		{
// 			sdkmath.LegacyNewDecWithPrec(5, 1), sdkmath.LegacyNewDecWithPrec(10, 2),
// 			sdk.OneDec().Sub(sdkmath.LegacyNewDecWithPrec(5, 1).Quo(params.GoalBonded)).Mul(params.InflationRateChange).Quo(blocksPerYr),
// 		},

// 		// test 7% minimum stop (testing with 100% bonded)
// 		{sdk.OneDec(), sdkmath.LegacyNewDecWithPrec(7, 2), sdk.ZeroDec()},
// 		{sdk.OneDec(), sdkmath.LegacyNewDecWithPrec(700000001, 10), sdkmath.LegacyNewDecWithPrec(-1, 10)},

// 		// test 20% maximum stop (testing with 0% bonded)
// 		{sdk.ZeroDec(), sdkmath.LegacyNewDecWithPrec(20, 2), sdk.ZeroDec()},
// 		{sdk.ZeroDec(), sdkmath.LegacyNewDecWithPrec(1999999999, 10), sdkmath.LegacyNewDecWithPrec(1, 10)},

// 		// perfect balance shouldn't change inflation
// 		{sdkmath.LegacyNewDecWithPrec(67, 2), sdkmath.LegacyNewDecWithPrec(15, 2), sdk.ZeroDec()},
// 	}
// 	for i, tc := range tests {
// 		minter.Inflation = tc.setInflation

// 		inflation := minter.NextInflationRate(params, tc.bondedRatio, )
// 		diffInflation := inflation.Sub(tc.setInflation)

// 		require.True(t, diffInflation.Equal(tc.expChange),
// 			"Test Index: %v\nDiff:  %v\nExpected: %v\n", i, diffInflation, tc.expChange)
// 	}
// }

// func TestBlockProvision(t *testing.T) {
// 	minter := InitialMinter(sdkmath.LegacyNewDecWithPrec(1, 1))
// 	params := DefaultParams()

// 	secondsPerYear := int64(60 * 60 * 8766)

// 	tests := []struct {
// 		annualProvisions int64
// 		expProvisions    int64
// 	}{
// 		{secondsPerYear / 5, 1},
// 		{secondsPerYear/5 + 1, 1},
// 		{(secondsPerYear / 5) * 2, 2},
// 		{(secondsPerYear / 5) / 2, 0},
// 	}
// 	for i, tc := range tests {
// 		minter.AnnualProvisions = sdkmath.LegacyNewDec(tc.annualProvisions)
// 		provisions := minter.BlockProvision(params)

// 		expProvisions := sdk.NewCoin(params.MintDenom,
// 			sdkmath.NewInt(tc.expProvisions))

// 		require.True(t, expProvisions.IsEqual(provisions),
// 			"test: %v\n\tExp: %v\n\tGot: %v\n",
// 			i, tc.expProvisions, provisions)
// 	}
// }

// // Benchmarking :)
// // previously using sdk.Int operations:
// // BenchmarkBlockProvision-4 5000000 220 ns/op
// //
// // using sdk.Dec operations: (current implementation)
// // BenchmarkBlockProvision-4 3000000 429 ns/op
// func BenchmarkBlockProvision(b *testing.B) {
// 	b.ReportAllocs()
// 	minter := InitialMinter(sdkmath.LegacyNewDecWithPrec(1, 1))
// 	params := DefaultParams()

// 	s1 := rand.NewSource(100)
// 	r1 := rand.New(s1)
// 	minter.AnnualProvisions = sdkmath.LegacyNewDec(r1.Int63n(1000000))

// 	// run the BlockProvision function b.N times
// 	for n := 0; n < b.N; n++ {
// 		minter.BlockProvision(params)
// 	}
// }

// // Next inflation benchmarking
// // BenchmarkNextInflation-4 1000000 1828 ns/op
// func BenchmarkNextInflation(b *testing.B) {
// 	b.ReportAllocs()
// 	minter := InitialMinter(sdkmath.LegacyNewDecWithPrec(1, 1))
// 	params := DefaultParams()
// 	bondedRatio := sdkmath.LegacyNewDecWithPrec(1, 1)

// 	// run the NextInflationRate function b.N times
// 	for n := 0; n < b.N; n++ {
// 		minter.NextInflationRate(params, bondedRatio)
// 	}
// }

// // Next annual provisions benchmarking
// // BenchmarkNextAnnualProvisions-4 5000000 251 ns/op
// func BenchmarkNextAnnualProvisions(b *testing.B) {
// 	b.ReportAllocs()
// 	minter := InitialMinter(sdkmath.LegacyNewDecWithPrec(1, 1))
// 	params := DefaultParams()
// 	totalSupply := sdkmath.NewInt(100000000000000)

// 	// run the NextAnnualProvisions function b.N times
// 	for n := 0; n < b.N; n++ {
// 		minter.NextAnnualProvisions(params, totalSupply)
// 	}
// }

func TestSimulateMint(t *testing.T) {
	minter := DefaultInitialMinter()
	params := DefaultParams()
	totalSupply := sdkmath.NewInt(1_000_000_000_000_000_000)
	totalStaked := sdkmath.NewInt(0)
	tokenMinted := sdk.NewCoin("stake", sdkmath.NewInt(0))

	for i := 1; i <= int(params.BlocksPerYear); i++ {

		stakingDiff := sdkmath.LegacyNewDec(int64(rand.Intn(10))).QuoInt(sdkmath.NewInt(1_000_000)).MulInt(totalSupply)
		if (rand.Float32() > 0.5 || totalStaked.Add(stakingDiff.RoundInt()).GT(totalSupply)) && !totalStaked.Sub(stakingDiff.RoundInt()).IsNegative() {
			stakingDiff = stakingDiff.Neg()
		}
		totalStaked = totalStaked.Add(stakingDiff.RoundInt())
		bondedRatio := sdkmath.LegacyNewDecFromInt(totalStaked).Quo(sdkmath.LegacyNewDecFromInt(totalSupply))
		minter.Inflation = minter.NextInflationRate(params, bondedRatio, totalStaked)
		minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalStaked)

		// mint coins, update supply
		mintedCoin := minter.BlockProvision(params)
		tokenMinted = tokenMinted.Add(mintedCoin)
		// if i%100000 == 0 {
		// 	fmt.Println(i, bondedRatio, tokenMinted, mintedCoin, minter.Inflation, minter.AnnualProvisions)
		// }
	}
	require.True(t, params.MaxTokenPerYear.GTE(tokenMinted.Amount))
	require.True(t, params.MinTokenPerYear.LTE(tokenMinted.Amount))
}
