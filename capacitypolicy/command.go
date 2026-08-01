package capacitypolicy

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "capacity-policy",
		Short: "Validate secret-free capacity qualification plans and evidence",
	}
	command.AddCommand(newValidateCommand(), newVerifyCommand())
	return command
}

func newValidateCommand() *cobra.Command {
	var file, output string
	command := &cobra.Command{
		Use: "validate --file <contract.json>", Args: noArgs,
		Short:        "Fail closed on unsafe or unbounded capacity plans",
		SilenceUsage: true, SilenceErrors: true,
		RunE: func(command *cobra.Command, _ []string) error {
			if file == "" {
				return fmt.Errorf("--file is required")
			}
			contract, err := LoadContract(file)
			if err != nil {
				return err
			}
			report := ValidateContract(contract)
			if err := printReport(command, output, report, "capacity contract"); err != nil {
				return err
			}
			if !report.Valid {
				return fmt.Errorf("capacity contract validation failed")
			}
			return nil
		},
	}
	command.Flags().StringVar(&file, "file", "", "strict JSON capacity contract")
	command.Flags().StringVar(&output, "output", "text", "output format: text or json")
	command.SetFlagErrorFunc(safeFlagError)
	_ = command.MarkFlagRequired("file")
	return command
}

func newVerifyCommand() *cobra.Command {
	var contractFile, evidenceFile, output string
	command := &cobra.Command{
		Use: "verify --contract <contract.json> --evidence <evidence.json>", Args: noArgs,
		Short:        "Fail closed on incomplete or unsafe capacity evidence",
		SilenceUsage: true, SilenceErrors: true,
		RunE: func(command *cobra.Command, _ []string) error {
			if contractFile == "" || evidenceFile == "" {
				return fmt.Errorf("--contract and --evidence are required")
			}
			contract, err := LoadContract(contractFile)
			if err != nil {
				return err
			}
			evidence, err := LoadEvidence(evidenceFile)
			if err != nil {
				return err
			}
			report := ValidateEvidence(contract, evidence)
			if err := printReport(command, output, report, "capacity evidence"); err != nil {
				return err
			}
			if !report.Valid {
				return fmt.Errorf("capacity evidence validation failed")
			}
			return nil
		},
	}
	command.Flags().StringVar(&contractFile, "contract", "", "strict JSON capacity contract")
	command.Flags().StringVar(&evidenceFile, "evidence", "", "strict JSON capacity evidence")
	command.Flags().StringVar(&output, "output", "text", "output format: text or json")
	command.SetFlagErrorFunc(safeFlagError)
	_ = command.MarkFlagRequired("contract")
	_ = command.MarkFlagRequired("evidence")
	return command
}

func noArgs(_ *cobra.Command, args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("command accepts no positional arguments")
	}
	return nil
}

func safeFlagError(_ *cobra.Command, _ error) error {
	return fmt.Errorf("invalid capacity policy flags")
}

func printReport(command *cobra.Command, output string, report any, label string) error {
	switch output {
	case "json":
		encoded, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("encode %s report", label)
		}
		command.Println(string(encoded))
	case "text":
		switch value := report.(type) {
		case Report:
			if value.Valid {
				command.Printf("OK: %s validates %d validators and %d transactions\n", label, value.ValidatorCount, value.TransactionCount)
			} else {
				printViolations(command, value.Violations)
			}
		case EvidenceReport:
			if value.Valid {
				command.Printf("OK: %s validates %d validators and %d committed transactions\n", label, value.ValidatorCount, value.CommittedCount)
			} else {
				printViolations(command, value.Violations)
			}
		}
	default:
		return fmt.Errorf("unknown --output: expected text or json")
	}
	return nil
}

func printViolations(command *cobra.Command, violations []Violation) {
	for _, violation := range violations {
		command.Printf("VIOLATION %s: %s\n", violation.Check, violation.Message)
	}
}
