package incidentpolicy

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// NewCommand builds the offline incident-rehearsal command group.
func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "incident-rehearsal",
		Short: "Validate a secret-free operations rehearsal contract",
	}
	command.AddCommand(newValidateCommand())
	return command
}

func newValidateCommand() *cobra.Command {
	var file, output string
	command := &cobra.Command{
		Use:   "validate --file <contract.json>",
		Short: "Fail closed on unsafe or incomplete incident rehearsal plans",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("validate accepts no positional arguments")
			}
			return nil
		},
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
					command.Printf("OK: incident rehearsal %s validates %d scenarios\n",
						report.Version, report.ScenarioCount)
				} else {
					for _, violation := range report.Violations {
						command.Printf("VIOLATION %s: %s\n", violation.Check, violation.Message)
					}
				}
			default:
				return fmt.Errorf("unknown --output: expected text or json")
			}
			if !report.Valid {
				return fmt.Errorf("incident rehearsal validation failed")
			}
			return nil
		},
	}
	command.Flags().StringVar(&file, "file", "", "strict JSON incident rehearsal contract")
	command.Flags().StringVar(&output, "output", "text", "output format: text or json")
	command.SetFlagErrorFunc(func(_ *cobra.Command, _ error) error {
		return fmt.Errorf("invalid incident rehearsal flags")
	})
	_ = command.MarkFlagRequired("file")
	return command
}
