package ibchooks

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/notional-labs/composable/v6/bech32-migration/utils"
)

func MigrateAddressBech32(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	ctx.Logger().Info("Migration of address bech32 for ibchooks module begin")
	totalAddr := uint64(0)
	store := ctx.KVStore(storeKey)
	channelKey := []byte("channel")
	iterator := sdk.KVStorePrefixIterator(store, channelKey)
	for ; iterator.Valid(); iterator.Next() {
		totalAddr++
		fullKey := iterator.Key()
		contract := string(store.Get(fullKey))
		contract = utils.SafeConvertAddress(contract)
		totalAddr++
		store.Set(fullKey, []byte(contract))
	}

	ctx.Logger().Info(
		"Migration of address bech32 for ibchooks module done",
		"totalAddr", totalAddr,
	)
}
