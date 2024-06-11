package keepers

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/params"

	circuitkeeper "cosmossdk.io/x/circuit/keeper"
	circuittypes "cosmossdk.io/x/circuit/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"

	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"

	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	evidencetypes "cosmossdk.io/x/evidence/types"

	"cosmossdk.io/x/feegrant"
	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"

	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"

	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	customstaking "github.com/notional-labs/composable/v6/custom/staking/keeper"

	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	upgradetypes "cosmossdk.io/x/upgrade/types"

	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	icahost "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer"

	icq "github.com/cosmos/ibc-apps/modules/async-icq/v8"
	icqkeeper "github.com/cosmos/ibc-apps/modules/async-icq/v8/keeper"
	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v8/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	customibctransferkeeper "github.com/notional-labs/composable/v6/custom/ibc-transfer/keeper"

	custombankkeeper "github.com/notional-labs/composable/v6/custom/bank/keeper"

	pfmkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/keeper"
	pfmtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"

	transfermiddleware "github.com/notional-labs/composable/v6/x/transfermiddleware"
	transfermiddlewarekeeper "github.com/notional-labs/composable/v6/x/transfermiddleware/keeper"
	transfermiddlewaretypes "github.com/notional-labs/composable/v6/x/transfermiddleware/types"

	txBoundaryKeeper "github.com/notional-labs/composable/v6/x/tx-boundary/keeper"
	txBoundaryTypes "github.com/notional-labs/composable/v6/x/tx-boundary/types"

	ratelimitmodule "github.com/notional-labs/composable/v6/x/ratelimit"
	ratelimitmodulekeeper "github.com/notional-labs/composable/v6/x/ratelimit/keeper"
	ratelimitmoduletypes "github.com/notional-labs/composable/v6/x/ratelimit/types"

	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"

	mintkeeper "github.com/notional-labs/composable/v6/x/mint/keeper"
	minttypes "github.com/notional-labs/composable/v6/x/mint/types"

	storetypes "cosmossdk.io/store/types"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	wasm08Keeper "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/keeper"
	wasm08types "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmvm "github.com/CosmWasm/wasmvm"

	ibc_hooks "github.com/notional-labs/composable/v6/x/ibc-hooks"
	ibchookskeeper "github.com/notional-labs/composable/v6/x/ibc-hooks/keeper"
	ibchookstypes "github.com/notional-labs/composable/v6/x/ibc-hooks/types"
	stakingmiddleware "github.com/notional-labs/composable/v6/x/stakingmiddleware/keeper"
	stakingmiddlewaretypes "github.com/notional-labs/composable/v6/x/stakingmiddleware/types"

	custompfm "github.com/notional-labs/composable/v6/custom/custompfm/keeper"
	ibctransfermiddleware "github.com/notional-labs/composable/v6/x/ibctransfermiddleware/keeper"
	ibctransfermiddlewaretypes "github.com/notional-labs/composable/v6/x/ibctransfermiddleware/types"
)

const (
	AccountAddressPrefix = "pica"
	authorityAddress     = "pica10556m38z4x6pqalr9rl5ytf3cff8q46nf36090" // convert from: centauri10556m38z4x6pqalr9rl5ytf3cff8q46nk85k9m
)

