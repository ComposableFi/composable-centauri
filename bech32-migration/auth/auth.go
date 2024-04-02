package auth

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"github.com/ComposableFi/composable-cosmos/v6/bech32-migration/utils"
)

func MigrateAddressBech32(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) {
	ctx.Logger().Info("Migration of address bech32 for auth module begin")
	migratedAccountCount := uint64(0)
	migratedAccountTypesStat := map[string]uint64{}
	utils.IterateStoreByPrefix(ctx, storeKey, types.AddressStoreKeyPrefix, func(bz []byte) []byte {
		var accountI types.AccountI
		err := cdc.UnmarshalInterface(bz, &accountI)
		if err != nil {
			panic(err)
		}
		switch acci := accountI.(type) {
		case *types.BaseAccount:
			acc := acci
			acc.Address = utils.ConvertAccAddr(acc.Address)
			migratedAccountTypesStat["BaseAccount"]++
		case *types.ModuleAccount:
			acc := acci
			acc.Address = utils.ConvertAccAddr(acc.Address)
			migratedAccountTypesStat["ModuleAccount"]++
		case *vestingtypes.BaseVestingAccount:
			acc := acci
			acc.Address = utils.ConvertAccAddr(acc.Address)
			migratedAccountTypesStat["BaseVestingAccount"]++
		case *vestingtypes.ContinuousVestingAccount:
			acc := acci
			acc.Address = utils.ConvertAccAddr(acc.Address)
			migratedAccountTypesStat["ContinuousVestingAccount"]++
		case *vestingtypes.DelayedVestingAccount:
			acc := acci
			acc.Address = utils.ConvertAccAddr(acc.Address)
			migratedAccountTypesStat["DelayedVestingAccount"]++
		case *vestingtypes.PeriodicVestingAccount:
			acc := acci
			acc.Address = utils.ConvertAccAddr(acc.Address)
			migratedAccountTypesStat["PeriodicVestingAccount"]++
		case *vestingtypes.PermanentLockedAccount:
			acc := acci
			acc.Address = utils.ConvertAccAddr(acc.Address)
			migratedAccountTypesStat["PermanentLockedAccount"]++
		default:
			ctx.Logger().Info(
				"Warning: unknown account type, skipping migration",
				"address", accountI.GetAddress().String(),
				"account_number", accountI.GetAccountNumber(),
				"public_key", accountI.GetPubKey(),
				"sequence", accountI.GetSequence(),
			)
			return bz
		}
		bz, err = cdc.MarshalInterface(accountI)
		if err != nil {
			panic(err)
		}
		migratedAccountCount++
		return bz
	})
	ctx.Logger().Info(
		"Migration of address bech32 for auth module done",
		"migrated_account_count", migratedAccountCount,
		"migrated_account_types_stat", migratedAccountTypesStat,
	)
}
