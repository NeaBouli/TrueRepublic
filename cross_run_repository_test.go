package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type repositoryCrossRunContract struct {
	Schema                  string `json:"schema"`
	CandidateContractSHA256 string `json:"candidate_contract_sha256"`
	Source                  struct {
		Repository    string `json:"repository"`
		WorkflowPath  string `json:"workflow_path"`
		Branch        string `json:"branch"`
		Event         string `json:"event"`
		RunIDPattern  string `json:"run_id_pattern"`
		RetentionDays int    `json:"retention_days"`
	} `json:"source"`
}

const crossRunJobMarker = "  cross-run-comparison:"

func TestCrossRunRepositoryContract(t *testing.T) {
	t.Parallel()

	contractBytes, err := os.ReadFile("configs/release/cross-run-rebuild.json")
	if err != nil {
		t.Fatal(err)
	}
	var contract repositoryCrossRunContract
	if err = json.Unmarshal(contractBytes, &contract); err != nil {
		t.Fatal(err)
	}
	makefileBytes, err := os.ReadFile("Makefile")
	if err != nil {
		t.Fatal(err)
	}
	makefile := string(makefileBytes)
	scripts := map[string]string{}
	for _, script := range []string{"scripts/generate-cross-run-evidence.sh", "scripts/test-cross-run-evidence.sh", "scripts/verify-cross-run-evidence.sh"} {
		contents, readErr := os.ReadFile(script)
		if readErr != nil {
			t.Fatal(readErr)
		}
		scripts[script] = string(contents)
		info, statErr := os.Stat(script)
		if statErr != nil {
			t.Fatal(statErr)
		}
		if info.Mode()&0111 == 0 {
			t.Fatalf("%s must be executable", script)
		}
	}
	cliBytes, err := os.ReadFile("cmd/cross-run-evidence/main.go")
	if err != nil {
		t.Fatal(err)
	}
	cli := string(cliBytes)
	workflowBytes, err := os.ReadFile(".github/workflows/reproducible-daemon.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(workflowBytes)
	candidateHash := repositoryFileSHA256(t, "configs/release/candidate-evidence.json")
	for _, fixture := range []string{
		"testdata/crossrunevidence/valid/cross-run-evidence.json",
		"testdata/crossrunevidence/invalid-claims/cross-run-evidence.json",
	} {
		if _, statErr := os.Stat(fixture); statErr != nil {
			t.Fatal(statErr)
		}
	}

	if violations := crossRunRepositoryViolations(contract, makefile, scripts, cli, workflow, candidateHash); len(violations) != 0 {
		t.Fatalf("cross-run repository contract violations:\n- %s", strings.Join(violations, "\n- "))
	}
	t.Run("isolates the cross-run job from later siblings", func(t *testing.T) {
		withSibling := workflow + "\n  future-release-job:\n    permissions:\n      contents: write\n"
		if violations := crossRunRepositoryViolations(contract, makefile, scripts, cli, withSibling, candidateHash); len(violations) != 0 {
			t.Fatalf("later sibling leaked into cross-run validation:\n- %s", strings.Join(violations, "\n- "))
		}
	})

	mutateJob := func(change func(string) string) string {
		start, end, ok := crossRunJobBounds(workflow)
		if !ok {
			t.Fatal("cross-run job missing")
		}
		return workflow[:start] + change(workflow[start:end]) + workflow[end:]
	}
	rejectJobMutation := func(name string, change func(string) string) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			if len(crossRunRepositoryViolations(contract, makefile, scripts, cli, mutateJob(change), candidateHash)) == 0 {
				t.Fatalf("workflow accepted %s", name)
			}
		})
	}

	t.Run("rejects contract drift", func(t *testing.T) {
		for _, mutate := range []func(*repositoryCrossRunContract){
			func(c *repositoryCrossRunContract) { c.Schema = "truerepublic.cross-run-rebuild/v2" },
			func(c *repositoryCrossRunContract) { c.Source.Repository = "example/fork" },
			func(c *repositoryCrossRunContract) { c.Source.WorkflowPath = ".github/workflows/other.yml" },
			func(c *repositoryCrossRunContract) { c.Source.Branch = "develop" },
			func(c *repositoryCrossRunContract) { c.Source.Event = "push" },
			func(c *repositoryCrossRunContract) { c.Source.RunIDPattern = "^.*$" },
			func(c *repositoryCrossRunContract) { c.Source.RetentionDays = 30 },
		} {
			mutated := contract
			mutate(&mutated)
			if len(crossRunRepositoryViolations(mutated, makefile, scripts, cli, workflow, candidateHash)) == 0 {
				t.Fatal("contract drift accepted")
			}
		}
	})
	t.Run("rejects pinned candidate digest drift", func(t *testing.T) {
		if len(crossRunRepositoryViolations(contract, makefile, scripts, cli, workflow, strings.Repeat("0", 64))) == 0 {
			t.Fatal("contract accepted a stale candidate contract digest")
		}
	})
	t.Run("rejects missing dispatch input", func(t *testing.T) {
		mutated := strings.Replace(workflow, `      baseline_run_id:
        description: 'Optional baseline run ID of an earlier workflow_dispatch execution of the same exact main commit for the GH-273 cross-run rebuild comparison'
        required: false
        type: string
`, "", 1)
		if len(crossRunRepositoryViolations(contract, makefile, scripts, cli, mutated, candidateHash)) == 0 {
			t.Fatal("workflow without the baseline_run_id dispatch input accepted")
		}
	})
	t.Run("rejects Makefile without target", func(t *testing.T) {
		mutated := strings.Replace(makefile, "cross-run-evidence-contract-test:", "removed-cross-run-target:", 1)
		if len(crossRunRepositoryViolations(contract, mutated, scripts, cli, workflow, candidateHash)) == 0 {
			t.Fatal("Makefile without the cross-run contract target accepted")
		}
	})
	t.Run("rejects verify script without CLI binding", func(t *testing.T) {
		mutated := map[string]string{}
		for name, contents := range scripts {
			mutated[name] = contents
		}
		mutated["scripts/verify-cross-run-evidence.sh"] = strings.Replace(
			mutated["scripts/verify-cross-run-evidence.sh"], "./cmd/cross-run-evidence", "./cmd/other", 1)
		if len(crossRunRepositoryViolations(contract, makefile, mutated, cli, workflow, candidateHash)) == 0 {
			t.Fatal("verify script without the cross-run CLI accepted")
		}
	})
	t.Run("rejects generator without hermetic claim", func(t *testing.T) {
		mutated := map[string]string{}
		for name, contents := range scripts {
			mutated[name] = contents
		}
		mutated["scripts/generate-cross-run-evidence.sh"] = strings.Replace(
			mutated["scripts/generate-cross-run-evidence.sh"], "long_term_hermetic:false", "long_term_hermetic:true", 1)
		if len(crossRunRepositoryViolations(contract, makefile, mutated, cli, workflow, candidateHash)) == 0 {
			t.Fatal("generator without the false long-term-hermetic claim accepted")
		}
	})

	rejectJobMutation("privilege expansion", func(job string) string {
		return strings.Replace(job, "    permissions:\n      contents: read\n      actions: read\n",
			"    permissions:\n      contents: write\n      actions: write\n", 1)
	})
	rejectJobMutation("a missing event guard", func(job string) string {
		return strings.Replace(job, "if: github.event_name == 'workflow_dispatch' && github.event.inputs.baseline_run_id != ''",
			"if: github.event.inputs.baseline_run_id != ''", 1)
	})
	rejectJobMutation("a missing input guard", func(job string) string {
		return strings.Replace(job, "if: github.event_name == 'workflow_dispatch' && github.event.inputs.baseline_run_id != ''",
			"if: github.event_name == 'workflow_dispatch'", 1)
	})
	rejectJobMutation("a missing needs gate", func(job string) string {
		return strings.Replace(job, "needs: [candidate-evidence]", "needs: [deterministic-build]", 1)
	})
	rejectJobMutation("a missing numeric input validation", func(job string) string {
		return strings.Replace(job, `          [[ "$BASELINE_RUN_ID" =~ ^[1-9][0-9]{0,18}$ ]]`+"\n", "", 1)
	})
	rejectJobMutation("a missing distinct-run guard", func(job string) string {
		return strings.Replace(job, `          test "$BASELINE_RUN_ID" != "$GITHUB_RUN_ID"`+"\n", "", 1)
	})
	rejectJobMutation("a missing baseline API guard", func(job string) string {
		return strings.Replace(job, `          gh api "repos/$GITHUB_REPOSITORY/actions/runs/$BASELINE_RUN_ID" >"$run_json"`+"\n", "", 1)
	})
	rejectJobMutation("a caller-selected ref checkout", func(job string) string {
		return strings.Replace(job, "          persist-credentials: false\n",
			"          persist-credentials: false\n          ref: ${{ github.event.inputs.baseline_run_id }}\n", 1)
	})
	rejectJobMutation("an arbitrary repository download", func(job string) string {
		return strings.Replace(job, "repository: ${{ github.repository }}", "repository: example/fork", 1)
	})
	rejectJobMutation("a caller-controlled run download", func(job string) string {
		return strings.Replace(job, "run-id: ${{ github.event.inputs.baseline_run_id }}", "run-id: 1", 1)
	})
	rejectJobMutation("an unpinned download action", func(job string) string {
		return strings.Replace(job, "actions/download-artifact@3e5f45b2cfb9172054b4087a40e8e0b5a5461e7c # v8.0.1",
			"actions/download-artifact@v8", 1)
	})
	rejectJobMutation("digest mismatch weakening", func(job string) string {
		return strings.Replace(job, "digest-mismatch: error", "digest-mismatch: warn", 1)
	})
	rejectJobMutation("a payload upload", func(job string) string {
		return strings.Replace(job, "            ${{ runner.temp }}/cross-run-report.json\n",
			"            ${{ runner.temp }}/cross-run-report.json\n            ${{ runner.temp }}/cross-run-input/baseline/daemon-linux-amd64-1.oci.tar\n", 1)
	})
	rejectJobMutation("retention drift", func(job string) string {
		return strings.Replace(job, "retention-days: 14", "retention-days: 30", 1)
	})
	rejectJobMutation("a missing offline contract gate", func(job string) string {
		return strings.Replace(job, "        run: ./scripts/test-cross-run-evidence.sh\n", "", 1)
	})
	t.Run("rejects a missing release-contract gate", func(t *testing.T) {
		mutated := strings.Replace(workflow, "        run: make cross-run-evidence-contract-test\n", "", 1)
		if len(crossRunRepositoryViolations(contract, makefile, scripts, cli, mutated, candidateHash)) == 0 {
			t.Fatal("workflow without the release-contract cross-run gate accepted")
		}
	})
	t.Run("rejects workflow without a path trigger", func(t *testing.T) {
		mutated := strings.ReplaceAll(workflow, "      - 'crossrunevidence/**'\n", "")
		if len(crossRunRepositoryViolations(contract, makefile, scripts, cli, mutated, candidateHash)) == 0 {
			t.Fatal("workflow accepted a missing cross-run path trigger")
		}
	})
}

