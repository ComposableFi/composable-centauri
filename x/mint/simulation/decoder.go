package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"

	"github.com/ComposableFi/composable-cosmos/v6/x/mint/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding mint type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		fmt.Println("kvA.Key[:1]:", kvA.Key[:1])
		switch {
		case bytes.Equal(kvA.Key[:1], types.MinterKey):
			fmt.Println("types.ParamsKey:", types.ParamsKey)
			var minterA, minterB types.Minter
			cdc.MustUnmarshal(kvA.Value, &minterA)
			cdc.MustUnmarshal(kvB.Value, &minterB)
			return fmt.Sprintf("%v\n%v", minterA, minterB)

		// case bytes.Equal(kvA.Key[:1], types.ParamsKey):
		// 	var paramsA, paramsB types.Params
		// 	cdc.MustUnmarshal(kvA.Value, &paramsA)
		// 	cdc.MustUnmarshal(kvB.Value, &paramsB)
		// 	return fmt.Sprintf("%v\n%v", paramsA, paramsB)

		default:
			panic(fmt.Sprintf("invalid mint key prefix %X", kvA.Key[:1]))
		}
	}
}
