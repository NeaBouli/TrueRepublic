package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// GH-313 trust/reporting contract: maintained guidance must not depend on the
// unregistered truerepublic.network infrastructure, and vulnerability
// reporting must follow only the SECURITY.md private-reporting flow.

const operatorGuidanceSecurityPolicyURL = "https://github.com/NeaBouli/TrueRepublic/security/policy"

const operatorGuidanceCaveatMarker = "**Non-production (recovery status).**"

// Guides touched by GH-313B that carry operator or security risk.
var operatorGuidanceCaveatGuides = []string{
	"docs/node-operators/configuration/network-config.md",
	"docs/node-operators/configuration/node-config.md",
	"docs/node-operators/operations/security.md",
	"docs/user-manual/troubleshooting.md",
	"wiki/operations/Node-Setup.md",
	"wiki/security/Security-Architecture.md",
}

var (
	operatorGuidanceFictionalDomainRE = regexp.MustCompile(`(?i)truerepublic\.network`)
	// Affirmative reporting promises that SECURITY.md excludes. Negated policy
	// sentences ("does not promise a bug-bounty payment") stay allowed.
	operatorGuidanceReportingPromiseRE = regexp.MustCompile(
		`(?i)bug[- ]bounty (program|submission)|response time:\s*\d|\bsecurity@[a-z0-9.-]+`)
	// Reward amounts are protocol economics in most guides (VoteToEarn, staking
	// rewards); only in security guidance do they promise a bounty payment.
	operatorGuidanceSecurityRewardRE = regexp.MustCompile(`(?i)\brewards?\s*[:=]?\s*\d`)
	operatorGuidanceTemplateRewardRE = regexp.MustCompile(`(?i)bounty|reward`)
)

// GH-313C1 executable-command contract. server_lifecycle.go registers tx/query
// subcommands only for these modules.
var operatorGuidanceRegisteredModules = map[string]bool{"truedemocracy": true, "dex": true}

// GH308-owned guides still carry their own quarantined contract recipes on this
// base; GH313D removes this exclusion after GH-308 lands.
var operatorGuidanceCommandRuleExclusions = map[string]bool{
	"README.md":                                   true,
	"INSTALLATION.md":                             true,
	"docs/INSTALL.md":                             true,
	"docs/QUICKSTART.md":                          true,
	"docs/ARCHITECTURE.md":                        true,
	"docs/developers/README.md":                   true,
	"docs/developers/smart-contracts/cosmwasm.md": true,
	"wiki/develop/Module-Deep-Dive.md":            true,
}

var (
	operatorGuidanceModuleCmdRE      = regexp.MustCompile(`truerepublicd\s+(tx|query|q)\s+([a-z][a-z0-9-]*)`)
	operatorGuidanceGenesisCmdRE     = regexp.MustCompile(`truerepublicd\s+genesis\s+(add-genesis-account|gentx|collect-gentxs)\b`)
	operatorGuidanceRegisterRE       = regexp.MustCompile(`truedemocracy\s+register-validator((?:[ \t]+(?:\\\r?\n)?[ \t]*[^\s\\-][^\s\\]*)*)`)
	operatorGuidanceStakeArgRE       = regexp.MustCompile(`^(\d+upnyx|[<\[][^>\]]*stake[^>\]]*[>\]](upnyx)?|\$\{?[A-Za-z_]+\}?(upnyx)?|"\$[A-Za-z_]+")$`)
	operatorGuidanceWithdrawSuffixRE = regexp.MustCompile(`withdraw-stake\s+(\d+|[<\[][^>\]\s]*[>\]]|\$\{?[A-Za-z_]+\}?)upnyx\b`)
	operatorGuidanceCmdSubstRE       = regexp.MustCompile(`\$\([^)]*\)`)
	operatorGuidancePnyxAmountRE     = regexp.MustCompile(`\b\d[\d,_]*pnyx\b`)
	operatorGuidanceFictionalEnvRE   = regexp.MustCompile(`(?m)^\s*(EXTERNAL_IP|RPC_PORT|REST_PORT|GRPC_PORT|DB_BACKEND|PRUNING[A-Z_]*|PROMETHEUS_PORT|SNAPSHOT_URL|STATE_SYNC_[A-Z_]+|LOG_LEVEL|LOG_FORMAT)=`)
	operatorGuidanceComposeV1RE      = regexp.MustCompile(`\bdocker-compose\s+(?:--profile(?:=|\s+)\S+\s+)?(up|down|logs|ps|pull|restart|stop|start|exec|build|--version)\b`)
	operatorGuidancePublishedImageRE = regexp.MustCompile(`(?i)ghcr\.io/neabouli`)
	operatorGuidanceRootHomeRE       = regexp.MustCompile(`/root/\.truerepublic`)
	operatorGuidanceDefaultGrafanaRE = regexp.MustCompile("(?i)admin\\s*/\\s*admin|password:\\s*`admin`")
	operatorGuidanceFaucetRE         = regexp.MustCompile(`(?i)/faucet\b|#faucet\b`)
)

