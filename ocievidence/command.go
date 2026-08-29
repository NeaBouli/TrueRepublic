package ocievidence

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
)

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "verify" {
		_, _ = fmt.Fprintln(stderr, "usage: oci-evidence verify --evidence <dir> [--contract <file>] [--output text|json]")
		return 2
	}
	flags := flag.NewFlagSet("verify", flag.ContinueOnError)
	flags.SetOutput(stderr)
	evidence := flags.String("evidence", "", "OCI evidence directory")
	contract := flags.String("contract", "configs/build/reproducible-oci.json", "OCI build contract")
	output := flags.String("output", "text", "text or json")
	if err := flags.Parse(args[1:]); err != nil || flags.NArg() != 0 || *evidence == "" || (*output != "text" && *output != "json") {
		return 2
	}
	report := VerifyDirectory(*evidence, *contract)
	if *output == "json" {
		encoder := json.NewEncoder(stdout)
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(report); err != nil {
			_, _ = fmt.Fprintln(stderr, "write OCI evidence report:", err)
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
