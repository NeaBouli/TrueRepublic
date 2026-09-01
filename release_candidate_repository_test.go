package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type repositoryCandidateContract struct {
	Schema string `json:"schema"`
	Source struct {
		CommitKind    string `json:"commit_kind"`
		CommitPattern string `json:"commit_pattern"`
		TagKind       string `json:"tag_kind"`
		TagPattern    string `json:"tag_pattern"`
	} `json:"source"`
	BinaryContractSHA256 string `json:"binary_contract_sha256"`
	OCIContractSHA256    string `json:"oci_contract_sha256"`
	BinaryTargets        []struct {
		ID         string `json:"id"`
		Artifact   string `json:"artifact"`
		CIRunner   string `json:"ci_runner"`
		RunnerArch string `json:"runner_arch"`
	} `json:"binary_targets"`
	OCIRepetitions int `json:"oci_repetitions"`
	OCITargets     []struct {
		ID       string `json:"id"`
		Platform string `json:"platform"`
	} `json:"oci_targets"`
}

func TestReleaseCandidateRepositoryContract(t *testing.T) {
	t.Parallel()

	contractBytes, err := os.ReadFile("configs/release/candidate-evidence.json")
	if err != nil {
		t.Fatal(err)
	}
	var contract repositoryCandidateContract
	if err = json.Unmarshal(contractBytes, &contract); err != nil {
		t.Fatal(err)
	}
	makefileBytes, err := os.ReadFile("Makefile")
	if err != nil {
		t.Fatal(err)
	}
	makefile := string(makefileBytes)
	scripts := map[string]string{}
	for _, script := range []string{"scripts/generate-candidate-evidence.sh", "scripts/test-candidate-evidence.sh", "scripts/verify-candidate-evidence.sh"} {
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
	cliBytes, err := os.ReadFile("cmd/candidate-evidence/main.go")
	if err != nil {
		t.Fatal(err)
	}
	cli := string(cliBytes)
	gatesBytes, err := os.ReadFile("configs/security/gates.json")
	if err != nil {
		t.Fatal(err)
	}
	var gates struct {
		Actions map[string]string `json:"actions"`
	}
	if err = json.Unmarshal(gatesBytes, &gates); err != nil {
		t.Fatal(err)
	}
	workflowEntries, err := os.ReadDir(".github/workflows")
	if err != nil {
		t.Fatal(err)
	}
	workflows := map[string]string{}
	for _, entry := range workflowEntries {
		if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".yml") && !strings.HasSuffix(entry.Name(), ".yaml")) {
			continue
		}
		contents, readErr := os.ReadFile(".github/workflows/" + entry.Name())
		if readErr != nil {
			t.Fatal(readErr)
		}
		workflows[entry.Name()] = string(contents)
	}
	for _, fixture := range []string{
		"testdata/candidateevidence/valid/candidate-evidence.json",
		"testdata/candidateevidence/invalid-claims/candidate-evidence.json",
	} {
		if _, statErr := os.Stat(fixture); statErr != nil {
			t.Fatal(statErr)
		}
	}
	binaryHash := repositoryFileSHA256(t, "configs/build/deterministic-linux-daemon.json")
	ociHash := repositoryFileSHA256(t, "configs/build/reproducible-oci.json")

	if violations := candidateRepositoryViolations(contract, makefile, scripts, cli, workflows, binaryHash, ociHash, gates.Actions); len(violations) != 0 {
		t.Fatalf("release candidate repository contract violations:\n- %s", strings.Join(violations, "\n- "))
	}

	t.Run("rejects tag grammar drift", func(t *testing.T) {
		mutated := contract
		mutated.Source.TagPattern = "^.*$"
		if len(candidateRepositoryViolations(mutated, makefile, scripts, cli, workflows, binaryHash, ociHash, gates.Actions)) == 0 {
			t.Fatal("contract accepted a weakened simulated-tag grammar")
		}
	})
	t.Run("rejects binary target drift", func(t *testing.T) {
		mutated := contract
		mutated.BinaryTargets[0].ID = "darwin-amd64"
		if len(candidateRepositoryViolations(mutated, makefile, scripts, cli, workflows, binaryHash, ociHash, gates.Actions)) == 0 {
			t.Fatal("contract accepted a binary target drift")
		}
	})
	t.Run("rejects OCI platform drift", func(t *testing.T) {
		mutated := contract
		mutated.OCITargets[0].Platform = "linux/s390x"
		if len(candidateRepositoryViolations(mutated, makefile, scripts, cli, workflows, binaryHash, ociHash, gates.Actions)) == 0 {
			t.Fatal("contract accepted an OCI platform drift")
		}
	})
	t.Run("rejects pinned digest drift", func(t *testing.T) {
		if len(candidateRepositoryViolations(contract, makefile, scripts, cli, workflows, strings.Repeat("0", 64), ociHash, gates.Actions)) == 0 {
			t.Fatal("contract accepted a stale deterministic build contract digest")
		}
	})
	t.Run("rejects workflow tag creation", func(t *testing.T) {
		for _, command := range []string{"git tag v1.0.0", "git update-ref refs/tags/v1.0.0 HEAD"} {
			mutated := map[string]string{}
			for name, contents := range workflows {
				mutated[name] = contents
			}
			mutated["reproducible-daemon.yml"] += "\n      - run: " + command + "\n"
			if len(candidateRepositoryViolations(contract, makefile, scripts, cli, mutated, binaryHash, ociHash, gates.Actions)) == 0 {
				t.Fatalf("workflow accepted tag creation through %q", command)
			}
		}
	})
	t.Run("rejects workflow tag push", func(t *testing.T) {
		for _, command := range []string{
			"git push origin --tags",
			"git push origin v1.0.0",
			"git push origin refs/tags/v1.0.0",
			"gh api --method POST repos/example/project/releases",
			"gh api -X POST repos/example/project/releases",
		} {
			mutated := map[string]string{}
			for name, contents := range workflows {
				mutated[name] = contents
			}
			mutated["reproducible-daemon.yml"] += "\n      - run: " + command + "\n"
			if len(candidateRepositoryViolations(contract, makefile, scripts, cli, mutated, binaryHash, ociHash, gates.Actions)) == 0 {
				t.Fatalf("workflow accepted a tag/release push through %q", command)
			}
		}
	})
	t.Run("rejects workflow payload upload", func(t *testing.T) {
		mutated := map[string]string{}
		for name, contents := range workflows {
			mutated[name] = contents
		}
		mutated["reproducible-daemon.yml"] += `
      - name: Upload payload
        uses: actions/upload-artifact@043fb46d1a93c77aae656e7c1c64a875d1fc6a0a
        with:
          name: payload
          path: evidence/daemon-linux-amd64-1.oci.tar
`
		if len(candidateRepositoryViolations(contract, makefile, scripts, cli, mutated, binaryHash, ociHash, gates.Actions)) == 0 {
			t.Fatal("workflow accepted an OCI payload upload")
		}
	})
	t.Run("rejects Makefile without target", func(t *testing.T) {
		mutated := strings.Replace(makefile, "candidate-evidence-contract-test:", "removed-candidate-target:", 1)
		if len(candidateRepositoryViolations(contract, mutated, scripts, cli, workflows, binaryHash, ociHash, gates.Actions)) == 0 {
			t.Fatal("Makefile without the candidate contract target accepted")
		}
	})
	t.Run("rejects script without CLI binding", func(t *testing.T) {
		mutated := map[string]string{}
		for name, contents := range scripts {
			mutated[name] = contents
		}
		mutated["scripts/verify-candidate-evidence.sh"] = strings.Replace(
			mutated["scripts/verify-candidate-evidence.sh"], "./cmd/candidate-evidence", "./cmd/other", 1)
		if len(candidateRepositoryViolations(contract, makefile, mutated, cli, workflows, binaryHash, ociHash, gates.Actions)) == 0 {
			t.Fatal("verify script without the candidate CLI accepted")
		}
	})
	t.Run("rejects generator without tag grammar", func(t *testing.T) {
		mutated := map[string]string{}
		for name, contents := range scripts {
			mutated[name] = contents
		}
		mutated["scripts/generate-candidate-evidence.sh"] = strings.Replace(
			mutated["scripts/generate-candidate-evidence.sh"], `^v(0|[1-9][0-9]*)`, "^v.*", 1)
		if len(candidateRepositoryViolations(contract, makefile, mutated, cli, workflows, binaryHash, ociHash, gates.Actions)) == 0 {
			t.Fatal("generator without the simulated-tag grammar accepted")
		}
	})
	t.Run("rejects security gate download pin drift", func(t *testing.T) {
		mutated := map[string]string{}
		for name, sha := range gates.Actions {
			mutated[name] = sha
		}
		mutated["actions/download-artifact"] = strings.Repeat("0", 40)
		if len(candidateRepositoryViolations(contract, makefile, scripts, cli, workflows, binaryHash, ociHash, mutated)) == 0 {
			t.Fatal("security gates accepted a drifting download-artifact pin")
		}
	})
	t.Run("rejects download-artifact pin drift", func(t *testing.T) {
		mutated := map[string]string{}
		for name, contents := range workflows {
			mutated[name] = contents
		}
		mutated["reproducible-daemon.yml"] = strings.Replace(
			mutated["reproducible-daemon.yml"], "3e5f45b2cfb9172054b4087a40e8e0b5a5461e7c", strings.Repeat("0", 40), 1)
		if len(candidateRepositoryViolations(contract, makefile, scripts, cli, mutated, binaryHash, ociHash, gates.Actions)) == 0 {
			t.Fatal("workflow accepted a drifting download-artifact commit")
		}
	})
	t.Run("rejects aggregation without both dependencies", func(t *testing.T) {
		mutated := map[string]string{}
		for name, contents := range workflows {
			mutated[name] = contents
		}
		mutated["reproducible-daemon.yml"] = strings.Replace(
			mutated["reproducible-daemon.yml"], "needs: [deterministic-build, reproducible-oci]", "needs: [deterministic-build]", 1)
		if len(candidateRepositoryViolations(contract, makefile, scripts, cli, mutated, binaryHash, ociHash, gates.Actions)) == 0 {
			t.Fatal("workflow accepted aggregation without both evidence dependencies")
		}
	})
	t.Run("rejects candidate upload beyond allowlist", func(t *testing.T) {
		mutated := map[string]string{}
		for name, contents := range workflows {
			mutated[name] = contents
		}
		mutated["reproducible-daemon.yml"] = strings.Replace(
			mutated["reproducible-daemon.yml"],
			"${{ runner.temp }}/candidate-artifact/candidate-evidence-report.json\n",
			"${{ runner.temp }}/candidate-artifact/candidate-evidence-report.json\n            ${{ runner.temp }}/candidate-artifact/sbom.cdx.json\n", 1)
		if len(candidateRepositoryViolations(contract, makefile, scripts, cli, mutated, binaryHash, ociHash, gates.Actions)) == 0 {
			t.Fatal("workflow accepted a candidate upload outside the metadata allowlist")
		}
	})
	t.Run("rejects workflow without generator path trigger", func(t *testing.T) {
		mutated := map[string]string{}
		for name, contents := range workflows {
			mutated[name] = contents
		}
		mutated["reproducible-daemon.yml"] = strings.ReplaceAll(
			mutated["reproducible-daemon.yml"], "      - 'scripts/generate-candidate-evidence.sh'\n", "")
		if len(candidateRepositoryViolations(contract, makefile, scripts, cli, mutated, binaryHash, ociHash, gates.Actions)) == 0 {
			t.Fatal("workflow accepted a missing generator path trigger")
		}
	})
}

