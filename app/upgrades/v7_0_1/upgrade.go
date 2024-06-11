package v7_0_1

import (
	"context"
	"encoding/hex"
	"fmt"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/notional-labs/composable/v6/app/keepers"
	"github.com/notional-labs/composable/v6/app/upgrades"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	_ upgrades.BaseAppParamManager,
	codec codec.Codec,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(goCtx context.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		fmt.Println("start v7.0.1 upgrade")
		ctx := sdk.UnwrapSDKContext(goCtx)
		store := runtime.NewKVStoreService(keepers.GetKVStoreKey()[types.StoreKey]).OpenKVStore(ctx)

		// list code in testnet
		listCode := []string{
			"58c7623a3ab78f4cb2e4c5d02876ac36c3b38bb472118173a7ec7faa688a66d2", // local key
			"13ef83deb0dd1140f2e171aa72abbedee604e5903c4fb2a99da8bb13eda4dfb1",
			"1cff60adf40895b5fccb1e9ce6305a65ae01400a02cc4ded2cf3669221905adc",
			"24d21d60aa3bc4e84ddf21493ad1ada891a9af527def252edbba277ee7f05276",
			"292a4db6c6ea2cd0cc9852b4313bfc8c3727de5762f46552b59a7df8a33f14b2",
			"391add0fb14814296d134709797e697043c297841b23b14077bcaa09d67e7957",
			"8efc525173fb23ca7d62e9d6bd0a1ee98fd8f3e2374ae1ea8bb59a3ccdf34295",
			"936c3a1931746f9471d7088b236445979aabfe3af5378cdca67a032f5c4e4ed0",
			"9636e7c7e357280260d6db81df78ec78226c746239636cc0e42e5e306ab8e199",
			"b2adbd22fc3c410e781baeba5a655a0f0f6a009705ffe02b128ae6b4eabe3cf8",
			"d2e8126bc2226fb57e4fa8462da2a3441c3bead05c1161e848c944f99d9119ab",
			"ee97c9bd49a83282c2be1cd8cef0f2f55feb6d2d4c63ec3d27d69c252bd78531",
			"ef52ef690dc5ec88fd4dd78dc8fd7582904492284b0c290a12ef343d8a541056",
			"fccfde77e9318b1316e545a34414b5fc3e6cf82a5fe432815b956153c2e655bc",
		}
		listCheckSum := [][]byte{}
		for _, code := range listCode {
			checksumStr, err := hex.DecodeString(code)
			if err != nil {
				panic(err)
			}
			listCheckSum = append(listCheckSum, checksumStr)
		}

		// Register SendEnabled for legacy subspace

		checksum := types.Checksums{Checksums: listCheckSum}
		bz, err := codec.Marshal(&checksum)
		if err != nil {
			panic(err)
		}
		err = store.Set([]byte(types.KeyChecksums), bz)
		if err != nil {
			panic(err)
		}
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
