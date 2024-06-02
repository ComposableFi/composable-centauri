package app

import (
	"testing"

	"github.com/CosmWasm/wasmd/x/wasm"
	ratelimitkeeper "github.com/notional-labs/composable/v6/x/ratelimit/keeper"
	tfmdkeeper "github.com/notional-labs/composable/v6/x/transfermiddleware/keeper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	wasm08 "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/keeper"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
)

type TestSupport struct {
	tb  testing.TB
	app *ComposableApp
}

func NewTestSupport(tb testing.TB, app *ComposableApp) *TestSupport {
	tb.Helper()
	return &TestSupport{tb: tb, app: app}
}

func (s TestSupport) IBCKeeper() *ibckeeper.Keeper {
	return s.app.IBCKeeper
}

func (s TestSupport) AppCodec() codec.Codec {
	return s.app.appCodec
}

func (s TestSupport) ScopeIBCKeeper() capabilitykeeper.ScopedKeeper {
	return s.app.ScopedIBCKeeper
}

func (s TestSupport) ScopedTransferKeeper() capabilitykeeper.ScopedKeeper {
	return s.app.ScopedTransferKeeper
}

func (s TestSupport) StakingKeeper() *stakingkeeper.Keeper {
	return s.app.StakingKeeper
}

func (s TestSupport) AccountKeeper() authkeeper.AccountKeeper {
	return s.app.AccountKeeper
}

func (s TestSupport) BankKeeper() bankkeeper.Keeper {
	return s.app.BankKeeper
}

func (s TestSupport) GovKeeper() govkeeper.Keeper {
	return s.app.GovKeeper
}

func (s TestSupport) TransferKeeper() ibctransferkeeper.Keeper {
	return s.app.TransferKeeper
}

func (s TestSupport) Wasm08Keeper() wasm08.Keeper {
	return s.app.Wasm08Keeper
}

func (s TestSupport) WasmdKeeper() wasm.Keeper {
	return s.app.WasmKeeper
}

func (s TestSupport) GetBaseApp() *baseapp.BaseApp {
	return s.app.BaseApp
}

func (s TestSupport) GetTxConfig() client.TxConfig {
	return s.app.GetTxConfig()
}

func (s TestSupport) TransferMiddleware() tfmdkeeper.Keeper {
	return s.app.TransferMiddlewareKeeper
}

func (s TestSupport) RateLimit() ratelimitkeeper.Keeper {
	return s.app.RatelimitKeeper
}
