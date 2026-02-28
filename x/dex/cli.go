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
		CmdRegisterAsset(),
		CmdUpdateAssetStatus(),
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
		CmdQueryRegisteredAssets(),
		CmdQueryAsset(),
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

// --- Asset registry tx commands ---

func CmdRegisterAsset() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-asset [ibc-denom] [symbol] [name] [decimals] [origin-chain] [ibc-channel]",
		Short: "Register a new IBC asset for DEX trading",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			decimals, err := strconv.ParseUint(args[3], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid decimals: %w", err)
			}
			msg := MsgRegisterAsset{
				Sender:      clientCtx.GetFromAddress(),
				IBCDenom:    args[0],
				Symbol:      args[1],
				Name:        args[2],
				Decimals:    uint32(decimals),
				OriginChain: args[4],
				IBCChannel:  args[5],
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdUpdateAssetStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-asset-status [ibc-denom] [enabled]",
		Short: "Enable or disable trading for a registered asset",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			enabled, err := strconv.ParseBool(args[1])
			if err != nil {
				return fmt.Errorf("invalid enabled value (use true/false): %w", err)
			}
			msg := MsgUpdateAssetStatus{
				Sender:   clientCtx.GetFromAddress(),
				IBCDenom: args[0],
				Enabled:  enabled,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// --- Asset registry query commands ---

func CmdQueryRegisteredAssets() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registered-assets",
		Short: "Query all registered trading assets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := NewQueryClient(clientCtx)
			resp, err := queryClient.RegisteredAssets(cmd.Context(), &QueryRegisteredAssetsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintObjectLegacy(json.RawMessage(resp.Result))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryAsset() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset [denom-or-symbol]",
		Short: "Query a registered asset by IBC denom or symbol",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := NewQueryClient(clientCtx)

			// Try by denom first, then by symbol.
			resp, err := queryClient.AssetByDenom(cmd.Context(), &QueryAssetByDenomRequest{IBCDenom: args[0]})
			if err == nil {
				return clientCtx.PrintObjectLegacy(json.RawMessage(resp.Result))
			}

			respSym, err := queryClient.AssetBySymbol(cmd.Context(), &QueryAssetBySymbolRequest{Symbol: args[0]})
			if err != nil {
				return fmt.Errorf("asset not found by denom or symbol: %s", args[0])
			}
			return clientCtx.PrintObjectLegacy(json.RawMessage(respSym.Result))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
