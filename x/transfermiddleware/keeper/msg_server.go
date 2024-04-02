package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"cosmossdk.io/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/ComposableFi/composable-cosmos/v6/x/transfermiddleware/types"
)

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{
		Keeper: keeper,
	}
}

type msgServer struct {
	Keeper
}

func (ms msgServer) AddParachainIBCTokenInfo(goCtx context.Context, req *types.MsgAddParachainIBCTokenInfo) (*types.MsgAddParachainIBCTokenInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if ms.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	err := ms.AddParachainIBCInfo(ctx, req.IbcDenom, req.ChannelID, req.NativeDenom, req.AssetId)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventAddParachainIBCTokenInfo,
			sdk.NewAttribute(types.AttributeKeyNativeDenom, req.NativeDenom),
			sdk.NewAttribute(types.AttributeKeyIbcDenom, req.IbcDenom),
			sdk.NewAttribute(types.AttributeKeyAssetID, req.AssetId),
		),
	})

	return &types.MsgAddParachainIBCTokenInfoResponse{}, nil
}

func (ms msgServer) RemoveParachainIBCTokenInfo(goCtx context.Context, req *types.MsgRemoveParachainIBCTokenInfo) (*types.MsgRemoveParachainIBCTokenInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	removeTime, err := ms.AddParachainIBCInfoToRemoveList(ctx, req.NativeDenom)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventRemoveParachainIBCTokenInfo,
			sdk.NewAttribute(types.AttributeKeyNativeDenom, req.NativeDenom),
			sdk.NewAttribute(types.AttributeKeyRemoveTime, removeTime.String()),
		),
	})

	return &types.MsgRemoveParachainIBCTokenInfoResponse{}, nil
}

func (ms msgServer) AddRlyAddress(goCtx context.Context, req *types.MsgAddRlyAddress) (*types.MsgAddRlyAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.authority != req.Authority {
		return nil, errors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, req.Authority)
	}

	found := ms.HasAllowRlyAddress(ctx, req.RlyAddress)
	if found {
		return nil, fmt.Errorf("address %v already registry in allow list", req.RlyAddress)
	}

	ms.SetAllowRlyAddress(ctx, req.RlyAddress)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventAddRlyToAllowList,
			sdk.NewAttribute(types.AttributeKeyRlyAdress, req.RlyAddress),
		),
	})

	return &types.MsgAddRlyAddressResponse{}, nil
}
