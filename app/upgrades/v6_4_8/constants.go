package v6_4_8

import (
	store "cosmossdk.io/store/types"
	"github.com/notional-labs/composable/v6/app/upgrades"
	ibctransfermiddleware "github.com/notional-labs/composable/v6/x/ibctransfermiddleware/types"
)

const (
	// UpgradeName defines the on-chain upgrade name for the composable upgrade.
	UpgradeName = "v6_4_8"
)

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateUpgradeHandler,
	StoreUpgrades: store.StoreUpgrades{
		Added:   []string{ibctransfermiddleware.StoreKey},
		Deleted: []string{},
	},
}
