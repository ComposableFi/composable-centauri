package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/ComposableFi/composable-cosmos/v6/x/transfermiddleware/types"
)

// GetTxCmd returns the tx commands for router
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		Short:                      "Registry and remove IBC dotsama chain information",
		Long:                       "Registry and remove IBC dotsama chain information",
	}

	txCmd.AddCommand(
		RegistryDotSamaChain(),
		RemoveDotSamaChain(),
		AddRlyAddress(),
	)

	return txCmd
}

// RegistryDotSamaChain returns the command handler for registry dotsame token info.
func RegistryDotSamaChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "registry",
		Short:   "registry dotsama chain information",
		Long:    "registry dotsama chain information",
		Args:    cobra.MatchAll(cobra.ExactArgs(4), cobra.OnlyValidArgs),
		Example: fmt.Sprintf("%s tx transfermiddleware registry [ibc_denom] [native_denom] [asset_id] [channel_id]", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			ibcDenom := args[0]
			nativeDenom := args[1]
			assetID := args[2]
			channelID := args[3]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			fromAddress := clientCtx.GetFromAddress().String()

			msg := types.NewMsgAddParachainIBCTokenInfo(
				fromAddress,
				ibcDenom,
				nativeDenom,
				assetID,
				channelID,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func RemoveDotSamaChain() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove",
		Short:   "remove dotsama chain information",
		Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Example: fmt.Sprintf("%s tx transfermiddleware remove [native_denom]", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			nativeDenom := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			fromAddress := clientCtx.GetFromAddress().String()

			msg := types.NewMsgRemoveParachainIBCTokenInfo(
				fromAddress,
				nativeDenom,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func AddRlyAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-rly [addr ]",
		Short:   "add address to whitelist relayer",
		Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Example: fmt.Sprintf("%s tx transfermiddleware add-rly [allowed_addr]", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			allowedAddress := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			fromAddress := clientCtx.GetFromAddress().String()

			msg := types.NewMsgAddRlyAddress(
				fromAddress,
				allowedAddress,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
