package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type threatModelContract struct {
	Model           threatModelIdentity   `json:"model"`
	ProductionReady *bool                 `json:"production_ready"`
	SafetyBoundary  string                `json:"safety_boundary"`
	Assets          []threatModelAsset    `json:"assets"`
	Actors          []threatModelActor    `json:"actors"`
	TrustBoundaries []threatModelBoundary `json:"trust_boundaries"`
	Assumptions     []string              `json:"assumptions"`
	NonGoals        []string              `json:"non_goals"`
	Threats         []threatModelThreat   `json:"threats"`
}

type threatModelIdentity struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Updated string `json:"updated"`
}

type threatModelAsset struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

type threatModelActor struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
}

type threatModelBoundary struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	DataFlows   []string `json:"data_flows"`
}

type threatModelThreat struct {
	ID               string   `json:"id"`
	Domain           string   `json:"domain"`
	Severity         string   `json:"severity"`
	Likelihood       string   `json:"likelihood"`
	Status           string   `json:"status"`
	AttackPath       string   `json:"attack_path"`
	Impact           string   `json:"impact"`
	Controls         []string `json:"controls"`
	Evidence         []string `json:"evidence"`
	ResidualRisk     string   `json:"residual_risk"`
	ResidualSeverity string   `json:"residual_severity"`
	Owner            string   `json:"owner"`
	NextGate         string   `json:"next_gate"`
	UmbrellaIssue    string   `json:"umbrella_issue"`
}

const (
	threatModelJSONPath = "configs/security/threat-model.json"
	threatModelDocPath  = "docs/security/THREAT_MODEL.md"
	threatModelVersion  = "truerepublic.threat-model/v1"
)

var threatModelIDPattern = regexp.MustCompile(`^TM-[A-Z]{3}-[0-9]{3}$`)
var threatModelDatePattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)

var threatModelDomains = []string{
	"consensus_p2p",
	"governance_identity",
	"token_treasury_dex",
	"zkp_privacy",
	"ibc_upgrades",
	"client_wallet_rpc",
	"operations_observability",
	"dependencies_ci",
	"release_artifacts",
}

var threatModelSeverities = []string{"critical", "high", "medium", "low"}
var threatModelLikelihoods = []string{"high", "medium", "low"}
var threatModelStatuses = []string{"mitigated", "deferred", "blocked", "accepted", "not_applicable"}
var threatModelUmbrellas = []string{"GH-7", "GH-29"}
var threatModelActorKinds = []string{"adversary", "legitimate"}

func TestThreatModelRepositoryContract(t *testing.T) {
	raw := readRepositoryFile(t, threatModelJSONPath)
	contract := parseThreatModelContract(t, []byte(raw))

	if violations := threatModelViolations(contract); len(violations) != 0 {
		t.Fatalf("threat model contract violations:\n- %s", strings.Join(violations, "\n- "))
	}

	for _, forbidden := range []string{
		"mnemonic", "private_key", "priv_validator", "password", "BEGIN ",
		"${", "/home/", "/Users/", "http://", "https://",
	} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("threat model register contains forbidden secret/host shape %q", forbidden)
		}
	}
	doc := readRepositoryFile(t, threatModelDocPath)
	for _, required := range []string{
		threatModelVersion,
		"production_ready=false",
		"not an independent external audit",
		"Review and update triggers",
		"Verified repository controls",
	} {
		if !strings.Contains(doc, required) {
			t.Fatalf("threat model document must contain %q", required)
		}
	}
	for _, domain := range threatModelDomains {
		if !strings.Contains(doc, domain) {
			t.Fatalf("threat model document must cover domain %q", domain)
		}
	}
	jsonIDs := make(map[string]bool, len(contract.Threats))
	for _, threat := range contract.Threats {
		jsonIDs[threat.ID] = true
		if !strings.Contains(doc, threat.ID) {
			t.Fatalf("threat model document must present threat %s", threat.ID)
		}
	}
	for _, docID := range threatModelIDPattern.FindAllString(doc, -1) {
		if !jsonIDs[docID] {
			t.Fatalf("threat model document presents %s which is absent from the JSON register", docID)
		}
	}

	mutations := threatModelMutations(raw)
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			mutated := mutation.apply([]byte(raw))
			parsed, err := decodeThreatModel(mutated)
			var violations []string
			if err != nil {
				violations = []string{fmt.Sprintf("parse threat model: %v", err)}
			} else {
				violations = threatModelViolations(parsed)
			}
			if !strings.Contains(strings.Join(violations, "\n"), mutation.want) {
				t.Fatalf("mutation %q missing rejection %q in %v", mutation.name, mutation.want, violations)
			}
		})
	}
}

