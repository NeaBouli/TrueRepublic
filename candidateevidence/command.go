package candidateevidence

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "verify" {
		_, _ = fmt.Fprintln(stderr, "usage: candidate-evidence verify --evidence <dir> [--contract <file>] [--output text|json]")
		return 2
	}
	flags := flag.NewFlagSet("verify", flag.ContinueOnError)
	flags.SetOutput(stderr)
	evidence := flags.String("evidence", "", "release-candidate evidence directory")
	contract := flags.String("contract", "configs/release/candidate-evidence.json", "release-candidate contract")
	output := flags.String("output", "text", "text or json")
	if err := flags.Parse(args[1:]); err != nil || flags.NArg() != 0 || *evidence == "" || (*output != "text" && *output != "json") {
		return 2
	}
	report := VerifyDirectory(*evidence, *contract)
	if *output == "json" {
		encoder := json.NewEncoder(stdout)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(report); err != nil {
			_, _ = fmt.Fprintln(stderr, "write candidate evidence report:", err)
			return 1
		}
	} else {
		_, _ = fmt.Fprintln(stdout, formatViolations(report))
	}
	if !report.Valid {
		return 1
	}
	return 0
}
