package keeper

import (
	"context"
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	custombankkeeper "github.com/notional-labs/composable/v6/custom/bank/keeper"
	ibctransfermiddleware "github.com/notional-labs/composable/v6/x/ibctransfermiddleware/keeper"
	ibctransfermiddlewaretypes "github.com/notional-labs/composable/v6/x/ibctransfermiddleware/types"
)

type Keeper struct {
	ibctransferkeeper.Keeper
	cdc                   codec.BinaryCodec
	IbcTransfermiddleware *ibctransfermiddleware.Keeper
	bank                  *custombankkeeper.Keeper
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
	bankKeeper *custombankkeeper.Keeper,
) Keeper {
	keeper := Keeper{
		Keeper:                ibctransferkeeper.NewKeeper(cdc, key, paramSpace, ics4Wrapper, channelKeeper, portKeeper, authKeeper, bk, scopedKeeper),
		IbcTransfermiddleware: ibcTransfermiddleware,
		cdc:                   cdc,
		bank:                  bankKeeper,
	}
	return keeper
}

// Transfer is the server API around the Transfer method of the IBC transfer module.
// It checks if the sender is allowed to transfer the token and if the channel has fees.
// If the channel has fees, it will charge the sender and send the fees to the fee address.
// If the sender is not allowed to transfer the token because this tokens does not exists in the allowed tokens list, it just return without doing anything.
// If the sender is allowed to transfer the token, it will call the original transfer method.
// If the transfer amount is less than the minimum fee, it will charge the full transfer amount.
// If the transfer amount is greater than the minimum fee, it will charge the minimum fee and the percentage fee.
func (k Keeper) Transfer(goCtx context.Context, msg *types.MsgTransfer) (*types.MsgTransferResponse, error) {
	return k.Keeper.Transfer(goCtx, msg)
}

func GetPriority(jsonString string) *string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonString), &data); err != nil {
		return nil
	}

	priority, ok := data["priority"].(string)
	if !ok {
		return nil
	}

	return &priority
}

func findPriority(priorities []*ibctransfermiddlewaretypes.TxPriorityFee, priority string) *ibctransfermiddlewaretypes.TxPriorityFee {
	for _, p := range priorities {
		if p.Priority == priority {
			return p
		}
	}
	return nil
}
