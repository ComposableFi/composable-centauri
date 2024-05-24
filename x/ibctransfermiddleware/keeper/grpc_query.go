package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/notional-labs/composable/v6/x/ibctransfermiddleware/types"
)

var _ types.QueryServer = Keeper{}

// Params returns params of the staking middleware module.
func (k Keeper) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

// ChannelFees returns channel fees of the staking middleware module.
func (k Keeper) FeeConfigByChannelAndDenom(c context.Context, req *types.QueryFeeConfigByChannelAndDenomRequest) (*types.QueryFeeConfigByChannelAndDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	feeConfig := k.GetCoin(ctx, req.Channel, req.Denom)
	if feeConfig == nil {
		return nil, fmt.Errorf("fee configuration not found for channel %s and denom %s", req.Channel, req.Denom)
	} else {
		ret_fee_config := types.CoinItem{
			MinFee:        feeConfig.MinFee,
			Percentage:    feeConfig.Percentage,
			TxPriorityFee: feeConfig.TxPriorityFee,
		}
		return &types.QueryFeeConfigByChannelAndDenomResponse{Fees: ret_fee_config}, nil
	}

}
