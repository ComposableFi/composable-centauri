package keeper

import (
	"context"
	"fmt"
	"time"

	custombankkeeper "github.com/ComposableFi/composable-cosmos/v6/custom/bank/keeper"
	ibctransfermiddlewaretypes "github.com/ComposableFi/composable-cosmos/v6/x/ibctransfermiddleware/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
)

type msgServer struct {
	Keeper
	bank      types.BankKeeper
	msgServer types.MsgServer
}

var _ types.MsgServer = msgServer{}

func NewMsgServerImpl(ibcKeeper Keeper, bankKeeper custombankkeeper.Keeper) types.MsgServer {
	return &msgServer{Keeper: ibcKeeper, bank: bankKeeper, msgServer: ibcKeeper.Keeper}
}

// Transfer is the server API around the Transfer method of the IBC transfer module.
// It checks if the sender is allowed to transfer the token and if the channel has fees.
// If the channel has fees, it will charge the sender and send the fees to the fee address.
// If the sender is not allowed to transfer the token because this tokens does not exists in the allowed tokens list, it just return without doing anything.
// If the sender is allowed to transfer the token, it will call the original transfer method.
// If the transfer amount is less than the minimum fee, it will charge the full transfer amount.
// If the transfer amount is greater than the minimum fee, it will charge the minimum fee and the percentage fee.
func (k msgServer) Transfer(goCtx context.Context, msg *types.MsgTransfer) (*types.MsgTransferResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := k.Keeper.IbcTransfermiddleware.GetParams(ctx)
	if params.ChannelFees != nil && len(params.ChannelFees) > 0 {
		channelFee := findChannelParams(params.ChannelFees, msg.SourceChannel)
		if channelFee != nil {
			if channelFee.MinTimeoutTimestamp > 0 {

				goCtx := sdk.UnwrapSDKContext(goCtx)
				blockTime := goCtx.BlockTime()

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

			send_err := k.bank.SendCoins(ctx, msgSender, feeAddress, sdk.NewCoins(sdk.NewCoin(msg.Token.Denom, charge)))
			if send_err != nil {
				return nil, send_err
			}

			if newAmount.LTE(sdk.ZeroInt()) {
				return &types.MsgTransferResponse{}, nil
			}
			msg.Token.Amount = newAmount
		}
	}
	return k.msgServer.Transfer(goCtx, msg)
}

func findChannelParams(channelFees []*ibctransfermiddlewaretypes.ChannelFee, targetChannelID string) *ibctransfermiddlewaretypes.ChannelFee {
	for _, fee := range channelFees {
		if fee.Channel == targetChannelID {
			return fee
		}
	}
	return nil // If the channel is not found
}

func findCoinByDenom(allowedTokens []*ibctransfermiddlewaretypes.CoinItem, denom string) *ibctransfermiddlewaretypes.CoinItem {
	for _, coin := range allowedTokens {
		if coin.MinFee.Denom == denom {
			return coin
		}
	}
	return nil // If the denom is not found
}
