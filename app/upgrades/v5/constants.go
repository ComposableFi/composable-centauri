package v5

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/notional-labs/centauri/v5/app/upgrades"
	txboundary "github.com/notional-labs/centauri/v5/x/tx-boundary/types"
)

const (
	// UpgradeName defines the on-chain upgrade name for the Centauri upgrade.
	UpgradeName = "v5"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added: []string{txboundary.ModuleName},
	},
}
