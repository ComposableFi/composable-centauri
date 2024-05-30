package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	abcicometbft "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"

	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/core/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	mintkeeper "github.com/notional-labs/composable/v6/x/mint/keeper"
	minttypes "github.com/notional-labs/composable/v6/x/mint/types"
	stakingmiddleware "github.com/notional-labs/composable/v6/x/stakingmiddleware/keeper"
)

type Keeper struct {
	stakingkeeper.Keeper
	cdc               codec.BinaryCodec
	Stakingmiddleware *stakingmiddleware.Keeper
	authority         string
	mintKeeper        mintkeeper.Keeper
	distrKeeper       distkeeper.Keeper
	accountKeeper     types.AccountKeeper
}

func (k Keeper) BlockValidatorUpdates(ctx context.Context) ([]abcicometbft.ValidatorUpdate, error) {
	// Calculate validator set changes.
	//
	// NOTE: ApplyAndReturnValidatorSetUpdates has to come before
	// UnbondAllMatureValidatorQueue.
	// This fixes a bug when the unbonding period is instant (is the case in
	// some of the tests). The test expected the validator to be completely
	// unbonded after the Endblocker (go from Bonded -> Unbonding during
	// ApplyAndReturnValidatorSetUpdates and then Unbonding -> Unbonded during
	// UnbondAllMatureValidatorQueue).
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params := k.Stakingmiddleware.GetParams(sdkCtx)
	height := sdkCtx.BlockHeight()
	shouldExecuteBatch := (height % int64(params.BlocksPerEpoch)) == 0
	var validatorUpdates []abcicometbft.ValidatorUpdate
	if shouldExecuteBatch {
		k.Logger(ctx).Info("Should Execute ApplyAndReturnValidatorSetUpdates at height : ", height)
		v, err := k.ApplyAndReturnValidatorSetUpdates(ctx)
		if err != nil {
			panic(err)
		}
		validatorUpdates = v
	}

	// unbond all mature validators from the unbonding queue
	err := k.UnbondAllMatureValidators(ctx)
	if err != nil {
		return nil, err
	}

	// Remove all mature unbonding delegations from the ubd queue.
	matureUnbonds, err := k.DequeueAllMatureUBDQueue(ctx, sdkCtx.BlockHeader().Time)
	if err != nil {
		return nil, err
	}
	for _, dvPair := range matureUnbonds {
		addr, err := sdk.ValAddressFromBech32(dvPair.ValidatorAddress)
		if err != nil {
			return nil, err
		}
		delegatorAddress := sdk.MustAccAddressFromBech32(dvPair.DelegatorAddress)

		balances, err := k.CompleteUnbonding(ctx, delegatorAddress, addr)
		if err != nil {
			continue
		}

		sdkCtx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCompleteUnbonding,
				sdk.NewAttribute(sdk.AttributeKeyAmount, balances.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, dvPair.ValidatorAddress),
				sdk.NewAttribute(types.AttributeKeyDelegator, dvPair.DelegatorAddress),
			),
		)
	}

	// Remove all mature redelegations from the red queue.
	matureRedelegations, err := k.DequeueAllMatureRedelegationQueue(ctx, sdkCtx.BlockHeader().Time)
	if err != nil {
		return nil, err
	}
	for _, dvvTriplet := range matureRedelegations {
		valSrcAddr, err := sdk.ValAddressFromBech32(dvvTriplet.ValidatorSrcAddress)
		if err != nil {
			return nil, err
		}
		valDstAddr, err := sdk.ValAddressFromBech32(dvvTriplet.ValidatorDstAddress)
		if err != nil {
			return nil, err
		}
		delegatorAddress := sdk.MustAccAddressFromBech32(dvvTriplet.DelegatorAddress)

		balances, err := k.CompleteRedelegation(
			ctx,
			delegatorAddress,
			valSrcAddr,
			valDstAddr,
		)
		if err != nil {
			continue
		}

		sdkCtx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCompleteRedelegation,
				sdk.NewAttribute(sdk.AttributeKeyAmount, balances.String()),
				sdk.NewAttribute(types.AttributeKeyDelegator, dvvTriplet.DelegatorAddress),
				sdk.NewAttribute(types.AttributeKeySrcValidator, dvvTriplet.ValidatorSrcAddress),
				sdk.NewAttribute(types.AttributeKeyDstValidator, dvvTriplet.ValidatorDstAddress),
			),
		)
	}

	return validatorUpdates, nil
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService storetypes.KVStoreService,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	authority string,
	stakingmiddleware *stakingmiddleware.Keeper,
	validatorAddressCodec, consensusAddressCodec address.Codec,
) *Keeper {
	keeper := Keeper{
		Keeper:            *stakingkeeper.NewKeeper(cdc, storeService, ak, bk, authority, validatorAddressCodec, consensusAddressCodec),
		authority:         authority,
		Stakingmiddleware: stakingmiddleware,
		cdc:               cdc,
		mintKeeper:        mintkeeper.Keeper{},
		distrKeeper:       distkeeper.Keeper{},
		accountKeeper:     ak,
	}
	return &keeper
}

func (k *Keeper) RegisterKeepers(dk distkeeper.Keeper, mk mintkeeper.Keeper) {
	k.distrKeeper = dk
	k.mintKeeper = mk
}

// SlashWithInfractionReason send coins to community pool
func (k Keeper) SlashWithInfractionReason(ctx context.Context, consAddr sdk.ConsAddress, infractionHeight, power int64, slashFactor math.LegacyDec, _ types.Infraction) (math.Int, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// keep slashing logic the same
	amountBurned, err := k.Slash(ctx, consAddr, infractionHeight, power, slashFactor)
	if err != nil {
		return math.ZeroInt(), err
	}

	// after usual slashing and burning is done, mint burned coinds into community pool
	denom, err := k.BondDenom(ctx)
	if err != nil {
		return math.ZeroInt(), err
	}

	coins := sdk.NewCoins(sdk.NewCoin(denom, amountBurned))
	err = k.mintKeeper.MintCoins(sdkCtx, coins)
	if err != nil {
		k.Logger(ctx).Error("Failed to mint slashed coins: ", amountBurned)
	} else {
		err = k.distrKeeper.FundCommunityPool(ctx, coins, k.accountKeeper.GetModuleAddress(minttypes.ModuleName))
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to fund community pool. Tokens minted to the staking module account: %d. ", amountBurned))
		} else {
			sdkCtx.EventManager().EmitEvent(
				sdk.NewEvent(
					minttypes.EventTypeMintSlashed,
					sdk.NewAttribute(sdk.AttributeKeyAmount, amountBurned.String()),
				),
			)
		}
	}
	return amountBurned, nil
}
