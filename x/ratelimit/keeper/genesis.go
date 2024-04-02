package keeper

import (
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ComposableFi/composable-cosmos/v6/x/ratelimit/types"
)

func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	for _, rateLimit := range genState.RateLimits {
		k.SetRateLimit(ctx, rateLimit)
	}
	for _, addressPair := range genState.WhitelistedAddressPairs {
		k.SetWhitelistedAddressPair(ctx, addressPair)
	}
	for _, pendingPacketID := range genState.PendingSendPacketSequenceNumbers {
		splits := strings.Split(pendingPacketID, "/")
		if len(splits) != 2 {
			panic("Invalid pending send packet, must be of form: {channelID}/{sequenceNumber}")
		}
		channelID := splits[0]
		sequence, err := strconv.ParseUint(splits[1], 10, 64)
		if err != nil {
			panic(err)
		}
		k.SetPendingSendPacket(ctx, channelID, sequence)
	}
	for _, epoch := range genState.Epochs {
		err := k.AddEpochInfo(ctx, epoch)
		if err != nil {
			panic(err)
		}
	}
}

func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesisState()

	genesis.Params = k.GetParams(ctx)
	genesis.RateLimits = k.GetAllRateLimits(ctx)
	genesis.WhitelistedAddressPairs = k.GetAllWhitelistedAddressPairs(ctx)
	genesis.PendingSendPacketSequenceNumbers = k.GetAllPendingSendPackets(ctx)

	return genesis
}
