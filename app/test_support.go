package app

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/keeper"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	wasm08keeper "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	customibctransferkeeper "github.com/notional-labs/composable/v6/custom/ibc-transfer/keeper"
	ratelimitmodulekeeper "github.com/notional-labs/composable/v6/x/ratelimit/keeper"
	transfermiddlewarekeeper "github.com/notional-labs/composable/v6/x/transfermiddleware/keeper"
)

func (app *ComposableApp) GetStakingKeeper() *stakingkeeper.Keeper {
	return &app.StakingKeeper.Keeper
}

func (app *ComposableApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

func (app *ComposableApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

func (app *ComposableApp) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

func (app *ComposableApp) GetBankKeeper() bankkeeper.Keeper {
	return app.BankKeeper
}

func (app *ComposableApp) GetAccountKeeper() authkeeper.AccountKeeper {
	return app.AccountKeeper
}

func (app *ComposableApp) GetWasmKeeper() wasmkeeper.Keeper {
	return app.WasmKeeper
}

func (app *ComposableApp) GetWasm08Keeper() wasm08keeper.Keeper {
	return app.Wasm08Keeper
}

func (app *ComposableApp) GetPfmKeeper() packetforwardkeeper.Keeper {
	return *app.PfmKeeper
}

func (app *ComposableApp) GetRateLimitKeeper() ratelimitmodulekeeper.Keeper {
	return app.RatelimitKeeper
}

// GetTransferKeeper implements the TestingApp interface.
func (app *ComposableApp) GetTransferKeeper() customibctransferkeeper.Keeper {
	return app.TransferKeeper
}

func (app *ComposableApp) GetTransferMiddlewareKeeper() transfermiddlewarekeeper.Keeper {
	return app.TransferMiddlewareKeeper
}

func (app *ComposableApp) GetGovKeeper() *govkeeper.Keeper {
	return &app.GovKeeper
}

// GetTxConfig implements the TestingApp interface.
func (app *ComposableApp) GetTxConfig() client.TxConfig {
	cfg := MakeEncodingConfig()
	return cfg.TxConfig
}
