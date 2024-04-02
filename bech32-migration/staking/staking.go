package staking

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/ComposableFi/composable-cosmos/v6/bech32-migration/utils"
)

func MigrateAddressBech32(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	ctx.Logger().Info("Migration of address bech32 for staking module begin")
	validatorCount := uint64(0)
	utils.IterateStoreByPrefix(ctx, storeKey, types.ValidatorsKey, func(bz []byte) []byte {
		validator := types.MustUnmarshalValidator(cdc, bz)
		validator.OperatorAddress = utils.ConvertValAddr(validator.OperatorAddress)
		validatorCount++
		return types.MustMarshalValidator(cdc, &validator)
	})
	delegationCount := uint64(0)
	utils.IterateStoreByPrefix(ctx, storeKey, types.DelegationKey, func(bz []byte) []byte {
		delegation := types.MustUnmarshalDelegation(cdc, bz)
		delegation.DelegatorAddress = utils.ConvertAccAddr(delegation.DelegatorAddress)
		delegation.ValidatorAddress = utils.ConvertValAddr(delegation.ValidatorAddress)
		delegationCount++
		return types.MustMarshalDelegation(cdc, delegation)
	})
	redelegationCount := uint64(0)
	utils.IterateStoreByPrefix(ctx, storeKey, types.RedelegationKey, func(bz []byte) []byte {
		redelegation := types.MustUnmarshalRED(cdc, bz)
		redelegation.DelegatorAddress = utils.ConvertAccAddr(redelegation.DelegatorAddress)
		redelegation.ValidatorSrcAddress = utils.ConvertValAddr(redelegation.ValidatorSrcAddress)
		redelegation.ValidatorDstAddress = utils.ConvertValAddr(redelegation.ValidatorDstAddress)
		redelegationCount++
		return types.MustMarshalRED(cdc, redelegation)
	})
	unbondingDelegationCount := uint64(0)
	utils.IterateStoreByPrefix(ctx, storeKey, types.UnbondingDelegationKey, func(bz []byte) []byte {
		unbonding := types.MustUnmarshalUBD(cdc, bz)
		unbonding.DelegatorAddress = utils.ConvertAccAddr(unbonding.DelegatorAddress)
		unbonding.ValidatorAddress = utils.ConvertValAddr(unbonding.ValidatorAddress)
		unbondingDelegationCount++
		return types.MustMarshalUBD(cdc, unbonding)
	})
	historicalInfoCount := uint64(0)
	utils.IterateStoreByPrefix(ctx, storeKey, types.HistoricalInfoKey, func(bz []byte) []byte {
		historicalInfo := types.MustUnmarshalHistoricalInfo(cdc, bz)
		for i := range historicalInfo.Valset {
			historicalInfo.Valset[i].OperatorAddress = utils.ConvertValAddr(historicalInfo.Valset[i].OperatorAddress)
		}
		historicalInfoCount++
		return cdc.MustMarshal(&historicalInfo)
	})
	ctx.Logger().Info(
		"Migration of address bech32 for staking module done",
		"validator_count", validatorCount,
		"delegation_count", delegationCount,
		"redelegation_count", redelegationCount,
		"unbonding_delegation_count", unbondingDelegationCount,
		"historical_info_count", historicalInfoCount,
	)
}

func MigrateUnbonding(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	unbondingQueueKeyCount := uint64(0)
	utils.IterateStoreByPrefix(ctx, storeKey, types.UnbondingQueueKey, func(bz []byte) []byte {
		pairs := types.DVPairs{}
		cdc.MustUnmarshal(bz, &pairs)
		for i, pair := range pairs.Pairs {
			pairs.Pairs[i].DelegatorAddress = utils.ConvertAccAddr(pair.DelegatorAddress)
			pairs.Pairs[i].ValidatorAddress = utils.ConvertValAddr(pair.ValidatorAddress)
		}
		unbondingQueueKeyCount++
		return cdc.MustMarshal(&pairs)
	})

	redelegationQueueKeyCount := uint64(0)
	utils.IterateStoreByPrefix(ctx, storeKey, types.RedelegationQueueKey, func(bz []byte) []byte {
		triplets := types.DVVTriplets{}
		cdc.MustUnmarshal(bz, &triplets)

		for i, triplet := range triplets.Triplets {
			triplets.Triplets[i].DelegatorAddress = utils.ConvertAccAddr(triplet.DelegatorAddress)
			triplets.Triplets[i].ValidatorDstAddress = utils.ConvertValAddr(triplet.ValidatorDstAddress)
			triplets.Triplets[i].ValidatorSrcAddress = utils.ConvertValAddr(triplet.ValidatorSrcAddress)
		}
		redelegationQueueKeyCount++
		return cdc.MustMarshal(&triplets)
	})

	validatorQueueKeyCount := uint(0)
	utils.IterateStoreByPrefix(ctx, storeKey, types.ValidatorQueueKey, func(bz []byte) []byte {
		addrs := types.ValAddresses{}
		cdc.MustUnmarshal(bz, &addrs)

		for i, valAddress := range addrs.Addresses {
			addrs.Addresses[i] = utils.ConvertValAddr(valAddress)
		}
		validatorQueueKeyCount++
		return cdc.MustMarshal(&addrs)
	})

	ctx.Logger().Info(
		"Migration of address bech32 for staking unboding done",
		"unbondingQueueKeyCount", unbondingQueueKeyCount,
		"redelegationQueueKeyCount", redelegationQueueKeyCount,
		"validatorQueueKeyCount", validatorQueueKeyCount,
	)
}
