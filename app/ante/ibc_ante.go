package ante

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/authz"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	tfmwKeeper "github.com/notional-labs/composable/v6/x/transfermiddleware/keeper"
)

type IBCPermissionDecorator struct {
	cdc        codec.BinaryCodec
	tfmwKeeper tfmwKeeper.Keeper
}

func NewIBCPermissionDecorator(cdc codec.BinaryCodec, keeper tfmwKeeper.Keeper) IBCPermissionDecorator {
	return IBCPermissionDecorator{
		cdc:        cdc,
		tfmwKeeper: keeper,
	}
}

func (g IBCPermissionDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx,
	simulate bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// run checks only on CheckTx or simulate
	if simulate {
		return next(ctx, tx, simulate)
	}

	msgs := tx.GetMsgs()
	if err = g.ValidateIBCUpdateClientMsg(ctx, msgs); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

// ValidateIBCUpdateClientMsg validate
func (g IBCPermissionDecorator) ValidateIBCUpdateClientMsg(ctx sdk.Context, msgs []sdk.Msg) error {
	for _, m := range msgs {
		if msg, ok := m.(*authz.MsgExec); ok {
			if err := g.validAuthz(ctx, msg); err != nil {
				return err
			}
			continue
		}

		// validate normal msgs
		if err := g.validMsg(ctx, m); err != nil {
			return err
		}
	}
	return nil
}

func (g IBCPermissionDecorator) validMsg(ctx sdk.Context, m sdk.Msg) error {
	if msg, ok := m.(*clienttypes.MsgUpdateClient); ok {
		if msg.ClientMessage.TypeUrl == "/ibc.lightclients.wasm.v1.Header" && !g.tfmwKeeper.HasAllowRlyAddress(ctx, msg.Signer) {
			return fmt.Errorf("permission denied, address %s don't have relay permission", msg.Signer)
		}
	}

	return nil
}

func (g IBCPermissionDecorator) validAuthz(ctx sdk.Context, execMsg *authz.MsgExec) error {
	for _, v := range execMsg.Msgs {
		var innerMsg sdk.Msg
		if err := g.cdc.UnpackAny(v, &innerMsg); err != nil {
			return errorsmod.Wrap(err, "cannot unmarshal authz exec msgs")
		}
		if err := g.validMsg(ctx, innerMsg); err != nil {
			return err
		}
	}

	return nil
}
