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
	// A crate-level cfg can compile the whole crate, guard included, out of a
	// wasm32 build; a [lib] path override can point the build at an unguarded
	// file. Both bypass the unconditional guard. Plain host-only features are
	// allowed.
	contractQuarantineCrateCfgRE = regexp.MustCompile(`(?m)^\s*#!\[cfg(_attr)?\(`)
	contractQuarantineLibPathRE  = regexp.MustCompile(`(?m)^\[lib\][^\[]*^\s*path\s*=`)
	// Wasm artifacts of the quarantined crates, and package/build commands
	// that name one of them.
	contractQuarantineArtifactRE = regexp.MustCompile(
		`\b(truerepublic_contracts|governance_dao|zkp_aggregator|dex_bot|token_vesting|governance|treasury)\.wasm\b`)
	contractQuarantinePackageCmdRE = regexp.MustCompile(
		`cargo[^\n]*\b(build|publish|wasm)\b[^\n]*(-p|--package)[ =](truerepublic-contracts|governance-dao|zkp-aggregator|dex-bot|token-vesting)\b`)
	// A wasm build or optimizer run is a packaging path for the quarantined
	// crates only when it targets the contracts workspace or one of them.
	contractQuarantineWasmBuildRE = regexp.MustCompile(`wasm32-unknown-unknown|workspace-optimizer|rust-optimizer|cargo wasm\b`)
	contractQuarantineWorkspaceRE = regexp.MustCompile(
		`(?m)working-directory:\s*\.?/?contracts\s*$|cd \.?/?contracts(/(core|examples/[a-z-]+))?\b|contracts/Cargo\.toml|contracts/(core|examples/(governance-dao|zkp-aggregator|dex-bot|token-vesting))\b`)
	contractQuarantineDeployRE = regexp.MustCompile(`tx wasm (store|instantiate2?|migrate)\b`)
)

type contractQuarantineInputs struct {
	cargo     map[string]string // crate dir -> Cargo.toml
	lib       map[string]string // crate dir -> src/lib.rs
	packaging map[string]string // CI/image/release file -> content
	guides    map[string]string // maintained guide -> content
	ignore    string            // .dockerignore
	verifier  string            // scripts/verify-contract-quarantine-wasm.sh
}

// The rust-ci workflow installs the wasm32 target only to prove the quarantine:
// this exact toolchain line is the one wasm32 reference allowed in CI, and only
// while the verifier step and trigger are present (GH308D1).
const (
	contractQuarantineWorkflowPath   = ".github/workflows/rust-ci.yml"
	contractQuarantineVerifierPath   = "scripts/verify-contract-quarantine-wasm.sh"
	contractQuarantineWasmTargetLine = "          targets: wasm32-unknown-unknown\n"
	contractQuarantineVerifierStep   = "        run: ../scripts/verify-contract-quarantine-wasm.sh\n"
	contractQuarantineVerifierPathRE = "      - 'scripts/verify-contract-quarantine-wasm.sh'\n"
)

var contractQuarantinePackageNames = []string{"truerepublic-contracts", "governance-dao", "zkp-aggregator", "dex-bot", "token-vesting"}

