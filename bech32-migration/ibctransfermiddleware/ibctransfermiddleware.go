package ibctransfermiddleware

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/notional-labs/composable/v6/bech32-migration/utils"
	"github.com/notional-labs/composable/v6/x/ibctransfermiddleware/types"
)

func MigrateAddressBech32(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	ctx.Logger().Info("Migration of address bech32 for ibctransfermiddleware module begin")
	totalAddr := uint64(0)
	store := ctx.KVStore(storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return
	}
	var params types.Params
	cdc.MustUnmarshal(bz, &params)
	for i := range params.ChannelFees {
		totalAddr++
		params.ChannelFees[i].FeeAddress = utils.SafeConvertAddress(params.ChannelFees[i].FeeAddress)
	}
	bz = cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)

	ctx.Logger().Info(
		"Migration of address bech32 for ibctransfermiddleware module done",
		"totalAddr", totalAddr,
	)
}
