package dex

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
)

// GetTxCmd returns the transaction commands for the dex module.
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        ModuleName,
		Short:                      "DEX transaction commands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		CmdCreatePool(),
		CmdSwap(),
		CmdAddLiquidity(),
		CmdRemoveLiquidity(),
	)
	return txCmd
}

// GetQueryCmd returns the query commands for the dex module.
func GetQueryCmd(cdc *codec.LegacyAmino) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        ModuleName,
		Short:                      "DEX query commands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		CmdQueryPool(),
		CmdQueryPools(),
	)
	return queryCmd
}

// --- Tx commands ---

func CmdCreatePool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-pool [asset-denom] [pnyx-amt] [asset-amt]",
		Short: "Create a new PNYX/<asset> liquidity pool",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			pnyxAmt, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pnyx amount: %w", err)
			}
			assetAmt, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid asset amount: %w", err)
			}
			msg := MsgCreatePool{
				Sender:     clientCtx.GetFromAddress(),
				AssetDenom: args[0],
				PnyxAmt:    pnyxAmt,
				AssetAmt:   assetAmt,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdSwap() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "swap [input-denom] [input-amt] [output-denom]",
		Short: "Swap tokens via the AMM",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			amt, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid input amount: %w", err)
			}
			msg := MsgSwap{
				Sender:      clientCtx.GetFromAddress(),
				InputDenom:  args[0],
				InputAmt:    amt,
				OutputDenom: args[2],
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdAddLiquidity() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-liquidity [asset-denom] [pnyx-amt] [asset-amt]",
		Short: "Add liquidity to an existing pool",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			pnyxAmt, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid pnyx amount: %w", err)
			}
			assetAmt, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid asset amount: %w", err)
			}
			msg := MsgAddLiquidity{
				Sender:     clientCtx.GetFromAddress(),
				AssetDenom: args[0],
				PnyxAmt:    pnyxAmt,
				AssetAmt:   assetAmt,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdRemoveLiquidity() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-liquidity [asset-denom] [shares]",
		Short: "Remove liquidity by burning LP shares",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			shares, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid shares: %w", err)
			}
			msg := MsgRemoveLiquidity{
				Sender:     clientCtx.GetFromAddress(),
				AssetDenom: args[0],
				Shares:     shares,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// --- Query commands ---

func CmdQueryPool() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool [asset-denom]",
		Short: "Query a specific liquidity pool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := NewQueryClient(clientCtx)
			resp, err := queryClient.Pool(cmd.Context(), &QueryPoolRequest{AssetDenom: args[0]})
			if err != nil {
				return err
			}
			return clientCtx.PrintObjectLegacy(json.RawMessage(resp.Result))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryPools() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools",
		Short: "Query all liquidity pools",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := NewQueryClient(clientCtx)
			resp, err := queryClient.Pools(cmd.Context(), &QueryPoolsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintObjectLegacy(json.RawMessage(resp.Result))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