// contractQuarantineVerifierViolations checks that the wasm32 verifier covers
// exactly the five quarantined packages, requires the marker and fails on success.
func contractQuarantineVerifierViolations(script string) []string {
	var out []string
	add := func(reason string) { out = append(out, contractQuarantineVerifierPath+": "+reason) }
	start := strings.Index(script, "PACKAGES=(\n")
	end := -1
	if start >= 0 {
		end = strings.Index(script[start:], "\n)\n")
	}
	if start < 0 || end < 0 {
		add("missing PACKAGES array")
	} else {
		listed := strings.Fields(script[start+len("PACKAGES=(\n") : start+end])
		if strings.Join(listed, ",") != strings.Join(contractQuarantinePackageNames, ",") {
			add("must cover exactly " + strings.Join(contractQuarantinePackageNames, ", ") + "; got " + strings.Join(listed, ", "))
		}
	}
	for _, required := range []string{
		`MARKER="QUARANTINED (TrueRepublic #308)"`,
		`TARGET="wasm32-unknown-unknown"`,
		`cargo check --locked --package "$package"`,
		`--target "$TARGET"`,
		`echo "FAIL $package compiled for $TARGET`,
		`grep -Fq "$MARKER" "$log"`,
		`trap cleanup EXIT`,
	} {
		if !strings.Contains(script, required) {
			add("must contain " + required)
		}
	}
	if strings.Count(script, "failures=$((failures + 1))") < 2 {
		add("both a successful build and a markerless failure must count as failures")
	}
	return out
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
		"lib path override": func(in *contractQuarantineInputs) {
			in.cargo["contracts/examples/governance-dao"] = strings.Replace(in.cargo["contracts/examples/governance-dao"],
				"[lib]\n", "[lib]\npath = \"src/unguarded.rs\"\n", 1)
		},
		"missing wasm32 guard": func(in *contractQuarantineInputs) {
			in.lib["contracts/examples/zkp-aggregator"] = strings.Replace(in.lib["contracts/examples/zkp-aggregator"], contractQuarantineGuardCfg, "", 1)
		},
		"feature-gated guard": func(in *contractQuarantineInputs) {
			in.lib["contracts/examples/dex-bot"] = strings.Replace(in.lib["contracts/examples/dex-bot"],
				contractQuarantineGuardCfg, `#[cfg(all(target_arch = "wasm32", not(feature = "unsafe")))]`, 1)
		},
		"crate-level cfg bypass": func(in *contractQuarantineInputs) {
			in.lib["contracts/examples/token-vesting"] = "#![cfg(not(target_arch = \"wasm32\"))]\n" + in.lib["contracts/examples/token-vesting"]
		},
		"wasm build in CI": func(in *contractQuarantineInputs) {
			in.packaging[".github/workflows/rust-ci.yml"] += "\n        run: cargo build --release --target wasm32-unknown-unknown\n"
		},
		"artifact in release contract": func(in *contractQuarantineInputs) {
			in.packaging["configs/release/compatibility.json"] += `{"artifact": "governance_dao.wasm"}`
		},
		"crate publish command": func(in *contractQuarantineInputs) {
			in.packaging["Dockerfile"] += "\nRUN cargo publish -p zkp-aggregator\n"
		},
		"deploy recipe in guide": func(in *contractQuarantineInputs) {
			in.guides["docs/QUICKSTART.md"] += "\n\n```bash\ntruerepublicd tx wasm store governance_dao.wasm --from alice\n```\n"
		},
		"wasm build recipe in guide": func(in *contractQuarantineInputs) {
			in.guides["INSTALLATION.md"] += "\n\n```bash\ncd contracts\ncargo build --release --target wasm32-unknown-unknown\n```\n"
		},
		"verifier target removed": func(in *contractQuarantineInputs) {
			in.packaging[contractQuarantineWorkflowPath] = strings.Replace(in.packaging[contractQuarantineWorkflowPath], contractQuarantineWasmTargetLine, "", 1)
		},
		"verifier step removed": func(in *contractQuarantineInputs) {
			in.packaging[contractQuarantineWorkflowPath] = strings.Replace(in.packaging[contractQuarantineWorkflowPath], contractQuarantineVerifierStep, "", 1)
		},
		"verifier trigger removed": func(in *contractQuarantineInputs) {
			in.packaging[contractQuarantineWorkflowPath] = strings.Replace(in.packaging[contractQuarantineWorkflowPath], contractQuarantineVerifierPathRE, "", 1)
		},
		"extra wasm build step in CI": func(in *contractQuarantineInputs) {
			in.packaging[contractQuarantineWorkflowPath] += "\n      - run: cargo build --release --target wasm32-unknown-unknown\n"
		},
		"verifier package omitted": func(in *contractQuarantineInputs) {
			in.verifier = strings.Replace(in.verifier, "  token-vesting\n", "", 1)
		},
		"verifier accepts success": func(in *contractQuarantineInputs) {
			in.verifier = strings.Replace(in.verifier, `echo "FAIL $package compiled for $TARGET; the quarantine guard is not effective" >&2
    failures=$((failures + 1))`, `echo "OK   $package compiled for $TARGET"`, 1)
		},
		"verifier marker weakened": func(in *contractQuarantineInputs) {
			in.verifier = strings.Replace(in.verifier, `MARKER="QUARANTINED (TrueRepublic #308)"`, `MARKER="QUARANTINED"`, 1)
		},
		"contracts in image context": func(in *contractQuarantineInputs) {
			in.ignore = strings.ReplaceAll(in.ignore, "contracts", "")
		},
	}
	allowed := map[string]func(*contractQuarantineInputs){
		"host-only feature": func(in *contractQuarantineInputs) {
			in.cargo["contracts/examples/governance-dao"] += "\n[features]\nhost-fixtures = []\n"
		},
		"crate name in review metadata": func(in *contractQuarantineInputs) {
			in.packaging["configs/release/compatibility.json"] += `{"note": "governance-dao and zkp-aggregator stay quarantined under #308"}`
		},
		"reviewed contract deploy guide": func(in *contractQuarantineInputs) {
			in.guides["docs/developers/smart-contracts/cosmwasm.md"] += "\n\n```bash\ntruerepublicd tx wasm store reviewed_escrow.wasm --from wallet\n```\n"
		},
	}
	allowedNames := make([]string, 0, len(allowed))
	for name := range allowed {
		allowedNames = append(allowedNames, name)
	}
	sort.Strings(allowedNames)
	for _, name := range allowedNames {
		t.Run("allows "+name, func(t *testing.T) {
			clone := cloneContractQuarantineInputs(inputs)
			allowed[name](&clone)
			if violations := contractQuarantineViolations(clone); len(violations) != 0 {
				t.Fatalf("harmless change rejected:\n- %s", strings.Join(violations, "\n- "))
			}
		})
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
		verifier:  read(contractQuarantineVerifierPath),
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
		if contractQuarantineLibPathRE.MatchString(cargo) {
			out = append(out, crate+": a [lib] path override can bypass the wasm32 guard")
		}
		if !strings.Contains(lib, contractQuarantineGuardCfg+"\n"+contractQuarantineGuardMarker) {
			out = append(out, crate+": src/lib.rs lacks the unconditional wasm32 quarantine compile_error guard")
		}
		if contractQuarantineCrateCfgRE.MatchString(lib) {
			out = append(out, crate+": a crate-level cfg can compile the wasm32 guard out")
		}
	}
	workflow := in.packaging[contractQuarantineWorkflowPath]
	if !strings.Contains(workflow, contractQuarantineWasmTargetLine) ||
		!strings.Contains(workflow, contractQuarantineVerifierStep) ||
		strings.Count(workflow, contractQuarantineVerifierPathRE) != 2 {
		out = append(out, contractQuarantineWorkflowPath+": must install wasm32-unknown-unknown, run the quarantine verifier and trigger on its changes")
	}
	out = append(out, contractQuarantineVerifierViolations(in.verifier)...)
	for _, path := range sortedContractQuarantineKeys(in.packaging) {
		content := in.packaging[path]
		if path == contractQuarantineWorkflowPath {
			// The verifier's toolchain target is the only allowed wasm32 use.
			content = strings.Replace(content, contractQuarantineWasmTargetLine, "", 1)
		}
		if match := contractQuarantineArtifactRE.FindString(content); match != "" {
			out = append(out, path+": packages a quarantined contract artifact ("+match+")")
		}
		if match := contractQuarantinePackageCmdRE.FindString(content); match != "" {
			out = append(out, path+": builds or publishes a quarantined crate ("+match+")")
		}
		if contractQuarantineWasmBuildRE.MatchString(content) && contractQuarantineWorkspaceRE.MatchString(content) {
			out = append(out, path+": builds the quarantined contracts workspace for wasm")
		}
	}
	if !regexp.MustCompile(`(?m)^contracts/?$`).MatchString(in.ignore) {
		out = append(out, ".dockerignore must exclude contracts from the daemon image context")
	}
	for _, path := range sortedContractQuarantineKeys(in.guides) {
		// Judge each paragraph/code block on its own, so documentation for a
		// separately reviewed contract elsewhere in a guide stays allowed.
		for _, block := range strings.Split(in.guides[path], "\n\n") {
			deploysPrototype := contractQuarantineDeployRE.MatchString(block) && contractQuarantineArtifactRE.MatchString(block)
			buildsPrototype := contractQuarantineWasmBuildRE.MatchString(block) && contractQuarantineWorkspaceRE.MatchString(block)
			if deploysPrototype || buildsPrototype {
				out = append(out, path+": maintained guide builds or deploys a quarantined prototype")
				break
			}
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
		verifier: in.verifier,
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
