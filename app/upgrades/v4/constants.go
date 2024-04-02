package v4

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"

	"github.com/ComposableFi/composable-cosmos/v6/app/upgrades"
	ibchookstypes "github.com/ComposableFi/composable-cosmos/v6/x/ibc-hooks/types"
	ratelimitmoduletypes "github.com/ComposableFi/composable-cosmos/v6/x/ratelimit/types"
)

const (
	// UpgradeName defines the on-chain upgrade name for the composable upgrade.
	UpgradeName = "v4"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{wasmtypes.StoreKey, ibchookstypes.StoreKey, ratelimitmoduletypes.StoreKey, icahosttypes.StoreKey},
		Deleted: []string{},
	},
}
