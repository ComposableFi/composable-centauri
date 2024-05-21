package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/notional-labs/composable/v6/app/upgrades/v6_6_4"

	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	tendermint "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	wasm08 "github.com/cosmos/ibc-go/v7/modules/light-clients/08-wasm"
	wasm08keeper "github.com/cosmos/ibc-go/v7/modules/light-clients/08-wasm/keeper"

	wasm08types "github.com/cosmos/ibc-go/v7/modules/light-clients/08-wasm/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank"

	"github.com/notional-labs/composable/v6/app/keepers"

	// bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/cosmos/cosmos-sdk/x/group"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	tmos "github.com/cometbft/cometbft/libs/os"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcclientclient "github.com/cosmos/ibc-go/v7/modules/core/02-client/client"
	ibchost "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	custompfm "github.com/notional-labs/composable/v6/custom/custompfm/keeper"
	customibctransfer "github.com/notional-labs/composable/v6/custom/ibc-transfer"
	customstaking "github.com/notional-labs/composable/v6/custom/staking"
	"github.com/spf13/cast"
	icq "github.com/strangelove-ventures/async-icq/v7"
	icqtypes "github.com/strangelove-ventures/async-icq/v7/types"

	router "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward"
	routertypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	alliancemodule "github.com/terra-money/alliance/x/alliance"
	alliancemoduleclient "github.com/terra-money/alliance/x/alliance/client"
	alliancemoduletypes "github.com/terra-money/alliance/x/alliance/types"

	custombankmodule "github.com/notional-labs/composable/v6/custom/bank"

	"github.com/notional-labs/composable/v6/app/ante"
	"github.com/notional-labs/composable/v6/x/ibctransfermiddleware"
	"github.com/notional-labs/composable/v6/x/stakingmiddleware"
	transfermiddleware "github.com/notional-labs/composable/v6/x/transfermiddleware"
	transfermiddlewaretypes "github.com/notional-labs/composable/v6/x/transfermiddleware/types"

	txBoundary "github.com/notional-labs/composable/v6/x/tx-boundary"
	txBoundaryTypes "github.com/notional-labs/composable/v6/x/tx-boundary/types"

	ratelimitmodule "github.com/notional-labs/composable/v6/x/ratelimit"
	ratelimitmoduletypes "github.com/notional-labs/composable/v6/x/ratelimit/types"

	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"

	"github.com/notional-labs/composable/v6/x/mint"
	minttypes "github.com/notional-labs/composable/v6/x/mint/types"

	ibctestingtypes "github.com/cosmos/ibc-go/v7/testing/types"

	ibc_hooks "github.com/notional-labs/composable/v6/x/ibc-hooks"
	ibchookstypes "github.com/notional-labs/composable/v6/x/ibc-hooks/types"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	upgrades "github.com/notional-labs/composable/v6/app/upgrades"
	ibctransfermiddlewaretypes "github.com/notional-labs/composable/v6/x/ibctransfermiddleware/types"
	stakingmiddlewaretypes "github.com/notional-labs/composable/v6/x/stakingmiddleware/types"
)

const (
	Name       = "pica"
	dirName    = "banksy"
	ForkHeight = 244008
)

var (
	// If EnabledSpecificProposals is "", and this is "true", then enable all x/wasm proposals.
	// If EnabledSpecificProposals is "", and this is not "true", then disable all x/wasm proposals.
	ProposalsEnabled = "false"
	// If set to non-empty string it must be comma-separated list of values that are all a subset
	// of "EnableAllProposals" (takes precedence over ProposalsEnabled)
	// https://github.com/CosmWasm/wasmd/blob/02a54d33ff2c064f3539ae12d75d027d9c665f05/x/wasm/internal/types/proposal.go#L28-L34
	EnableSpecificProposals = ""

	Upgrades = []upgrades.Upgrade{v6_6_4.Upgrade}
	Forks    = []upgrades.Fork{}
)

// GetEnabledProposals parses the ProposalsEnabled / EnableSpecificProposals values to
// produce a list of enabled proposals to pass into wasmd app.
func GetEnabledProposals() []wasm.ProposalType {
	if EnableSpecificProposals == "" {
		if ProposalsEnabled == "true" {
			return wasm.EnableAllProposals
		}
		return wasm.DisableAllProposals
	}
	chunks := strings.Split(EnableSpecificProposals, ",")
	proposals, err := wasm.ConvertToProposals(chunks)
	if err != nil {
		panic(err)
	}
	return proposals
}

