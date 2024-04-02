package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ComposableFi/composable-cosmos/v6/x/tx-boundary/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// GetQueryCmd returns the cli query commands for the tx-boundary module.
func GetQueryCmd() *cobra.Command {
	txboundaryQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the tx-boundary module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txboundaryQueryCmd.AddCommand(
		GetCmdQueryDelegateBoundary(),
		GetCmdQueryRedelegateBoundary(),
	)

	return txboundaryQueryCmd
}

// GetCmdQueryDelegateBoundary implements a command to return the current delegate boundary value
func GetCmdQueryDelegateBoundary() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegate-boundary",
		Short: "Query the current DelegateBoundary value",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryDelegateBoundaryRequest{}
			res, err := queryClient.DelegateBoundary(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%s\n", &res.Boundary))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryDelegateBoundary implements a command to return the current delegate boundary value
func GetCmdQueryRedelegateBoundary() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redelegate-boundary",
		Short: "Query the current RedelegateBoundary value",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryRedelegateBoundaryRequest{}
			res, err := queryClient.RedelegateBoundary(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintString(fmt.Sprintf("%s\n", &res.Boundary))
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
