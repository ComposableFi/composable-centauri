package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	helpers "github.com/ComposableFi/composable-cosmos/v6/app/helpers"
	"github.com/ComposableFi/composable-cosmos/v6/x/transfermiddleware/types"
)

func TestTFMInitGenesis(t *testing.T) {
	app := helpers.SetupComposableAppWithValSet(t)
	ctx := helpers.NewContextForApp(*app)

	tokenInfos := make([]types.ParachainIBCTokenInfo, 1)
	tokenInfos[0] = types.ParachainIBCTokenInfo{
		IbcDenom:    "ibc-test",
		ChannelID:   "channel-0",
		NativeDenom: "pica",
		AssetId:     "1",
	}

	app.TransferMiddlewareKeeper.InitGenesis(ctx, types.GenesisState{
		TokenInfos: tokenInfos,
	})

	info := app.TransferMiddlewareKeeper.GetParachainIBCTokenInfoByNativeDenom(ctx, "pica")
	require.Equal(t, info, app.TransferMiddlewareKeeper.GetParachainIBCTokenInfoByNativeDenom(ctx, "pica"))
	require.Equal(t, "1", info.AssetId)
	require.Equal(t, "pica", info.NativeDenom)
	require.Equal(t, "ibc-test", info.IbcDenom)
	require.Equal(t, "channel-0", info.ChannelID)
}

func TestTFMExportGenesis(t *testing.T) {
	app := helpers.SetupComposableAppWithValSet(t)
	ctx := helpers.NewContextForApp(*app)

	err := app.TransferMiddlewareKeeper.AddParachainIBCInfo(ctx, "ibc-test", "channel-0", "pica", "1")
	require.NoError(t, err)
	err = app.TransferMiddlewareKeeper.AddParachainIBCInfo(ctx, "ibc-test2", "channel-1", "poke", "2")
	require.NoError(t, err)
	genesis := app.TransferMiddlewareKeeper.ExportGenesis(ctx)

	require.Equal(t, "1", genesis.TokenInfos[0].AssetId)
	require.Equal(t, "pica", genesis.TokenInfos[0].NativeDenom)
	require.Equal(t, "channel-0", genesis.TokenInfos[0].ChannelID)
	require.Equal(t, "ibc-test", genesis.TokenInfos[0].IbcDenom)

	require.Equal(t, "2", genesis.TokenInfos[1].AssetId)
	require.Equal(t, "poke", genesis.TokenInfos[1].NativeDenom)
	require.Equal(t, "channel-1", genesis.TokenInfos[1].ChannelID)
	require.Equal(t, "ibc-test2", genesis.TokenInfos[1].IbcDenom)
}

func TestIterateParaTokenInfos(t *testing.T) {
	app := helpers.SetupComposableAppWithValSet(t)
	ctx := helpers.NewContextForApp(*app)

	err := app.TransferMiddlewareKeeper.AddParachainIBCInfo(ctx, "ibc-test", "channel-0", "pica", "1")
	require.NoError(t, err)
	err = app.TransferMiddlewareKeeper.AddParachainIBCInfo(ctx, "ibc-test2", "channel-1", "poke", "2")
	require.NoError(t, err)

	infos := []types.ParachainIBCTokenInfo{}

	app.TransferMiddlewareKeeper.IterateParaTokenInfos(ctx, func(index int64, info types.ParachainIBCTokenInfo) (stop bool) {
		infos = append(infos, info)
		return false
	})

	require.Equal(t, "1", infos[0].AssetId)
	require.Equal(t, "pica", infos[0].NativeDenom)
	require.Equal(t, "channel-0", infos[0].ChannelID)
	require.Equal(t, "ibc-test", infos[0].IbcDenom)

	require.Equal(t, "2", infos[1].AssetId)
	require.Equal(t, "poke", infos[1].NativeDenom)
	require.Equal(t, "channel-1", infos[1].ChannelID)
	require.Equal(t, "ibc-test2", infos[1].IbcDenom)
}