type threatModelMutation struct {
	name  string
	apply func([]byte) []byte
	want  string
}

func threatModelMutations(raw string) []threatModelMutation {
	edit := func(name, want string, mutate func(map[string]interface{})) threatModelMutation {
		return threatModelMutation{
			name: name,
			apply: func(original []byte) []byte {
				var root map[string]interface{}
				if err := json.Unmarshal(original, &root); err != nil {
					panic(err)
				}
				mutate(root)
				rendered, err := json.Marshal(root)
				if err != nil {
					panic(err)
				}
				return rendered
			},
			want: want,
		}
	}
	threat := func(root map[string]interface{}, id string) map[string]interface{} {
		for _, candidate := range root["threats"].([]interface{}) {
			entry := candidate.(map[string]interface{})
			if entry["id"] == id {
				return entry
			}
		}
		panic("missing mutation target " + id)
	}
	return []threatModelMutation{
		{
			name: "rejects malformed JSON",
			apply: func(original []byte) []byte {
				return original[:len(original)/2]
			},
			want: "parse threat model",
		},
		{
			name: "rejects trailing JSON value",
			apply: func(original []byte) []byte {
				return append(append([]byte{}, original...), []byte(`{"unexpected":true}`)...)
			},
			want: "trailing",
		},
		edit("rejects duplicate threat IDs", "duplicate threat ID", func(root map[string]interface{}) {
			threat(root, "TM-CON-002")["id"] = threat(root, "TM-CON-001")["id"]
		}),
		edit("rejects unknown domain", "unknown domain", func(root map[string]interface{}) {
			threat(root, "TM-CON-001")["domain"] = "quantum"
		}),
		edit("rejects unknown status", "unknown status", func(root map[string]interface{}) {
			threat(root, "TM-CON-001")["status"] = "resolved"
		}),
		edit("rejects unknown severity", "unknown severity", func(root map[string]interface{}) {
			threat(root, "TM-CON-001")["severity"] = "catastrophic"
		}),
		edit("rejects unknown likelihood", "unknown likelihood", func(root map[string]interface{}) {
			threat(root, "TM-CON-001")["likelihood"] = "certain"
		}),
		edit("rejects missing evidence", "at least one evidence path", func(root map[string]interface{}) {
			threat(root, "TM-CON-001")["evidence"] = []interface{}{}
		}),
		edit("rejects nonexistent evidence path", "does not exist", func(root map[string]interface{}) {
			threat(root, "TM-CON-001")["evidence"] = []interface{}{"docs/security/missing-evidence.md"}
		}),
		edit("rejects parent-traversal evidence path", "unsafe evidence path", func(root map[string]interface{}) {
			threat(root, "TM-CON-001")["evidence"] = []interface{}{"../BRIDGE.md"}
		}),
		edit("rejects absolute evidence path", "unsafe evidence path", func(root map[string]interface{}) {
			threat(root, "TM-CON-001")["evidence"] = []interface{}{"/etc/passwd"}
		}),
		edit("rejects missing controls", "at least one control", func(root map[string]interface{}) {
			threat(root, "TM-CON-001")["controls"] = []interface{}{}
		}),
		edit("rejects missing next gate", "next_gate", func(root map[string]interface{}) {
			threat(root, "TM-CON-001")["next_gate"] = ""
		}),
		edit("rejects unsafe production-ready claim", "production_ready must be explicitly false", func(root map[string]interface{}) {
			root["production_ready"] = true
		}),
		edit("rejects unknown model version", "model identity or version", func(root map[string]interface{}) {
			root["model"].(map[string]interface{})["version"] = "truerepublic.threat-model/v2"
		}),
		edit("rejects unmapped high residual", "umbrella_issue", func(root map[string]interface{}) {
			threat(root, "TM-CON-001")["umbrella_issue"] = ""
		}),
		edit("rejects mitigated threat with high residual", "residual_severity", func(root map[string]interface{}) {
			threat(root, "TM-GOV-001")["residual_severity"] = "high"
		}),
		edit("rejects empty threat ID", "non-empty stable ID", func(root map[string]interface{}) {
			threat(root, "TM-CON-001")["id"] = ""
		}),
		edit("rejects empty owner", "owner", func(root map[string]interface{}) {
			threat(root, "TM-CON-001")["owner"] = ""
		}),
	}
}

