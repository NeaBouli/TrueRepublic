package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const releaseCompatibilityPath = "configs/release/compatibility.json"

type releaseCompatibilityContract struct {
	Schema           string              `json:"schema"`
	Candidate        releaseCandidate    `json:"candidate"`
	SupportedTargets []releaseTarget     `json:"supported_targets"`
	Toolchains       releaseToolchains   `json:"toolchains"`
	MaintainedClient releaseClient       `json:"maintained_client"`
	InstallLifecycle releaseLifecycle    `json:"install_lifecycle"`
	GovernedUpgrade  releaseUpgrade      `json:"governed_upgrade"`
	Protocol         releaseProtocol     `json:"protocol"`
	Changes          []releaseChange     `json:"changes"`
	Historical       []releaseHistorical `json:"historical"`
	Documents        releaseDocuments    `json:"documents"`
}
type releaseCandidate struct {
	ID                  string   `json:"id"`
	RecoveryLineLabel   string   `json:"recovery_line_label"`
	SourceIdentity      string   `json:"source_identity"`
	BinaryVersionLDFlag string   `json:"binary_version_ldflag"`
	DefaultVersion      string   `json:"default_version"`
	ReleaseStatus       string   `json:"release_status"`
	Unreleased          *bool    `json:"unreleased"`
	Production          *bool    `json:"production"`
	Tagged              *bool    `json:"tagged"`
	Published           *bool    `json:"published"`
	Signed              *bool    `json:"signed"`
	Evidence            []string `json:"evidence"`
}

type releaseTarget struct {
	ID       string   `json:"id"`
	Runtime  string   `json:"runtime"`
	Evidence []string `json:"evidence"`
}
type releaseToolchains struct {
	Go       string   `json:"go"`
	Node     string   `json:"node"`
	NPM      string   `json:"npm"`
	Rust     string   `json:"rust"`
	Evidence []string `json:"evidence"`
}
type releaseClient struct {
	Path           string   `json:"path"`
	VersionLabel   string   `json:"version_label"`
	PackageVersion string   `json:"package_version"`
	Status         string   `json:"status"`
	Evidence       []string `json:"evidence"`
}
type releaseLifecycle struct {
	Schema     string   `json:"schema"`
	Contract   string   `json:"contract"`
	Operations []string `json:"operations"`
	Evidence   []string `json:"evidence"`
}
type releaseUpgrade struct {
	PlanName  string   `json:"plan_name"`
	Authority string   `json:"authority"`
	Approval  string   `json:"approval"`
	Boundary  string   `json:"boundary"`
	Evidence  []string `json:"evidence"`
}
type releaseProtocol struct {
	Supported   []releaseSurface `json:"supported"`
	Unsupported []releaseSurface `json:"unsupported"`
}
type releaseSurface struct {
	ID       string   `json:"id"`
	Summary  string   `json:"summary"`
	Reason   string   `json:"reason"`
	Evidence []string `json:"evidence"`
}
type releaseChange struct {
	ID              string   `json:"id"`
	Category        string   `json:"category"`
	Summary         string   `json:"summary"`
	OperatorActions []string `json:"operator_actions"`
	UserActions     []string `json:"user_actions"`
	Evidence        []string `json:"evidence"`
}
type releaseHistorical struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	Marker string `json:"marker"`
}
type releaseDocuments struct {
	ReleaseNotes  string `json:"release_notes"`
	Compatibility string `json:"compatibility"`
}

type releaseBuildContract struct {
	GoVersion string `json:"go_version"`
	SourceRef struct {
		Kind string `json:"kind"`
	} `json:"source_ref"`
	Targets []struct {
		ID string `json:"id"`
	} `json:"targets"`
	BuildFlags struct {
		LDFlags []string `json:"ldflags"`
	} `json:"build_flags"`
}

