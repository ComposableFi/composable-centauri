package v45

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/ComposableFi/composable-cosmos/v6/app/keepers"
)

func RunForkLogic(ctx sdk.Context, appKeepers *keepers.AppKeepers) {
	for i := 0; i < 100; i++ {
		fmt.Println("Switching to v4_5 code")
	}

	// Specifying the whole list instead of adding and removing. Less fragile.
	hostParams := icahosttypes.Params{
		HostEnabled: true,
		AllowMessages: []string{
			// Change: Normal Msg
			sdk.MsgTypeURL(&banktypes.MsgSend{}),
			sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
			sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
			sdk.MsgTypeURL(&stakingtypes.MsgCreateValidator{}),
			sdk.MsgTypeURL(&stakingtypes.MsgEditValidator{}),
			sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}),
			sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}),
			sdk.MsgTypeURL(&distrtypes.MsgWithdrawValidatorCommission{}),
			sdk.MsgTypeURL(&distrtypes.MsgFundCommunityPool{}),
			sdk.MsgTypeURL(&govtypes.MsgVote{}),
			sdk.MsgTypeURL(&authz.MsgExec{}),
			sdk.MsgTypeURL(&authz.MsgGrant{}),
			sdk.MsgTypeURL(&authz.MsgRevoke{}),

			// Change: Added MsgTrasnsfer
			sdk.MsgTypeURL(&ibctransfertypes.MsgTransfer{}),
			sdk.MsgTypeURL(&banktypes.MsgSend{}),
			sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
			sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
			sdk.MsgTypeURL(&stakingtypes.MsgCreateValidator{}),
			sdk.MsgTypeURL(&stakingtypes.MsgEditValidator{}),

			// Change: Added MsgUndelegate
			sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
			sdk.MsgTypeURL(&distrtypes.MsgWithdrawDelegatorReward{}),
			sdk.MsgTypeURL(&distrtypes.MsgSetWithdrawAddress{}),
			sdk.MsgTypeURL(&distrtypes.MsgWithdrawValidatorCommission{}),
			sdk.MsgTypeURL(&distrtypes.MsgFundCommunityPool{}),
			sdk.MsgTypeURL(&govtypes.MsgVote{}),
		},
	}
	appKeepers.ICAHostKeeper.SetParams(ctx, hostParams)
}
