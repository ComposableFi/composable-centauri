package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	accountkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	banktypes "github.com/ComposableFi/composable-cosmos/v6/custom/bank/types"

	transfermiddlewarekeeper "github.com/ComposableFi/composable-cosmos/v6/x/transfermiddleware/keeper"

	alliancekeeper "github.com/terra-money/alliance/x/alliance/keeper"
)

type Keeper struct {
	bankkeeper.BaseKeeper

	tfmk banktypes.TransferMiddlewareKeeper
	ak   alliancekeeper.Keeper
	sk   banktypes.StakingKeeper
	acck accountkeeper.AccountKeeper
}

var _ bankkeeper.Keeper = Keeper{}

func NewBaseKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ak accountkeeper.AccountKeeper,
	blockedAddrs map[string]bool,
	tfmk *transfermiddlewarekeeper.Keeper,
	authority string,
) Keeper {
	keeper := Keeper{
		BaseKeeper: bankkeeper.NewBaseKeeper(cdc, storeKey, ak, blockedAddrs, authority),
		ak:         alliancekeeper.Keeper{},
		sk:         stakingkeeper.Keeper{},
		tfmk:       tfmk,
		acck:       ak,
	}
	return keeper
}

func (k *Keeper) RegisterKeepers(ak alliancekeeper.Keeper, sk banktypes.StakingKeeper) {
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