func TestReleaseCompatibilityRepositoryContract(t *testing.T) {
	raw := []byte(readRepositoryFile(t, releaseCompatibilityPath))
	contract, err := decodeReleaseCompatibility(raw)
	if err != nil {
		t.Fatalf("decode contract: %v", err)
	}
	if violations := releaseCompatibilityViolations(contract); len(violations) != 0 {
		t.Fatalf("contract violations:\n- %s", strings.Join(violations, "\n- "))
	}
	if len(evidenceViolations([]string{"."})) == 0 {
		t.Fatal("repository root accepted as evidence")
	}
	if sourceIdentityLDFlagConfigured(contract.Candidate, contract.Candidate.SourceIdentity, []string{"-s"}) {
		t.Fatal("missing source-identity ldflag accepted")
	}
	if sourceIdentityLDFlagConfigured(contract.Candidate, "tag", []string{"main.version={{source_ref}}"}) {
		t.Fatal("mismatched source identity accepted")
	}
	for _, doc := range []string{contract.Documents.ReleaseNotes, contract.Documents.Compatibility} {
		body := readRepositoryFile(t, doc)
		for _, label := range []string{"production", "tagged", "published", "signed"} {
			documentLabel := strings.ToUpper(label[:1]) + label[1:]
			falseClaim := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(label) + `:\s*` + "`false`")
			contradictory := falseClaim.ReplaceAllString(body, documentLabel+": `true`")
			if contradictory == body || len(documentStatusViolations(contradictory, contract.Candidate)) == 0 {
				t.Fatalf("%s true claim accepted in %s", label, doc)
			}
		}
	}

	for name, mutated := range map[string][]byte{
		"malformed":     []byte(`{"schema":`),
		"trailing":      append(append([]byte{}, raw...), []byte(` {}`)...),
		"duplicate key": bytes.Replace(raw, []byte(`"schema":`), []byte(`"schema":"duplicate","schema":`), 1),
		"unknown field": bytes.Replace(raw, []byte(`"candidate":`), []byte(`"unknown":true,"candidate":`), 1),
	} {
		t.Run("rejects "+name, func(t *testing.T) {
			if _, err := decodeReleaseCompatibility(mutated); err == nil {
				t.Fatal("invalid JSON accepted")
			}
		})
	}

	mutations := map[string]func(*releaseCompatibilityContract){
		"production claim":    func(c *releaseCompatibilityContract) { v := true; c.Candidate.Production = &v },
		"published claim":     func(c *releaseCompatibilityContract) { v := true; c.Candidate.Published = &v },
		"missing safety flag": func(c *releaseCompatibilityContract) { c.Candidate.Signed = nil },
		"stale target":        func(c *releaseCompatibilityContract) { c.SupportedTargets[0].ID = "darwin-amd64" },
		"stale runtime":       func(c *releaseCompatibilityContract) { c.SupportedTargets[0].Runtime = "darwin" },
		"stale Go":            func(c *releaseCompatibilityContract) { c.Toolchains.Go = "0.0.0" },
		"stale client":        func(c *releaseCompatibilityContract) { c.MaintainedClient.PackageVersion = "0.0.0" },
		"stale lifecycle":     func(c *releaseCompatibilityContract) { c.InstallLifecycle.Schema = "v0" },
		"stale upgrade":       func(c *releaseCompatibilityContract) { c.GovernedUpgrade.PlanName = "v0.4.2" },
		"unsafe evidence":     func(c *releaseCompatibilityContract) { c.Changes[0].Evidence = []string{"../secret"} },
		"missing evidence":    func(c *releaseCompatibilityContract) { c.Changes[0].Evidence = []string{"does-not-exist"} },
		"duplicate change":    func(c *releaseCompatibilityContract) { c.Changes[1].ID = c.Changes[0].ID },
		"unknown category":    func(c *releaseCompatibilityContract) { c.Changes[0].Category = "maybe" },
		"unsupported surface promoted": func(c *releaseCompatibilityContract) {
			c.Protocol.Supported = append(c.Protocol.Supported, c.Protocol.Unsupported[0])
			c.Protocol.Unsupported = c.Protocol.Unsupported[1:]
		},
		"missing breaking action": func(c *releaseCompatibilityContract) {
			c.Changes[0].OperatorActions = nil
			c.Changes[0].UserActions = nil
		},
	}
	for name, mutate := range mutations {
		t.Run("rejects "+name, func(t *testing.T) {
			clone := cloneReleaseCompatibility(t, contract)
			mutate(&clone)
			if violations := releaseCompatibilityViolations(clone); len(violations) == 0 {
				t.Fatal("drift accepted")
			}
		})
	}
}

