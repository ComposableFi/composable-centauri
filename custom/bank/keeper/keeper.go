package keeper

import (
	"context"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	accountkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	banktypes "github.com/notional-labs/composable/v6/custom/bank/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	transfermiddlewarekeeper "github.com/notional-labs/composable/v6/x/transfermiddleware/keeper"
)

type Keeper struct {
	bankkeeper.BaseKeeper
	tfmk banktypes.TransferMiddlewareKeeper
	sk   banktypes.StakingKeeper
	ak   accountkeeper.AccountKeeper
}

var _ bankkeeper.Keeper = Keeper{}

func NewBaseKeeper(
	logger log.Logger,
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	ak accountkeeper.AccountKeeper,
	blockedAddrs map[string]bool,
	tfmk *transfermiddlewarekeeper.Keeper,
	authority string,
) Keeper {
	keeper := Keeper{
		BaseKeeper: bankkeeper.NewBaseKeeper(cdc, storeService, ak, blockedAddrs, authority, logger),
		tfmk:       tfmk,
		ak:         ak,
	}
	return keeper
}

func (k *Keeper) RegisterKeepers(ak accountkeeper.AccountKeeper, sk banktypes.StakingKeeper) {
	k.ak = ak
	k.sk = sk
}

// SupplyOf implements the Query/SupplyOf gRPC method
func (k Keeper) SupplyOf(c context.Context, req *types.QuerySupplyOfRequest) (*types.QuerySupplyOfResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}

	ctx := sdk.UnwrapSDKContext(c)
	supply := k.GetSupply(ctx, req.Denom)

	return &types.QuerySupplyOfResponse{Amount: sdk.NewCoin(req.Denom, supply.Amount)}, nil
}

// TotalSupply implements the Query/TotalSupply gRPC method
func (k Keeper) TotalSupply(ctx context.Context, req *types.QueryTotalSupplyRequest) (*types.QueryTotalSupplyResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	totalSupply, pageRes, err := k.GetPaginatedTotalSupply(sdkCtx, req.Pagination)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Get duplicate token from transfermiddeware
	duplicateCoins := k.tfmk.GetTotalEscrowedToken(sdkCtx)
	totalSupply = totalSupply.Sub(duplicateCoins...)

	return &types.QueryTotalSupplyResponse{Supply: totalSupply, Pagination: pageRes}, nil
}
