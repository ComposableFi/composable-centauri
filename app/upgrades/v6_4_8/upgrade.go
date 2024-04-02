package v6_4_8

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/ComposableFi/composable-cosmos/v6/app/keepers"
	"github.com/ComposableFi/composable-cosmos/v6/app/upgrades"
	ibctransfermiddleware "github.com/ComposableFi/composable-cosmos/v6/x/ibctransfermiddleware/types"
	"github.com/cosmos/cosmos-sdk/codec"
)

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	_ upgrades.BaseAppParamManager,
	_ codec.Codec,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, plan upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {
		custommiddlewareparams := ibctransfermiddleware.DefaultGenesisState()
		keepers.IbcTransferMiddlewareKeeper.SetParams(ctx, custommiddlewareparams.Params)

		// remove broken proposals
		BrokenProposals := [3]uint64{2, 6, 11}
		for _, proposal_id := range BrokenProposals {
			_, ok := keepers.GovKeeper.GetProposal(ctx, proposal_id)
			if ok {
				keepers.GovKeeper.DeleteProposal(ctx, proposal_id)
			}

		}

		// burn extra ppica in escrow account
		// this ppica is unused because it is a native token stored in escrow account
		// it was unnecessarily minted to match pica escrowed on picasso to ppica minted
		// in genesis, to make initial native ppica transferrable to picasso
		amount, ok := sdk.NewIntFromString("1066669217167120000000")
		if ok {
			coins := sdk.Coins{sdk.NewCoin("ppica", amount)}
			keepers.BankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromBech32("centauri12k2pyuylm9t7ugdvz67h9pg4gmmvhn5vmvgw48"), "gov", coins)
			keepers.BankKeeper.BurnCoins(ctx, "gov", coins)
		}
		return mm.RunMigrations(ctx, configurator, vm)
	}
}