func decodeReleaseCompatibility(raw []byte) (releaseCompatibilityContract, error) {
	if err := rejectDuplicateJSONKeys(raw); err != nil {
		return releaseCompatibilityContract{}, err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var c releaseCompatibilityContract
	if err := dec.Decode(&c); err != nil {
		return c, err
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return c, fmt.Errorf("trailing JSON")
	}
	return c, nil
}

func rejectDuplicateJSONKeys(raw []byte) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	var walk func() error
	walk = func() error {
		tok, err := d.Token()
		if err != nil {
			return err
		}
		switch v := tok.(type) {
		case json.Delim:
			if v == '{' {
				seen := map[string]bool{}
				for d.More() {
					k, err := d.Token()
					if err != nil {
						return err
					}
					key, ok := k.(string)
					if !ok {
						return fmt.Errorf("object key is not a string")
					}
					if seen[key] {
						return fmt.Errorf("duplicate key %q", key)
					}
					seen[key] = true
					if err := walk(); err != nil {
						return err
					}
				}
				_, err = d.Token()
				return err
			}
			if v == '[' {
				for d.More() {
					if err := walk(); err != nil {
						return err
					}
				}
				_, err = d.Token()
				return err
			}
		}
		return nil
	}
	if err := walk(); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return fmt.Errorf("trailing JSON")
	}
	return nil
}

func cloneReleaseCompatibility(t *testing.T, c releaseCompatibilityContract) releaseCompatibilityContract {
	t.Helper()
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	clone, err := decodeReleaseCompatibility(b)
	if err != nil {
		t.Fatal(err)
	}
	return clone
}

