package releaseevidence

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	root := &cobra.Command{Use: "release-evidence", Short: "Verify an offline release-evidence bundle"}
	var bundle, buildContract, toolContract, output string
	verify := &cobra.Command{
		Use: "verify", Short: "Fail closed on incomplete or cross-unbound release evidence",
		Args: cobra.NoArgs, SilenceUsage: true, SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if bundle == "" || buildContract == "" || toolContract == "" {
				return fmt.Errorf("--bundle, --build-contract, and --tool-contract are required")
			}
			report := VerifyDirectory(bundle, buildContract, toolContract)
			switch output {
			case "json":
				b, _ := json.MarshalIndent(report, "", "  ")
				cmd.Println(string(b))
			case "text":
				if report.Valid {
					cmd.Println("OK: release evidence verifies two targets and two SBOMs")
				} else {
					for _, v := range report.Violations {
						cmd.Printf("VIOLATION: %s\n", v)
					}
				}
			default:
				return fmt.Errorf("unknown --output: expected text or json")
			}
			if !report.Valid {
				return fmt.Errorf("release evidence verification failed with %d violation(s)", len(report.Violations))
			}
			return nil
		},
	}
	verify.Flags().StringVar(&bundle, "bundle", "", "release bundle directory")
	verify.Flags().StringVar(&buildContract, "build-contract", "", "deterministic build contract")
	verify.Flags().StringVar(&toolContract, "tool-contract", "", "release tool/platform contract")
	verify.Flags().StringVar(&output, "output", "text", "output format: text or json")
	root.AddCommand(verify)
	return root
}
