package topologypolicy

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// NewCommand builds the offline topology-policy command group.
func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "topology-policy",
		Short: "Validate a secret-free multi-node topology qualification contract",
	}
	command.AddCommand(newValidateCommand())
	return command
}

func newValidateCommand() *cobra.Command {
	var file, output string
	command := &cobra.Command{
		Use:           "validate --file <contract.json>",
		Short:         "Fail closed on unsafe topology relationships or ingress controls",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(command *cobra.Command, _ []string) error {
			if file == "" {
				return fmt.Errorf("--file is required")
			}
			contract, err := Load(file)
			if err != nil {
				return err
			}
			report := Validate(contract)
			switch output {
			case "json":
				encoded, err := json.MarshalIndent(report, "", "  ")
				if err != nil {
					return err
				}
				command.Println(string(encoded))
			case "text":
				if report.Valid {
					command.Printf("OK: topology contract %s validates %d nodes\n",
						report.Version, report.NodeCount)
				} else {
					for _, violation := range report.Violations {
						command.Printf("VIOLATION %s: %s\n",
							violation.Check, violation.Message)
					}
				}
			default:
				return fmt.Errorf("unknown --output %q: expected text or json", output)
			}
			if !report.Valid {
				return fmt.Errorf("topology policy validation failed with %d violation(s)",
					len(report.Violations))
			}
			return nil
		},
	}
	command.Flags().StringVar(&file, "file", "", "strict JSON topology contract")
	command.Flags().StringVar(&output, "output", "text", "output format: text or json")
	_ = command.MarkFlagRequired("file")
	return command
}
