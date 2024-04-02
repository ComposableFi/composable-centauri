package gov

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/ComposableFi/composable-cosmos/v6/bech32-migration/utils"
)

func MigrateAddressBech32(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	ctx.Logger().Info("Migration of address bech32 for gov module begin")
	voteCount := uint64(0)
	utils.IterateStoreByPrefix(ctx, storeKey, types.VotesKeyPrefix, func(bz []byte) []byte {
		vote := v1beta1.Vote{}
		err := cdc.Unmarshal(bz, &vote)
		if err != nil {
			vote := v1.Vote{}
			cdc.MustUnmarshal(bz, &vote)
			vote.Voter = utils.ConvertAccAddr(vote.Voter)
			voteCount++
			return cdc.MustMarshal(&vote)
		}
		vote.Voter = utils.ConvertAccAddr(vote.Voter)
		voteCount++
		return cdc.MustMarshal(&vote)
	})
	depositCount := uint64(0)
	utils.IterateStoreByPrefix(ctx, storeKey, types.DepositsKeyPrefix, func(bz []byte) []byte {
		deposit := v1beta1.Deposit{}
		err := cdc.Unmarshal(bz, &deposit)
		if err != nil {
			vote := v1.Deposit{}
			cdc.MustUnmarshal(bz, &vote)
			deposit.Depositor = utils.ConvertAccAddr(deposit.Depositor)
			depositCount++
			return cdc.MustMarshal(&deposit)
		}
		deposit.Depositor = utils.ConvertAccAddr(deposit.Depositor)
		depositCount++
		return cdc.MustMarshal(&deposit)
	})
}
