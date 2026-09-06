package toolbootstrapevidence

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "verify" {
		_, _ = fmt.Fprintln(stderr, "usage: tool-bootstrap-evidence verify --evidence <dir> --artifacts <dir> [--contract <file>] [--gates <file>] [--locks-root <dir>] [--output text|json]")
		return 2
	}
	flags := flag.NewFlagSet("verify", flag.ContinueOnError)
	flags.SetOutput(stderr)
	evidence := flags.String("evidence", "", "tool-bootstrap evidence directory")
	artifacts := flags.String("artifacts", "", "built tool artifacts directory")
	contract := flags.String("contract", "configs/release/tool-platform.json", "tool-platform contract")
	gates := flags.String("gates", "configs/security/gates.json", "security gate contract")
	locksRoot := flags.String("locks-root", ".", "repository root the contract lock paths resolve against")
	output := flags.String("output", "text", "text or json")
	if err := flags.Parse(args[1:]); err != nil || flags.NArg() != 0 || *evidence == "" || *artifacts == "" ||
		*contract == "" || *gates == "" || *locksRoot == "" || (*output != "text" && *output != "json") {
		return 2
	}
	report := Verify(*evidence, *artifacts, *contract, *gates, *locksRoot)
	if *output == "json" {
		encoder := json.NewEncoder(stdout)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(report); err != nil {
			_, _ = fmt.Fprintln(stderr, "write tool-bootstrap evidence report:", err)
			return 1
		}
	} else {
		if _, err := fmt.Fprintln(stdout, formatViolations(report)); err != nil {
			_, _ = fmt.Fprintln(stderr, "write tool-bootstrap evidence report:", err)
			return 1
		}
	}
	if !report.Valid {
		return 1
	}
	return 0
}
