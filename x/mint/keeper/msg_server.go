package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/ComposableFi/composable-cosmos/v6/x/mint/types"
)

var _ types.MsgServer = msgServer{}

// msgServer is a wrapper of Keeper.
type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the x/mint MsgServer interface.
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{
		Keeper: k,
	}
}

// UpdateParams updates the params.
func (ms msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (ms msgServer) FundModuleAccount(goCtx context.Context, req *types.MsgFundModuleAccount) (*types.MsgFundModuleAccountResponse, error) {
	// Unwrap context
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check sender address
	sender, err := sdk.AccAddressFromBech32(req.FromAddress)
	if err != nil {
		return nil, err
	}

	if !ms.IsAllowedAddress(ctx, req.FromAddress) {
		return nil, errorsmod.Wrapf(types.ErrInvalidAddress, "Invalid send address")
	}

	params := ms.GetParams(ctx)

	if len(req.Amount.Denoms()) > 1 || req.Amount[0].Denom != params.MintDenom {
		return nil, errorsmod.Wrapf(types.ErrInvalidCoin, "Invalid fund")
	}

	// Send Fund to account module
	moduleAccountAccAddress := ms.GetModuleAccountAccAddress(ctx)
	err = ms.bankKeeper.SendCoins(ctx, sender, moduleAccountAccAddress, req.Amount)
	if err != nil {
		return nil, err
	}

	return &types.MsgFundModuleAccountResponse{}, nil
}

func (ms msgServer) AddAccountToFundModuleSet(goCtx context.Context, req *types.MsgAddAccountToFundModuleSet) (*types.MsgAddAccountToFundModuleSetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if ms.authority != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	ms.SetAllowedAddress(ctx, req.AllowedAddress)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventAddAllowedFundAddress,
			sdk.NewAttribute(types.AttributeKeyAllowedAddress, req.AllowedAddress),
		),
	})

	return &types.MsgAddAccountToFundModuleSetResponse{}, nil
}
