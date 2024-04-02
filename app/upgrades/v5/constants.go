package v5

import (
	"github.com/ComposableFi/composable-cosmos/v6/app/upgrades"
	txboundary "github.com/ComposableFi/composable-cosmos/v6/x/tx-boundary/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
)

const (
	// UpgradeName defines the on-chain upgrade name for the composable upgrade.
	UpgradeName = "v5"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{txboundary.ModuleName},
	},
}
