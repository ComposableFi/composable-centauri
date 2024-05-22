package keeper

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/armon/go-metrics"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	router "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/keeper"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	custombankkeeper "github.com/notional-labs/composable/v6/custom/bank/keeper"
	ibctransfermiddlewarekeeper "github.com/notional-labs/composable/v6/x/ibctransfermiddleware/keeper"
)

var _ porttypes.Middleware = &IBCMiddleware{}

// IBCMiddleware implements the ICS26 callbacks for the forward middleware given the
// forward keeper and the underlying application.
type IBCMiddleware struct {
	router.IBCMiddleware

	app1    porttypes.IBCModule
	keeper1 *keeper.Keeper

	retriesOnTimeout1 uint8
	forwardTimeout1   time.Duration
	refundTimeout1    time.Duration
	ibcfeekeeper      *ibctransfermiddlewarekeeper.Keeper
	bank              *custombankkeeper.Keeper
}

func NewIBCMiddleware(
	app porttypes.IBCModule,
	k *keeper.Keeper,
	retriesOnTimeout uint8,
	forwardTimeout time.Duration,
	refundTimeout time.Duration,
	ibcfeekeeper *ibctransfermiddlewarekeeper.Keeper,
	bankkeeper *custombankkeeper.Keeper,
) IBCMiddleware {
	return IBCMiddleware{
		IBCMiddleware: router.NewIBCMiddleware(app, k, retriesOnTimeout, forwardTimeout, refundTimeout),
		ibcfeekeeper:  ibcfeekeeper,

		app1:              app,
		keeper1:           k,
		retriesOnTimeout1: retriesOnTimeout,
		forwardTimeout1:   forwardTimeout,
		refundTimeout1:    refundTimeout,
		bank:              bankkeeper,
	}
}

func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	logger := im.keeper1.Logger(ctx)

	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		logger.Debug(fmt.Sprintf("packetForwardMiddleware OnRecvPacket payload is not a FungibleTokenPacketData: %s", err.Error()))
		return im.IBCMiddleware.OnRecvPacket(ctx, packet, relayer)
	}

	logger.Debug("packetForwardMiddleware OnRecvPacket",
		"sequence", packet.Sequence,
		"src-channel", packet.SourceChannel, "src-port", packet.SourcePort,
		"dst-channel", packet.DestinationChannel, "dst-port", packet.DestinationPort,
		"amount", data.Amount, "denom", data.Denom, "memo", data.Memo,
	)

	d := make(map[string]interface{})
	err := json.Unmarshal([]byte(data.Memo), &d)
	if err != nil || d["forward"] == nil {
		// not a packet that should be forwarded
		logger.Debug("packetForwardMiddleware OnRecvPacket forward metadata does not exist")
		return im.app1.OnRecvPacket(ctx, packet, relayer)
	}
	m := &types.PacketMetadata{}
	err = json.Unmarshal([]byte(data.Memo), m)
	if err != nil {
		logger.Error("packetForwardMiddleware OnRecvPacket error parsing forward metadata", "error", err)
		return newErrorAcknowledgement(fmt.Errorf("error parsing forward metadata: %w", err))
	}

	metadata := m.Forward

	goCtx := ctx.Context()
	processed := getBoolFromAny(goCtx.Value(types.ProcessedKey{}))
	nonrefundable := getBoolFromAny(goCtx.Value(types.NonrefundableKey{}))
	disableDenomComposition := getBoolFromAny(goCtx.Value(types.DisableDenomCompositionKey{}))

	if err := metadata.Validate(); err != nil {
		logger.Error("packetForwardMiddleware OnRecvPacket forward metadata is invalid", "error", err)
		return newErrorAcknowledgement(err)
	}

	// override the receiver so that senders cannot move funds through arbitrary addresses.
	overrideReceiver, err := getReceiver(packet.DestinationChannel, data.Sender)
	if err != nil {
		logger.Error("packetForwardMiddleware OnRecvPacket failed to construct override receiver", "error", err)
		return newErrorAcknowledgement(fmt.Errorf("failed to construct override receiver: %w", err))
	}

	// if this packet has been handled by another middleware in the stack there may be no need to call into the
	// underlying app, otherwise the transfer module's OnRecvPacket callback could be invoked more than once
	// which would mint/burn vouchers more than once
	if !processed {
		if err := im.receiveFunds(ctx, packet, data, overrideReceiver, relayer); err != nil {
			logger.Error("packetForwardMiddleware OnRecvPacket error receiving packet", "error", err)
			return newErrorAcknowledgement(fmt.Errorf("error receiving packet: %w", err))
		}
	}

	// if this packet's token denom is already the base denom for some native token on this chain,
	// we do not need to do any further composition of the denom before forwarding the packet
	denomOnThisChain := data.Denom
	// Check if the packet was sent from Picasso
	paraChainIBCTokenInfo, found := im.keeper1.GetParachainTokenInfoByAssetID(ctx, data.Denom)
	if found && (paraChainIBCTokenInfo.ChannelID == packet.DestinationChannel) {
		disableDenomComposition = true
		denomOnThisChain = paraChainIBCTokenInfo.NativeDenom
	}

	if !disableDenomComposition {
		denomOnThisChain = getDenomForThisChain(
			packet.DestinationPort, packet.DestinationChannel,
			packet.SourcePort, packet.SourceChannel,
			data.Denom,
		)
	}

	amountInt, ok := sdk.NewIntFromString(data.Amount)
	if !ok {
		logger.Error("packetForwardMiddleware OnRecvPacket error parsing amount for forward", "amount", data.Amount)
		return newErrorAcknowledgement(fmt.Errorf("error parsing amount for forward: %s", data.Amount))
	}

	token := sdk.NewCoin(denomOnThisChain, amountInt)

	timeout := time.Duration(metadata.Timeout)

	if timeout.Nanoseconds() <= 0 {
		timeout = im.forwardTimeout1
	}

	var retries uint8
	if metadata.Retries != nil {
		retries = *metadata.Retries
	} else {
		retries = im.retriesOnTimeout1
	}

	// im.ibcfeekeeper.Transfer()

	feeAmount := sdk.NewDecFromInt(token.Amount).Mul(im.keeper1.GetFeePercentage(ctx)).RoundInt()
	packetAmount := token.Amount.Sub(feeAmount)
	packetCoin := sdk.NewCoin(token.Denom, packetAmount)

	memo := ""

	// set memo for next transfer with next from this transfer.
	if metadata.Next != nil {
		memoBz, err := json.Marshal(metadata.Next)
		if err != nil {
			im.keeper1.Logger(ctx).Error("packetForwardMiddleware error marshaling next as JSON",
				"error", err,
			)
			// return errorsmod.Wrapf(sdkerrors.ErrJSONMarshal, err.Error())
		}
		memo = string(memoBz)
	}

	tr := transfertypes.NewMsgTransfer(
		metadata.Port,
		metadata.Channel,
		packetCoin,
		overrideReceiver,
		metadata.Receiver,
		clienttypes.Height{
			RevisionNumber: 0,
			RevisionHeight: 0,
		},
		uint64(ctx.BlockTime().UnixNano())+uint64(timeout.Nanoseconds()),
		memo,
	)

	result, err := im.ibcfeekeeper.ChargeFee(ctx, tr)
	if err != nil {
		logger.Error("packetForwardMiddleware OnRecvPacket error charging fee", "error", err)
		return newErrorAcknowledgement(fmt.Errorf("error charging fee: %w", err))
	}
	if result != nil {
		if result.Fee.Amount.LT(token.Amount) {
			token = token.SubAmount(result.Fee.Amount)
		} else {
			send_err := im.bank.SendCoins(ctx, result.Sender, result.Receiver, sdk.NewCoins(result.Fee))
			if send_err != nil {
				logger.Error("packetForwardMiddleware OnRecvPacket error sending fee", "error", send_err)
				return newErrorAcknowledgement(fmt.Errorf("error charging fee: %w", send_err))
			}
			ack := channeltypes.NewResultAcknowledgement([]byte{byte(1)})
			return ack
		}
	}

	err = im.keeper1.ForwardTransferPacket(ctx, nil, packet, data.Sender, overrideReceiver, metadata, token, retries, timeout, []metrics.Label{}, nonrefundable)
	if err != nil {
		logger.Error("packetForwardMiddleware OnRecvPacket error forwarding packet", "error", err)
		return newErrorAcknowledgement(err)
	}

	// returning nil ack will prevent WriteAcknowledgement from occurring for forwarded packet.
	// This is intentional so that the acknowledgement will be written later based on the ack/timeout of the forwarded packet.
	return nil
}