type operatorGuidanceInputs struct {
	maintained map[string]string // maintained guidance/config text -> content
	templates  map[string]string // .github/ISSUE_TEMPLATE/* except config.yml
	config     string            // .github/ISSUE_TEMPLATE/config.yml
}

func TestOperatorGuidanceTrustAndReportingContract(t *testing.T) {
	inputs := loadOperatorGuidanceInputs(t)
	if violations := operatorGuidanceViolations(inputs); len(violations) != 0 {
		t.Fatalf("operator guidance violations:\n- %s", strings.Join(violations, "\n- "))
	}

	rejected := map[string]func(*operatorGuidanceInputs){
		"fictional seed": func(in *operatorGuidanceInputs) {
			in.maintained["docs/node-operators/configuration/node-config.md"] += "\nseeds = \"id@seed1.TrueRepublic.network:26656\"\n"
		},
		"fictional status page": func(in *operatorGuidanceInputs) {
			in.maintained["docs/user-manual/troubleshooting.md"] += "\nCheck https://status.truerepublic.network\n"
		},
		"security email": func(in *operatorGuidanceInputs) {
			in.maintained["wiki/security/Security-Architecture.md"] += "\n- Email: security@example.org\n"
		},
		"response SLA": func(in *operatorGuidanceInputs) {
			in.maintained["wiki/security/Security-Architecture.md"] += "\n- Response time: 24 hours\n"
		},
		"bounty promise": func(in *operatorGuidanceInputs) {
			in.maintained["wiki/security/Best-Practices.md"] += "\nBug bounty program with Rewards: 100 PNYX\n"
		},
		"reward amount in security guidance": func(in *operatorGuidanceInputs) {
			in.maintained["wiki/security/Audit-Reports.md"] += "\nValid reports earn rewards: 500 PNYX\n"
		},
		"unregistered bank command": func(in *operatorGuidanceInputs) {
			in.maintained["wiki/operations/Validator-Guide.md"] += "\n```bash\ntruerepublicd query bank balances cosmos1abc\n```\n"
		},
		"unregistered staking command": func(in *operatorGuidanceInputs) {
			in.maintained["docs/VALIDATOR_GUIDE.md"] += "\ntruerepublicd tx staking edit-validator --from v\n"
		},
		"unregistered ibc command": func(in *operatorGuidanceInputs) {
			in.maintained["docs/IBC_RELAYER_SETUP.md"] += "\ntruerepublicd query ibc channel channels\n"
		},
		"unregistered genesis command": func(in *operatorGuidanceInputs) {
			in.maintained["docs/IBC_RELAYER_SETUP.md"] += "\ntruerepublicd genesis add-genesis-account a 1upnyx\n"
		},
		"two-argument register-validator": func(in *operatorGuidanceInputs) {
			in.maintained["wiki/users/Installation-Wizards.md"] += "\ntruerepublicd tx truedemocracy register-validator \\\n    my-domain \\\n    100000000000upnyx \\\n    --from validator\n"
		},
		"withdraw-stake with denom": func(in *operatorGuidanceInputs) {
			in.maintained["docs/validators/README.md"] += "\ntruerepublicd tx truedemocracy withdraw-stake 5upnyx\n"
		},
		"pnyx amount": func(in *operatorGuidanceInputs) {
			in.maintained["docs/API_REFERENCE.md"] += "\ndeposit-to-domain my-domain 1000pnyx\n"
		},
		"fictional env variable": func(in *operatorGuidanceInputs) {
			in.maintained["wiki/operations/Node-Setup.md"] += "\n```bash\nEXTERNAL_IP=1.2.3.4\n```\n"
		},
		"compose v1 profiled command": func(in *operatorGuidanceInputs) {
			in.maintained["docs/V0.4.0_OPTIONAL_INDEXER_STACK.md"] += "\ndocker-compose --profile indexer up -d\ndocker-compose --profile=indexer up -d\n"
		},
		"compose v1 command": func(in *operatorGuidanceInputs) {
			in.maintained["wiki/operations/Troubleshooting.md"] += "\ndocker-compose logs -f truerepublic-node\n"
		},
		"published image": func(in *operatorGuidanceInputs) {
			in.maintained["wiki/operations/Deployment-Options.md"] += "\nimage: ghcr.io/neabouli/truerepublic:latest\n"
		},
		"root home volume": func(in *operatorGuidanceInputs) {
			in.maintained["docs/DEPLOYMENT.md"] += "\n- ./data:/root/.truerepublic\n"
		},
		"default grafana password": func(in *operatorGuidanceInputs) {
			in.maintained["wiki/users/Installation-Wizards.md"] += "\nLogin admin/admin\n"
		},
		"faucet": func(in *operatorGuidanceInputs) {
			in.maintained["wiki/users/Installation-Wizards.md"] += "\nType `/faucet <address>`\n"
		},
		"public bounty template": func(in *operatorGuidanceInputs) {
			in.templates[".github/ISSUE_TEMPLATE/bug_bounty.md"] = "# Bug Bounty Submission\nBTC or PNYX address (for Reward)\n"
		},
		"security routing removed": func(in *operatorGuidanceInputs) {
			in.config = strings.ReplaceAll(in.config, operatorGuidanceSecurityPolicyURL, "https://example.org/report")
		},
		"missing caveat": func(in *operatorGuidanceInputs) {
			in.maintained["wiki/operations/Node-Setup.md"] = strings.ReplaceAll(in.maintained["wiki/operations/Node-Setup.md"], operatorGuidanceCaveatMarker, "")
		},
		"security contact without policy link": func(in *operatorGuidanceInputs) {
			in.maintained["wiki/security/Security-Architecture.md"] = strings.ReplaceAll(in.maintained["wiki/security/Security-Architecture.md"], operatorGuidanceSecurityPolicyURL, "")
		},
	}
	allowed := map[string]func(*operatorGuidanceInputs){
		"negated bounty policy sentence": func(in *operatorGuidanceInputs) {
			in.maintained["SECURITY.md"] += "\nThe repository does not promise a bug-bounty payment or response SLA.\n"
		},
		"ordinary bug report template": func(in *operatorGuidanceInputs) {
			in.templates[".github/ISSUE_TEMPLATE/bug_report.md"] += "\n## Steps to reproduce\n1. Run the node\n"
		},
		"protocol reward economics": func(in *operatorGuidanceInputs) {
			in.maintained["docs/user-manual/stones-voting-guide.md"] += "\nVoteToEarn reward = 5% of the domain treasury\n"
		},
		"protocol staking terminology": func(in *operatorGuidanceInputs) {
			in.maintained["wiki/users/How-It-Works.md"] += "\nStaking rewards follow the distribution of stones; no bank module query.\n"
		},
		"negated unregistered command": func(in *operatorGuidanceInputs) {
			in.maintained["wiki/operations/Validator-Guide.md"] += "\nThere is no `query bank` or `tx staking` command in truerepublicd.\n"
		},
		"compose file name and v2 command": func(in *operatorGuidanceInputs) {
			in.maintained["wiki/operations/Deployment-Options.md"] += "\nSee docker-compose.yml; run docker compose up -d\n"
		},
		"three-argument register-validator": func(in *operatorGuidanceInputs) {
			in.maintained["docs/VALIDATOR_GUIDE.md"] += "\ntruerepublicd tx truedemocracy register-validator \\\n    \"$PUBKEY_HEX\" \\\n    100000000000upnyx \\\n    my-domain \\\n    --from validator\n"
		},
		"native rpc and upnyx amounts": func(in *operatorGuidanceInputs) {
			in.maintained["docs/validators/README.md"] += "\ncurl http://127.0.0.1:26657/status\ntruerepublicd tx dex swap upnyx 1000 atom\n"
		},
		"operator-supplied endpoint placeholder": func(in *operatorGuidanceInputs) {
			in.maintained["wiki/operations/Node-Setup.md"] += "\nseeds = \"<node-id>@<qualified-seed-host>:26656\"\n"
		},
	}
	runOperatorGuidanceFixtures(t, inputs, rejected, true)
	runOperatorGuidanceFixtures(t, inputs, allowed, false)
}

