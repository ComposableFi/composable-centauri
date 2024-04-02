package v4_5_1

import "github.com/ComposableFi/composable-cosmos/v6/app/upgrades"

const (
	// UpgradeName defines the on-chain upgrade name for the Composable v5 upgrade.
	UpgradeName = "v4_5_1"

	// UpgradeHeight defines the block height at which the Composable v6 upgrade is
	// triggered.
	UpgradeHeight = 1127000
)

var Fork = upgrades.Fork{
	UpgradeName:    UpgradeName,
	UpgradeHeight:  UpgradeHeight,
	BeginForkLogic: RunForkLogic,
}
