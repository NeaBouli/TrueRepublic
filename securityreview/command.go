package securityreview

import (
	"flag"
	"fmt"
	"io"
)

func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("security-review", flag.ContinueOnError)
	flags.SetOutput(stderr)
	repoRoot := flags.String("repo-root", "", "repository root")
	scopePath := flags.String("scope", "", "repository-relative review scope JSON")
	findingsPath := flags.String("findings", "", "repository-relative findings JSON")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 || *repoRoot == "" || *scopePath == "" || *findingsPath == "" {
		_, _ = fmt.Fprintln(stderr, "usage: security-review --repo-root DIR --scope PATH --findings PATH")
		return 2
	}
	if err := VerifyFiles(*repoRoot, *scopePath, *findingsPath); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	if _, err := fmt.Fprintln(stdout, "security review readiness contract verified; independent review and production claims remain false"); err != nil {
		return 1
	}
	return 0
}