func parseThreatModelContract(t *testing.T, raw []byte) threatModelContract {
	t.Helper()
	contract, err := decodeThreatModel(raw)
	if err != nil {
		t.Fatalf("parse threat model: %v", err)
	}
	return contract
}

func decodeThreatModel(raw []byte) (threatModelContract, error) {
	var contract threatModelContract
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&contract); err != nil {
		return threatModelContract{}, err
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return threatModelContract{}, fmt.Errorf("unexpected trailing JSON value")
		}
		return threatModelContract{}, fmt.Errorf("unexpected trailing data: %w", err)
	}
	return contract, nil
}

func threatModelViolations(contract threatModelContract) []string {
	var violations []string
	if contract.Model.Name != "truerepublic-cross-system-threat-model" ||
		contract.Model.Version != threatModelVersion ||
		!threatModelDatePattern.MatchString(contract.Model.Updated) {
		violations = append(violations, "invalid model identity or version")
	}
	if contract.ProductionReady == nil || *contract.ProductionReady {
		violations = append(violations, "production_ready must be explicitly false")
	}
	for _, boundary := range []string{"Recovery-only", "not an independent external audit"} {
		if !strings.Contains(contract.SafetyBoundary, boundary) {
			violations = append(violations, "safety_boundary must state the "+boundary+" boundary")
		}
	}
	violations = append(violations, threatModelSectionViolations(contract)...)
	violations = append(violations, threatModelThreatViolations(contract)...)
	return violations
}

func threatModelSectionViolations(contract threatModelContract) []string {
	var violations []string
	assetIDs := make(map[string]bool, len(contract.Assets))
	if len(contract.Assets) == 0 {
		violations = append(violations, "maintained assets must not be empty")
	}
	for _, asset := range contract.Assets {
		if asset.ID == "" || asset.Description == "" || assetIDs[asset.ID] {
			violations = append(violations, "assets need unique non-empty IDs and descriptions")
		}
		assetIDs[asset.ID] = true
	}
	actorIDs := make(map[string]bool, len(contract.Actors))
	if len(contract.Actors) == 0 {
		violations = append(violations, "actors must not be empty")
	}
	for _, actor := range contract.Actors {
		if actor.ID == "" || actor.Description == "" || actorIDs[actor.ID] {
			violations = append(violations, "actors need unique non-empty IDs and descriptions")
		}
		if !threatModelContains(threatModelActorKinds, actor.Kind) {
			violations = append(violations, "unknown actor kind "+actor.Kind)
		}
		actorIDs[actor.ID] = true
	}
	boundaryIDs := make(map[string]bool, len(contract.TrustBoundaries))
	if len(contract.TrustBoundaries) == 0 {
		violations = append(violations, "trust_boundaries must not be empty")
	}
	for _, boundary := range contract.TrustBoundaries {
		if boundary.ID == "" || boundary.Description == "" || boundaryIDs[boundary.ID] {
			violations = append(violations, "trust_boundaries need unique non-empty IDs and descriptions")
		}
		if len(boundary.DataFlows) == 0 {
			violations = append(violations, "trust boundary "+boundary.ID+" must describe at least one data flow")
		}
		boundaryIDs[boundary.ID] = true
	}
	if len(contract.Assumptions) == 0 || len(contract.NonGoals) == 0 {
		violations = append(violations, "assumptions and non_goals must not be empty")
	}
	return violations
}

