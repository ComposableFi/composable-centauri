package ica

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	"github.com/notional-labs/composable/v6/bech32-migration/utils"
)

func MigrateAddressBech32(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	ctx.Logger().Info("Migration of address bech32 for ica host module begin")
	interchainAccountCount := uint64(0)

	store := ctx.KVStore(storeKey)
	iterator := sdk.KVStorePrefixIterator(store, []byte(icatypes.OwnerKeyPrefix))

	for ; iterator.Valid(); iterator.Next() {
		keySplit := strings.Split(string(iterator.Key()), "/")
		interchainAccountCount++
		connectionID := keySplit[2]
		portID := keySplit[1]
		address := utils.ConvertAccAddr(string(iterator.Value()))
		store.Set(icatypes.KeyOwnerAccount(portID, connectionID), []byte(address))
	}

	ctx.Logger().Info(
		"Migration of address bech32 for ica host module done",
		"interchain_account_count", interchainAccountCount,
	)
}