func releaseCompatibilityViolations(c releaseCompatibilityContract) []string {
	var out []string
	add := func(ok bool, msg string) {
		if !ok {
			out = append(out, msg)
		}
	}
	add(c.Schema == "truerepublic.release-compatibility/v1", "schema mismatch")
	add(c.Candidate.ID != "", "candidate ID missing")
	add(c.Candidate.Unreleased != nil && *c.Candidate.Unreleased, "candidate must be explicitly unreleased")
	for label, v := range map[string]*bool{"production": c.Candidate.Production, "tagged": c.Candidate.Tagged, "published": c.Candidate.Published, "signed": c.Candidate.Signed} {
		add(v != nil && !*v, label+" must be explicitly false")
	}
	add(c.Candidate.SourceIdentity == "git-commit" && c.Candidate.BinaryVersionLDFlag == "main.version" && c.Candidate.DefaultVersion == version, "source identity mismatch")

	var status struct {
		Version       string `json:"version"`
		ReleaseStatus string `json:"release_status"`
		Recovery      struct {
			ProductionReady    bool   `json:"production_ready"`
			CanonicalWebClient string `json:"canonical_web_client"`
		} `json:"recovery"`
		Rollout struct {
			ProductionReady bool `json:"production_ready"`
		} `json:"rollout"`
		WebClient struct {
			Version string `json:"version"`
		} `json:"web_client"`
	}
	mustReadJSON(&out, "docs/status.json", &status)
	add(c.Candidate.RecoveryLineLabel == status.Version && c.Candidate.ReleaseStatus == status.ReleaseStatus && !status.Recovery.ProductionReady && !status.Rollout.ProductionReady, "public status mismatch")
	var platform struct {
		Tools     map[string]string `json:"tools"`
		Platforms []struct {
			ID string `json:"id"`
		} `json:"platforms"`
	}
	mustReadJSON(&out, "configs/release/tool-platform.json", &platform)
	var lifecycle struct {
		Schema           string            `json:"schema"`
		SupportedTargets []string          `json:"supported_targets"`
		TargetRuntimes   map[string]string `json:"target_runtimes"`
	}
	add(c.InstallLifecycle.Contract == "configs/release/install-lifecycle.json", "lifecycle contract path mismatch")
	mustReadJSON(&out, "configs/release/install-lifecycle.json", &lifecycle)
	var gates struct {
		Toolchains map[string]string `json:"toolchains"`
		Tools      map[string]string `json:"tools"`
	}
	mustReadJSON(&out, "configs/security/gates.json", &gates)
	var build releaseBuildContract
	mustReadJSON(&out, "configs/build/deterministic-linux-daemon.json", &build)
	add(sourceIdentityLDFlagConfigured(c.Candidate, build.SourceRef.Kind, build.BuildFlags.LDFlags), "source identity ldflag mismatch")
	ids := func(ts []releaseTarget) []string {
		r := make([]string, len(ts))
		for i, t := range ts {
			r[i] = t.ID
		}
		sort.Strings(r)
		return r
	}
	want := append([]string(nil), lifecycle.SupportedTargets...)
	sort.Strings(want)
	add(strings.Join(ids(c.SupportedTargets), ",") == strings.Join(want, ","), "target set mismatch")
	var platformIDs, buildIDs []string
	for _, target := range platform.Platforms {
		platformIDs = append(platformIDs, target.ID)
	}
	for _, target := range build.Targets {
		buildIDs = append(buildIDs, target.ID)
	}
	sort.Strings(platformIDs)
	sort.Strings(buildIDs)
	add(strings.Join(platformIDs, ",") == strings.Join(want, ",") && strings.Join(buildIDs, ",") == strings.Join(want, ","), "platform/build target set mismatch")
	for _, t := range c.SupportedTargets {
		add(t.Runtime == lifecycle.TargetRuntimes[t.ID], "target runtime mismatch")
		out = append(out, evidenceViolations(t.Evidence)...)
	}
	add(c.Toolchains.Go == build.GoVersion && c.Toolchains.Go == gates.Toolchains["go"] && c.Toolchains.Node == platform.Tools["node"] && c.Toolchains.Node == gates.Toolchains["node"] && c.Toolchains.NPM == platform.Tools["npm"] && c.Toolchains.NPM == gates.Tools["npm"] && c.Toolchains.Rust == gates.Toolchains["rust"], "toolchain mismatch")
	var pkg struct {
		Version string `json:"version"`
	}
	mustReadJSON(&out, "client-web/package.json", &pkg)
	add(c.MaintainedClient.Path == status.Recovery.CanonicalWebClient && c.MaintainedClient.PackageVersion == pkg.Version && c.MaintainedClient.VersionLabel == status.WebClient.Version, "client identity mismatch")
	add(c.MaintainedClient.Status == "maintained_beta", "maintained client status mismatch")
	add(c.InstallLifecycle.Schema == lifecycle.Schema, "lifecycle schema mismatch")
	add(releaseStringsEqual(c.InstallLifecycle.Operations, []string{"install", "status", "pre-start", "upgrade", "rollback", "uninstall"}), "lifecycle operations mismatch")
	add(c.GovernedUpgrade.PlanName == governedUpgradePlanV041 && strings.Contains(strings.Join(build.BuildFlags.LDFlags, " "), "main.upgradePlan="+c.GovernedUpgrade.PlanName), "governed upgrade mismatch")
	add(c.GovernedUpgrade.Authority != "" && c.GovernedUpgrade.Approval != "" && c.GovernedUpgrade.Boundary != "", "governed upgrade statement incomplete")
	validCategories := map[string]bool{"breaking_chain_state": true, "breaking_api": true, "breaking_client": true, "breaking_operations": true, "compatible": true, "unsupported_surface": true}
	seen := map[string]bool{}
	idRE := regexp.MustCompile(`^COMPAT-[0-9]{3}$`)
	add(len(c.Changes) > 0, "compatibility changes missing")
	for _, ch := range c.Changes {
		add(idRE.MatchString(ch.ID) && !seen[ch.ID], "invalid or duplicate change ID")
		seen[ch.ID] = true
		add(validCategories[ch.Category], "unknown category")
		if strings.HasPrefix(ch.Category, "breaking_") {
			add(len(ch.OperatorActions)+len(ch.UserActions) > 0, "breaking change lacks action")
		}
		out = append(out, evidenceViolations(ch.Evidence)...)
	}
	surfaceSeen := map[string]bool{}
	for _, s := range append(c.Protocol.Supported, c.Protocol.Unsupported...) {
		add(s.ID != "" && !surfaceSeen[s.ID] && (s.Summary != "" || s.Reason != ""), "invalid or duplicate protocol surface")
		surfaceSeen[s.ID] = true
		out = append(out, evidenceViolations(s.Evidence)...)
	}
	unsupportedSeen := map[string]bool{}
	for _, s := range c.Protocol.Unsupported {
		unsupportedSeen[s.ID] = true
	}
	for _, required := range []string{"x_staking", "x_distribution", "ibc_staking", "ibc_client_upgrade", "external_relayers", "anonymous_zkp_submission", "legacy_custom_abci_queries"} {
		add(unsupportedSeen[required], "required unsupported surface missing")
	}
	for _, paths := range [][]string{c.Candidate.Evidence, c.Toolchains.Evidence, c.MaintainedClient.Evidence, c.InstallLifecycle.Evidence, c.GovernedUpgrade.Evidence} {
		out = append(out, evidenceViolations(paths)...)
	}
	add(len(c.Historical) == 2, "historical quarantine entries missing")
	historicalSeen := map[string]bool{}
	for _, h := range c.Historical {
		historicalSeen[h.ID] = true
		pathViolations := evidenceViolations([]string{h.Path})
		out = append(out, pathViolations...)
		if len(pathViolations) != 0 {
			continue
		}
		body, err := os.ReadFile(h.Path)
		add(err == nil && strings.Contains(string(body), h.Marker), "historical marker mismatch")
	}
	add(historicalSeen["daemon-v0.3.0-draft"] && historicalSeen["client-v0.4.0-changelog"], "required historical IDs missing")
	for _, doc := range []string{c.Documents.ReleaseNotes, c.Documents.Compatibility} {
		pathViolations := evidenceViolations([]string{doc})
		out = append(out, pathViolations...)
		if len(pathViolations) != 0 {
			continue
		}
		body, err := os.ReadFile(doc)
		add(err == nil, "release document missing")
		if err == nil {
			txt := string(body)
			add(strings.Contains(txt, c.Schema), "document schema missing")
			out = append(out, documentStatusViolations(txt, c.Candidate)...)
			for _, ch := range c.Changes {
				add(strings.Contains(txt, ch.ID), "document change ID missing")
			}
		}
	}
	for _, workflow := range []string{".github/workflows/docs-check.yml", ".github/workflows/reproducible-daemon.yml"} {
		body, err := os.ReadFile(workflow)
		add(err == nil && strings.Contains(string(body), "make release-compatibility-contract-test"), "protected workflow compatibility gate missing")
	}
	return out
}

