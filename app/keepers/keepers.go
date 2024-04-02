package keepers

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"

	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"

	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"

	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"

	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"

	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"

	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	customstaking "github.com/ComposableFi/composable-cosmos/v6/custom/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	icahost "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"

	customibctransferkeeper "github.com/ComposableFi/composable-cosmos/v6/custom/ibc-transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclient "github.com/cosmos/ibc-go/v7/modules/core/02-client"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibchost "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	icq "github.com/strangelove-ventures/async-icq/v7"
	icqkeeper "github.com/strangelove-ventures/async-icq/v7/keeper"
	icqtypes "github.com/strangelove-ventures/async-icq/v7/types"

	custombankkeeper "github.com/ComposableFi/composable-cosmos/v6/custom/bank/keeper"

	router "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward"
	routerkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/keeper"
	routertypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"

	alliancemodule "github.com/terra-money/alliance/x/alliance"
	alliancemodulekeeper "github.com/terra-money/alliance/x/alliance/keeper"
	alliancemoduletypes "github.com/terra-money/alliance/x/alliance/types"

	transfermiddleware "github.com/ComposableFi/composable-cosmos/v6/x/transfermiddleware"
	transfermiddlewarekeeper "github.com/ComposableFi/composable-cosmos/v6/x/transfermiddleware/keeper"
	transfermiddlewaretypes "github.com/ComposableFi/composable-cosmos/v6/x/transfermiddleware/types"

	txBoundaryKeeper "github.com/ComposableFi/composable-cosmos/v6/x/tx-boundary/keeper"
	txBoundaryTypes "github.com/ComposableFi/composable-cosmos/v6/x/tx-boundary/types"

	ratelimitmodule "github.com/ComposableFi/composable-cosmos/v6/x/ratelimit"
	ratelimitmodulekeeper "github.com/ComposableFi/composable-cosmos/v6/x/ratelimit/keeper"
	ratelimitmoduletypes "github.com/ComposableFi/composable-cosmos/v6/x/ratelimit/types"

	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"

	mintkeeper "github.com/ComposableFi/composable-cosmos/v6/x/mint/keeper"
	minttypes "github.com/ComposableFi/composable-cosmos/v6/x/mint/types"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	wasm08Keeper "github.com/cosmos/ibc-go/v7/modules/light-clients/08-wasm/keeper"
	wasm08types "github.com/cosmos/ibc-go/v7/modules/light-clients/08-wasm/types"

	ibc_hooks "github.com/ComposableFi/composable-cosmos/v6/x/ibc-hooks"
	ibchookskeeper "github.com/ComposableFi/composable-cosmos/v6/x/ibc-hooks/keeper"
	ibchookstypes "github.com/ComposableFi/composable-cosmos/v6/x/ibc-hooks/types"
	stakingmiddleware "github.com/ComposableFi/composable-cosmos/v6/x/stakingmiddleware/keeper"
	stakingmiddlewaretypes "github.com/ComposableFi/composable-cosmos/v6/x/stakingmiddleware/types"

	ibctransfermiddleware "github.com/ComposableFi/composable-cosmos/v6/x/ibctransfermiddleware/keeper"
	ibctransfermiddlewaretypes "github.com/ComposableFi/composable-cosmos/v6/x/ibctransfermiddleware/types"
)

const (
	AccountAddressPrefix = "composable"
	authorityAddress     = "centauri10556m38z4x6pqalr9rl5ytf3cff8q46nk85k9m" // convert from: centauri10556m38z4x6pqalr9rl5ytf3cff8q46nk85k9m
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
	WasmKeeper       wasm.Keeper
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
	RouterKeeper                *routerkeeper.Keeper
	RatelimitKeeper             ratelimitmodulekeeper.Keeper
	AllianceKeeper              alliancemodulekeeper.Keeper
	StakingMiddlewareKeeper     stakingmiddleware.Keeper
	IbcTransferMiddlewareKeeper ibctransfermiddleware.Keeper
}