type AppKeepers struct {
	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper    authkeeper.AccountKeeper
	BankKeeper       custombankkeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	StakingKeeper    *customstaking.Keeper
	SlashingKeeper   slashingkeeper.Keeper
	MintKeeper       mintkeeper.Keeper
	DistrKeeper      distrkeeper.Keeper
	GovKeeper        govkeeper.Keeper
	CrisisKeeper     *crisiskeeper.Keeper
	UpgradeKeeper    *upgradekeeper.Keeper
	ParamsKeeper     paramskeeper.Keeper
	IBCKeeper        *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	EvidenceKeeper   evidencekeeper.Keeper
	TransferKeeper   customibctransferkeeper.Keeper
	ICQKeeper        icqkeeper.Keeper
	ICAHostKeeper    icahostkeeper.Keeper
	FeeGrantKeeper   feegrantkeeper.Keeper
	AuthzKeeper      authzkeeper.Keeper
	GroupKeeper      groupkeeper.Keeper
	Wasm08Keeper     wasm08Keeper.Keeper // TODO: use this name ?
	WasmKeeper       wasmkeeper.Keeper
	IBCHooksKeeper   *ibchookskeeper.Keeper
	Ics20WasmHooks   *ibc_hooks.WasmHooks
	HooksICS4Wrapper ibc_hooks.ICS4Middleware
	// make scoped keepers public for test purposes
	ScopedIBCKeeper       capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper  capabilitykeeper.ScopedKeeper
	ScopedWasmKeeper      capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper   capabilitykeeper.ScopedKeeper
	ScopedRateLimitKeeper capabilitykeeper.ScopedKeeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	// this line is used by starport scaffolding # stargate/app/keeperDeclaration
	TransferMiddlewareKeeper    transfermiddlewarekeeper.Keeper
	TxBoundaryKeepper           txBoundaryKeeper.Keeper
	PfmKeeper                   *pfmkeeper.Keeper
	RatelimitKeeper             ratelimitmodulekeeper.Keeper
	StakingMiddlewareKeeper     stakingmiddleware.Keeper
	IbcTransferMiddlewareKeeper ibctransfermiddleware.Keeper

	CircuitKeeper circuitkeeper.Keeper
}

