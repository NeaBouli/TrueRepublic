package networkpolicy

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// NewCommand builds the `network-policy` command group. It is deliberately
// self-contained (own --home flag, no reliance on daemon pre-run config
// interception) so it can run against any initialized home without touching
// it. Validation is read-only.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "network-policy",
		Short: "Validate node home network configuration against a role policy",
	}
	cmd.AddCommand(newValidateCommand())
	return cmd
}

func newValidateCommand() *cobra.Command {
	var (
		roleFlag   string
		homeFlag   string
		outputFlag string
	)
	cmd := &cobra.Command{
		Use:   "validate --role <seed|sentry|validator|rpc|private> --home <path>",
		Short: "Fail-closed validation of an initialized node home against a role profile",
		Long: "Inspects the effective CometBFT config.toml and Cosmos app.toml of an initialized\n" +
			"node home and enforces the least-privilege network boundary for the given role.\n" +
			"The check is read-only, deterministic, and prints no secret material.\n" +
			"Exit code is non-zero when any policy violation is found.",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			role, err := ParseRole(roleFlag)
			if err != nil {
				return err
			}
			if homeFlag == "" {
				return fmt.Errorf("--home is required and must point at an initialized node home")
			}
			report := Validate(homeFlag, role, Options{})
			switch outputFlag {
			case "json":
				encoded, err := json.MarshalIndent(report, "", "  ")
				if err != nil {
					return err
				}
				cmd.Println(string(encoded))
			case "text", "":
				if report.Valid {
					cmd.Printf("OK: initialized home satisfies the %s network policy\n", report.Role)
				} else {
					for _, violation := range report.Violations {
						cmd.Printf("VIOLATION %s: %s\n", violation.Check, violation.Message)
					}
				}
			default:
				return fmt.Errorf("unknown --output %q: expected text or json", outputFlag)
			}
			if !report.Valid {
				return fmt.Errorf("network policy validation failed with %d violation(s)", len(report.Violations))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&roleFlag, "role", "", "role profile to enforce: seed, sentry, validator, rpc, or private")
	cmd.Flags().StringVar(&homeFlag, "home", "", "initialized node home containing config/config.toml and config/app.toml")
	cmd.Flags().StringVar(&outputFlag, "output", "text", "output format: text or json")
	_ = cmd.MarkFlagRequired("role")
	_ = cmd.MarkFlagRequired("home")
	return cmd
}
