package ibctesting

import (
	"testing"

	tmtypes "github.com/cometbft/cometbft/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/notional-labs/composable/v6/app"
)

// DefaultComposableAppFactory instantiates and sets up the default Composable App
func DefaultComposableAppFactory(t *testing.T, valSet *tmtypes.ValidatorSet, genAccs []authtypes.GenesisAccount, chainID string, balances ...banktypes.Balance) ChainApp {
	t.Helper()
	return app.SetupWithGenesisValSet(t, valSet, genAccs, chainID, balances...)
}