// this line is used by starport scaffolding # stargate/wasm/app/enabledProposals

func getGovProposalHandlers() []govclient.ProposalHandler {
	var govProposalHandlers []govclient.ProposalHandler
	// this line is used by starport scaffolding # stargate/app/govProposalHandlers

	govProposalHandlers = append(govProposalHandlers,
		paramsclient.ProposalHandler,
		upgradeclient.LegacyProposalHandler,
		upgradeclient.LegacyCancelProposalHandler,
		ibcclientclient.UpdateClientProposalHandler,
		ibcclientclient.UpgradeProposalHandler,
		alliancemoduleclient.CreateAllianceProposalHandler,
		alliancemoduleclient.UpdateAllianceProposalHandler,
		alliancemoduleclient.DeleteAllianceProposalHandler,
		// this line is used by starport scaffolding # stargate/app/govProposalHandler
	)

	return govProposalHandlers
}

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(getGovProposalHandlers()),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		groupmodule.AppModuleBasic{},
		ibc.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		transfer.AppModuleBasic{},
		icq.AppModuleBasic{},
		vesting.AppModuleBasic{},
		tendermint.AppModuleBasic{},
		mint.AppModuleBasic{},
		wasm08.AppModuleBasic{},
		wasm.AppModuleBasic{},
		router.AppModuleBasic{},
		ica.AppModuleBasic{},
		ibc_hooks.AppModuleBasic{},
		transfermiddleware.AppModuleBasic{},
		txBoundary.AppModuleBasic{},
		ratelimitmodule.AppModuleBasic{},
		consensus.AppModuleBasic{},
		alliancemodule.AppModuleBasic{},
		stakingmiddleware.AppModuleBasic{},
		ibctransfermiddleware.AppModuleBasic{},
		// this line is used by starport scaffolding # stargate/app/moduleBasic
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName: nil,
		distrtypes.ModuleName:      nil,
		// mint module needs burn access to remove excess validator tokens (it overallocates, then burns)
		minttypes.ModuleName:                {authtypes.Minter},
		stakingtypes.BondedPoolName:         {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:      {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:                 {authtypes.Burner},
		transfermiddlewaretypes.ModuleName:  {authtypes.Minter, authtypes.Burner},
		ibctransfertypes.ModuleName:         {authtypes.Minter, authtypes.Burner},
		alliancemoduletypes.ModuleName:      {authtypes.Minter, authtypes.Burner},
		alliancemoduletypes.RewardsPoolName: nil,
		icatypes.ModuleName:                 nil,
		// this line is used by starport scaffolding # stargate/app/maccPerms
	}
)

var _ servertypes.Application = (*ComposableApp)(nil)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+dirName)
	// manually update the power reduction by replacing micro (u) -> pico (p)pica
	sdk.DefaultPowerReduction = PowerReduction
}

// ComposableApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type ComposableApp struct {
	*baseapp.BaseApp
	keepers.AppKeepers

	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	mm           *module.Manager
	sm           *module.SimulationManager
	configurator module.Configurator
}

