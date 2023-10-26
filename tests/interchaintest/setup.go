package interchaintest

import (
	"os"
	"strings"

	"github.com/strangelove-ventures/interchaintest/v7/ibc"
)

var (
	CentauriMainRepo   = "ghcr.io/notional-labs/composable"
	CentauriICTestRepo = "ghcr.io/notional-labs/composable-ictest"

	repo, version = GetDockerImageInfo()

	CentauriImage = ibc.DockerImage{
		Repository: repo,
		Version:    version,
		UidGid:     "1025:1025",
	}

	composableConfig = ibc.ChainConfig{
		Type:                "cosmos",
		Name:                "composable",
		ChainID:             "composable-2",
		Images:              []ibc.DockerImage{CentauriImage},
		Bin:                 "composabled",
		Bech32Prefix:        "composable",
		Denom:               "stake",
		CoinType:            "118",
		GasPrices:           "0.0stake",
		GasAdjustment:       1.1,
		TrustingPeriod:      "112h",
		NoHostMount:         false,
		ModifyGenesis:       nil,
		ConfigFileOverrides: nil,
	}
)

// GetDockerImageInfo returns the appropriate repo and branch version string for integration with the CI pipeline.
// The remote runner sets the BRANCH_CI env var. If present, interchaintest will use the docker image pushed up to the repo.
// If testing locally, user should run `make docker-build-debug` and interchaintest will use the local image.
func GetDockerImageInfo() (repo, version string) {
	branchVersion, found := os.LookupEnv("BRANCH_CI")
	repo = CentauriICTestRepo
	if !found {
		// make local-image
		repo = "composable"
		branchVersion = "debug"
	}

	// github converts / to - for pushed docker images
	branchVersion = strings.ReplaceAll(branchVersion, "/", "-")
	return repo, branchVersion
}
