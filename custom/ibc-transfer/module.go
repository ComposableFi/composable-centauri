package bank

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"

	ibctransfermodule "github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	custombankkeeper "github.com/notional-labs/composable/v6/custom/bank/keeper"
	customibctransferkeeper "github.com/notional-labs/composable/v6/custom/ibc-transfer/keeper"
)

// AppModule wraps around the bank module and the bank keeper to return the right total supply
type AppModule struct {
	ibctransfermodule.AppModule
	keeper customibctransferkeeper.Keeper
	bank   custombankkeeper.Keeper
	// subspace  exported.Subspace
	msgServer types.MsgServer
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper customibctransferkeeper.Keeper, bank custombankkeeper.Keeper) AppModule {
	ibctransferModule := ibctransfermodule.NewAppModule(keeper.Keeper)
	return AppModule{
		AppModule: ibctransferModule,
		keeper:    keeper,
		bank:      bank,
		// subspace:  ss,
		msgServer: keeper.Keeper,
	}
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	msgServer := customibctransferkeeper.NewMsgServerImpl(am.keeper, am.bank)
	types.RegisterMsgServer(cfg.MsgServer(), msgServer)
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper.Keeper)

	m := ibctransferkeeper.NewMigrator(am.keeper.Keeper)
	if err := cfg.RegisterMigration(types.ModuleName, 1, m.MigrateTraces); err != nil {
		panic(fmt.Sprintf("failed to migrate transfer app from version 1 to 2: %v", err))
	}

	if err := cfg.RegisterMigration(types.ModuleName, 2, m.MigrateTotalEscrowForDenom); err != nil {
		panic(fmt.Sprintf("failed to migrate transfer app from version 2 to 3: %v", err))
	}

	if err := cfg.RegisterMigration(types.ModuleName, 3, m.MigrateParams); err != nil {
		panic(fmt.Errorf("failed to migrate transfer app version 3 to 4 (self-managed params migration): %v", err))
	}

	if err := cfg.RegisterMigration(types.ModuleName, 4, m.MigrateDenomMetadata); err != nil {
		panic(fmt.Errorf("failed to migrate transfer app from version 4 to 5 (set denom metadata migration): %v", err))
	}
}
