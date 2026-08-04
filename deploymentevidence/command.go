package deploymentevidence

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// NewCommand builds the offline deployment-evidence command group. It reads
// only the two explicit files named by flags: no DNS, network, environment,
// or home-directory access.
func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "deployment-evidence",
		Short: "Verify a secret-free offline deployment-evidence manifest",
	}
	command.AddCommand(newVerifyCommand())
	return command
}

func newVerifyCommand() *cobra.Command {
	var contractFile, manifestFile, at, output string
	command := &cobra.Command{
		Use: "verify --contract <topology.json> --manifest <manifest.json>", Args: noArgs,
		Short:        "Fail closed on unbound, stale, or incomplete deployment evidence",
		SilenceUsage: true, SilenceErrors: true,
		RunE: func(command *cobra.Command, _ []string) error {
			if contractFile == "" || manifestFile == "" {
				return fmt.Errorf("--contract and --manifest are required")
			}
			evaluation := time.Now().UTC()
			if at != "" {
				parsed, ok := parseStrictTimestamp(at)
				if !ok {
					return fmt.Errorf("invalid --at: expected a strict UTC timestamp")
				}
				evaluation = time.Unix(parsed, 0).UTC()
			}
			topology, err := LoadTopology(contractFile)
			if err != nil {
				return err
			}
			manifest, err := LoadManifest(manifestFile)
			if err != nil {
				return err
			}
			report := Verify(manifest, topology, evaluation)
			switch output {
			case "json":
				encoded, err := json.MarshalIndent(report, "", "  ")
				if err != nil {
					return fmt.Errorf("encode deployment evidence report")
				}
				command.Println(string(encoded))
			case "text":
				if report.Valid {
					command.Printf("OK: deployment evidence validates %d gates across %d nodes\n",
						report.GateCount, report.NodeCount)
				} else {
					for _, violation := range report.Violations {
						command.Printf("VIOLATION %s: %s\n", violation.Check, violation.Message)
					}
				}
			default:
				return fmt.Errorf("unknown --output: expected text or json")
			}
			if !report.Valid {
				return fmt.Errorf("deployment evidence verification failed with %d violation(s)",
					len(report.Violations))
			}
			return nil
		},
	}
	command.Flags().StringVar(&contractFile, "contract", "", "strict JSON topology contract")
	command.Flags().StringVar(&manifestFile, "manifest", "", "strict JSON deployment-evidence manifest")
	command.Flags().StringVar(&at, "at", "", "evaluation time as a strict UTC timestamp")
	command.Flags().StringVar(&output, "output", "text", "output format: text or json")
	command.SetFlagErrorFunc(safeFlagError)
	_ = command.MarkFlagRequired("contract")
	_ = command.MarkFlagRequired("manifest")
	return command
}

func noArgs(_ *cobra.Command, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("command accepts no positional arguments")
	}
	return nil
}

// safeFlagError keeps rejected flag values out of command errors.
func safeFlagError(_ *cobra.Command, _ error) error {
	return fmt.Errorf("invalid deployment evidence flags")
}