// TestOperatorGuidanceExcludesHistoricalRecords proves the scan boundary: dated
// audit reports and append-only coordination logs may quote the historical
// domain and are never treated as maintained guidance.
func TestOperatorGuidanceExcludesHistoricalRecords(t *testing.T) {
	for _, path := range []string{
		"docs/community-audits/trr-integration-surfaces-audit-2026-09-15.md",
		"docs/agent-bridge/ACTION_LOG.md",
		"docs/archive/v0.3.0/notes.md",
		"BRIDGE.md",
	} {
		if operatorGuidanceMaintained(path) {
			t.Fatalf("historical record %s treated as maintained guidance", path)
		}
	}
	for _, path := range []string{"wiki/operations/Node-Setup.md", "docs/FAQ.md", ".github/SECURITY.md", "README.md"} {
		if !operatorGuidanceMaintained(path) {
			t.Fatalf("maintained guidance %s excluded from the scan", path)
		}
	}
}

func runOperatorGuidanceFixtures(t *testing.T, inputs operatorGuidanceInputs, fixtures map[string]func(*operatorGuidanceInputs), wantViolation bool) {
	t.Helper()
	names := make([]string, 0, len(fixtures))
	for name := range fixtures {
		names = append(names, name)
	}
	sort.Strings(names)
	prefix := "allows "
	if wantViolation {
		prefix = "rejects "
	}
	for _, name := range names {
		t.Run(prefix+name, func(t *testing.T) {
			clone := cloneOperatorGuidanceInputs(inputs)
			fixtures[name](&clone)
			violations := operatorGuidanceViolations(clone)
			if wantViolation && len(violations) == 0 {
				t.Fatal("trust/reporting drift accepted")
			}
			if !wantViolation && len(violations) != 0 {
				t.Fatalf("harmless change rejected:\n- %s", strings.Join(violations, "\n- "))
			}
		})
	}
}

