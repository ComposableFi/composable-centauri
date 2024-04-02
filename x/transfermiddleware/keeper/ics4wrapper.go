package keeper

import (
	"fmt"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/ComposableFi/composable-cosmos/v6/x/transfermiddleware/types"
)

func (keeper Keeper) hasParachainIBCTokenInfo(ctx sdk.Context, nativeDenom string) bool {
	store := ctx.KVStore(keeper.storeKey)
	return store.Has(types.GetKeyParachainIBCTokenInfoByNativeDenom(nativeDenom))
}

func (keeper Keeper) handleOverrideSendPacketTransferLogic(
	ctx sdk.Context,
	_ *capabilitytypes.Capability,
	sourcePort, sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	var fungibleTokenPacketData transfertypes.FungibleTokenPacketData

	err = transfertypes.ModuleCdc.UnmarshalJSON(data, &fungibleTokenPacketData)
	if err != nil {
		return 0, err
	}
	sender, err := sdk.AccAddressFromBech32(fungibleTokenPacketData.Sender)
	if err != nil {
		return 0, err
	}

	// check if denom in fungibleTokenPacketData is native denom in parachain info and
	parachainInfo := keeper.GetParachainIBCTokenInfoByNativeDenom(ctx, fungibleTokenPacketData.Denom)

	// burn native token in escrow address
	transferAmount, ok := sdk.NewIntFromString(fungibleTokenPacketData.Amount)

	// TODO: remove this panic and replace by err hanlde
	if !ok {
		panic("cannot parse string amount")
	}
	nativeTransferToken := sdk.NewCoin(fungibleTokenPacketData.Denom, transferAmount)
	ibcTransferToken := sdk.NewCoin(parachainInfo.IbcDenom, transferAmount)

	escrowAddress := transfertypes.GetEscrowAddress(sourcePort, sourceChannel)
	err = keeper.bankKeeper.SendCoinsFromAccountToModule(ctx, escrowAddress, transfertypes.ModuleName, sdk.NewCoins(nativeTransferToken))
	if err != nil {
		return 0, err
	}
	// burn native token
	// Get Coin from excrow address
	keeper.bankKeeper.BurnCoins(ctx, transfertypes.ModuleName, sdk.NewCoins(nativeTransferToken))

	// release lock IBC token and send it to sender
	// TODO: should we use a module address for this ?
	err = keeper.bankKeeper.SendCoins(ctx, escrowAddress, sender, sdk.NewCoins(ibcTransferToken))
	if err != nil {
		return 0, err
	}

	// new msg transfer from transfer to parachain
	transferMsg := transfertypes.MsgTransfer{
		SourcePort:       transfertypes.PortID,
		SourceChannel:    sourceChannel,
		Token:            ibcTransferToken,
		Sender:           fungibleTokenPacketData.Sender,
		Receiver:         fungibleTokenPacketData.Receiver,
		TimeoutHeight:    timeoutHeight,
		TimeoutTimestamp: timeoutTimestamp,
		Memo:             fungibleTokenPacketData.Memo,
	}
	res, err := keeper.executeTransferMsg(ctx, &transferMsg)
	if err != nil {
		return 0, err
	}

	return res.Sequence, nil
}

func (keeper Keeper) executeTransferMsg(ctx sdk.Context, transferMsg *transfertypes.MsgTransfer) (*transfertypes.MsgTransferResponse, error) {
	if err := transferMsg.ValidateBasic(); err != nil {
		return nil, fmt.Errorf("bad msg %v", err.Error())
	}
	return keeper.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), transferMsg)
}

// SendPacket wraps IBC ChannelKeeper's SendPacket function
func (keeper Keeper) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort, sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (sequence uint64, err error) {
	var fungibleTokenPacketData transfertypes.FungibleTokenPacketData

	err = transfertypes.ModuleCdc.UnmarshalJSON(data, &fungibleTokenPacketData)
	if err != nil {
		return 0, err
	}

	// check if denom in fungibleTokenPacketData is native denom in parachain info and
	parachainInfo := keeper.GetParachainIBCTokenInfoByNativeDenom(ctx, fungibleTokenPacketData.Denom)

	if parachainInfo.ChannelID != sourceChannel || parachainInfo.NativeDenom != fungibleTokenPacketData.Denom {
		return keeper.ICS4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
	}

	return keeper.handleOverrideSendPacketTransferLogic(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

// WriteAcknowledgement wraps IBC ICS4Wrapper WriteAcknowledgement function.
// ICS29 WriteAcknowledgement is used for asynchronous acknowledgements.
func (keeper *Keeper) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, acknowledgement ibcexported.Acknowledgement) error {
	return keeper.ICS4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, acknowledgement)
}

// WriteAcknowledgement wraps IBC ICS4Wrapper GetAppVersion function.
func (keeper *Keeper) GetAppVersion(
	ctx sdk.Context,
	portID,
	channelID string,
) (string, bool) {
	return keeper.ICS4Wrapper.GetAppVersion(ctx, portID, channelID)
}

func (keeper *Keeper) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	_ sdk.AccAddress,
) error {
	var ack channeltypes.Acknowledgement
	if err := transfertypes.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return errors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}
	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		return errors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	switch ack.Response.(type) {
	case *channeltypes.Acknowledgement_Error:
		return keeper.refundToken(ctx, packet, data)
	default:
		// the acknowledgement succeeded on the receiving chain so nothing
		// needs to be executed and no error needs to be returned
		return nil
	}
}

func (keeper Keeper) refundToken(ctx sdk.Context, packet channeltypes.Packet, data transfertypes.FungibleTokenPacketData) error {
	// parse the denomination from the full denom path
	trace := transfertypes.ParseDenomTrace(data.Denom)
	// parse the transfer amount
	transferAmount, ok := sdk.NewIntFromString(data.Amount)
	if !ok {
		return errors.Wrapf(transfertypes.ErrInvalidAmount, "unable to parse transfer amount (%s) into math.Int", data.Amount)
	}
	token := sdk.NewCoin(trace.IBCDenom(), transferAmount)

	// decode the sender address
	sender, err := sdk.AccAddressFromBech32(data.Sender)
	if err != nil {
		return err
	}
	if transfertypes.SenderChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), data.Denom) {
		// Do nothing
		// This case should never happened
		return nil
	}
	nativeDenom := keeper.GetNativeDenomByIBCDenomSecondaryIndex(ctx, trace.IBCDenom())
	paraTokenInfo := keeper.GetParachainIBCTokenInfoByNativeDenom(ctx, nativeDenom)

	// only trigger if source channel is from parachain.
	if !keeper.hasParachainIBCTokenInfo(ctx, nativeDenom) {
		return nil
	}

	if packet.GetSourceChannel() == paraTokenInfo.ChannelID {
		nativeToken := sdk.NewCoin(paraTokenInfo.NativeDenom, transferAmount)
		// send IBC token to escrow address ibc token
		escrowAddress := transfertypes.GetEscrowAddress(transfertypes.PortID, paraTokenInfo.ChannelID)
		if err := keeper.bankKeeper.SendCoins(ctx, sender, escrowAddress, sdk.NewCoins(token)); err != nil {
			panic(fmt.Sprintf("unable to send coins from account to module despite previously minting coins to module account: %v", err))
		}

		// mint native token and send back to sender
		if err := keeper.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(nativeToken)); err != nil {
			return err
		}

		if err := keeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, sdk.NewCoins(nativeToken)); err != nil {
			panic(fmt.Sprintf("unable to send coins from module to account despite previously minting coins to module account: %v", err))
		}
	}

	return nil
}
