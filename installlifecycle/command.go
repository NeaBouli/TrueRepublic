package installlifecycle

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return newCommandForHost(runtime.GOOS + "-" + runtime.GOARCH)
}

func newCommandForHost(hostTarget string) *cobra.Command {
	root := &cobra.Command{Use: "install-lifecycle", Short: "Manage a verified TrueRepublic daemon installation", SilenceUsage: true, SilenceErrors: true}
	var contractPath, prefix, operatorState, sha, sourceRef, target, runtime string
	root.PersistentFlags().StringVar(&contractPath, "contract", "", "absolute path to lifecycle contract")
	root.PersistentFlags().StringVar(&prefix, "prefix", "", "absolute isolated installation prefix")
	root.PersistentFlags().StringVar(&operatorState, "operator-state", "", "absolute operator state path outside prefix")
	root.PersistentFlags().StringVar(&sha, "sha256", "", "exact candidate or installed artifact digest")
	root.PersistentFlags().StringVar(&sourceRef, "source-ref", "", "exact 40-character source commit")
	root.PersistentFlags().StringVar(&target, "target", "", "contract-supported target")
	root.PersistentFlags().StringVar(&runtime, "runtime", "", "runtime ABI binding")
	load := func() (Contract, error) {
		if contractPath == "" {
			return Contract{}, fmt.Errorf("--contract is required")
		}
		template, err := LoadContract(contractPath)
		if err != nil {
			return Contract{}, err
		}
		if target != hostTarget {
			return Contract{}, fmt.Errorf("--target must match host target %s", hostTarget)
		}
		return Bind(template, prefix, operatorState, sha, sourceRef, target, runtime)
	}
	var artifact string
	install := &cobra.Command{Use: "install", Args: cobra.NoArgs, RunE: func(*cobra.Command, []string) error {
		c, err := load()
		if err != nil {
			return err
		}
		return Install(c, artifact)
	}}
	install.Flags().StringVar(&artifact, "artifact", "", "absolute path to verified artifact")
	_ = install.MarkFlagRequired("artifact")
	var upgradeArtifact, expected, expectedSource string
	upgrade := &cobra.Command{Use: "upgrade", Args: cobra.NoArgs, RunE: func(*cobra.Command, []string) error {
		c, err := load()
		if err != nil {
			return err
		}
		return Upgrade(c, upgradeArtifact, expected, expectedSource)
	}}
	upgrade.Flags().StringVar(&upgradeArtifact, "artifact", "", "absolute path to verified artifact")
	upgrade.Flags().StringVar(&expected, "expected-current-sha256", "", "expected installed binary digest")
	upgrade.Flags().StringVar(&expectedSource, "expected-current-source-ref", "", "expected installed source commit")
	_ = upgrade.MarkFlagRequired("artifact")
	_ = upgrade.MarkFlagRequired("expected-current-sha256")
	_ = upgrade.MarkFlagRequired("expected-current-source-ref")
	rollback := &cobra.Command{Use: "rollback", Args: cobra.NoArgs, RunE: func(*cobra.Command, []string) error {
		c, err := load()
		if err != nil {
			return err
		}
		return Rollback(c)
	}}
	var uninstallExpected string
	uninstall := &cobra.Command{Use: "uninstall", Args: cobra.NoArgs, RunE: func(*cobra.Command, []string) error {
		c, err := load()
		if err != nil {
			return err
		}
		return Uninstall(c, uninstallExpected)
	}}
	uninstall.Flags().StringVar(&uninstallExpected, "expected-current-sha256", "", "expected installed binary digest")
	_ = uninstall.MarkFlagRequired("expected-current-sha256")
	preStart := &cobra.Command{Use: "pre-start", Args: cobra.NoArgs, RunE: func(*cobra.Command, []string) error {
		c, err := load()
		if err != nil {
			return err
		}
		return PreStart(c)
	}}
	status := &cobra.Command{Use: "status", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		c, err := load()
		if err != nil {
			return err
		}
		s, checkErr := Check(c)
		b, _ := json.MarshalIndent(s, "", "  ")
		cmd.Println(string(b))
		return checkErr
	}}
	root.AddCommand(install, status, preStart, upgrade, rollback, uninstall)
	return root
}
