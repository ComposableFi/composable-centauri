package keepers

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"

	// bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	evidencetypes "cosmossdk.io/x/evidence/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"

	"cosmossdk.io/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/group"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	icqtypes "github.com/cosmos/ibc-apps/modules/async-icq/v8/types"
	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibchost "github.com/cosmos/ibc-go/v8/modules/core/exported"

	routertypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"

	ibchookstypes "github.com/notional-labs/composable/v6/x/ibc-hooks/types"
	ratelimitmoduletypes "github.com/notional-labs/composable/v6/x/ratelimit/types"
	transfermiddlewaretypes "github.com/notional-labs/composable/v6/x/transfermiddleware/types"
	txBoundaryTypes "github.com/notional-labs/composable/v6/x/tx-boundary/types"

	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"

	storetypes "cosmossdk.io/store/types"

	minttypes "github.com/notional-labs/composable/v6/x/mint/types"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasm08types "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
)

// GenerateKeys generates new keys (KV Store, Transient store, and memory store).
func (appKeepers *AppKeepers) GenerateKeys() {
	// Define what keys will be used in the cosmos-sdk key/value store.
	// Cosmos-SDK modules each have a "key" that allows the application to reference what they've stored on the chain.
	appKeepers.keys = storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibchost.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey, icqtypes.StoreKey, capabilitytypes.StoreKey, consensusparamtypes.StoreKey, wasm08types.StoreKey,
		crisistypes.StoreKey, routertypes.StoreKey, transfermiddlewaretypes.StoreKey, group.StoreKey, minttypes.StoreKey, wasm.StoreKey, ibchookstypes.StoreKey, icahosttypes.StoreKey, ratelimitmoduletypes.StoreKey, txBoundaryTypes.StoreKey,
		authzkeeper.StoreKey,
	)

	// Define transient store keys
	appKeepers.tkeys = storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)

	// MemKeys are for information that is stored only in RAM.
	appKeepers.memKeys = sdkstoretypes.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)
}

// GetKVStoreKey gets KV Store keys.
func (appKeepers *AppKeepers) GetKVStoreKey() map[string]*storetypes.KVStoreKey {
	return appKeepers.keys
}

// GetTransientStoreKey gets Transient Store keys.
func (appKeepers *AppKeepers) GetTransientStoreKey() map[string]*storetypes.TransientStoreKey {
	return appKeepers.tkeys
}

// GetMemoryStoreKey get memory Store keys.
func (appKeepers *AppKeepers) GetMemoryStoreKey() map[string]*storetypes.MemoryStoreKey {
	return appKeepers.memKeys
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (appKeepers *AppKeepers) GetKey(storeKey string) *storetypes.KVStoreKey {
	return appKeepers.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (appKeepers *AppKeepers) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return appKeepers.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (appKeepers *AppKeepers) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return appKeepers.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (appKeepers *AppKeepers) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := appKeepers.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}
