package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkquery "github.com/cosmos/cosmos-sdk/types/query"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/notional-labs/composable/v6/x/transfermiddleware/types"
)

func (k Keeper) ParaTokenInfo(c context.Context, req *types.QueryParaTokenInfoRequest) (*types.QueryParaTokenInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	info := k.GetParachainIBCTokenInfoByNativeDenom(ctx, req.NativeDenom)

	return &types.QueryParaTokenInfoResponse{
		IbcDenom:    info.IbcDenom,
		NativeDenom: info.NativeDenom,
		ChannelID:   info.ChannelID,
		AssetId:     info.AssetId,
	}, nil
}

func (k Keeper) EscrowAddress(_ context.Context, req *types.QueryEscrowAddressRequest) (*types.QueryEscrowAddressResponse, error) {
	escrowAddress := transfertypes.GetEscrowAddress(transfertypes.PortID, req.ChannelID)

	return &types.QueryEscrowAddressResponse{
		EscrowAddress: escrowAddress.String(),
	}, nil
}

func (k Keeper) RelayerAccount(c context.Context, req *types.QueryIBCWhiteListRequest) (*types.QueryIBCWhiteListResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)

	var whiteList []string

	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyRlyAddress)

	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	pageRes, err := sdkquery.FilteredPaginate(prefixStore, req.Pagination, func(key, _ []byte, accumulate bool) (bool, error) {
		if accumulate {
			whiteList = append(whiteList, string(key))
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryIBCWhiteListResponse{
		WhiteList:  whiteList,
		Pagination: pageRes,
	}, nil
}
