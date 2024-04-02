//nolint:all
package v5_2_0

import (
	"github.com/ComposableFi/composable-cosmos/v6/app/keepers"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	wasm08types "github.com/cosmos/ibc-go/v7/modules/light-clients/08-wasm/types"
)

const (
	newWasmCodeID      = "ad84ee3292e28b4e46da16974c118d40093e1a6e28a083f2f045f68fde7fb575"
	subjectClientId    = "08-wasm-5"
	substituteClientId = "08-wasm-133"
)

func RunForkLogic(ctx sdk.Context, keepers *keepers.AppKeepers) {
	ctx.Logger().Info("Applying v5_2_0 upgrade" +
		"Upgrade 08-wasm contract",
	)

	UpdateWasmContract(ctx, keepers.IBCKeeper)

	err := ClientUpdate(ctx, keepers.IBCKeeper.Codec(), keepers.IBCKeeper, subjectClientId, substituteClientId)
	if err != nil {
		panic(err)
	}
}

func UpdateWasmContract(ctx sdk.Context, ibckeeper *ibckeeper.Keeper) {
	unknownClientState, found := ibckeeper.ClientKeeper.GetClientState(ctx, subjectClientId)
	if !found {
		panic("substitute client client not found ")
	}

	clientState, ok := unknownClientState.(*wasm08types.ClientState)
	if !ok {
		panic("cannot update client")
	}

	code, err := transfertypes.ParseHexHash(newWasmCodeID)
	if err != nil {
		panic(err)
	}

	clientState.CodeId = code

	ibckeeper.ClientKeeper.SetClientState(ctx, subjectClientId, clientState)
}

func ClientUpdate(ctx sdk.Context, codec codec.BinaryCodec, ibckeeper *ibckeeper.Keeper, subjectClientId string, substituteClientId string) error {
	subjectClientState, found := ibckeeper.ClientKeeper.GetClientState(ctx, subjectClientId)
	if !found {
		return sdkerrors.Wrapf(clienttypes.ErrClientNotFound, "subject client with ID %s", subjectClientId)
	}

	subjectClientStore := ibckeeper.ClientKeeper.ClientStore(ctx, subjectClientId)

	substituteClientState, found := ibckeeper.ClientKeeper.GetClientState(ctx, substituteClientId)
	if !found {
		return sdkerrors.Wrapf(clienttypes.ErrClientNotFound, "substitute client with ID %s", substituteClientId)
	}

	substituteClientStore := ibckeeper.ClientKeeper.ClientStore(ctx, substituteClientId)

	if status := ibckeeper.ClientKeeper.GetClientStatus(ctx, substituteClientState, substituteClientId); status != exported.Active {
		return sdkerrors.Wrapf(clienttypes.ErrClientNotActive, "substitute client is not Active, status is %s", status)
	}

	if err := subjectClientState.CheckSubstituteAndUpdateState(ctx, codec, subjectClientStore, substituteClientStore, substituteClientState); err != nil {
		return err
	}

	ctx.Logger().Info("client updated after hark fork passed", "client-id", subjectClientId)

	return nil
}
