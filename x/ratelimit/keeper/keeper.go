package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"

	"github.com/ComposableFi/composable-cosmos/v6/x/ratelimit/types"
	tfmwkeeper "github.com/ComposableFi/composable-cosmos/v6/x/transfermiddleware/keeper"
)

type Keeper struct {
	storeKey   storetypes.StoreKey
	cdc        codec.BinaryCodec
	paramstore paramtypes.Subspace

	bankKeeper    types.BankKeeper
	channelKeeper types.ChannelKeeper
	ics4Wrapper   porttypes.ICS4Wrapper
	tfmwKeeper    tfmwkeeper.Keeper

	// the address capable of executing a AddParachainIBCTokenInfo and RemoveParachainIBCTokenInfo message. Typically, this
	// should be the x/gov module account.
	authority string
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	ps paramtypes.Subspace,
	bankKeeper types.BankKeeper,
	channelKeeper types.ChannelKeeper,
	ics4Wrapper porttypes.ICS4Wrapper,
	tfmwKeeper tfmwkeeper.Keeper,
	authority string,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:           cdc,
		storeKey:      key,
		paramstore:    ps,
		bankKeeper:    bankKeeper,
		channelKeeper: channelKeeper,
		ics4Wrapper:   ics4Wrapper,
		tfmwKeeper:    tfmwKeeper,
		authority:     authority,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
