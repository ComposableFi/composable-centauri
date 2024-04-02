package ante

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	txBoundaryKeeper "github.com/ComposableFi/composable-cosmos/v6/x/tx-boundary/keeper"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type StakingPermissionDecorator struct {
	cdc        codec.BinaryCodec
	txBoundary txBoundaryKeeper.Keeper
}

func NewStakingPermissionDecorator(cdc codec.BinaryCodec, keeper txBoundaryKeeper.Keeper) StakingPermissionDecorator {
	return StakingPermissionDecorator{
		cdc:        cdc,
		txBoundary: keeper,
	}
}

func (g StakingPermissionDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx,
	simulate bool, next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// run checks only on CheckTx or simulate
	if simulate {
		return next(ctx, tx, simulate)
	}

	msgs := tx.GetMsgs()
	if err = g.ValidateStakingMsgs(ctx, msgs); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

// ValidateStakingMsg validate
func (g StakingPermissionDecorator) ValidateStakingMsgs(ctx sdk.Context, msgs []sdk.Msg) error {
	for _, m := range msgs {
		err := g.ValidateStakingMsg(ctx, m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g StakingPermissionDecorator) ValidateStakingMsg(ctx sdk.Context, msg sdk.Msg) error {
	switch msg := msg.(type) {

	case *stakingtypes.MsgDelegate:
		if err := g.validDelegateMsg(ctx, msg); err != nil {
			return err
		}
	case *stakingtypes.MsgBeginRedelegate:
		if err := g.validRedelegateMsg(ctx, msg); err != nil {
			return err
		}
	case *authz.MsgExec:
		if err := g.validAuthz(ctx, msg); err != nil {
			return err
		}
	default:
		return nil
	}
	return nil
}

func (g StakingPermissionDecorator) validDelegateMsg(ctx sdk.Context, msg *stakingtypes.MsgDelegate) error {
	boundary := g.txBoundary.GetDelegateBoundary(ctx)
	delegator, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return err
	}
	g.txBoundary.UpdateLimitPerAddr(ctx, delegator)

	if boundary.TxLimit == 0 {
		return nil
	} else if g.txBoundary.GetLimitPerAddr(ctx, delegator).DelegateCount >= boundary.TxLimit {
		return fmt.Errorf("delegate tx denied, excess tx limit")
	}
	g.txBoundary.IncrementDelegateCount(ctx, delegator)
	return nil
}

func (g StakingPermissionDecorator) validRedelegateMsg(ctx sdk.Context, msg *stakingtypes.MsgBeginRedelegate) error {
	boundary := g.txBoundary.GetRedelegateBoundary(ctx)
	delegator, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return err
	}

	g.txBoundary.UpdateLimitPerAddr(ctx, delegator)
	if boundary.TxLimit == 0 {
		return nil
	} else if g.txBoundary.GetLimitPerAddr(ctx, delegator).ReledegateCount >= boundary.TxLimit {
		return fmt.Errorf("redelegate tx denied, excess tx limit")
	}
	g.txBoundary.IncrementRedelegateCount(ctx, delegator)
	return nil
}

func (g StakingPermissionDecorator) validAuthz(ctx sdk.Context, execMsg *authz.MsgExec) error {
	for _, v := range execMsg.Msgs {
		var innerMsg sdk.Msg
		if err := g.cdc.UnpackAny(v, &innerMsg); err != nil {
			return errorsmod.Wrap(err, "cannot unmarshal authz exec msgs")
		}
		if err := g.ValidateStakingMsg(ctx, innerMsg); err != nil {
			return err
		}
	}
	return nil
}
