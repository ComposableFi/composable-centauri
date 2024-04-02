package v6_4_8

import (
	"github.com/ComposableFi/composable-cosmos/v6/app/upgrades"
	ibctransfermiddleware "github.com/ComposableFi/composable-cosmos/v6/x/ibctransfermiddleware/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
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