func candidateRepositoryViolations(contract repositoryCandidateContract, makefile string, scripts map[string]string, cli string, workflows map[string]string, binaryHash, ociHash string, actions map[string]string) []string {
	var violations []string
	if actions["actions/download-artifact"] != "3e5f45b2cfb9172054b4087a40e8e0b5a5461e7c" {
		violations = append(violations, "security gates do not pin the exact actions/download-artifact v8.0.1 commit")
	}
	if contract.Schema != "truerepublic.release-candidate/v1" ||
		contract.Source.CommitKind != "git-commit" || contract.Source.CommitPattern != "^[0-9a-f]{40}$" ||
		contract.Source.TagKind != "simulated-future-tag" ||
		contract.Source.TagPattern != `^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$` {
		violations = append(violations, "contract identity mismatch")
	}
	if contract.BinaryContractSHA256 != binaryHash || contract.OCIContractSHA256 != ociHash {
		violations = append(violations, "pinned build/OCI contract digest mismatch")
	}
	expectedBinary := []struct{ id, artifact, runner, arch string }{
		{"linux-amd64", "truerepublicd-linux-amd64", "ubuntu-24.04", "x86_64"},
		{"linux-arm64", "truerepublicd-linux-arm64", "ubuntu-24.04-arm", "aarch64"},
	}
	if len(contract.BinaryTargets) != len(expectedBinary) {
		violations = append(violations, "binary target count mismatch")
	} else {
		for index, want := range expectedBinary {
			got := contract.BinaryTargets[index]
			if got.ID != want.id || got.Artifact != want.artifact || got.CIRunner != want.runner || got.RunnerArch != want.arch {
				violations = append(violations, "binary target contract mismatch: "+want.id)
			}
		}
	}
	expectedOCI := []struct{ id, platform string }{
		{"daemon-linux-amd64", "linux/amd64"},
		{"client-web-linux-amd64", "linux/amd64"},
		{"daemon-linux-arm64", "linux/arm64"},
		{"client-web-linux-arm64", "linux/arm64"},
	}
	if contract.OCIRepetitions != 2 || len(contract.OCITargets) != len(expectedOCI) {
		violations = append(violations, "OCI target contract mismatch")
	} else {
		for index, want := range expectedOCI {
			if got := contract.OCITargets[index]; got.ID != want.id || got.Platform != want.platform {
				violations = append(violations, "OCI target contract mismatch: "+want.id)
			}
		}
	}

	generateScript, ok := scripts["scripts/generate-candidate-evidence.sh"]
	if !ok {
		violations = append(violations, "generate-candidate-evidence.sh is missing")
	} else {
		for _, required := range []string{
			"set -euo pipefail",
			"configs/release/candidate-evidence.json",
			"--tag", "--commit", "--amd64-dir", "--arm64-dir", "--oci-amd64-dir", "--oci-arm64-dir", "--output-dir",
			`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`,
			"^[0-9a-f]{40}$",
			"truerepublic.release-candidate-evidence/v1",
			"real_tag_created:false",
			"jq -n -S",
			"verify-candidate-evidence.sh",
		} {
			if !strings.Contains(generateScript, required) {
				violations = append(violations, "generate script lacks "+required)
			}
		}
	}
	verifyScript, ok := scripts["scripts/verify-candidate-evidence.sh"]
	if !ok {
		violations = append(violations, "verify-candidate-evidence.sh is missing")
	} else {
		for _, required := range []string{"go run ./cmd/candidate-evidence verify", "configs/release/candidate-evidence.json", "GOPROXY=off"} {
			if !strings.Contains(verifyScript, required) {
				violations = append(violations, "verify script lacks "+required)
			}
		}
	}
	testScript, ok := scripts["scripts/test-candidate-evidence.sh"]
	if !ok {
		violations = append(violations, "test-candidate-evidence.sh is missing")
	} else {
		for _, required := range []string{"go test ./candidateevidence ./cmd/candidate-evidence", "verify-candidate-evidence.sh", "testdata/candidateevidence", "invalid-claims", "generate-candidate-evidence.sh", "bash -n"} {
			if !strings.Contains(testScript, required) {
				violations = append(violations, "test script lacks "+required)
			}
		}
	}
	for _, required := range []string{"candidate-evidence-contract-test:", "./scripts/test-candidate-evidence.sh", "TestReleaseCandidateRepositoryContract"} {
		if !strings.Contains(makefile, required) {
			violations = append(violations, "Makefile lacks "+required)
		}
	}
	for _, required := range []string{"truerepublic/candidateevidence", "candidateevidence.Run"} {
		if !strings.Contains(cli, required) {
			violations = append(violations, "candidate CLI lacks "+required)
		}
	}

	daemonWorkflow := workflows["reproducible-daemon.yml"]
	for _, required := range []string{
		"candidate-evidence:",
		"needs: [deterministic-build, reproducible-oci]",
		"actions/download-artifact@3e5f45b2cfb9172054b4087a40e8e0b5a5461e7c # v8.0.1",
		"name: truerepublic-linux-amd64-${{ github.sha }}",
		"name: truerepublic-linux-arm64-${{ github.sha }}",
		"name: truerepublic-oci-amd64-${{ github.sha }}",
		"name: truerepublic-oci-arm64-${{ github.sha }}",
		`tee "${evidence}/oci-evidence-report-${PLATFORM_SUFFIX}.json"`,
		"${{ runner.temp }}/oci-evidence-${{ matrix.suffix }}/oci-evidence-report-${{ matrix.suffix }}.json",
		"CANDIDATE_TAG: v0.5.0",
		`--tag "$CANDIDATE_TAG"`,
		`--commit "$GITHUB_SHA"`,
		"./scripts/generate-candidate-evidence.sh",
		`cp "${RUNNER_TEMP}/candidate-evidence/candidate-evidence.json"`,
		`"${RUNNER_TEMP}/candidate-artifact/candidate-evidence.json"`,
		`cp "${RUNNER_TEMP}/candidate-evidence-report.json"`,
		`"${RUNNER_TEMP}/candidate-artifact/candidate-evidence-report.json"`,
		`test "$(find "${RUNNER_TEMP}/candidate-artifact" -mindepth 1 -maxdepth 1 -type f | wc -l)" -eq 2`,
		"name: truerepublic-candidate-${{ github.sha }}",
	} {
		if !strings.Contains(daemonWorkflow, required) {
			violations = append(violations, "reproducible-daemon workflow lacks candidate aggregation binding "+required)
		}
	}
	for _, trigger := range []string{
		"- 'candidateevidence/**'",
		"- 'cmd/candidate-evidence/**'",
		"- 'testdata/candidateevidence/**'",
		"- 'scripts/generate-candidate-evidence.sh'",
		"- 'scripts/test-candidate-evidence.sh'",
		"- 'scripts/verify-candidate-evidence.sh'",
		"- 'release_candidate_repository_test.go'",
	} {
		if strings.Count(daemonWorkflow, trigger) != 2 {
			violations = append(violations, "reproducible-daemon workflow lacks push and pull_request trigger "+trigger)
		}
	}
	for _, block := range candidateUploadBlocks(daemonWorkflow) {
		if !strings.Contains(block, "name: truerepublic-candidate-") {
			continue
		}
		allowed := map[string]bool{
			"${{ runner.temp }}/candidate-artifact/candidate-evidence.json":        false,
			"${{ runner.temp }}/candidate-artifact/candidate-evidence-report.json": false,
		}
		for _, line := range strings.Split(block, "\n") {
			trimmed := strings.TrimSpace(line)
			if !strings.HasPrefix(trimmed, "${{ runner.temp }}/") {
				continue
			}
			if _, ok := allowed[trimmed]; !ok {
				violations = append(violations, "candidate upload includes an unallowed path: "+trimmed)
				continue
			}
			allowed[trimmed] = true
		}
		for path, seen := range allowed {
			if !seen {
				violations = append(violations, "candidate upload is missing "+path)
			}
		}
	}

	for _, forbidden := range []string{
		"git tag ", "git update-ref refs/tags/", "git push --tags", "git push origin --tags",
		"git push origin v", "git push origin refs/tags/", "git push origin tag ",
		"gh release create", "gh api --method POST", "gh api -X POST",
		"softprops/action-gh-release", "goreleaser", "cosign", "sigstore",
		"docker login", "docker/login-action", "--push", "type=registry",
	} {
		for name, workflow := range workflows {
			if strings.Contains(workflow, forbidden) {
				violations = append(violations, "workflow "+name+" contains forbidden release operation: "+forbidden)
			}
		}
	}
	for name, workflow := range workflows {
		for _, block := range candidateUploadBlocks(workflow) {
			for _, payload := range []string{".oci.tar", "/truerepublicd-linux-amd64", "/truerepublicd-linux-arm64"} {
				if strings.Contains(block, payload) {
					violations = append(violations, "workflow "+name+" uploads a binary/image payload: "+payload)
				}
			}
		}
	}
	return violations
}

// candidateUploadBlocks extracts the step bodies of every upload-artifact use
// so payload paths can be rejected without forbidding metadata-only uploads.
func candidateUploadBlocks(workflow string) []string {
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

func repositoryFileSHA256(t *testing.T, path string) string {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:])
}