// RUN GOSEC
// New returns a reference to an initialized blockchain app
func NewComposableApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	enabledProposals []wasm.ProposalType,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig EncodingConfig,
	appOpts servertypes.AppOptions,
	wasmOpts []wasm.Option,
	baseAppOptions ...func(*baseapp.BaseApp),
) *ComposableApp {
	appCodec := encodingConfig.Marshaler
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(Name, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(encodingConfig.TxConfig.TxEncoder())

	app := &ComposableApp{
		BaseApp:           bApp,
		AppKeepers:        keepers.AppKeepers{},
		cdc:               cdc,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
	}

	app.InitSpecialKeepers(
		appCodec,
		cdc,
		bApp,
		invCheckPeriod,
		skipUpgradeHeights,
		homePath,
	)
	app.setupUpgradeStoreLoaders()
	app.InitNormalKeepers(
		appCodec,
		cdc,
		bApp,
		maccPerms,
		invCheckPeriod,
		skipUpgradeHeights,
		homePath,
		appOpts,
		wasmOpts,
		enabledProposals,
	)

	custompfm.NewIBCMiddleware()
	// transferModule := transfer.NewAppModule(app.TransferKeeper)
	transferModule := customibctransfer.NewAppModule(appCodec, app.TransferKeeper, app.BankKeeper)
	routerModule := router.NewAppModule(app.RouterKeeper)
	transfermiddlewareModule := transfermiddleware.NewAppModule(&app.TransferMiddlewareKeeper)
	txBoundaryModule := txBoundary.NewAppModule(appCodec, app.TxBoundaryKeepper)
	ratelimitModule := ratelimitmodule.NewAppModule(&app.RatelimitKeeper)
	icqModule := icq.NewAppModule(app.ICQKeeper)
	ibcHooksModule := ibc_hooks.NewAppModule()
	icaModule := ica.NewAppModule(nil, &app.ICAHostKeeper) // Only ICA Host
	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.mm = module.NewManager(
		genutil.NewAppModule(
			app.AccountKeeper, app.StakingKeeper, app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		custombankmodule.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
		gov.NewAppModule(appCodec, &app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName)),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		customstaking.NewAppModule(appCodec, *app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		stakingmiddleware.NewAppModule(appCodec, app.StakingMiddlewareKeeper),
		ibctransfermiddleware.NewAppModule(appCodec, app.IbcTransferMiddlewareKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		upgrade.NewAppModule(app.UpgradeKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		params.NewAppModule(app.ParamsKeeper),
		transferModule,
		icqModule,
		ibcHooksModule,
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		wasm08.NewAppModule(app.Wasm08Keeper),
		wasm.NewAppModule(appCodec, &app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)),
		routerModule,
		transfermiddlewareModule,
		txBoundaryModule,
		icaModule,
		ratelimitModule,
		alliancemodule.NewAppModule(appCodec, app.AllianceKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		// this line is used by starport scaffolding # stargate/app/appModule
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		vestingtypes.ModuleName,
		ibchost.ModuleName,
		ibctransfertypes.ModuleName,
		routertypes.ModuleName,
		transfermiddlewaretypes.ModuleName,
		txBoundaryTypes.ModuleName,
		ratelimitmoduletypes.ModuleName,
		ibchookstypes.ModuleName,
		icqtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		consensusparamtypes.ModuleName,
		wasm08types.ModuleName,
		icatypes.ModuleName,
		wasm.ModuleName,
		alliancemoduletypes.ModuleName,
		stakingmiddlewaretypes.ModuleName,
		ibctransfermiddlewaretypes.ModuleName,
		// this line is used by starport scaffolding # stargate/app/beginBlockers
	)

	app.mm.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		vestingtypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibchost.ModuleName,
		routertypes.ModuleName,
		transfermiddlewaretypes.ModuleName,
		txBoundaryTypes.ModuleName,
		ratelimitmoduletypes.ModuleName,
		ibchookstypes.ModuleName,
		ibctransfertypes.ModuleName,
		icqtypes.ModuleName,
		consensusparamtypes.ModuleName,
		wasm08types.ModuleName,
		icatypes.ModuleName,
		wasm.ModuleName,
		alliancemoduletypes.ModuleName,
		stakingmiddlewaretypes.ModuleName,
		ibctransfermiddlewaretypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		vestingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		ibchost.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibctransfertypes.ModuleName,
		icqtypes.ModuleName,
		routertypes.ModuleName,
		transfermiddlewaretypes.ModuleName,
		txBoundaryTypes.ModuleName,
		ratelimitmoduletypes.ModuleName,
		ibchookstypes.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		consensusparamtypes.ModuleName,
		wasm08types.ModuleName,
		icatypes.ModuleName,
		wasm.ModuleName,
		alliancemoduletypes.ModuleName,
		stakingmiddlewaretypes.ModuleName,
		ibctransfermiddlewaretypes.ModuleName,
		// this line is used by starport scaffolding # stargate/app/initGenesis
	)

	app.mm.RegisterInvariants(app.CrisisKeeper)
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	app.setupUpgradeHandlers()

	// create the simulation manager and define the order of the modules for deterministic simulations
	// app.sm = module.NewSimulationManager(
	// 	auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts),
	// 	bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
	// 	capability.NewAppModule(appCodec, *app.CapabilityKeeper),
	// 	feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
	// 	gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
	// 	mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper),
	// 	staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
	// 	distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
	// 	slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
	// 	params.NewAppModule(app.ParamsKeeper),
	// 	evidence.NewAppModule(app.EvidenceKeeper),
	// 	ibc.NewAppModule(app.IBCKeeper),
	// 	transferModule,
	// 	// monitoringModule,
	// 	stakeibcModule,
	// 	epochsModule,
	// 	interchainQueryModule,
	// 	recordsModule,
	// icacallbacksModule,
	// this line is used by starport scaffolding # stargate/app/appModule
	// )
	// app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(app.GetKVStoreKey())
	app.MountTransientStores(app.GetTransientStoreKey())
	app.MountMemoryStores(app.GetMemoryStoreKey())

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)

	app.SetAnteHandler(ante.NewAnteHandler(
		appOpts,
		app.AccountKeeper,
		app.BankKeeper,
		app.FeeGrantKeeper,
		nil,
		authante.DefaultSigVerificationGasConsumer,
		encodingConfig.TxConfig.SignModeHandler(),
		app.IBCKeeper,
		app.TransferMiddlewareKeeper,
		app.TxBoundaryKeepper,
		appCodec,
	))
	app.SetEndBlocker(app.EndBlocker)

	if manager := app.SnapshotManager(); manager != nil {
		err := manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.WasmKeeper),
			wasm08keeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.Wasm08Keeper),
		)
		if err != nil {
			panic(fmt.Errorf("failed to register snapshot extension: %s", err))
		}
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
	}

	return app
}

