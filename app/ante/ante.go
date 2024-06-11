package ante

import (
	circuitante "cosmossdk.io/x/circuit/ante"
	circuitkeeper "cosmossdk.io/x/circuit/keeper"
	ibcante "github.com/cosmos/ibc-go/v8/modules/core/ante"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	"cosmossdk.io/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	tfmwKeeper "github.com/notional-labs/composable/v6/x/transfermiddleware/keeper"
	txBoundaryAnte "github.com/notional-labs/composable/v6/x/tx-boundary/ante"
	txBoundaryKeeper "github.com/notional-labs/composable/v6/x/tx-boundary/keeper"
)

// Link to default ante handler used by cosmos sdk:
// https://github.com/cosmos/cosmos-sdk/blob/v0.43.0/x/auth/ante/ante.go#L41
func NewAnteHandler(
	options servertypes.AppOptions,
	ak ante.AccountKeeper,
	bk authtypes.BankKeeper,
	feegrantKeeper ante.FeegrantKeeper,
	txFeeChecker ante.TxFeeChecker,
	sigGasConsumer ante.SignatureVerificationGasConsumer,
	signModeHandler *signing.HandlerMap,
	channelKeeper *ibckeeper.Keeper,
	tfmwKeeper tfmwKeeper.Keeper,
	txBoundaryKeeper txBoundaryKeeper.Keeper,
	ck *circuitkeeper.Keeper,
	codec codec.BinaryCodec,
) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		ante.NewSetUpContextDecorator(), //  // outermost AnteDecorator. SetUpContext must be called first
		ante.NewValidateBasicDecorator(),
		ante.NewConsumeGasForTxSizeDecorator(ak),
		ante.NewDeductFeeDecorator(ak, bk, feegrantKeeper, txFeeChecker),
		circuitante.NewCircuitBreakerDecorator(ck),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(ak),
		ante.NewConsumeGasForTxSizeDecorator(ak),
		NewIBCPermissionDecorator(codec, tfmwKeeper),
		txBoundaryAnte.NewStakingPermissionDecorator(codec, txBoundaryKeeper),
		ante.NewSetPubKeyDecorator(ak), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(ak),
		ante.NewSigGasConsumeDecorator(ak, sigGasConsumer),
		ante.NewSigVerificationDecorator(ak, signModeHandler),
		ante.NewIncrementSequenceDecorator(ak),
		ibcante.NewRedundantRelayDecorator(channelKeeper),
	)
}
