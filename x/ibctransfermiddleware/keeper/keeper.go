package keeper

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/notional-labs/composable/v6/x/ibctransfermiddleware/types"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
)

// Keeper of the staking middleware store
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string

	addresses []string
}

// NewKeeper creates a new middleware Keeper instance
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	authority string,
	addresses []string,
) Keeper {
	return Keeper{
		cdc:       cdc,
		storeKey:  key,
		authority: authority,
		addresses: addresses,
	}
}

// GetAuthority returns the x/ibctransfermiddleware module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SetParams sets the x/ibctransfermiddleware module parameters.
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&p)
	store.Set(types.ParamsKey, bz)
	return nil
}

// GetParams returns the current x/ibctransfermiddleware module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return p
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p
}

func (k Keeper) GetSequenceFee(ctx sdk.Context, sequence uint64) (coin sdk.Coin, found bool) {
	store := ctx.KVStore(k.storeKey)

	value := store.Get(types.GetSequenceKey(sequence))
	if value == nil {
		return sdk.Coin{}, false
	}

	fee := types.MustUnmarshalCoin(k.cdc, value)
	return fee, true
}

func (k Keeper) SetSequenceFee(ctx sdk.Context, sequence uint64, coin sdk.Coin) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetSequenceKey(sequence), types.MustMarshalCoin(k.cdc, &coin))
}

func (k Keeper) DeleteSequenceFee(ctx sdk.Context, sequence uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetSequenceKey(sequence))
}

func (k Keeper) GetCoin(ctx sdk.Context, targetChannelID, denom string) *types.CoinItem {
	params := k.GetParams(ctx)
	channelFee := findChannelParams(params.ChannelFees, targetChannelID)
	if channelFee == nil {
		return nil
	}
	return findCoinByDenom(channelFee.AllowedTokens, denom)
}

func (k Keeper) GetChannelFeeAddress(ctx sdk.Context, targetChannelID string) string {
	params := k.GetParams(ctx)
	channelFee := findChannelParams(params.ChannelFees, targetChannelID)
	if channelFee == nil {
		return ""
	}
	return channelFee.FeeAddress
}

type BridgeFee struct {
	Fee      sdk.Coin
	Sender   sdk.AccAddress
	Receiver sdk.AccAddress
}

func (k Keeper) GetBridgeFeeBasedOnConfigForChannelAndDenom(ctx sdk.Context, msg *ibctypes.MsgTransfer) (*BridgeFee, error) {
	params := k.GetParams(ctx)
	// charge_coin := sdk.NewCoin(msg.Token.Denom, sdk.ZeroInt())
	if params.ChannelFees != nil && len(params.ChannelFees) > 0 {
		channelFee := findChannelParams(params.ChannelFees, msg.SourceChannel)
		if channelFee != nil {
			if channelFee.MinTimeoutTimestamp > 0 {

				blockTime := ctx.BlockTime()

				timeoutTimeInFuture := time.Unix(0, int64(msg.TimeoutTimestamp))
				if timeoutTimeInFuture.Before(blockTime) {
					return nil, fmt.Errorf("incorrect timeout timestamp found during ibc transfer. timeout timestamp is in the past")
				}

				difference := timeoutTimeInFuture.Sub(blockTime).Nanoseconds()
				if difference < channelFee.MinTimeoutTimestamp {
					return nil, fmt.Errorf("incorrect timeout timestamp found during ibc transfer. too soon")
				}
			}
			coin := findCoinByDenom(channelFee.AllowedTokens, msg.Token.Denom)
			if coin == nil {
				return nil, fmt.Errorf("token not allowed to be transferred in this channel")
			}

			minFee := coin.MinFee.Amount
			priority := GetPriority(msg.Memo)
			if priority != nil {
				p := findPriority(coin.TxPriorityFee, *priority)
				if p != nil && coin.MinFee.Denom == p.PriorityFee.Denom {
					minFee = minFee.Add(p.PriorityFee.Amount)
				}
			}

			charge := minFee
			if charge.GT(msg.Token.Amount) {
				charge = msg.Token.Amount
			}

			newAmount := msg.Token.Amount.Sub(charge)

			if newAmount.IsPositive() {
				percentageCharge := newAmount.QuoRaw(coin.Percentage)
				newAmount = newAmount.Sub(percentageCharge)
				charge = charge.Add(percentageCharge)
			}

			msgSender, err := sdk.AccAddressFromBech32(msg.Sender)
			if err != nil {
				return nil, err
			}

			feeAddress, err := sdk.AccAddressFromBech32(channelFee.FeeAddress)
			if err != nil {
				return nil, err
			}

			charge_coin := sdk.NewCoin(msg.Token.Denom, charge)
			// send_err := k.bank.SendCoins(ctx, msgSender, feeAddress, sdk.NewCoins(charge_coin))
			// if send_err != nil {
			// 	return nil, send_err
			// }
			msg.Token.Amount = newAmount
			return &BridgeFee{Fee: charge_coin, Sender: msgSender, Receiver: feeAddress}, nil

			// if newAmount.LTE(sdk.ZeroInt()) {
			// 	zeroTransfer := sdk.NewCoin(msg.Token.Denom, sdk.ZeroInt())
			// 	return &zeroTransfer, nil
			// }
		}
	}
	// ret, err := k.Keeper.Transfer(goCtx, msg)
	// if err == nil && ret != nil && !charge_coin.IsZero() {
	// if !charge_coin.IsZero() {
	// 	k.SetSequenceFee(ctx, ret.Sequence, charge_coin)
	// }
	return nil, nil
}

func GetPriority(jsonString string) *string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonString), &data); err != nil {
		return nil
	}

	priority, ok := data["priority"].(string)
	if !ok {
		return nil
	}

	return &priority
}

func findPriority(priorities []*types.TxPriorityFee, priority string) *types.TxPriorityFee {
	for _, p := range priorities {
		if p.Priority == priority {
			return p
		}
	}
	return nil
}
