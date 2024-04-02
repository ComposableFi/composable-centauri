package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"

	ibctransfermiddleware "github.com/ComposableFi/composable-cosmos/v6/x/ibctransfermiddleware/keeper"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
)

type Keeper struct {
	ibctransferkeeper.Keeper
	cdc                   codec.BinaryCodec
	IbcTransfermiddleware *ibctransfermiddleware.Keeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	ics4Wrapper porttypes.ICS4Wrapper,
	channelKeeper types.ChannelKeeper,
	portKeeper types.PortKeeper,
	authKeeper types.AccountKeeper,
	bk types.BankKeeper,
	scopedKeeper exported.ScopedKeeper,
	ibcTransfermiddleware *ibctransfermiddleware.Keeper,
) Keeper {
	keeper := Keeper{
		Keeper:                ibctransferkeeper.NewKeeper(cdc, key, paramSpace, ics4Wrapper, channelKeeper, portKeeper, authKeeper, bk, scopedKeeper),
		IbcTransfermiddleware: ibcTransfermiddleware,
		cdc:                   cdc,
	}
	return keeper
}
