package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/ComposableFi/composable-cosmos/v6/x/ibctransfermiddleware/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var _ types.MsgServer = msgServer{}

// msgServer is a wrapper of Keeper.
type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the x/ibctransfermiddleware MsgServer interface.
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{
		Keeper: k,
	}
}

// UpdateParams updates the params.
func (ms msgServer) UpdateCustomIbcParams(goCtx context.Context, req *types.MsgUpdateCustomIbcParams) (*types.MsgUpdateParamsCustomIbcResponse, error) {
	if ms.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsCustomIbcResponse{}, nil
}

// AddIBCFeeConfig(MsgAddIBCFeeConfig) returns (MsgAddIBCFeeConfigResponse);
func (ms msgServer) AddIBCFeeConfig(goCtx context.Context, req *types.MsgAddIBCFeeConfig) (*types.MsgAddIBCFeeConfigResponse, error) {
	if !contains(ms.addresses, req.Authority) && ms.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected of this addresses from list: %s, got %s", ms.addresses, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := sdk.AccAddressFromBech32(req.FeeAddress)
	if err != nil {
		return nil, err
	}

	params := ms.Keeper.GetParams(ctx)
	channelFee := findChannelParams(params.ChannelFees, req.ChannelID)
	if channelFee != nil {
		channelFee.FeeAddress = req.FeeAddress
		channelFee.MinTimeoutTimestamp = req.MinTimeoutTimestamp
	} else {
		channelFee := &types.ChannelFee{
			Channel:             req.ChannelID,
			FeeAddress:          req.FeeAddress,
			MinTimeoutTimestamp: req.MinTimeoutTimestamp,
			AllowedTokens:       []*types.CoinItem{},
		}
		params.ChannelFees = append(params.ChannelFees, channelFee)
	}
	errSetParams := ms.Keeper.SetParams(ctx, params)
	if errSetParams != nil {
		return nil, errSetParams
	}
	return &types.MsgAddIBCFeeConfigResponse{}, nil
}

func (ms msgServer) RemoveIBCFeeConfig(goCtx context.Context, req *types.MsgRemoveIBCFeeConfig) (*types.MsgRemoveIBCFeeConfigResponse, error) {
	if !contains(ms.addresses, req.Authority) && ms.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected of this addresses from list: %s, got %s", ms.addresses, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := ms.Keeper.GetParams(ctx)
	for i, fee := range params.ChannelFees {
		if fee.Channel == req.ChannelID {
			params.ChannelFees = append(params.ChannelFees[:i], params.ChannelFees[i+1:]...)
			break
		}
	}
	errSetParams := ms.Keeper.SetParams(ctx, params)
	if errSetParams != nil {
		return nil, errSetParams
	}

	return &types.MsgRemoveIBCFeeConfigResponse{}, nil
}

func (ms msgServer) AddAllowedIbcToken(goCtx context.Context, req *types.MsgAddAllowedIbcToken) (*types.MsgAddAllowedIbcTokenResponse, error) {
	if !contains(ms.addresses, req.Authority) && ms.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected of this addresses from list: %s, got %s", ms.addresses, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := ms.Keeper.GetParams(ctx)
	channelFee := findChannelParams(params.ChannelFees, req.ChannelID)
	if channelFee != nil {
		coin := findCoinByDenom(channelFee.AllowedTokens, req.Denom)
		if coin != nil {
			coin_c := sdk.Coin{
				Denom:  req.Denom,
				Amount: sdk.NewInt(req.Amount),
			}
			coin.MinFee = coin_c
			coin.Percentage = req.Percentage
		} else {
			coin_c := sdk.Coin{
				Denom:  req.Denom,
				Amount: sdk.NewInt(req.Amount),
			}
			coin := &types.CoinItem{
				MinFee:     coin_c,
				Percentage: req.Percentage,
			}
			channelFee.AllowedTokens = append(channelFee.AllowedTokens, coin)
		}
	} else {
		return nil, errorsmod.Wrapf(types.ErrChannelFeeNotFound, "channel fee not found for channel %s", req.ChannelID)
	}
	errSetParams := ms.Keeper.SetParams(ctx, params)
	if errSetParams != nil {
		return nil, errSetParams
	}

	return &types.MsgAddAllowedIbcTokenResponse{}, nil
}

func (ms msgServer) RemoveAllowedIbcToken(goCtx context.Context, req *types.MsgRemoveAllowedIbcToken) (*types.MsgRemoveAllowedIbcTokenResponse, error) {
	if !contains(ms.addresses, req.Authority) && ms.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected of this addresses from list: %s, got %s", ms.addresses, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := ms.Keeper.GetParams(ctx)
	channelFee := findChannelParams(params.ChannelFees, req.ChannelID)
	if channelFee != nil {
		for i, coin := range channelFee.AllowedTokens {
			if coin.MinFee.Denom == req.Denom {
				channelFee.AllowedTokens = append(channelFee.AllowedTokens[:i], channelFee.AllowedTokens[i+1:]...)
				break
			}
		}
	} else {
		return nil, errorsmod.Wrapf(types.ErrChannelFeeNotFound, "channel fee not found for channel %s", req.ChannelID)
	}

	errSetParams := ms.Keeper.SetParams(ctx, params)
	if errSetParams != nil {
		return nil, errSetParams
	}

	return &types.MsgRemoveAllowedIbcTokenResponse{}, nil
}

func findChannelParams(channelFees []*types.ChannelFee, targetChannelID string) *types.ChannelFee {
	for _, fee := range channelFees {
		if fee.Channel == targetChannelID {
			return fee
		}
	}
	return nil // If the channel is not found
}

func findCoinByDenom(allowedTokens []*types.CoinItem, denom string) *types.CoinItem {
	for _, coin := range allowedTokens {
		if coin.MinFee.Denom == denom {
			return coin
		}
	}
	return nil // If the denom is not found
}

func contains(arr []string, element string) bool {
	for _, v := range arr {
		if v == element {
			return true
		}
	}
	return false
}

// rpc AddAllowedIbcToken(MsgAddAllowedIbcToken)
//       returns (MsgAddAllowedIbcTokenResponse);
//   rpc RemoveAllowedIbcToken(MsgRemoveAllowedIbcToken)
//     returns (MsgRemoveAllowedIbcTokenResponse);
