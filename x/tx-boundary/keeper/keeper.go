package keeper

import (
	"fmt"

	"cosmossdk.io/log"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/notional-labs/composable/v6/x/tx-boundary/types"
)

// Keeper struct
type Keeper struct {
	cdc      codec.Codec
	storeKey storetypes.StoreKey

	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper returns keeper
func NewKeeper(cdc codec.Codec, storeKey storetypes.StoreKey, authority string) Keeper {
	return Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		authority: authority,
	}
}

// GetAuthority returns the x/tx-boundary module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns logger
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetDelegateBoundary sets the delegate boundary.
func (k Keeper) SetDelegateBoundary(ctx sdk.Context, boundary types.Boundary) error {
	store := ctx.KVStore(k.storeKey)
	if boundary.BlocksPerGeneration == 0 {
		return fmt.Errorf("BlocksPerGeneration must not be zero")
	}
	bz := k.cdc.MustMarshal(&boundary)
	store.Set(types.DelegateBoundaryKey, bz)
	return nil
}

// GetDelegateBoundary sets the delegate boundary.
func (k Keeper) GetDelegateBoundary(ctx sdk.Context) (boundary types.Boundary) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.DelegateBoundaryKey)
	if bz == nil {
		panic("stored delegate boundary should not have been nil")
	}

	k.cdc.MustUnmarshal(bz, &boundary)
	return
}

// SetRedelegateBoundary sets the delegate boundary.
func (k Keeper) SetRedelegateBoundary(ctx sdk.Context, boundary types.Boundary) error {
	store := ctx.KVStore(k.storeKey)
	if boundary.BlocksPerGeneration == 0 {
		return fmt.Errorf("BlocksPerGeneration must not be zero")
	}
	bz := k.cdc.MustMarshal(&boundary)
	store.Set(types.RedelegateBoundaryKey, bz)
	return nil
}

// GetRedelegateBoundary sets the delegate boundary.
func (k Keeper) GetRedelegateBoundary(ctx sdk.Context) (boundary types.Boundary) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.RedelegateBoundaryKey)
	if bz == nil {
		panic("stored redelegate boundary should not have been nil")
	}

	k.cdc.MustUnmarshal(bz, &boundary)
	return
}

// SetDelegateCount set the number of delegate tx for a given address
func (k Keeper) SetLimitPerAddr(ctx sdk.Context, addr sdk.AccAddress, limitPerAddr types.LimitPerAddr) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&limitPerAddr)
	store.Set(addr, bz)
}

func (k Keeper) IncrementDelegateCount(ctx sdk.Context, addr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(addr) {
		k.SetLimitPerAddr(ctx, addr, types.LimitPerAddr{
			DelegateCount:     1,
			ReledegateCount:   0,
			LatestUpdateBlock: ctx.BlockHeight(),
		})
		return
	}
	bz := store.Get(addr)
	var limitPerAddr types.LimitPerAddr
	k.cdc.MustUnmarshal(bz, &limitPerAddr)
	limitPerAddr.DelegateCount++
	k.SetLimitPerAddr(ctx, addr, limitPerAddr)
}

func (k Keeper) IncrementRedelegateCount(ctx sdk.Context, addr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(addr) {
		k.SetLimitPerAddr(ctx, addr, types.LimitPerAddr{
			DelegateCount:     0,
			ReledegateCount:   1,
			LatestUpdateBlock: ctx.BlockHeight(),
		})
		return
	}
	bz := store.Get(addr)
	var limitPerAddr types.LimitPerAddr
	k.cdc.MustUnmarshal(bz, &limitPerAddr)
	limitPerAddr.ReledegateCount++
	k.SetLimitPerAddr(ctx, addr, limitPerAddr)
}

// GetDelegateCount get the number of delegate tx for a given address
func (k Keeper) GetLimitPerAddr(ctx sdk.Context, addr sdk.AccAddress) (limitPerAddr types.LimitPerAddr) {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(addr) {
		return types.LimitPerAddr{
			DelegateCount:     0,
			ReledegateCount:   0,
			LatestUpdateBlock: 0,
		}
	}
	bz := store.Get(addr)
	k.cdc.MustUnmarshal(bz, &limitPerAddr)
	return
}

func (k Keeper) UpdateLimitPerAddr(ctx sdk.Context, addr sdk.AccAddress) {
	limitPerAddr := k.GetLimitPerAddr(ctx, addr)
	if limitPerAddr.LatestUpdateBlock == 0 {
		return
	}
	boundary := k.GetDelegateBoundary(ctx)
	if limitPerAddr.LatestUpdateBlock+int64(boundary.BlocksPerGeneration) <= ctx.BlockHeight() {
		// Calculate the generated tx number from the duration between latest update block and current block height
		var generatedTx uint64

		duration := uint64(ctx.BlockHeight()) - uint64(limitPerAddr.LatestUpdateBlock)
		if duration/boundary.BlocksPerGeneration > 5 {
			generatedTx = 5
		} else {
			generatedTx = duration / boundary.BlocksPerGeneration
		}

		// Update the delegate tx limit
		if generatedTx > limitPerAddr.DelegateCount {
			limitPerAddr.DelegateCount = 0
		} else {
			limitPerAddr.DelegateCount -= generatedTx
		}
		// Update the redelegate tx limit
		if generatedTx > limitPerAddr.ReledegateCount {
			limitPerAddr.ReledegateCount = 0
		} else {
			limitPerAddr.ReledegateCount -= generatedTx
		}
		// Update LatestUpdateBlock
		limitPerAddr.LatestUpdateBlock = ctx.BlockHeight()

		// Set to store
		k.SetLimitPerAddr(ctx, addr, limitPerAddr)
		return
	}
}
