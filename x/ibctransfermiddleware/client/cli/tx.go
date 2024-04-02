package cli

import (
	"fmt"
	"strconv"

	"github.com/ComposableFi/composable-cosmos/v6/x/ibctransfermiddleware/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the tx commands for staking middleware module.
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Exp transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		AddIBCFeeConfig(),
		RemoveIBCFeeConfig(),
		AddAllowedIbcToken(),
		RemoveAllowedIbcToken(),
	)

	return txCmd
}

func AddIBCFeeConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-config [channel] [feeAddress] [minTimeoutTimestamp]",
		Short:   "add ibc fee config",
		Args:    cobra.MatchAll(cobra.ExactArgs(3), cobra.OnlyValidArgs),
		Example: fmt.Sprintf("%s tx ibctransfermiddleware add-config [channel] [feeAddress] [minTimeoutTimestamp]", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			channel := args[0]
			feeAddress := args[1]
			minTimeoutTimestamp := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			fromAddress := clientCtx.GetFromAddress().String()

			minTimeoutTimestampInt, err := strconv.ParseInt(minTimeoutTimestamp, 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgAddIBCFeeConfig(
				fromAddress,
				channel,
				feeAddress,
				minTimeoutTimestampInt,
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

func AddAllowedIbcToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-allowed-ibc-token [channel] [denom] [amount] [percentage]",
		Short:   "add allowed ibc token",
		Args:    cobra.MatchAll(cobra.ExactArgs(4), cobra.OnlyValidArgs),
		Example: fmt.Sprintf("%s tx ibctransfermiddleware add-allowed-ibc-token [channel] [denom] [amount] [percentage]  (percentage '5' means 1/5 of amount will be taken as fee) ", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			channel := args[0]
			denom := args[1]
			amount := args[2]
			percentage := args[3]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			fromAddress := clientCtx.GetFromAddress().String()

			amountInt, err := strconv.ParseInt(amount, 10, 64)
			if err != nil {
				return err
			}

			percentageInt, errPercentage := strconv.ParseInt(percentage, 10, 64)
			if errPercentage != nil {
				return errPercentage
			}

			msg := types.NewMsgAddAllowedIbcToken(
				fromAddress,
				channel,
				denom,
				amountInt,
				percentageInt,
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

func RemoveIBCFeeConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-config [channel]",
		Short:   "remove ibc fee config",
		Args:    cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Example: fmt.Sprintf("%s tx ibctransfermiddleware remove-config [channel]", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			channel := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			fromAddress := clientCtx.GetFromAddress().String()

			msg := types.NewMsgRemoveIBCFeeConfig(
				fromAddress,
				channel,
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

func RemoveAllowedIbcToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-allowed-ibc-token [channel] [denom]",
		Short:   "remove allowed ibc token",
		Args:    cobra.MatchAll(cobra.ExactArgs(2), cobra.OnlyValidArgs),
		Example: fmt.Sprintf("%s tx ibctransfermiddleware remove-allowed-ibc-token [channel] [denom]", version.AppName),
		RunE: func(cmd *cobra.Command, args []string) error {
			channel := args[0]
			denom := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			fromAddress := clientCtx.GetFromAddress().String()

			msg := types.NewMsgRemoveAllowedIbcToken(
				fromAddress,
				channel,
				denom,
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
