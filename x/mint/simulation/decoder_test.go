package simulation_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/ComposableFi/composable-cosmos/v6/x/mint/simulation"
	composableminttypes "github.com/ComposableFi/composable-cosmos/v6/x/mint/types"
)

func TestDecodeStore(t *testing.T) {
	cdc := testutil.MakeTestEncodingConfig().Codec
	dec := simulation.NewDecodeStore(cdc)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: composableminttypes.MinterKey, Value: cdc.MustMarshal(&composableminttypes.Minter{Inflation: sdk.NewDec(13), AnnualProvisions: sdk.NewDec(1)})},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectedLog string
	}{
		{"Minter", fmt.Sprintf("%v\n%v", composableminttypes.Minter{Inflation: sdk.NewDec(13), AnnualProvisions: sdk.NewDec(1)}, composableminttypes.Minter{Inflation: sdk.NewDec(13), AnnualProvisions: sdk.NewDec(1)})},
		{"other", ""},
	}
	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			switch i {
			case len(tests) - 1:
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			default:
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
