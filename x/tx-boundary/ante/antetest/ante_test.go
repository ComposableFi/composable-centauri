package antetest

import (
	"testing"

	txboundaryAnte "github.com/ComposableFi/composable-cosmos/v6/x/tx-boundary/ante"
	"github.com/ComposableFi/composable-cosmos/v6/x/tx-boundary/types"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, new(AnteTestSuite))
}

func (s *AnteTestSuite) TestStakingAnteBasic() {
	_, _, addr1 := testdata.KeyTestPubAddr()
	delegateMsg := stakingtypes.NewMsgDelegate(s.delegator, s.validators[0].GetOperator(), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000)))
	msgDelegateAny, err := cdctypes.NewAnyWithValue(delegateMsg)
	require.NoError(s.T(), err)

	addr2 := s.delegator

	for _, tc := range []struct {
		desc      string
		txMsg     sdk.Msg
		malleate  func() error
		expErr    bool
		expErrStr string
	}{
		{
			desc:  "Case delegate success",
			txMsg: stakingtypes.NewMsgDelegate(s.delegator, s.validators[0].GetOperator(), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000))),
			malleate: func() error {
				return nil
			},
			expErr: false,
		},
		{
			desc:  "Case redelegate success",
			txMsg: stakingtypes.NewMsgBeginRedelegate(s.delegator, s.validators[0].GetOperator(), s.newvalidators[0].GetOperator(), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000))),
			malleate: func() error {
				return nil
			},
			expErr: false,
		},
		{
			desc:  "Case authz success",
			txMsg: &authz.MsgExec{Grantee: addr1.String(), Msgs: []*cdctypes.Any{msgDelegateAny}},
			malleate: func() error {
				return nil
			},
			expErr: false,
		},
		{
			desc:  "Case delegate failed",
			txMsg: stakingtypes.NewMsgDelegate(s.delegator, s.validators[0].GetOperator(), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000))),
			malleate: func() error {
				s.app.TxBoundaryKeepper.SetLimitPerAddr(s.ctx, addr2, types.LimitPerAddr{
					DelegateCount:     5,
					ReledegateCount:   5,
					LatestUpdateBlock: s.ctx.BlockHeight(),
				})
				return nil
			},
			expErr: true,
		},
		{
			desc:  "Case redelegate failed",
			txMsg: stakingtypes.NewMsgBeginRedelegate(s.delegator, s.validators[0].GetOperator(), s.newvalidators[0].GetOperator(), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000))),
			malleate: func() error {
				s.app.TxBoundaryKeepper.SetLimitPerAddr(s.ctx, addr2, types.LimitPerAddr{
					DelegateCount:     5,
					ReledegateCount:   5,
					LatestUpdateBlock: s.ctx.BlockHeight(),
				})
				return nil
			},
			expErr: true,
		},
		{
			desc:  "Case authz failed",
			txMsg: &authz.MsgExec{Grantee: addr1.String(), Msgs: []*cdctypes.Any{msgDelegateAny}},
			malleate: func() error {
				s.app.TxBoundaryKeepper.SetLimitPerAddr(s.ctx, addr2, types.LimitPerAddr{
					DelegateCount:     5,
					ReledegateCount:   5,
					LatestUpdateBlock: s.ctx.BlockHeight(),
				})
				return nil
			},
			expErr: true,
		},
	} {
		tc := tc
		s.SetupTest()
		tc.malleate()
		s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()
		priv1, _, _ := testdata.KeyTestPubAddr()
		privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}

		mfd := txboundaryAnte.NewStakingPermissionDecorator(s.app.AppCodec(), s.app.TxBoundaryKeepper)
		antehandler := sdk.ChainAnteDecorators(mfd)
		s.Require().NoError(s.txBuilder.SetMsgs(tc.txMsg))

		tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
		s.Require().NoError(err)
		_, err = antehandler(s.ctx, tx, false)
		if !tc.expErr {
			s.Require().NoError(err)
		} else {
			s.Require().Error(err)
		}
	}
}