func crossRunRepositoryViolations(contract repositoryCrossRunContract, makefile string, scripts map[string]string, cli string, workflow string, candidateHash string) []string {
	var violations []string
	if contract.Schema != "truerepublic.cross-run-rebuild/v1" ||
		contract.Source.Repository != "NeaBouli/TrueRepublic" ||
		contract.Source.WorkflowPath != ".github/workflows/reproducible-daemon.yml" ||
		contract.Source.Branch != "main" ||
		contract.Source.Event != "workflow_dispatch" ||
		contract.Source.RunIDPattern != `^[1-9][0-9]{0,18}$` ||
		contract.Source.RetentionDays != 14 {
		violations = append(violations, "cross-run contract identity mismatch")
	}
	if contract.CandidateContractSHA256 != candidateHash {
		violations = append(violations, "pinned candidate contract digest mismatch")
	}

	generateScript, ok := scripts["scripts/generate-cross-run-evidence.sh"]
	if !ok {
		violations = append(violations, "generate-cross-run-evidence.sh is missing")
	} else {
		for _, required := range []string{
			"set -euo pipefail",
			"configs/release/cross-run-rebuild.json",
			"--baseline-receipt", "--current-receipt", "--baseline-dir", "--current-dir", "--output-dir",
			"--repository", "--workflow-path", "--branch", "--commit", "--baseline-run-id", "--current-run-id",
			`^[1-9][0-9]{0,18}$`,
			"^[0-9a-f]{40}$",
			"truerepublic.cross-run-evidence/v1",
			"long_term_hermetic:false",
			"jq -n -S",
			"verify-cross-run-evidence.sh",
		} {
			if !strings.Contains(generateScript, required) {
				violations = append(violations, "generate script lacks "+required)
			}
		}
	}
	verifyScript, ok := scripts["scripts/verify-cross-run-evidence.sh"]
	if !ok {
		violations = append(violations, "verify-cross-run-evidence.sh is missing")
	} else {
		for _, required := range []string{
			"go run ./cmd/cross-run-evidence compare",
			"configs/release/cross-run-rebuild.json",
			"configs/release/candidate-evidence.json",
			"--expected-repository", "--expected-workflow", "--expected-branch", "--expected-commit",
			"--expected-baseline-run-id", "--expected-current-run-id",
			"GOPROXY=off",
		} {
			if !strings.Contains(verifyScript, required) {
				violations = append(violations, "verify script lacks "+required)
			}
		}
	}
	testScript, ok := scripts["scripts/test-cross-run-evidence.sh"]
	if !ok {
		violations = append(violations, "test-cross-run-evidence.sh is missing")
	} else {
		for _, required := range []string{"go test ./crossrunevidence ./cmd/cross-run-evidence", "verify-cross-run-evidence.sh", "testdata/crossrunevidence", "invalid-claims", "generate-cross-run-evidence.sh", "bash -n"} {
			if !strings.Contains(testScript, required) {
				violations = append(violations, "test script lacks "+required)
			}
		}
	}
	for _, required := range []string{"cross-run-evidence-contract-test:", "./scripts/test-cross-run-evidence.sh", "TestCrossRunRepositoryContract"} {
		if !strings.Contains(makefile, required) {
			violations = append(violations, "Makefile lacks "+required)
		}
	}
	for _, required := range []string{"truerepublic/crossrunevidence", "crossrunevidence.Run"} {
		if !strings.Contains(cli, required) {
			violations = append(violations, "cross-run CLI lacks "+required)
		}
	}

	for _, required := range []string{
		"      baseline_run_id:\n",
		"        required: false\n",
		"        type: string\n",
		"make cross-run-evidence-contract-test",
	} {
		if !strings.Contains(workflow, required) {
			violations = append(violations, "reproducible-daemon workflow lacks "+required)
		}
	}
	for _, trigger := range []string{
		"- 'configs/release/cross-run-rebuild.json'",
		"- 'crossrunevidence/**'",
		"- 'cmd/cross-run-evidence/**'",
		"- 'testdata/crossrunevidence/**'",
		"- 'scripts/generate-cross-run-evidence.sh'",
		"- 'scripts/test-cross-run-evidence.sh'",
		"- 'scripts/verify-cross-run-evidence.sh'",
		"- 'cross_run_repository_test.go'",
	} {
		if strings.Count(workflow, trigger) != 2 {
			violations = append(violations, "reproducible-daemon workflow lacks push and pull_request trigger "+trigger)
		}
	}

	start, end, ok := crossRunJobBounds(workflow)
	if !ok {
		violations = append(violations, "cross-run comparison job is missing")
		return violations
	}
	job := workflow[start:end]
	for _, required := range []string{
		"needs: [candidate-evidence]",
		"if: github.event_name == 'workflow_dispatch' && github.event.inputs.baseline_run_id != ''",
		"    permissions:\n      contents: read\n      actions: read\n",
		`[[ "$BASELINE_RUN_ID" =~ ^[1-9][0-9]{0,18}$ ]]`,
		`test "$BASELINE_RUN_ID" != "$GITHUB_RUN_ID"`,
		`gh api "repos/$GITHUB_REPOSITORY/actions/runs/$BASELINE_RUN_ID" >"$run_json"`,
		`gh api "repos/$GITHUB_REPOSITORY/actions/runs/$BASELINE_RUN_ID/artifacts?name=$artifact_name&per_page=100" >"$artifacts_json"`,
		`test "$(jq -er '.total_count' "$artifacts_json")" = "1"`,
		`test "$(jq -er '.head_sha' "$run_json")" = "$GITHUB_SHA"`,
		`test "$(jq -er '.status' "$run_json")" = "completed"`,
		`test "$(jq -er '.conclusion' "$run_json")" = "success"`,
		`test "$(jq -er '.head_branch' "$run_json")" = "main"`,
		`test "$(jq -er '.event' "$run_json")" = "workflow_dispatch"`,
		"./scripts/test-cross-run-evidence.sh",
		"./scripts/generate-cross-run-evidence.sh",
		"./scripts/verify-cross-run-evidence.sh",
		`--repository "$GITHUB_REPOSITORY"`,
		`--commit "$GITHUB_SHA"`,
		`--baseline-run-id "$BASELINE_RUN_ID"`,
		`--current-run-id "$GITHUB_RUN_ID"`,
	} {
		if !strings.Contains(job, required) {
			violations = append(violations, "cross-run job lacks "+required)
		}
	}
	for pinned, count := range map[string]int{
		"actions/download-artifact@3e5f45b2cfb9172054b4087a40e8e0b5a5461e7c # v8.0.1": 2,
		"github-token: ${{ github.token }}":                                           1,
		"repository: ${{ github.repository }}":                                        1,
		"run-id: ${{ github.event.inputs.baseline_run_id }}":                          1,
		"digest-mismatch: error":                                                      2,
		`gh api "repos/$GITHUB_REPOSITORY/actions/runs/$BASELINE_RUN_ID/artifacts?name=$artifact_name&per_page=100"`: 1,
		`gh api "repos/$GITHUB_REPOSITORY/actions/runs/$GITHUB_RUN_ID/artifacts?name=$artifact_name&per_page=100"`:   1,
		`test "$(jq -er '.total_count' "$artifacts_json")" = "1"`:                                                    2,
	} {
		if strings.Count(job, pinned) != count {
			violations = append(violations, "cross-run job binding drift: "+pinned)
		}
	}
	for _, forbidden := range []string{"ref:", "contents: write", "actions: write", "packages: write", "id-token:", "environment:"} {
		if strings.Contains(job, forbidden) {
			violations = append(violations, "cross-run job contains forbidden capability "+forbidden)
		}
	}

	uploadFound := false
	for _, block := range crossRunUploadBlocks(job) {
		if !strings.Contains(block, "name: truerepublic-cross-run-") {
			continue
		}
		uploadFound = true
		if !strings.Contains(block, "if-no-files-found: error") || !strings.Contains(block, "retention-days: 14") {
			violations = append(violations, "cross-run upload lacks the error or 14-day retention policy")
		}
		allowed := map[string]bool{
			"${{ runner.temp }}/cross-run-evidence/cross-run-evidence.json":          false,
			"${{ runner.temp }}/cross-run-evidence/receipt-baseline.json":            false,
			"${{ runner.temp }}/cross-run-evidence/receipt-current.json":             false,
			"${{ runner.temp }}/cross-run-evidence/candidate-manifest-baseline.json": false,
			"${{ runner.temp }}/cross-run-evidence/candidate-report-baseline.json":   false,
			"${{ runner.temp }}/cross-run-evidence/candidate-manifest-current.json":  false,
			"${{ runner.temp }}/cross-run-evidence/candidate-report-current.json":    false,
			"${{ runner.temp }}/cross-run-report.json":                               false,
		}
		for _, line := range strings.Split(block, "\n") {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "${{ runner.temp }}/") {
				continue
			}
			if _, ok := allowed[trimmed]; !ok {
				violations = append(violations, "cross-run upload includes an unallowed path: "+trimmed)
				continue
			}
			allowed[trimmed] = true
		}
		for path, seen := range allowed {
			if !seen {
				violations = append(violations, "cross-run upload is missing "+path)
			}
		}
	}
	if !uploadFound {
		violations = append(violations, "cross-run comparison upload is missing")
	}
	return violations
}