// InitNormalKeepers initializes all 'normal' keepers.
func (appKeepers *AppKeepers) InitNormalKeepers(
	logger log.Logger,
	appCodec codec.Codec,
	cdc *codec.LegacyAmino,
	bApp *baseapp.BaseApp,
	maccPerms map[string][]string,
	invCheckPeriod uint,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	appOpts servertypes.AppOptions,
) {
	govModAddress := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	// add keepers
	appKeepers.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec, runtime.NewKVStoreService(appKeepers.keys[authtypes.StoreKey]), authtypes.ProtoBaseAccount, maccPerms, authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()), AccountAddressPrefix, govModAddress,
	)

	appKeepers.BankKeeper = custombankkeeper.NewBaseKeeper(
		logger,
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[banktypes.StoreKey]),
		appKeepers.AccountKeeper,
		appKeepers.BlacklistedModuleAccountAddrs(maccPerms),
		&appKeepers.TransferMiddlewareKeeper,
		govModAddress,
	)

	appKeepers.AuthzKeeper = authzkeeper.NewKeeper(
		runtime.NewKVStoreService(appKeepers.keys[authzkeeper.StoreKey]),
		appCodec,
		bApp.MsgServiceRouter(),
		appKeepers.AccountKeeper,
	)

	appKeepers.StakingMiddlewareKeeper = stakingmiddleware.NewKeeper(appCodec, appKeepers.keys[stakingmiddlewaretypes.StoreKey], govModAddress)
	appKeepers.IbcTransferMiddlewareKeeper = ibctransfermiddleware.NewKeeper(appCodec, appKeepers.keys[ibctransfermiddlewaretypes.StoreKey], govModAddress,
		[]string{
			"pica1ay9y5uns9khw2kzaqr3r33v2pkuptfnnunlt5x",
			"pica14lz7gaw92valqjearnye4shex7zg2p05yfguqm",
			"pica1r2zlh2xn85v8ljmwymnfrnsmdzjl7k6w9f2ja8",
			"pica10556m38z4x6pqalr9rl5ytf3cff8q46nf36090",
		})

	appKeepers.StakingKeeper = customstaking.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[stakingtypes.StoreKey]),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		govModAddress,
		&appKeepers.StakingMiddlewareKeeper,
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	)

	appKeepers.MintKeeper = mintkeeper.NewKeeper(
		appCodec, appKeepers.keys[minttypes.StoreKey], appKeepers.StakingKeeper,
		appKeepers.AccountKeeper, appKeepers.BankKeeper, authtypes.FeeCollectorName, govModAddress,
	)

	appKeepers.DistrKeeper = distrkeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(appKeepers.keys[distrtypes.StoreKey]), appKeepers.AccountKeeper, appKeepers.BankKeeper,
		appKeepers.StakingKeeper, authtypes.FeeCollectorName, govModAddress,
	)
	appKeepers.StakingKeeper.RegisterKeepers(appKeepers.DistrKeeper, appKeepers.MintKeeper)
	appKeepers.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, cdc, runtime.NewKVStoreService(appKeepers.keys[slashingtypes.StoreKey]), appKeepers.StakingKeeper, govModAddress,
	)

	appKeepers.CircuitKeeper = circuitkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[circuittypes.StoreKey]),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		appKeepers.AccountKeeper.AddressCodec(),
	)
	bApp.SetCircuitBreaker(&appKeepers.CircuitKeeper)

	appKeepers.CrisisKeeper = crisiskeeper.NewKeeper(appCodec, runtime.NewKVStoreService(appKeepers.keys[crisistypes.StoreKey]),
		invCheckPeriod, appKeepers.BankKeeper, authtypes.FeeCollectorName, govModAddress, appKeepers.AccountKeeper.AddressCodec(),
	)

	groupConfig := group.DefaultConfig()
	/*
		Example of setting group params:
		groupConfig.MaxMetadataLen = 1000
	*/
	appKeepers.GroupKeeper = groupkeeper.NewKeeper(
		appKeepers.keys[group.StoreKey],
		appCodec,
		bApp.MsgServiceRouter(),
		appKeepers.AccountKeeper,
		groupConfig,
	)

	appKeepers.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(appKeepers.keys[feegrant.StoreKey]), appKeepers.AccountKeeper)
	appKeepers.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, runtime.NewKVStoreService(appKeepers.keys[upgradetypes.StoreKey]), appCodec, homePath, bApp, govModAddress)

	appKeepers.BankKeeper.RegisterKeepers(appKeepers.AccountKeeper, appKeepers.StakingKeeper)
	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	appKeepers.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(appKeepers.DistrKeeper.Hooks(), appKeepers.SlashingKeeper.Hooks()),
	)

	// ... other modules keepers

	// Create IBC Keeper
	appKeepers.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		appKeepers.keys[ibcexported.StoreKey],
		appKeepers.GetSubspace(ibcexported.ModuleName),
		appKeepers.StakingKeeper,
		appKeepers.UpgradeKeeper,
		appKeepers.ScopedIBCKeeper,
		govModAddress,
	)
	// ICA Host keeper
	appKeepers.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec, appKeepers.keys[icahosttypes.StoreKey], appKeepers.GetSubspace(icahosttypes.SubModuleName),
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.ChannelKeeper, appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper, appKeepers.ScopedICAHostKeeper, bApp.MsgServiceRouter(), govModAddress,
	)

	icaHostStack := icahost.NewIBCModule(appKeepers.ICAHostKeeper)

	hooksKeeper := ibchookskeeper.NewKeeper(
		appKeepers.keys[ibchookstypes.StoreKey],
	)
	appKeepers.IBCHooksKeeper = &hooksKeeper

	composablePrefix := sdk.GetConfig().GetBech32AccountAddrPrefix()
	wasmHooks := ibc_hooks.NewWasmHooks(&hooksKeeper, nil, composablePrefix) // The contract keeper needs to be set later
	appKeepers.Ics20WasmHooks = &wasmHooks
	appKeepers.HooksICS4Wrapper = ibc_hooks.NewICS4Middleware(
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.Ics20WasmHooks,
	)

	appKeepers.TransferMiddlewareKeeper = transfermiddlewarekeeper.NewKeeper(
		appKeepers.keys[transfermiddlewaretypes.StoreKey],
		appKeepers.GetSubspace(transfermiddlewaretypes.ModuleName),
		appCodec,
		&appKeepers.RatelimitKeeper,
		&appKeepers.TransferKeeper,
		appKeepers.BankKeeper,
		&appKeepers.IbcTransferMiddlewareKeeper,
		govModAddress,
	)

	appKeepers.TxBoundaryKeepper = txBoundaryKeeper.NewKeeper(
		appCodec,
		appKeepers.keys[txBoundaryTypes.StoreKey],
		govModAddress,
	)

	appKeepers.PfmKeeper = pfmkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[pfmtypes.StoreKey],
		appKeepers.TransferKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.DistrKeeper,
		appKeepers.BankKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.TransferMiddlewareKeeper,
		govModAddress,
	)

	appKeepers.TransferKeeper = customibctransferkeeper.NewKeeper(
		appCodec, appKeepers.keys[ibctransfertypes.StoreKey],
		appKeepers.GetSubspace(ibctransfertypes.ModuleName),
		&appKeepers.TransferMiddlewareKeeper, // ICS4Wrapper
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.ScopedTransferKeeper,
		&appKeepers.IbcTransferMiddlewareKeeper,
		&appKeepers.BankKeeper,
		govModAddress,
	)
	appKeepers.PfmKeeper.SetTransferKeeper(appKeepers.TransferKeeper)

	appKeepers.RatelimitKeeper = *ratelimitmodulekeeper.NewKeeper(
		appCodec,
		appKeepers.keys[ratelimitmoduletypes.StoreKey],
		appKeepers.GetSubspace(ratelimitmoduletypes.ModuleName),
		appKeepers.BankKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		// TODO: Implement ICS4Wrapper in Records and pass records keeper here
		&appKeepers.HooksICS4Wrapper, // ICS4Wrapper
		appKeepers.TransferMiddlewareKeeper,
		govModAddress,
	)

	scopedICQKeeper := appKeepers.CapabilityKeeper.ScopeToModule(icqtypes.ModuleName)

	appKeepers.ICQKeeper = icqkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[icqtypes.StoreKey],
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.PortKeeper,
		scopedICQKeeper, bApp.GRPCQueryRouter(), govModAddress,
	)

	icqIBCModule := icq.NewIBCModule(appKeepers.ICQKeeper)

	// Create Transfer Keepers
	// * SendPacket. Originates from the transferKeeper and goes up the stack:
	// transferKeeper.SendPacket -> transfermiddleware.SendPacket -> ibc_rate_limit.SendPacket -> ibc_hooks.SendPacket -> channel.SendPacket
	// * RecvPacket, message that originates from core IBC and goes down to app, the flow is the other way
	// channel.RecvPacket -> ibc_hooks.OnRecvPacket -> ibc_rate_limit.OnRecvPacket -> forward.OnRecvPacket -> transfermiddleware_OnRecvPacket -> transfer.OnRecvPacket
	//

	transferIBCModule := transfer.NewIBCModule(appKeepers.TransferKeeper.Keeper)

	transfermiddlewareStack := transfermiddleware.NewIBCMiddleware(
		transferIBCModule,
		appKeepers.TransferMiddlewareKeeper,
	)

	ibcMiddlewareStack := custompfm.NewIBCMiddleware(
		transfermiddlewareStack,
		appKeepers.PfmKeeper,
		0,
		pfmkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
		pfmkeeper.DefaultRefundTransferPacketTimeoutTimestamp,
		&appKeepers.IbcTransferMiddlewareKeeper,
		&appKeepers.BankKeeper,
	)
	ratelimitMiddlewareStack := ratelimitmodule.NewIBCMiddleware(appKeepers.RatelimitKeeper, ibcMiddlewareStack)
	hooksTransferMiddleware := ibc_hooks.NewIBCMiddleware(ratelimitMiddlewareStack, &appKeepers.HooksICS4Wrapper)

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(appKeepers.keys[evidencetypes.StoreKey]), appKeepers.StakingKeeper, appKeepers.SlashingKeeper, appKeepers.AccountKeeper.AddressCodec(), runtime.ProvideCometInfoService(),
	)
	// If evidence needs to be handled for the app, set routes in pfm here and seal
	appKeepers.EvidenceKeeper = *evidenceKeeper

	wasmDir := filepath.Join(homePath, "wasm")
	wasmConfig, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}

	// increase default wasm size in all wasmd related codes (as on Neutorn/Osmosis)
	wasmtypes.MaxWasmSize *= 2

	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	availableCapabilities := strings.Join(AllCapabilities(), ",")

	wasmer, err := wasmvm.NewVM(
		wasmDir,
		strings.Join(AllCapabilities(), ","),
		32, // default of 32
		false,
		0,
	)
	if err != nil {
		panic(err)
	}

	wasmOpts := []wasmkeeper.Option{
		wasmkeeper.WithWasmEngine(wasmer),
	}

	appKeepers.WasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[wasmtypes.StoreKey]),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		distrkeeper.NewQuerier(appKeepers.DistrKeeper),
		appKeepers.IBCKeeper.ChannelKeeper, // ISC4 Wrapper: fee IBC middleware
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.PortKeeper,
		appKeepers.ScopedWasmKeeper,
		appKeepers.TransferKeeper,
		bApp.MsgServiceRouter(),
		bApp.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		availableCapabilities,
		govModAddress,
		wasmOpts...,
	)
	appKeepers.Ics20WasmHooks.ContractKeeper = &appKeepers.WasmKeeper

	wasmDataDir := filepath.Join(homePath, "wasm_client_data")
	wasmSupportedFeatures := strings.Join([]string{"storage", "iterator"}, ",")
	wasmMemoryLimitMb := uint32(math.Pow(2, 12))
	wasmPrintDebug := true
	wasmCacheSizeMb := uint32(math.Pow(2, 8))

	vm, err := wasmvm.NewVM(wasmDataDir, wasmSupportedFeatures, wasmMemoryLimitMb, wasmPrintDebug, wasmCacheSizeMb)
	if err != nil {
		panic(err)
	}

	// use same VM for wasm
	appKeepers.Wasm08Keeper = wasm08Keeper.NewKeeperWithVM(
		appCodec,
		runtime.NewKVStoreService(appKeepers.keys[wasm08types.StoreKey]),
		&appKeepers.IBCKeeper.ClientKeeper,
		authorityAddress,
		vm,
		bApp.GRPCQueryRouter(),
	)

	// Register Gov (must be registered after stakeibc)
	govRouter := govtypesv1beta1.NewRouter()
	// Register the proposal types
	// Deprecated: Avoid adding new handlers, instead use the new proposal flow
	// by granting the governance module the right to execute the message.
	// See: https://docs.cosmos.network/main/modules/gov#proposal-messages
	govRouter.AddRoute(govtypes.RouterKey, govtypesv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(appKeepers.ParamsKeeper))

	govKeeper := *govkeeper.NewKeeper(
		appCodec, runtime.NewKVStoreService(appKeepers.keys[govtypes.StoreKey]), appKeepers.AccountKeeper, appKeepers.BankKeeper,
		appKeepers.StakingKeeper, appKeepers.DistrKeeper, bApp.MsgServiceRouter(), govtypes.DefaultConfig(),
		govModAddress,
	)

	govKeeper.SetLegacyRouter(govRouter)

	appKeepers.GovKeeper = *govKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
		// register the governance hooks
		),
	)

	ibcRouter := porttypes.NewRouter()
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, hooksTransferMiddleware)
	ibcRouter.AddRoute(icqtypes.ModuleName, icqIBCModule)
	ibcRouter.AddRoute(wasmtypes.ModuleName, wasm.NewIBCHandler(appKeepers.WasmKeeper, appKeepers.IBCKeeper.ChannelKeeper, appKeepers.IBCKeeper.ChannelKeeper))
	ibcRouter.AddRoute(icahosttypes.SubModuleName, icaHostStack)

	// this line is used by starport scaffolding # ibc/app/pfm
	appKeepers.IBCKeeper.SetRouter(ibcRouter)
}

