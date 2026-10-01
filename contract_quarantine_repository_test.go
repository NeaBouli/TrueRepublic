package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// quarantinedContractCrates are the prototype CosmWasm crates from issue #308
// (six contract implementations in five crates). They must stay unpublished,
// reject wasm32 compilation and never be packaged or deployed.
var quarantinedContractCrates = []string{
	"contracts/core",
	"contracts/examples/dex-bot",
	"contracts/examples/governance-dao",
	"contracts/examples/token-vesting",
	"contracts/examples/zkp-aggregator",
}

const contractQuarantineGuardCfg = `#[cfg(target_arch = "wasm32")]`

const contractQuarantineGuardMarker = `compile_error!(
    "QUARANTINED (TrueRepublic #308):`

var (
	contractQuarantinePublishRE = regexp.MustCompile(`(?m)^publish = false$`)
	// Any packaging reference to a wasm build, an optimizer or a quarantined
	// crate name in CI, images or release contracts.
	contractQuarantinePackagingRE = regexp.MustCompile(
		`wasm32-unknown-unknown|workspace-optimizer|rust-optimizer|optimizer-arm64|` +
			`truerepublic-contracts|truerepublic_contracts|governance-dao|governance_dao|` +
			`zkp-aggregator|zkp_aggregator|dex-bot|dex_bot|token-vesting|token_vesting`)
	contractQuarantineDeployRE = regexp.MustCompile(`tx wasm (store|instantiate2?|migrate)\b|cargo wasm\b`)
)

type contractQuarantineInputs struct {
	cargo     map[string]string // crate dir -> Cargo.toml
	lib       map[string]string // crate dir -> src/lib.rs
	packaging map[string]string // CI/image/release file -> content
	guides    map[string]string // maintained guide -> content
	ignore    string            // .dockerignore
}

func TestContractQuarantineRepositoryContract(t *testing.T) {
	inputs := loadContractQuarantineInputs(t)
	if violations := contractQuarantineViolations(inputs); len(violations) != 0 {
		t.Fatalf("contract quarantine violations:\n- %s", strings.Join(violations, "\n- "))
	}

	mutations := map[string]func(*contractQuarantineInputs){
		"published crate": func(in *contractQuarantineInputs) {
			in.cargo["contracts/core"] = strings.Replace(in.cargo["contracts/core"], "publish = false\n", "", 1)
		},
		"opt-in feature": func(in *contractQuarantineInputs) {
			in.cargo["contracts/examples/governance-dao"] += "\n[features]\nunsafe-prototype-entrypoints = []\n"
		},
		"missing wasm32 guard": func(in *contractQuarantineInputs) {
			in.lib["contracts/examples/zkp-aggregator"] = strings.Replace(in.lib["contracts/examples/zkp-aggregator"], contractQuarantineGuardCfg, "", 1)
		},
		"feature-gated guard": func(in *contractQuarantineInputs) {
			in.lib["contracts/examples/dex-bot"] = strings.Replace(in.lib["contracts/examples/dex-bot"],
				contractQuarantineGuardCfg, `#[cfg(all(target_arch = "wasm32", not(feature = "unsafe")))]`, 1)
		},
		"wasm build in CI": func(in *contractQuarantineInputs) {
			in.packaging[".github/workflows/rust-ci.yml"] += "\n        run: cargo build --release --target wasm32-unknown-unknown\n"
		},
		"crate in release contract": func(in *contractQuarantineInputs) {
			in.packaging["configs/release/compatibility.json"] += `{"artifact": "governance_dao.wasm"}`
		},
		"deploy recipe in guide": func(in *contractQuarantineInputs) {
			in.guides["docs/QUICKSTART.md"] += "\ntruerepublicd tx wasm store governance_dao.wasm --from alice\n"
		},
		"contracts in image context": func(in *contractQuarantineInputs) {
			in.ignore = strings.ReplaceAll(in.ignore, "contracts", "")
		},
	}
	names := make([]string, 0, len(mutations))
	for name := range mutations {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		t.Run("rejects "+name, func(t *testing.T) {
			clone := cloneContractQuarantineInputs(inputs)
			mutations[name](&clone)
			if len(contractQuarantineViolations(clone)) == 0 {
				t.Fatal("quarantine drift accepted")
			}
		})
	}
}

func loadContractQuarantineInputs(t *testing.T) contractQuarantineInputs {
	t.Helper()
	read := func(path string) string {
		t.Helper()
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return string(content)
	}
	in := contractQuarantineInputs{
		cargo:     map[string]string{},
		lib:       map[string]string{},
		packaging: map[string]string{},
		guides:    map[string]string{},
		ignore:    read(".dockerignore"),
	}
	for _, crate := range quarantinedContractCrates {
		in.cargo[crate] = read(filepath.Join(crate, "Cargo.toml"))
		in.lib[crate] = read(filepath.Join(crate, "src", "lib.rs"))
	}
	var packaging []string
	for _, pattern := range []string{".github/workflows/*.yml", "configs/build/*.json", "configs/release/*.json"} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		packaging = append(packaging, matches...)
	}
	packaging = append(packaging, "Dockerfile", "client-web/Dockerfile")
	for _, path := range packaging {
		in.packaging[path] = read(path)
	}
	guides := []string{"README.md", "INSTALLATION.md", "CONTRIBUTING.md"}
	for _, root := range []string{"docs", "wiki"} {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				// Dated audit reports and append-only agent logs are
				// historical records, not maintained guidance.
				if path == filepath.Join("docs", "community-audits") || path == filepath.Join("docs", "agent-bridge") {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(path, ".md") {
				guides = append(guides, path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range guides {
		in.guides[filepath.ToSlash(path)] = read(path)
	}
	return in
}

func contractQuarantineViolations(in contractQuarantineInputs) []string {
	var out []string
	for _, crate := range quarantinedContractCrates {
		cargo, lib := in.cargo[crate], in.lib[crate]
		if !contractQuarantinePublishRE.MatchString(cargo) {
			out = append(out, crate+": Cargo.toml must set publish = false")
		}
		if strings.Contains(cargo, "[features]") {
			out = append(out, crate+": no feature may re-enable a deployable prototype build")
		}
		guard := strings.Index(lib, contractQuarantineGuardCfg+"\n"+contractQuarantineGuardMarker)
		if guard < 0 {
			out = append(out, crate+": src/lib.rs lacks the unconditional wasm32 quarantine compile_error guard")
		}
	}
	for _, path := range sortedContractQuarantineKeys(in.packaging) {
		if match := contractQuarantinePackagingRE.FindString(in.packaging[path]); match != "" {
			out = append(out, path+": packages or builds quarantined contracts ("+match+")")
		}
	}
	if !regexp.MustCompile(`(?m)^contracts/?$`).MatchString(in.ignore) {
		out = append(out, ".dockerignore must exclude contracts from the daemon image context")
	}
	for _, path := range sortedContractQuarantineKeys(in.guides) {
		if match := contractQuarantineDeployRE.FindString(in.guides[path]); match != "" {
			out = append(out, path+": maintained guide instructs contract deployment ("+match+")")
		}
	}
	return out
}

func cloneContractQuarantineInputs(in contractQuarantineInputs) contractQuarantineInputs {
	clone := func(source map[string]string) map[string]string {
		out := make(map[string]string, len(source))
		for key, value := range source {
			out[key] = value
		}
		return out
	}
	return contractQuarantineInputs{
		cargo: clone(in.cargo), lib: clone(in.lib), packaging: clone(in.packaging), guides: clone(in.guides), ignore: in.ignore,
	}
}

func sortedContractQuarantineKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