// crossRunJobBounds isolates the cross-run job from any later top-level job,
// so security assertions cannot be satisfied or tripped by a sibling block.
func crossRunJobBounds(workflow string) (int, int, bool) {
	start := strings.Index(workflow, crossRunJobMarker)
	if start < 0 {
		return 0, 0, false
	}
	rest := workflow[start+len(crossRunJobMarker):]
	offset := len(crossRunJobMarker)
	for _, line := range strings.SplitAfter(rest, "\n") {
		offset += len(line)
		trimmed := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		if !strings.HasPrefix(trimmed, "  ") || strings.HasPrefix(trimmed, "   ") || !strings.HasSuffix(trimmed, ":") {
			continue
		}
		key := strings.TrimSuffix(strings.TrimSpace(trimmed), ":")
		if key != "" && !strings.ContainsAny(key, " \t") {
			return start, start + offset - len(line), true
		}
	}
	return start, len(workflow), true
}

// crossRunUploadBlocks extracts the step bodies of every upload-artifact use
// so payload paths can be rejected without forbidding metadata-only uploads.
func crossRunUploadBlocks(workflow string) []string {
	const marker = "uses: actions/upload-artifact"
	var blocks []string
	rest := workflow
	for {
		index := strings.Index(rest, marker)
		if index < 0 {
			return blocks
		}
		rest = rest[index:]
		end := strings.Index(rest, "\n      - ")
		if end < 0 {
			blocks = append(blocks, rest)
			return blocks
		}
		blocks = append(blocks, rest[:end])
		rest = rest[end+1:]
	}
}
