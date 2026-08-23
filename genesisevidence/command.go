package genesisevidence

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	root := &cobra.Command{Use: "genesis-evidence", Short: "Verify a rollout genesis offline", SilenceUsage: true, SilenceErrors: true}
	root.AddCommand(newVerifyCommand())
	return root
}

func newVerifyCommand() *cobra.Command {
	var manifestPath, genesisPath, output string
	cmd := &cobra.Command{Use: "verify", Args: func(_ *cobra.Command, args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("verify accepts no positional arguments")
		}
		return nil
	}, RunE: func(cmd *cobra.Command, _ []string) error {
		manifest, err := readBounded(manifestPath, MaxManifestBytes)
		if err != nil {
			return fmt.Errorf("read manifest: unavailable")
		}
		genesis, err := readBounded(genesisPath, MaxGenesisBytes)
		if err != nil {
			return fmt.Errorf("read genesis: unavailable")
		}
		evidence := Verify(manifest, genesis)
		switch output {
		case "json":
			data, err := MarshalJSON(evidence)
			if err != nil {
				return fmt.Errorf("encode evidence")
			}
			cmd.Println(string(data))
		case "text":
			for _, c := range evidence.Checks {
				if c.Pass {
					cmd.Printf("PASS %s\n", c.Name)
				} else {
					for _, v := range c.Violations {
						cmd.Printf("FAIL %s %s\n", c.Name, v)
					}
				}
			}
		default:
			return fmt.Errorf("--output must be json or text")
		}
		if !evidence.Valid {
			return fmt.Errorf("genesis evidence verification failed")
		}
		return nil
	}}
	cmd.Flags().StringVar(&manifestPath, "manifest", "", "rollout genesis manifest JSON")
	cmd.Flags().StringVar(&genesisPath, "genesis", "", "candidate Cosmos genesis JSON")
	cmd.Flags().StringVar(&output, "output", "text", "output format: text or json")
	_ = cmd.MarkFlagRequired("manifest")
	_ = cmd.MarkFlagRequired("genesis")
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, _ error) error { return fmt.Errorf("invalid genesis evidence flags") })
	return cmd
}
func readBounded(path string, limit int) ([]byte, error) {
	pathInfo, err := os.Lstat(path)
	if err != nil || !pathInfo.Mode().IsRegular() {
		return nil, fmt.Errorf("invalid input file")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() || !os.SameFile(pathInfo, info) || info.Size() > int64(limit) {
		return nil, fmt.Errorf("invalid size")
	}
	data, err := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	if err != nil || len(data) > limit {
		return nil, fmt.Errorf("invalid size")
	}
	return data, nil
}