func mustReadJSON(out *[]string, path string, dst any) {
	b, err := os.ReadFile(path)
	if err != nil {
		*out = append(*out, "cannot read "+path)
		return
	}
	if err := json.Unmarshal(b, dst); err != nil {
		*out = append(*out, "cannot parse "+path)
	}
}
func evidenceViolations(paths []string) []string {
	var out []string
	if len(paths) == 0 {
		return []string{"missing evidence"}
	}
	for _, p := range paths {
		clean := filepath.Clean(p)
		if p == "" || p == "." || clean != p || filepath.IsAbs(p) || strings.Contains(p, "\\") || strings.HasPrefix(p, "../") {
			out = append(out, "unsafe evidence path")
			continue
		}
		current := ""
		unsafeLink := false
		for _, part := range strings.Split(p, string(filepath.Separator)) {
			current = filepath.Join(current, part)
			info, err := os.Lstat(current)
			if err != nil {
				out = append(out, "missing evidence path")
				unsafeLink = true
				break
			}
			if info.Mode()&os.ModeSymlink != 0 {
				out = append(out, "symlink evidence path")
				unsafeLink = true
				break
			}
		}
		if unsafeLink {
			continue
		}
		if _, err := os.Stat(p); err != nil {
			out = append(out, "missing evidence path")
		}
	}
	return out
}

func sourceIdentityLDFlagConfigured(candidate releaseCandidate, sourceKind string, ldflags []string) bool {
	if candidate.SourceIdentity == "" || sourceKind != candidate.SourceIdentity {
		return false
	}
	want := candidate.BinaryVersionLDFlag + "={{source_ref}}"
	for _, flag := range ldflags {
		if flag == want {
			return true
		}
	}
	return false
}

func documentStatusViolations(text string, candidate releaseCandidate) []string {
	var out []string
	lower := strings.ToLower(text)
	if candidate.Unreleased == nil || !*candidate.Unreleased || !regexp.MustCompile(`\bunreleased\b`).MatchString(lower) || regexp.MustCompile(`\breleased\b`).MatchString(lower) {
		out = append(out, "document release status mismatch")
	}
	claims := map[string]*bool{
		"production": candidate.Production,
		"tagged":     candidate.Tagged,
		"published":  candidate.Published,
		"signed":     candidate.Signed,
	}
	for label, expected := range claims {
		if expected == nil {
			out = append(out, "document "+label+" status missing")
			continue
		}
		value := "false"
		if *expected {
			value = "true"
		}
		claim := regexp.MustCompile(`\b` + regexp.QuoteMeta(label) + `:\s*` + "`?" + value + `\b` + "`?")
		contradiction := "false"
		if value == "false" {
			contradiction = "true"
		}
		forbidden := regexp.MustCompile(`\b` + regexp.QuoteMeta(label) + `:\s*` + "`?" + contradiction + `\b` + "`?")
		if !claim.MatchString(lower) || forbidden.MatchString(lower) {
			out = append(out, "document "+label+" status mismatch")
		}
	}
	if strings.Contains(text, "Release Date:") || strings.Contains(text, "Production Ready") {
		out = append(out, "document contains forbidden release claim")
	}
	return out
}

func releaseStringsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
