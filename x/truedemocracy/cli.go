package truedemocracy

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetTxCmd returns the transaction commands for the truedemocracy module.
func GetTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        ModuleName,
		Short:                      "TrueDemocracy transaction commands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		CmdCreateDomain(),
		CmdSubmitProposal(),
		CmdRegisterValidator(),
		CmdWithdrawStake(),
		CmdRemoveValidator(),
		CmdUnjail(),
		CmdJoinPermissionRegister(),
		CmdPurgePermissionRegister(),
		CmdPlaceStoneOnIssue(),
		CmdPlaceStoneOnSuggestion(),
		CmdPlaceStoneOnMember(),
		CmdVoteToExclude(),
		CmdVoteToDelete(),
		CmdRateProposal(),
		CmdCastElectionVote(),
	)
	return txCmd
}

// GetQueryCmd returns the query commands for the truedemocracy module.
func GetQueryCmd(cdc *codec.LegacyAmino) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        ModuleName,
		Short:                      "TrueDemocracy query commands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	queryCmd.AddCommand(
		CmdQueryDomain(cdc),
		CmdQueryDomains(cdc),
		CmdQueryValidator(cdc),
		CmdQueryValidators(cdc),
	)
	return queryCmd
}

// --- Transaction Commands ---

func CmdCreateDomain() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-domain [name] [initial-coins]",
		Short: "Create a new domain",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			coins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}
			msg := MsgCreateDomain{
				Name:         args[0],
				Admin:        clientCtx.GetFromAddress(),
				InitialCoins: coins,
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdSubmitProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-proposal [domain] [issue] [suggestion] [fee] [external-link]",
		Short: "Submit a new proposal (issue + suggestion)",
		Args:  cobra.RangeArgs(4, 5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			fee, err := sdk.ParseCoinsNormalized(args[3])
			if err != nil {
				return err
			}
			link := ""
			if len(args) > 4 {
				link = args[4]
			}
			msg := MsgSubmitProposal{
				Sender:         clientCtx.GetFromAddress(),
				DomainName:     args[0],
				IssueName:      args[1],
				SuggestionName: args[2],
				Creator:        clientCtx.GetFromAddress().String(),
				Fee:            fee,
				ExternalLink:   link,
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdRegisterValidator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register-validator [pubkey-hex] [stake] [domain]",
		Short: "Register a new Proof of Domain validator",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			stake, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}
			msg := MsgRegisterValidator{
				Sender:       clientCtx.GetFromAddress(),
				OperatorAddr: clientCtx.GetFromAddress().String(),
				PubKey:       args[0],
				Stake:        stake,
				DomainName:   args[2],
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdWithdrawStake() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-stake [amount]",
		Short: "Withdraw staked PNYX (subject to 10% transfer limit)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			amount, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid amount: %w", err)
			}
			msg := MsgWithdrawStake{
				Sender:       clientCtx.GetFromAddress(),
				OperatorAddr: clientCtx.GetFromAddress().String(),
				Amount:       amount,
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdRemoveValidator() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-validator [operator-addr]",
		Short: "Remove a validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := MsgRemoveValidator{
				Sender:       clientCtx.GetFromAddress(),
				OperatorAddr: args[0],
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdUnjail() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unjail",
		Short: "Unjail a validator after jail period",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := MsgUnjail{
				Sender:       clientCtx.GetFromAddress(),
				OperatorAddr: clientCtx.GetFromAddress().String(),
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdJoinPermissionRegister() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join-permission-register [domain] [domain-pubkey-hex]",
		Short: "Register a domain public key for anonymous voting",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := MsgJoinPermissionRegister{
				Sender:       clientCtx.GetFromAddress(),
				DomainName:   args[0],
				MemberAddr:   clientCtx.GetFromAddress().String(),
				DomainPubKey: args[1],
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdPurgePermissionRegister() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "purge-permission-register [domain]",
		Short: "Purge domain permission register (admin only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := MsgPurgePermissionRegister{
				Caller:     clientCtx.GetFromAddress(),
				DomainName: args[0],
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdPlaceStoneOnIssue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "place-stone-issue [domain] [issue]",
		Short: "Place a stone on an issue (VoteToEarn reward)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := MsgPlaceStoneOnIssue{
				Sender:     clientCtx.GetFromAddress(),
				DomainName: args[0],
				IssueName:  args[1],
				MemberAddr: clientCtx.GetFromAddress().String(),
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdPlaceStoneOnSuggestion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "place-stone-suggestion [domain] [issue] [suggestion]",
		Short: "Place a stone on a suggestion",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := MsgPlaceStoneOnSuggestion{
				Sender:         clientCtx.GetFromAddress(),
				DomainName:     args[0],
				IssueName:      args[1],
				SuggestionName: args[2],
				MemberAddr:     clientCtx.GetFromAddress().String(),
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdPlaceStoneOnMember() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "place-stone-member [domain] [target-member]",
		Short: "Place a stone on a domain member for admin election",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := MsgPlaceStoneOnMember{
				Sender:       clientCtx.GetFromAddress(),
				DomainName:   args[0],
				TargetMember: args[1],
				VoterAddr:    clientCtx.GetFromAddress().String(),
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdVoteToExclude() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote-exclude [domain] [target-member]",
		Short: "Vote to exclude a member from domain (2/3 majority)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := MsgVoteToExclude{
				Sender:       clientCtx.GetFromAddress(),
				DomainName:   args[0],
				TargetMember: args[1],
				VoterAddr:    clientCtx.GetFromAddress().String(),
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdVoteToDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote-delete [domain] [issue] [suggestion]",
		Short: "Vote to fast-delete a suggestion (2/3 majority)",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := MsgVoteToDelete{
				Sender:         clientCtx.GetFromAddress(),
				DomainName:     args[0],
				IssueName:      args[1],
				SuggestionName: args[2],
				MemberAddr:     clientCtx.GetFromAddress().String(),
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdRateProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rate-proposal [domain] [issue] [suggestion] [rating] [domain-pubkey-hex] [signature-hex]",
		Short: "Rate a suggestion (-5 to +5) using anonymous domain key",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			rating, err := strconv.ParseInt(args[3], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid rating: %w", err)
			}
			msg := MsgRateProposal{
				Sender:         clientCtx.GetFromAddress(),
				DomainName:     args[0],
				IssueName:      args[1],
				SuggestionName: args[2],
				Rating:         int32(rating),
				DomainPubKey:   args[4],
				Signature:      args[5],
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdCastElectionVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cast-election-vote [domain] [issue] [candidate] [choice: 0=approve|1=abstain]",
		Short: "Cast a vote in a person election",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			choice, err := strconv.ParseInt(args[3], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid choice: %w", err)
			}
			msg := MsgCastElectionVote{
				Sender:        clientCtx.GetFromAddress(),
				DomainName:    args[0],
				IssueName:     args[1],
				CandidateName: args[2],
				VoterAddr:     clientCtx.GetFromAddress().String(),
				Choice:        int32(choice),
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// --- Query Commands ---

func CmdQueryDomain(cdc *codec.LegacyAmino) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "domain [name]",
		Short: "Query a domain by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := NewQueryClient(clientCtx)
			resp, err := queryClient.Domain(cmd.Context(), &QueryDomainRequest{Name: args[0]})
			if err != nil {
				return err
			}
			return clientCtx.PrintObjectLegacy(json.RawMessage(resp.Result))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryDomains(cdc *codec.LegacyAmino) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "domains",
		Short: "List all domains",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := NewQueryClient(clientCtx)
			resp, err := queryClient.Domains(cmd.Context(), &QueryDomainsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintObjectLegacy(json.RawMessage(resp.Result))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryValidator(cdc *codec.LegacyAmino) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator [operator-addr]",
		Short: "Query a validator by operator address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := NewQueryClient(clientCtx)
			resp, err := queryClient.Validator(cmd.Context(), &QueryValidatorRequest{OperatorAddr: args[0]})
			if err != nil {
				return err
			}
			return clientCtx.PrintObjectLegacy(json.RawMessage(resp.Result))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryValidators(cdc *codec.LegacyAmino) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validators",
		Short: "List all validators",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := NewQueryClient(clientCtx)
			resp, err := queryClient.Validators(cmd.Context(), &QueryValidatorsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintObjectLegacy(json.RawMessage(resp.Result))
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