func (s *AnteTestSuite) TestStakingAnteUpdateLimit() {
	_, _, addr1 := testdata.KeyTestPubAddr()
	delegateMsg := stakingtypes.NewMsgDelegate(s.delegator, s.validators[0].GetOperator(), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000)))

	msgDelegateAny, err := cdctypes.NewAnyWithValue(delegateMsg)
	require.NoError(s.T(), err)

	addr2 := s.delegator

	for _, tc := range []struct {
		desc        string
		txMsg       sdk.Msg
		malleate    func() error
		blocksAdded int64
		expErr      bool
		expErrStr   string
	}{
		{
			desc:  "Case delegate success update limit",
			txMsg: stakingtypes.NewMsgDelegate(s.delegator, s.validators[0].GetOperator(), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000))),
			malleate: func() error {
				s.app.TxBoundaryKeepper.SetLimitPerAddr(s.ctx, addr2, types.LimitPerAddr{
					DelegateCount:     5,
					ReledegateCount:   5,
					LatestUpdateBlock: s.ctx.BlockHeight(),
				})
				return nil
			},
			blocksAdded: 5,
			expErr:      false,
		},
		{
			desc:  "Case redelegate success update limit",
			txMsg: stakingtypes.NewMsgBeginRedelegate(s.delegator, s.validators[0].GetOperator(), s.newvalidators[0].GetOperator(), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000))),
			malleate: func() error {
				s.app.TxBoundaryKeepper.SetLimitPerAddr(s.ctx, addr2, types.LimitPerAddr{
					DelegateCount:     5,
					ReledegateCount:   5,
					LatestUpdateBlock: s.ctx.BlockHeight(),
				})
				return nil
			},
			blocksAdded: 5,
			expErr:      false,
		},
		{
			desc:  "Case authz success update limit",
			txMsg: &authz.MsgExec{Grantee: addr1.String(), Msgs: []*cdctypes.Any{msgDelegateAny}},
			malleate: func() error {
				s.app.TxBoundaryKeepper.SetLimitPerAddr(s.ctx, addr2, types.LimitPerAddr{
					DelegateCount:     5,
					ReledegateCount:   5,
					LatestUpdateBlock: s.ctx.BlockHeight(),
				})
				return nil
			},
			blocksAdded: 5,
			expErr:      false,
		},
		{
			desc:  "Case delegate fail update limit",
			txMsg: stakingtypes.NewMsgDelegate(s.delegator, s.validators[0].GetOperator(), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000))),
			malleate: func() error {
				s.app.TxBoundaryKeepper.SetLimitPerAddr(s.ctx, addr2, types.LimitPerAddr{
					DelegateCount:     5,
					ReledegateCount:   5,
					LatestUpdateBlock: s.ctx.BlockHeight(),
				})
				return nil
			},
			blocksAdded: 4,
			expErr:      true,
		},
		{
			desc:  "Case redelegate fail update limit",
			txMsg: stakingtypes.NewMsgBeginRedelegate(s.delegator, s.validators[0].GetOperator(), s.newvalidators[0].GetOperator(), sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000000))),
			malleate: func() error {
				s.app.TxBoundaryKeepper.SetLimitPerAddr(s.ctx, addr2, types.LimitPerAddr{
					DelegateCount:     5,
					ReledegateCount:   5,
					LatestUpdateBlock: s.ctx.BlockHeight(),
				})
				return nil
			},
			blocksAdded: 4,
			expErr:      true,
		},
		{
			desc:  "Case authz success update limit",
			txMsg: &authz.MsgExec{Grantee: addr1.String(), Msgs: []*cdctypes.Any{msgDelegateAny}},
			malleate: func() error {
				s.app.TxBoundaryKeepper.SetLimitPerAddr(s.ctx, addr2, types.LimitPerAddr{
					DelegateCount:     5,
					ReledegateCount:   5,
					LatestUpdateBlock: s.ctx.BlockHeight(),
				})
				return nil
			},
			blocksAdded: 4,
			expErr:      true,
		},
	} {
		tc := tc
		s.SetupTest()
		tc.malleate()
		s.txBuilder = s.clientCtx.TxConfig.NewTxBuilder()
		priv1, _, _ := testdata.KeyTestPubAddr()
		privs, accNums, accSeqs := []cryptotypes.PrivKey{priv1}, []uint64{0}, []uint64{0}

		mfd := txboundaryAnte.NewStakingPermissionDecorator(s.app.AppCodec(), s.app.TxBoundaryKeepper)
		antehandler := sdk.ChainAnteDecorators(mfd)
		s.Require().NoError(s.txBuilder.SetMsgs(tc.txMsg))

		tx, err := s.CreateTestTx(privs, accNums, accSeqs, s.ctx.ChainID())
		s.Require().NoError(err)
		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + tc.blocksAdded)
		_, err = antehandler(s.ctx, tx, false)
		if !tc.expErr {
			s.Require().NoError(err)
		} else {
			s.Require().Error(err)
		}
	}
}