func threatModelThreatViolations(contract threatModelContract) []string {
	var violations []string
	seen := make(map[string]bool, len(contract.Threats))
	covered := make(map[string]bool, len(threatModelDomains))
	if len(contract.Threats) == 0 {
		violations = append(violations, "threat register must not be empty")
	}
	for _, threat := range contract.Threats {
		label := threat.ID
		if label == "" {
			label = "(empty ID)"
		}
		if !threatModelIDPattern.MatchString(threat.ID) {
			violations = append(violations, label+" must use a non-empty stable ID matching TM-XXX-NNN")
		}
		if seen[threat.ID] {
			violations = append(violations, "duplicate threat ID "+threat.ID)
		}
		seen[threat.ID] = true
		if !threatModelContains(threatModelDomains, threat.Domain) {
			violations = append(violations, threat.ID+" uses unknown domain "+threat.Domain)
		}
		covered[threat.Domain] = true
		if !threatModelContains(threatModelSeverities, threat.Severity) {
			violations = append(violations, threat.ID+" uses unknown severity "+threat.Severity)
		}
		if !threatModelContains(threatModelLikelihoods, threat.Likelihood) {
			violations = append(violations, threat.ID+" uses unknown likelihood "+threat.Likelihood)
		}
		if !threatModelContains(threatModelStatuses, threat.Status) {
			violations = append(violations, threat.ID+" uses unknown status "+threat.Status)
		}
		if !threatModelContains(threatModelSeverities, threat.ResidualSeverity) {
			violations = append(violations, threat.ID+" uses unknown residual_severity "+threat.ResidualSeverity)
		}
		if threat.AttackPath == "" || threat.Impact == "" || threat.ResidualRisk == "" {
			violations = append(violations, threat.ID+" needs non-empty attack_path, impact, and residual_risk")
		}
		if threat.Owner == "" {
			violations = append(violations, threat.ID+" needs a non-empty owner")
		}
		if threat.NextGate == "" {
			violations = append(violations, threat.ID+" needs a non-empty next_gate")
		}
		if len(threat.Controls) == 0 {
			violations = append(violations, threat.ID+" must list at least one control")
		}
		for _, control := range threat.Controls {
			if strings.TrimSpace(control) == "" {
				violations = append(violations, threat.ID+" lists an empty control")
			}
		}
		if len(threat.Evidence) == 0 {
			violations = append(violations, threat.ID+" must reference at least one evidence path")
		}
		for _, path := range threat.Evidence {
			violations = append(violations, threatModelEvidenceViolations(threat.ID, path)...)
		}
		violations = append(violations, threatModelUmbrellaViolations(threat)...)
	}
	for _, domain := range threatModelDomains {
		if !covered[domain] {
			violations = append(violations, "required domain "+domain+" has no threat entry")
		}
	}
	sort.Strings(violations)
	return violations
}

func threatModelEvidenceViolations(id, path string) []string {
	label := id + " evidence path " + fmt.Sprintf("%q", path)
	if path == "" || strings.HasPrefix(path, "/") || strings.Contains(path, "\\") {
		return []string{label + " is an unsafe evidence path"}
	}
	elements := strings.Split(path, "/")
	for _, element := range elements {
		if element == "" || element == "." || element == ".." {
			return []string{label + " is an unsafe evidence path"}
		}
	}
	if _, err := os.Stat(path); err != nil {
		return []string{label + " does not exist in the repository"}
	}
	return nil
}

func threatModelUmbrellaViolations(threat threatModelThreat) []string {
	var violations []string
	closed := threat.Status == "mitigated" || threat.Status == "not_applicable"
	highResidual := threat.ResidualSeverity == "critical" || threat.ResidualSeverity == "high"
	if closed {
		if highResidual {
			violations = append(violations, threat.ID+" with status "+threat.Status+" must not keep a critical/high residual_severity")
		}
		if threat.UmbrellaIssue != "" {
			violations = append(violations, threat.ID+" with status "+threat.Status+" must leave umbrella_issue empty")
		}
		return violations
	}
	if threat.UmbrellaIssue != "" && !threatModelContains(threatModelUmbrellas, threat.UmbrellaIssue) {
		violations = append(violations, threat.ID+" maps to unknown umbrella_issue "+threat.UmbrellaIssue)
	}
	if highResidual && threat.UmbrellaIssue == "" {
		violations = append(violations, threat.ID+" with a critical/high residual must map to umbrella_issue GH-7 or GH-29")
	}
	return violations
}

func threatModelContains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
