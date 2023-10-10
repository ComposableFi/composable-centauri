package v5_1_0

import "github.com/notional-labs/composable/v5/app/upgrades"

const (
	// UpgradeName defines the on-chain upgrade name for the Composable v5 upgrade.
	UpgradeName = "v5_1_0"

	// UpgradeHeight defines the block height at which the Composable v6 upgrade is
	// triggered.
	UpgradeHeight = 1536990
)

var Fork = upgrades.Fork{
	UpgradeName:    UpgradeName,
	UpgradeHeight:  UpgradeHeight,
	BeginForkLogic: RunForkLogic,
}