// InitNormalKeepers initializes all 'normal' keepers.
func (appKeepers *AppKeepers) InitNormalKeepers(
	appCodec codec.Codec,
	cdc *codec.LegacyAmino,
	bApp *baseapp.BaseApp,
	maccPerms map[string][]string,
	invCheckPeriod uint,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	appOpts servertypes.AppOptions,
	wasmOpts []wasm.Option,
	enabledProposals []wasm.ProposalType,
) {
	// add keepers
	appKeepers.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec, appKeepers.keys[authtypes.StoreKey], authtypes.ProtoBaseAccount, maccPerms, AccountAddressPrefix, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	appKeepers.BankKeeper = custombankkeeper.NewBaseKeeper(
		appCodec, appKeepers.keys[banktypes.StoreKey], appKeepers.AccountKeeper, appKeepers.BlacklistedModuleAccountAddrs(maccPerms), &appKeepers.TransferMiddlewareKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	appKeepers.AuthzKeeper = authzkeeper.NewKeeper(
		appKeepers.keys[authzkeeper.StoreKey],
		appCodec,
		bApp.MsgServiceRouter(),
		appKeepers.AccountKeeper,
	)

	appKeepers.StakingMiddlewareKeeper = stakingmiddleware.NewKeeper(appCodec, appKeepers.keys[stakingmiddlewaretypes.StoreKey], authtypes.NewModuleAddress(govtypes.ModuleName).String())
	appKeepers.IbcTransferMiddlewareKeeper = ibctransfermiddleware.NewKeeper(appCodec, appKeepers.keys[ibctransfermiddlewaretypes.StoreKey], authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		[]string{"centauri1ay9y5uns9khw2kzaqr3r33v2pkuptfnnr93j5j",
			"centauri14lz7gaw92valqjearnye4shex7zg2p05mlx9q0",
			"centauri1r2zlh2xn85v8ljmwymnfrnsmdzjl7k6w6lytan",
			"centauri10556m38z4x6pqalr9rl5ytf3cff8q46nk85k9m",

			// "centauri1wkjvpgkuchq0r8425g4z4sf6n85zj5wtmqzjv9",

			// "centauri1hj5fveer5cjtn4wd6wstzugjfdxzl0xpzxlwgs",
		})

	appKeepers.StakingKeeper = customstaking.NewKeeper(
		appCodec, appKeepers.keys[stakingtypes.StoreKey], appKeepers.AccountKeeper, appKeepers.BankKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String(), &appKeepers.StakingMiddlewareKeeper,
	)

	appKeepers.MintKeeper = mintkeeper.NewKeeper(
		appCodec, appKeepers.keys[minttypes.StoreKey], appKeepers.StakingKeeper,
		appKeepers.AccountKeeper, appKeepers.BankKeeper, authtypes.FeeCollectorName, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	appKeepers.DistrKeeper = distrkeeper.NewKeeper(
		appCodec, appKeepers.keys[distrtypes.StoreKey], appKeepers.AccountKeeper, appKeepers.BankKeeper,
		appKeepers.StakingKeeper, authtypes.FeeCollectorName, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	appKeepers.StakingKeeper.RegisterKeepers(appKeepers.DistrKeeper, appKeepers.MintKeeper)
	appKeepers.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, cdc, appKeepers.keys[slashingtypes.StoreKey], appKeepers.StakingKeeper, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	appKeepers.CrisisKeeper = crisiskeeper.NewKeeper(appCodec, appKeepers.keys[crisistypes.StoreKey],
		invCheckPeriod, appKeepers.BankKeeper, authtypes.FeeCollectorName, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
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

	appKeepers.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, appKeepers.keys[feegrant.StoreKey], appKeepers.AccountKeeper)
	appKeepers.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, appKeepers.keys[upgradetypes.StoreKey], appCodec, homePath, bApp, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	appKeepers.AllianceKeeper = alliancemodulekeeper.NewKeeper(
		appCodec,
		appKeepers.keys[alliancemoduletypes.StoreKey],
		appKeepers.GetSubspace(alliancemoduletypes.ModuleName),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		appKeepers.DistrKeeper,
	)

	appKeepers.BankKeeper.RegisterKeepers(appKeepers.AllianceKeeper, appKeepers.StakingKeeper)
	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	appKeepers.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(appKeepers.DistrKeeper.Hooks(), appKeepers.SlashingKeeper.Hooks(), appKeepers.AllianceKeeper.StakingHooks()),
	)

	// ... other modules keepers

	// Create IBC Keeper
	appKeepers.IBCKeeper = ibckeeper.NewKeeper(
		appCodec, appKeepers.keys[ibchost.StoreKey], appKeepers.GetSubspace(ibchost.ModuleName), appKeepers.StakingKeeper, appKeepers.UpgradeKeeper, appKeepers.ScopedIBCKeeper,
	)

	appKeepers.Wasm08Keeper = wasm08Keeper.NewKeeper(appCodec, appKeepers.keys[wasm08types.StoreKey], authorityAddress, homePath, &appKeepers.IBCKeeper.ClientKeeper)

	// ICA Host keeper
	appKeepers.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec, appKeepers.keys[icahosttypes.StoreKey], appKeepers.GetSubspace(icahosttypes.SubModuleName),
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.ChannelKeeper, &appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper, appKeepers.ScopedICAHostKeeper, bApp.MsgServiceRouter(),
	)

	icaHostStack := icahost.NewIBCModule(appKeepers.ICAHostKeeper)

	// Create Transfer Keepers
	// * SendPacket. Originates from the transferKeeper and goes up the stack:
	// transferKeeper.SendPacket -> transfermiddleware.SendPacket -> ibc_rate_limit.SendPacket -> ibc_hooks.SendPacket -> channel.SendPacket
	// * RecvPacket, message that originates from core IBC and goes down to app, the flow is the other way
	// channel.RecvPacket -> ibc_hooks.OnRecvPacket -> ibc_rate_limit.OnRecvPacket -> forward.OnRecvPacket -> transfermiddleware_OnRecvPacket -> transfer.OnRecvPacket
	//
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
		&appKeepers.TransferKeeper.Keeper,
		appKeepers.BankKeeper,
		authorityAddress,
	)

	appKeepers.TxBoundaryKeepper = txBoundaryKeeper.NewKeeper(
		appCodec,
		appKeepers.keys[txBoundaryTypes.StoreKey],
		authorityAddress,
	)

	appKeepers.TransferKeeper = customibctransferkeeper.NewKeeper(
		appCodec, appKeepers.keys[ibctransfertypes.StoreKey],
		appKeepers.GetSubspace(ibctransfertypes.ModuleName),
		&appKeepers.TransferMiddlewareKeeper, // ICS4Wrapper
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.ScopedTransferKeeper,
		&appKeepers.IbcTransferMiddlewareKeeper,
	)

	appKeepers.RouterKeeper = routerkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[routertypes.StoreKey],
		appKeepers.GetSubspace(routertypes.ModuleName),
		appKeepers.TransferKeeper.Keeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.DistrKeeper,
		appKeepers.BankKeeper,
		appKeepers.TransferMiddlewareKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
	)

	appKeepers.RatelimitKeeper = *ratelimitmodulekeeper.NewKeeper(
		appCodec,
		appKeepers.keys[ratelimitmoduletypes.StoreKey],
		appKeepers.GetSubspace(ratelimitmoduletypes.ModuleName),
		appKeepers.BankKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		// TODO: Implement ICS4Wrapper in Records and pass records keeper here
		&appKeepers.HooksICS4Wrapper, // ICS4Wrapper
		appKeepers.TransferMiddlewareKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	transferIBCModule := transfer.NewIBCModule(appKeepers.TransferKeeper.Keeper)
	scopedICQKeeper := appKeepers.CapabilityKeeper.ScopeToModule(icqtypes.ModuleName)

	appKeepers.ICQKeeper = icqkeeper.NewKeeper(
		appCodec, appKeepers.keys[icqtypes.StoreKey], appKeepers.GetSubspace(icqtypes.ModuleName),
		appKeepers.IBCKeeper.ChannelKeeper, appKeepers.IBCKeeper.ChannelKeeper, &appKeepers.IBCKeeper.PortKeeper,
		scopedICQKeeper, bApp,
	)

	icqIBCModule := icq.NewIBCModule(appKeepers.ICQKeeper)
	transfermiddlewareStack := transfermiddleware.NewIBCMiddleware(
		transferIBCModule,
		appKeepers.TransferMiddlewareKeeper,
	)

	ibcMiddlewareStack := router.NewIBCMiddleware(
		transfermiddlewareStack,
		appKeepers.RouterKeeper,
		0,
		routerkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
		routerkeeper.DefaultRefundTransferPacketTimeoutTimestamp,
	)
	ratelimitMiddlewareStack := ratelimitmodule.NewIBCMiddleware(appKeepers.RatelimitKeeper, ibcMiddlewareStack)
	hooksTransferMiddleware := ibc_hooks.NewIBCMiddleware(ratelimitMiddlewareStack, &appKeepers.HooksICS4Wrapper)

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec, appKeepers.keys[evidencetypes.StoreKey], appKeepers.StakingKeeper, appKeepers.SlashingKeeper,
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
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
	appKeepers.WasmKeeper = wasm.NewKeeper(
		appCodec,
		appKeepers.keys[wasmtypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		distrkeeper.NewQuerier(appKeepers.DistrKeeper),
		appKeepers.IBCKeeper.ChannelKeeper, // ISC4 Wrapper: fee IBC middleware
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.IBCKeeper.PortKeeper,
		appKeepers.ScopedWasmKeeper,
		appKeepers.TransferKeeper.Keeper,
		bApp.MsgServiceRouter(),
		bApp.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		availableCapabilities,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		wasmOpts...,
	)

	appKeepers.Ics20WasmHooks.ContractKeeper = &appKeepers.WasmKeeper

	// Register Gov (must be registered after stakeibc)
	govRouter := govtypesv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypesv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(appKeepers.ParamsKeeper)).
		// AddRoute(distrtypes.RouterKey, distr.NewCommunityPoolSpendProposalHandler(appKeepers.DistrKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(appKeepers.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(appKeepers.IBCKeeper.ClientKeeper)).
		AddRoute(alliancemoduletypes.RouterKey, alliancemodule.NewAllianceProposalHandler(appKeepers.AllianceKeeper))

	// The gov proposal types can be individually enabled
	if len(enabledProposals) != 0 {
		govRouter.AddRoute(wasm.RouterKey, wasm.NewWasmProposalHandler(appKeepers.WasmKeeper, enabledProposals))
	}

	govKeeper := *govkeeper.NewKeeper(
		appCodec, appKeepers.keys[govtypes.StoreKey], appKeepers.AccountKeeper, appKeepers.BankKeeper,
		appKeepers.StakingKeeper, bApp.MsgServiceRouter(), govtypes.DefaultConfig(), authtypes.NewModuleAddress(govtypes.ModuleName).String(),
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
	ibcRouter.AddRoute(wasm.ModuleName, wasm.NewIBCHandler(appKeepers.WasmKeeper, appKeepers.IBCKeeper.ChannelKeeper, appKeepers.IBCKeeper.ChannelKeeper))
	ibcRouter.AddRoute(icahosttypes.SubModuleName, icaHostStack)

	// this line is used by starport scaffolding # ibc/app/router
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
	appKeepers.GenerateKeys()
	appKeepers.ParamsKeeper = appKeepers.initParamsKeeper(appCodec, cdc, appKeepers.keys[paramstypes.StoreKey], appKeepers.tkeys[paramstypes.TStoreKey])
	appKeepers.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, appKeepers.keys[capabilitytypes.StoreKey], appKeepers.memKeys[capabilitytypes.MemStoreKey])

	// set the BaseApp's parameter store
	appKeepers.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(appCodec, appKeepers.keys[consensusparamtypes.StoreKey], authtypes.NewModuleAddress(govtypes.ModuleName).String())
	bApp.SetParamStore(&appKeepers.ConsensusParamsKeeper)

	// grant capabilities for the ibc and ibc-transfer modules
	appKeepers.ScopedIBCKeeper = appKeepers.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	appKeepers.ScopedTransferKeeper = appKeepers.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	appKeepers.ScopedWasmKeeper = appKeepers.CapabilityKeeper.ScopeToModule(wasm.ModuleName)
	appKeepers.ScopedICAHostKeeper = appKeepers.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	appKeepers.ScopedRateLimitKeeper = appKeepers.CapabilityKeeper.ScopeToModule(ratelimitmoduletypes.ModuleName)

	appKeepers.UpgradeKeeper = upgradekeeper.NewKeeper(skipUpgradeHeights, appKeepers.keys[upgradetypes.StoreKey], appCodec, homePath, bApp, authtypes.NewModuleAddress(govtypes.ModuleName).String())
}

// initParamsKeeper init params keeper and its subspaces
func (appKeepers *AppKeepers) initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(routertypes.ModuleName).WithKeyTable(routertypes.ParamKeyTable()) // TODO:
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govtypesv1.ParamKeyTable())     //nolint:staticcheck
	paramsKeeper.Subspace(minttypes.ModuleName).WithKeyTable(minttypes.ParamKeyTable())
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ratelimitmoduletypes.ModuleName)
	paramsKeeper.Subspace(icqtypes.ModuleName)
	paramsKeeper.Subspace(ibchost.ModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(alliancemoduletypes.ModuleName)
	paramsKeeper.Subspace(wasm.ModuleName)
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
