package packetforward

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	router "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/keeper"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
)

var _ porttypes.Middleware = &IBCMiddleware{}

// IBCMiddleware implements the ICS26 callbacks for the forward middleware given the
// forward keeper and the underlying application.
type IBCMiddleware struct {
	router.IBCMiddleware
}

func NewIBCMiddleware(
	app porttypes.IBCModule,
	k *keeper.Keeper,
	retriesOnTimeout uint8,
	forwardTimeout time.Duration,
	refundTimeout time.Duration,
) IBCMiddleware {
	return IBCMiddleware{
		IBCMiddleware: router.NewIBCMiddleware(app, k, retriesOnTimeout, forwardTimeout, refundTimeout),
	}
}

func (im IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	return im.IBCMiddleware.OnRecvPacket(ctx, packet, relayer)
}