// operatorGuidanceMaintained reports whether a repository path is maintained
// guidance. Historical audits, archives and append-only coordination records
// are excluded.
func operatorGuidanceMaintained(path string) bool {
	path = filepath.ToSlash(path)
	for _, prefix := range []string{"docs/community-audits/", "docs/agent-bridge/", "docs/archive/"} {
		if strings.HasPrefix(path, prefix) {
			return false
		}
	}
	// Release notes describe a past version and are historical records.
	if strings.HasPrefix(path, "RELEASE_NOTES_") {
		return false
	}
	return path != "BRIDGE.md"
}

// operatorGuidanceSecurityScope reports whether a path is security guidance,
// where a reward amount can only mean a bounty promise.
func operatorGuidanceSecurityScope(path string) bool {
	return strings.HasPrefix(path, "wiki/security/") || strings.HasPrefix(path, "docs/security/") ||
		path == "SECURITY.md" || path == ".github/SECURITY.md"
}

func loadOperatorGuidanceInputs(t *testing.T) operatorGuidanceInputs {
	t.Helper()
	in := operatorGuidanceInputs{maintained: map[string]string{}, templates: map[string]string{}}
	textExt := map[string]bool{".md": true, ".yml": true, ".yaml": true, ".json": true, ".html": true, ".conf": true, ".toml": true}
	add := func(path string) {
		if !operatorGuidanceMaintained(path) || !textExt[filepath.Ext(path)] {
			return
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		in.maintained[filepath.ToSlash(path)] = string(content)
	}
	for _, root := range []string{"docs", "wiki", ".github", "configs", "nginx"} {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if !operatorGuidanceMaintained(path + "/") {
					return filepath.SkipDir
				}
				return nil
			}
			add(path)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	rootDocs, err := filepath.Glob("*.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range rootDocs {
		add(path)
	}
	templates, err := filepath.Glob(".github/ISSUE_TEMPLATE/*")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range templates {
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if filepath.Base(path) == "config.yml" {
			in.config = string(content)
			continue
		}
		in.templates[filepath.ToSlash(path)] = string(content)
	}
	return in
}

func operatorGuidanceViolations(in operatorGuidanceInputs) []string {
	var out []string
	for _, path := range sortedOperatorGuidanceKeys(in.maintained) {
		content := in.maintained[path]
		if match := operatorGuidanceFictionalDomainRE.FindString(content); match != "" {
			out = append(out, path+": depends on unregistered infrastructure ("+match+")")
		}
		if match := operatorGuidanceReportingPromiseRE.FindString(content); match != "" {
			out = append(out, path+": reporting promise or contact outside SECURITY.md ("+match+")")
		}
		if strings.HasSuffix(path, ".md") && !operatorGuidanceCommandRuleExclusions[path] {
			out = append(out, operatorGuidanceCommandViolations(path, content)...)
		}
		if operatorGuidanceSecurityScope(path) {
			if match := operatorGuidanceSecurityRewardRE.FindString(content); match != "" {
				out = append(out, path+": security guidance promises a reward ("+match+")")
			}
		}
	}
	for _, path := range sortedOperatorGuidanceKeys(in.templates) {
		if match := operatorGuidanceTemplateRewardRE.FindString(in.templates[path]); match != "" {
			out = append(out, path+": public issue template invites bounty/reward reports ("+match+")")
		}
	}
	if !strings.Contains(in.config, "blank_issues_enabled: true") || !strings.Contains(in.config, "url: "+operatorGuidanceSecurityPolicyURL) {
		out = append(out, ".github/ISSUE_TEMPLATE/config.yml must keep community issues open and route security reports to "+operatorGuidanceSecurityPolicyURL)
	}
	if !strings.Contains(in.maintained["wiki/security/Security-Architecture.md"], operatorGuidanceSecurityPolicyURL) {
		out = append(out, "wiki/security/Security-Architecture.md must point security reporting to SECURITY.md")
	}
	for _, path := range operatorGuidanceCaveatGuides {
		if !strings.Contains(in.maintained[path], operatorGuidanceCaveatMarker) {
			out = append(out, path+": missing non-production caveat")
		}
	}
	return out
}

// operatorGuidanceCommandViolations checks one maintained guide against the
// registered CLI and repository-owned runtime configuration.
func operatorGuidanceCommandViolations(path, content string) []string {
	var out []string
	add := func(reason, match string) { out = append(out, path+": "+reason+" ("+match+")") }
	for _, match := range operatorGuidanceModuleCmdRE.FindAllStringSubmatch(content, -1) {
		if !operatorGuidanceRegisteredModules[match[2]] {
			add("command for an unregistered module", match[0])
		}
	}
	if match := operatorGuidanceGenesisCmdRE.FindString(content); match != "" {
		add("unregistered genesis command", match)
	}
	for _, match := range operatorGuidanceRegisterRE.FindAllStringSubmatch(content, -1) {
		normalized := operatorGuidanceCmdSubstRE.ReplaceAllString(match[1], "$$SUBST")
		args := strings.Fields(strings.ReplaceAll(normalized, "\\", " "))
		if len(args) == 0 || strings.HasPrefix(args[0], "[pubkey") && len(args) < 3 {
			continue // prose mention or the bare usage line
		}
		if len(args) < 3 || !operatorGuidanceStakeArgRE.MatchString(args[1]) {
			add("register-validator must take [pubkey-hex] [stake] [domain]", strings.Join(args, " "))
		}
	}
	checks := []struct {
		re     *regexp.Regexp
		reason string
	}{
		{operatorGuidanceWithdrawSuffixRE, "withdraw-stake takes a plain integer upnyx amount"},
		{operatorGuidancePnyxAmountRE, "amount must use the upnyx base denomination"},
		{operatorGuidanceFictionalEnvRE, "environment variable not defined by .env.example"},
		{operatorGuidanceComposeV1RE, "Compose v1 command; use docker compose"},
		{operatorGuidancePublishedImageRE, "no published image exists"},
		{operatorGuidanceRootHomeRE, "the image runs as non-root /home/truerepublic"},
		{operatorGuidanceDefaultGrafanaRE, "Grafana has no default password"},
		{operatorGuidanceFaucetRE, "no faucet is operated"},
	}
	for _, check := range checks {
		if match := check.re.FindString(content); match != "" {
			add(check.reason, match)
		}
	}
	return out
}

func cloneOperatorGuidanceInputs(in operatorGuidanceInputs) operatorGuidanceInputs {
	clone := func(source map[string]string) map[string]string {
		out := make(map[string]string, len(source))
		for key, value := range source {
			out[key] = value
		}
		return out
	}
	return operatorGuidanceInputs{maintained: clone(in.maintained), templates: clone(in.templates), config: in.config}
}

func sortedOperatorGuidanceKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
