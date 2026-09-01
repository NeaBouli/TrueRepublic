package crossrunevidence

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "compare" {
		_, _ = fmt.Fprintln(stderr, "usage: cross-run-evidence compare --evidence <dir> --expected-repository <owner/repo> --expected-workflow <path> --expected-branch <branch> --expected-commit <40-hex> --expected-baseline-run-id <id> --expected-current-run-id <id> [--contract <file>] [--candidate-contract <file>] [--output text|json]")
		return 2
	}
	flags := flag.NewFlagSet("compare", flag.ContinueOnError)
	flags.SetOutput(stderr)
	evidence := flags.String("evidence", "", "cross-run comparison evidence directory")
	contract := flags.String("contract", "configs/release/cross-run-rebuild.json", "cross-run rebuild contract")
	candidateContract := flags.String("candidate-contract", "configs/release/candidate-evidence.json", "release-candidate contract")
	repository := flags.String("expected-repository", "", "expected repository (owner/repo)")
	workflow := flags.String("expected-workflow", "", "expected workflow path")
	branch := flags.String("expected-branch", "", "expected branch")
	commit := flags.String("expected-commit", "", "expected exact commit")
	baselineRunID := flags.String("expected-baseline-run-id", "", "expected baseline run ID")
	currentRunID := flags.String("expected-current-run-id", "", "expected current run ID")
	output := flags.String("output", "text", "text or json")
	if err := flags.Parse(args[1:]); err != nil || flags.NArg() != 0 || *evidence == "" ||
		*repository == "" || *workflow == "" || *branch == "" || *commit == "" ||
		*baselineRunID == "" || *currentRunID == "" || (*output != "text" && *output != "json") {
		return 2
	}
	expected := Expected{
		Repository:    *repository,
		WorkflowPath:  *workflow,
		Branch:        *branch,
		Commit:        *commit,
		BaselineRunID: *baselineRunID,
		CurrentRunID:  *currentRunID,
	}
	report := Compare(*evidence, *contract, *candidateContract, expected)
	if *output == "json" {
		encoder := json.NewEncoder(stdout)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(report); err != nil {
			_, _ = fmt.Fprintln(stderr, "write cross-run evidence report:", err)
			return 1
		}
	} else {
		if _, err := fmt.Fprintln(stdout, formatViolations(report)); err != nil {
			_, _ = fmt.Fprintln(stderr, "write cross-run evidence report:", err)
			return 1
		}
	}
	if !report.Valid {
		return 1
	}
	return 0
}
