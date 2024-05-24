package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/notional-labs/composable/v6/x/ibctransfermiddleware/types"
)

// GetQueryCmd returns the cli query commands for the staking middleware module.
func GetQueryCmd() *cobra.Command {
	ibctransfermiddlewareParamsQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the staking middleware module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	ibctransfermiddlewareParamsQueryCmd.AddCommand(
		GetCmdQueryParams(),
		GetFeeConfigByChannelAndDenom(),
	)

	return ibctransfermiddlewareParamsQueryCmd
}

// GetCmdQueryParams implements a command to return the current staking middleware's params
// parameters.
func GetCmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the current ibc middleware parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryParamsRequest{}
			res, err := queryClient.Params(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func GetFeeConfigByChannelAndDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "FeeConfigByChannelAndDenom",
		Short:   "Query bridge fee config by channel and denom",
		Args:    cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs),
		Example: fmt.Sprintf("%s query ibctransfermiddleware FeeConfigByChannelAndDenom [channel] [denom]", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryFeeConfigByChannelAndDenomRequest{
				Channel: args[0],
				Denom:   args[1],
			}
			res, err := queryClient.FeeConfigByChannelAndDenom(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Fees)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
