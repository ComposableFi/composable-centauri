package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibctmtypes "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"

	"github.com/notional-labs/composable/v6/x/ratelimit/types"
)

var _ types.QueryServer = Keeper{}

// Query all rate limits
func (k Keeper) AllRateLimits(goCtx context.Context, _ *types.QueryAllRateLimitsRequest) (*types.QueryAllRateLimitsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	rateLimits := k.GetAllRateLimits(ctx)
	return &types.QueryAllRateLimitsResponse{RateLimits: rateLimits}, nil
}

// Query a rate limit by denom and channelID
func (k Keeper) RateLimit(goCtx context.Context, req *types.QueryRateLimitRequest) (*types.QueryRateLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	rateLimit, found := k.GetRateLimit(ctx, req.Denom, req.ChannelID)
	if !found {
		return &types.QueryRateLimitResponse{}, nil
	}
	return &types.QueryRateLimitResponse{RateLimit: &rateLimit}, nil
}

// Query all rate limits for a given chain
func (k Keeper) RateLimitsByChainID(goCtx context.Context, req *types.QueryRateLimitsByChainIDRequest) (*types.QueryRateLimitsByChainIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	rateLimits := []types.RateLimit{}
	for _, rateLimit := range k.GetAllRateLimits(ctx) {

		// Determine the client state from the channel Id
		_, clientState, err := k.channelKeeper.GetChannelClientState(ctx, transfertypes.PortID, rateLimit.Path.ChannelID)
		if err != nil {
			return &types.QueryRateLimitsByChainIDResponse{}, errorsmod.Wrapf(types.ErrInvalidClientState, "Unable to fetch client state from channelID")
		}
		client, ok := clientState.(*ibctmtypes.ClientState)
		if !ok {
			return &types.QueryRateLimitsByChainIDResponse{}, errorsmod.Wrapf(types.ErrInvalidClientState, "Client state is not tendermint")
		}

		// If the chain ID matches, add the rate limit to the returned list
		if client.ChainId == req.ChainId {
			rateLimits = append(rateLimits, rateLimit)
		}
	}

	return &types.QueryRateLimitsByChainIDResponse{RateLimits: rateLimits}, nil
}

// Query all rate limits for a given channel
func (k Keeper) RateLimitsByChannelID(goCtx context.Context, req *types.QueryRateLimitsByChannelIDRequest) (*types.QueryRateLimitsByChannelIDResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	rateLimits := []types.RateLimit{}
	for _, rateLimit := range k.GetAllRateLimits(ctx) {
		// If the channel ID matches, add the rate limit to the returned list
		if rateLimit.Path.ChannelID == req.ChannelID {
			rateLimits = append(rateLimits, rateLimit)
		}
	}

	return &types.QueryRateLimitsByChannelIDResponse{RateLimits: rateLimits}, nil
}

// Query all whitelisted addresses
func (k Keeper) AllWhitelistedAddresses(goCtx context.Context, _ *types.QueryAllWhitelistedAddressesRequest) (*types.QueryAllWhitelistedAddressesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	whitelistedAddresses := k.GetAllWhitelistedAddressPairs(ctx)
	return &types.QueryAllWhitelistedAddressesResponse{AddressPairs: whitelistedAddresses}, nil
}