// InitSpecialKeepers initiates special keepers (upgradekeeper, params keeper)
func (appKeepers *AppKeepers) InitSpecialKeepers(
	appCodec codec.Codec,
	cdc *codec.LegacyAmino,
	bApp *baseapp.BaseApp,
	_ uint, // invCheckPeriod
	skipUpgradeHeights map[int64]bool,
	homePath string,
) {
	govModAddress := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	appKeepers.GenerateKeys()
	appKeepers.ParamsKeeper = appKeepers.initParamsKeeper(appCodec, cdc, appKeepers.keys[paramstypes.StoreKey], appKeepers.tkeys[paramstypes.TStoreKey])
	appKeepers.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, appKeepers.keys[capabilitytypes.StoreKey], appKeepers.memKeys[capabilitytypes.MemStoreKey])

	// set the BaseApp's parameter store
	appKeepers.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(appKeepers.keys[consensusparamtypes.StoreKey]), govModAddress, runtime.EventService{})
	bApp.SetParamStore(&appKeepers.ConsensusParamsKeeper.ParamsStore)

	// grant capabilities for the ibc and ibc-transfer modules
	appKeepers.ScopedIBCKeeper = appKeepers.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	appKeepers.ScopedTransferKeeper = appKeepers.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	appKeepers.ScopedWasmKeeper = appKeepers.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)
	appKeepers.ScopedICAHostKeeper = appKeepers.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	appKeepers.ScopedRateLimitKeeper = appKeepers.CapabilityKeeper.ScopeToModule(ratelimitmoduletypes.ModuleName)

	appKeepers.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, runtime.NewKVStoreService(appKeepers.keys[upgradetypes.StoreKey]), appCodec, homePath, bApp, authtypes.NewModuleAddress(govtypes.ModuleName).String())
}

