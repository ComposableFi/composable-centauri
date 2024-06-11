package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/notional-labs/composable/v6/app"
	"github.com/notional-labs/composable/v6/app/helpers"
	"github.com/notional-labs/composable/v6/x/tx-boundary/types"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx sdk.Context
	// querier sdk.Querier
	app *app.ComposableApp
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.app = helpers.SetupComposableAppWithValSet(suite.T())
	suite.ctx = suite.app.BaseApp.NewContext(false)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

var (
	defaultBoundary = types.Boundary{
		TxLimit:             5,
		BlocksPerGeneration: 5,
	}
	newBoundary = types.Boundary{
		TxLimit:             10,
		BlocksPerGeneration: 5,
	}
	failBoundary = types.Boundary{
		TxLimit:             10,
		BlocksPerGeneration: 0,
	}
)

func (suite *KeeperTestSuite) TestSetDelegateBoundary() {
	for _, tc := range []struct {
		desc             string
		expectedBoundary types.Boundary
		malleate         func() error
		shouldErr        bool
		expectedErr      string
	}{
		{
			desc:             "Case success",
			expectedBoundary: newBoundary,
			malleate: func() error {
				return suite.app.TxBoundaryKeepper.SetDelegateBoundary(suite.ctx, newBoundary)
			},
			shouldErr: false,
		},
		{
			desc:             "Case fail",
			expectedBoundary: failBoundary,
			malleate: func() error {
				return suite.app.TxBoundaryKeepper.SetDelegateBoundary(suite.ctx, failBoundary)
			},
			shouldErr:   true,
			expectedErr: "BlocksPerGeneration must not be zero",
		},
		{
			desc:             "Do no thing",
			expectedBoundary: defaultBoundary,
			malleate: func() error {
				return nil
			},
			shouldErr: false,
		},
	} {
		tc := tc
		suite.Run(tc.desc, func() {
			suite.SetupTest()
			err := tc.malleate()
			if !tc.shouldErr {
				res := suite.app.TxBoundaryKeepper.GetDelegateBoundary(suite.ctx)
				suite.Equal(res, tc.expectedBoundary)
			} else {
				suite.Equal(err.Error(), tc.expectedErr)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestSetRedelegateBoundary() {
	suite.app.TxBoundaryKeepper.SetRedelegateBoundary(suite.ctx, types.Boundary{
		TxLimit:             10,
		BlocksPerGeneration: 5,
	})

	for _, tc := range []struct {
		desc             string
		expectedBoundary types.Boundary
		malleate         func() error
		shouldErr        bool
		expectedErr      string
	}{
		{
			desc:             "Case success",
			expectedBoundary: newBoundary,
			malleate: func() error {
				return suite.app.TxBoundaryKeepper.SetRedelegateBoundary(suite.ctx, newBoundary)
			},
			shouldErr: false,
		},
		{
			desc:             "Success",
			expectedBoundary: failBoundary,
			malleate: func() error {
				return suite.app.TxBoundaryKeepper.SetRedelegateBoundary(suite.ctx, failBoundary)
			},
			shouldErr:   true,
			expectedErr: "BlocksPerGeneration must not be zero",
		},
		{
			desc:             "Do no thing",
			expectedBoundary: defaultBoundary,
			malleate: func() error {
				return nil
			},
			shouldErr: false,
		},
	} {
		tc := tc
		suite.Run(tc.desc, func() {
			suite.SetupTest()
			err := tc.malleate()
			if !tc.shouldErr {
				res := suite.app.TxBoundaryKeepper.GetRedelegateBoundary(suite.ctx)
				suite.Equal(res, tc.expectedBoundary)
			} else {
				suite.Equal(err.Error(), tc.expectedErr)
			}
		})
	}
}
