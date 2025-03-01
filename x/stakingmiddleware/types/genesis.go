package types

// NewGenesisState creates a new GenesisState object
func NewGenesisState(params Params) *GenesisState {
	return &GenesisState{
		Params: params,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: Params{BlocksPerEpoch: 360, AllowUnbondAfterEpochProgressBlockNumber: 0},
	}
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants holds.
func ValidateGenesis(data GenesisState) error {
	return nil
}
