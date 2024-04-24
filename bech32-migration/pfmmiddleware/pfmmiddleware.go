package pfmmiddleware

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	routertypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/notional-labs/composable/v6/app/keepers"
	"github.com/notional-labs/composable/v6/bech32-migration/utils"
)

func MigrateAddressBech32(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec, keepers *keepers.AppKeepers) {
	ctx.Logger().Info("Migration of address bech32 for pfmmiddleware module begin")
	totalAddr := uint64(0)

	store := ctx.KVStore(storeKey)

	channelKey := []byte("channel")
	iterator := sdk.KVStorePrefixIterator(store, channelKey)
	for ; iterator.Valid(); iterator.Next() {
		totalAddr++
		fullKey := iterator.Key()
		if !store.Has(fullKey) {
			continue
		}
		bz := store.Get(fullKey)
		var inFlightPacket routertypes.InFlightPacket
		cdc.MustUnmarshal(bz, &inFlightPacket)
		inFlightPacket.OriginalSenderAddress = utils.SafeConvertAddress(inFlightPacket.OriginalSenderAddress)
		var data transfertypes.FungibleTokenPacketData
		if err := transfertypes.ModuleCdc.UnmarshalJSON(inFlightPacket.PacketData, &data); err != nil {
			continue
		}
		data.Receiver = utils.SafeConvertAddress(data.Receiver)
		data.Sender = utils.SafeConvertAddress(data.Sender)

		d := make(map[string]interface{})
		err := json.Unmarshal([]byte(data.Memo), &d)
		// parse memo
		if err == nil && d["forward"] != nil {
			var m routertypes.PacketMetadata
			err = json.Unmarshal([]byte(data.Memo), &m)
			if err != nil {
				continue
			}
			m.Forward.Receiver = utils.SafeConvertAddress(m.Forward.Receiver)
			bzM, err := json.Marshal(m)
			if err != nil {
				continue
			}
			data.Memo = string(bzM)
		}
		bz = cdc.MustMarshal(&data)
		inFlightPacket.PacketData = bz
		bz = cdc.MustMarshal(&inFlightPacket)
		totalAddr++
		store.Set(fullKey, bz)
	}

	ctx.Logger().Info(
		"Migration of address bech32 for pfmmiddleware module done",
		"totalAddr", totalAddr,
	)
}