// initParamsKeeper init params keeper and its subspaces
func (appKeepers *AppKeepers) initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	// register the IBC key tables for legacy param subspaces

	// register the key tables for legacy param subspaces
	keyTable := ibcclienttypes.ParamKeyTable()
	keyTable.RegisterParamSet(&ibcconnectiontypes.Params{})
	paramsKeeper.Subspace(ibcexported.ModuleName).WithKeyTable(keyTable)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName).WithKeyTable(ibctransfertypes.ParamKeyTable())
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName).WithKeyTable(icacontrollertypes.ParamKeyTable())
	paramsKeeper.Subspace(icahosttypes.SubModuleName).WithKeyTable(icahosttypes.ParamKeyTable())

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(pfmtypes.ModuleName).WithKeyTable(pfmtypes.ParamKeyTable()) // TODO:
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govtypesv1.ParamKeyTable())
	paramsKeeper.Subspace(minttypes.ModuleName).WithKeyTable(minttypes.ParamKeyTable())
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ratelimitmoduletypes.ModuleName)
	paramsKeeper.Subspace(icqtypes.ModuleName).WithKeyTable(icqtypes.ParamKeyTable())
	paramsKeeper.Subspace(wasm08types.ModuleName)
	paramsKeeper.Subspace(wasmtypes.ModuleName)
	paramsKeeper.Subspace(transfermiddlewaretypes.ModuleName)
	paramsKeeper.Subspace(stakingmiddlewaretypes.ModuleName)
	paramsKeeper.Subspace(ibctransfermiddlewaretypes.ModuleName)

	return paramsKeeper
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (appKeepers *AppKeepers) BlacklistedModuleAccountAddrs(maccPerms map[string][]string) map[string]bool {
	modAccAddrs := make(map[string]bool)
	// DO NOT REMOVE: StringMapKeys fixes non-deterministic map iteration
	for acc := range maccPerms {
		if acc != authtypes.FeeCollectorName {
			modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
		}
	}
	return modAccAddrs
}
