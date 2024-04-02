package cli

import (
	"strconv"

	"github.com/ComposableFi/composable-cosmos/v6/x/tx-boundary/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the tx commands for tx-boundary
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Exp transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		GetCmdUpdateDelegateBoundary(),
		GetCmdQueryRedelegateBoundary(),
	)
	return txCmd
}

func GetCmdUpdateDelegateBoundary() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "update-delegate [txLimit] [blockPerGeneration]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txLimit, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			blockPerGeneration, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateDelegateBoundary(
				types.Boundary{
					TxLimit:             txLimit,
					BlocksPerGeneration: blockPerGeneration,
				},
				clientCtx.GetFromAddress().String(),
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func GetCmdUpdateRedelegateBoundary() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "update-redelegate [txLimit] [blockPerGeneration]",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txLimit, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			blockPerGeneration, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgUpdateRedelegateBoundary(
				types.Boundary{
					TxLimit:             txLimit,
					BlocksPerGeneration: blockPerGeneration,
				},
				clientCtx.GetFromAddress().String(),
			)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
