package healthcheck

import (
	"context"
	"time"

	"github.com/spf13/cobra"
)

// NewCommand builds the `healthcheck` command group with the `live` and
// `ready` subcommands. Each probe prints exactly its own name followed by a
// newline on success and nothing on failure; failures surface only through
// the returned, operator-safe error.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "healthcheck",
		Short: "Probe the local node RPC for liveness or readiness",
	}
	cmd.AddCommand(
		newProbeCommand("live", "Liveness: local RPC answers /health with a valid result", Live),
		newProbeCommand("ready", "Readiness: RPC status shows a positive height and catching_up=false", Ready),
	)
	return cmd
}

func newProbeCommand(name, short string, probe func(context.Context, string, time.Duration) error) *cobra.Command {
	var (
		rpcURL  string
		timeout time.Duration
	)
	cmd := &cobra.Command{
		Use:           name,
		Short:         short,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := probe(cmd.Context(), rpcURL, timeout); err != nil {
				return err
			}
			cmd.Println(name)
			return nil
		},
	}
	cmd.Flags().StringVar(&rpcURL, "rpc-url", DefaultRPCURL, "local CometBFT RPC base URL (plain http, literal loopback only)")
	cmd.Flags().DurationVar(&timeout, "timeout", DefaultTimeout, "probe timeout (must be positive and at most 10s)")
	return cmd
}
