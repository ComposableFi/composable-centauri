package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type StakingKeeper interface {
	BondDenom(ctx context.Context) (res string, err error)
}

type TransferMiddlewareKeeper interface {
	GetTotalEscrowedToken(ctx sdk.Context) (coins sdk.Coins)
}
