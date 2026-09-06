//go:build tools

// Package tools pins the exact repository-owned CI/release tool module
// versions. The blank imports keep the tool commands and their transitive
// dependencies locked in go.mod/go.sum so scripts/build-ci-tool.sh can build
// them with go mod verify and go build -mod=readonly instead of live
// go install ...@version resolution. This nested module is deliberately
// excluded from the root module's package selection and builds no repository
// code.
package tools

import (
	_ "github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod"
	_ "github.com/zricethezav/gitleaks/v8"
	_ "golang.org/x/vuln/cmd/govulncheck"
	_ "honnef.co/go/tools/cmd/staticcheck"
)