func newErrorAcknowledgement(err error) channeltypes.Acknowledgement {
	return channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Error{
			Error: fmt.Sprintf("packet-forward-middleware error: %s", err.Error()),
		},
	}
}

func getBoolFromAny(value any) bool {
	if value == nil {
		return false
	}
	boolVal, ok := value.(bool)
	if !ok {
		return false
	}
	return boolVal
}

func getReceiver(channel string, originalSender string) (string, error) {
	senderStr := fmt.Sprintf("%s/%s", channel, originalSender)
	senderHash32 := address.Hash(types.ModuleName, []byte(senderStr))
	sender := sdk.AccAddress(senderHash32[:20])
	bech32Prefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
	return sdk.Bech32ifyAddressBytes(bech32Prefix, sender)
}

func (im IBCMiddleware) receiveFunds(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data transfertypes.FungibleTokenPacketData,
	overrideReceiver string,
	relayer sdk.AccAddress,
) error {
	overrideData := transfertypes.FungibleTokenPacketData{
		Denom:    data.Denom,
		Amount:   data.Amount,
		Sender:   data.Sender,
		Receiver: overrideReceiver, // override receiver
		// Memo explicitly zeroed
	}
	overrideDataBz := transfertypes.ModuleCdc.MustMarshalJSON(&overrideData)
	overridePacket := channeltypes.Packet{
		Sequence:           packet.Sequence,
		SourcePort:         packet.SourcePort,
		SourceChannel:      packet.SourceChannel,
		DestinationPort:    packet.DestinationPort,
		DestinationChannel: packet.DestinationChannel,
		Data:               overrideDataBz, // override data
		TimeoutHeight:      packet.TimeoutHeight,
		TimeoutTimestamp:   packet.TimeoutTimestamp,
	}

	ack := im.app1.OnRecvPacket(ctx, overridePacket, relayer)

	if ack == nil {
		return fmt.Errorf("ack is nil")
	}

	if !ack.Success() {
		return fmt.Errorf("ack error: %s", string(ack.Acknowledgement()))
	}

	return nil
}

func getDenomForThisChain(port, channel, counterpartyPort, counterpartyChannel, denom string) string {
	counterpartyPrefix := transfertypes.GetDenomPrefix(counterpartyPort, counterpartyChannel)
	if strings.HasPrefix(denom, counterpartyPrefix) {
		// unwind denom
		unwoundDenom := denom[len(counterpartyPrefix):]
		denomTrace := transfertypes.ParseDenomTrace(unwoundDenom)
		if denomTrace.Path == "" {
			// denom is now unwound back to native denom
			return unwoundDenom
		}
		// denom is still IBC denom
		return denomTrace.IBCDenom()
	}
	// append port and channel from this chain to denom
	prefixedDenom := transfertypes.GetDenomPrefix(port, channel) + denom
	return transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()
}