// Name returns the name of the App
func (app *ComposableApp) Name() string { return app.BaseApp.Name() }

// GetBaseApp returns the base app of the application
func (app *ComposableApp) GetBaseApp() *baseapp.BaseApp { return app.BaseApp }

// GetStakingKeeper implements the TestingApp interface.
func (app *ComposableApp) GetStakingKeeper() ibctestingtypes.StakingKeeper {
	return app.StakingKeeper
}

// GetIBCKeeper implements the TestingApp interface.
func (app *ComposableApp) GetTransferKeeper() *ibctransferkeeper.Keeper {
	return &app.TransferKeeper.Keeper
}

// GetIBCKeeper implements the TestingApp interface.
func (app *ComposableApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

// GetScopedIBCKeeper implements the TestingApp interface.
func (app *ComposableApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

// GetTxConfig implements the TestingApp interface.
func (app *ComposableApp) GetTxConfig() client.TxConfig {
	cfg := MakeEncodingConfig()
	return cfg.TxConfig
}

// BeginBlocker application updates every begin block
func (app *ComposableApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	BeginBlockForks(ctx, app)
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *ComposableApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *ComposableApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *ComposableApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *ComposableApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	// DO NOT REMOVE: StringMapKeys fixes non-deterministic map iteration
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *ComposableApp) LegacyAmino() *codec.LegacyAmino {
	return app.cdc
}

// AppCodec returns an app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *ComposableApp) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns an InterfaceRegistry
func (app *ComposableApp) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *ComposableApp) RegisterAPIRoutes(apiSvr *api.Server, _ config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *ComposableApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *ComposableApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(clientCtx, app.BaseApp.GRPCQueryRouter(), app.interfaceRegistry, app.Query)
}

// RegisterNodeService registers the node gRPC Query service.
func (app *ComposableApp) RegisterNodeService(clientCtx client.Context) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

// SimulationManager implements the SimulationApp interface
func (app *ComposableApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// configure store loader that checks if version == upgradeHeight and applies store upgrades
func (app *ComposableApp) setupUpgradeStoreLoaders() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}

	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	currentHeight := app.CommitMultiStore().LastCommitID().Version

	if upgradeInfo.Height == currentHeight+1 {
		app.customPreUpgradeHandler(upgradeInfo)
	}

	for _, upgrade := range Upgrades {
		upgrade := upgrade
		if upgradeInfo.Name == upgrade.UpgradeName {
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &upgrade.StoreUpgrades))
		}
	}
}

func (app *ComposableApp) customPreUpgradeHandler(_ upgradetypes.Plan) {
	// switch upgradeInfo.Name {
	// default:
	// }
}

func (app *ComposableApp) setupUpgradeHandlers() {
	for _, upgrade := range Upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgrade.UpgradeName,
			upgrade.CreateUpgradeHandler(
				app.mm,
				app.configurator,
				app.BaseApp,
				app.AppCodec(),
				&app.AppKeepers,
			),
		)
	}
}
