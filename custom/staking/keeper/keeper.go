package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	abcicometbft "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"

	mintkeeper "github.com/ComposableFi/composable-cosmos/v6/x/mint/keeper"
	minttypes "github.com/ComposableFi/composable-cosmos/v6/x/mint/types"
	stakingmiddleware "github.com/ComposableFi/composable-cosmos/v6/x/stakingmiddleware/keeper"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type Keeper struct {
	stakingkeeper.Keeper
	cdc               codec.BinaryCodec
	Stakingmiddleware *stakingmiddleware.Keeper
	authority         string
	mintKeeper        mintkeeper.Keeper
	distrKeeper       distkeeper.Keeper
	authKeeper        minttypes.AccountKeeper
}

func (k Keeper) BlockValidatorUpdates(ctx sdk.Context, height int64) []abcicometbft.ValidatorUpdate {
	// Calculate validator set changes.
	//
	// NOTE: ApplyAndReturnValidatorSetUpdates has to come before
	// UnbondAllMatureValidatorQueue.
	// This fixes a bug when the unbonding period is instant (is the case in
	// some of the tests). The test expected the validator to be completely
	// unbonded after the Endblocker (go from Bonded -> Unbonding during
	// ApplyAndReturnValidatorSetUpdates and then Unbonding -> Unbonded during
	// UnbondAllMatureValidatorQueue).
	params := k.Stakingmiddleware.GetParams(ctx)
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
	k.UnbondAllMatureValidators(ctx)

	// Remove all mature unbonding delegations from the ubd queue.
	matureUnbonds := k.DequeueAllMatureUBDQueue(ctx, ctx.BlockHeader().Time)
	for _, dvPair := range matureUnbonds {
		addr, err := sdk.ValAddressFromBech32(dvPair.ValidatorAddress)
		if err != nil {
			panic(err)
		}
		delegatorAddress := sdk.MustAccAddressFromBech32(dvPair.DelegatorAddress)

		balances, err := k.CompleteUnbonding(ctx, delegatorAddress, addr)
		if err != nil {
			continue
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCompleteUnbonding,
				sdk.NewAttribute(sdk.AttributeKeyAmount, balances.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, dvPair.ValidatorAddress),
				sdk.NewAttribute(types.AttributeKeyDelegator, dvPair.DelegatorAddress),
			),
		)
	}

	// Remove all mature redelegations from the red queue.
	matureRedelegations := k.DequeueAllMatureRedelegationQueue(ctx, ctx.BlockHeader().Time)
	for _, dvvTriplet := range matureRedelegations {
		valSrcAddr, err := sdk.ValAddressFromBech32(dvvTriplet.ValidatorSrcAddress)
		if err != nil {
			panic(err)
		}
		valDstAddr, err := sdk.ValAddressFromBech32(dvvTriplet.ValidatorDstAddress)
		if err != nil {
			panic(err)
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

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCompleteRedelegation,
				sdk.NewAttribute(sdk.AttributeKeyAmount, balances.String()),
				sdk.NewAttribute(types.AttributeKeyDelegator, dvvTriplet.DelegatorAddress),
				sdk.NewAttribute(types.AttributeKeySrcValidator, dvvTriplet.ValidatorSrcAddress),
				sdk.NewAttribute(types.AttributeKeyDstValidator, dvvTriplet.ValidatorDstAddress),
			),
		)
	}

	return validatorUpdates
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	authority string,
	stakingmiddleware *stakingmiddleware.Keeper,
) *Keeper {
	keeper := Keeper{
		Keeper:            *stakingkeeper.NewKeeper(cdc, key, ak, bk, authority),
		authority:         authority,
		Stakingmiddleware: stakingmiddleware,
		cdc:               cdc,
		mintKeeper:        mintkeeper.Keeper{},
		distrKeeper:       distkeeper.Keeper{},
		authKeeper:        ak,
	}
	return &keeper
}

func (k *Keeper) RegisterKeepers(dk distkeeper.Keeper, mk mintkeeper.Keeper) {
	k.distrKeeper = dk
	k.mintKeeper = mk
}

// SlashWithInfractionReason send coins to community pool
func (k Keeper) SlashWithInfractionReason(ctx sdk.Context, consAddr sdk.ConsAddress, infractionHeight, power int64, slashFactor sdk.Dec, _ types.Infraction) math.Int {
	// keep slashing logic the same
	amountBurned := k.Slash(ctx, consAddr, infractionHeight, power, slashFactor)
	// after usual slashing and burning is done, mint burned coinds into community pool
	coins := sdk.NewCoins(sdk.NewCoin(k.BondDenom(ctx), amountBurned))
	err := k.mintKeeper.MintCoins(ctx, coins)
	if err != nil {
		k.Logger(ctx).Error("Failed to mint slashed coins: ", amountBurned)
	} else {
		err = k.distrKeeper.FundCommunityPool(ctx, coins, k.authKeeper.GetModuleAddress(minttypes.ModuleName))
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to fund community pool. Tokens minted to the staking module account: %d. ", amountBurned))
		} else {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					minttypes.EventTypeMintSlashed,
					sdk.NewAttribute(sdk.AttributeKeyAmount, amountBurned.String()),
				),
			)
		}
	}
	return amountBurned
}
