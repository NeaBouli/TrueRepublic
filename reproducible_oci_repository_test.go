package main

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

type repositoryOCIContract struct {
	Schema    string `json:"schema"`
	SourceRef struct {
		Kind    string `json:"kind"`
		Pattern string `json:"pattern"`
	} `json:"source_ref"`
	Repetitions int `json:"repetitions"`
	BuildKit    struct {
		Driver           string `json:"driver"`
		BuilderIdentity  string `json:"builder_identity"`
		NoCache          bool   `json:"no_cache"`
		Pull             bool   `json:"pull"`
		Provenance       bool   `json:"provenance"`
		SBOM             bool   `json:"sbom"`
		RewriteTimestamp bool   `json:"rewrite_timestamp"`
		SourceDateEpoch  string `json:"source_date_epoch"`
		Output           string `json:"output"`
	} `json:"buildkit"`
	Targets []struct {
		ID         string            `json:"id"`
		Dockerfile string            `json:"dockerfile"`
		Context    string            `json:"context"`
		Platform   string            `json:"platform"`
		Runner     string            `json:"runner"`
		RunnerArch string            `json:"runner_arch"`
		BuildArgs  map[string]string `json:"build_args"`
		BaseImages []string          `json:"base_images"`
	} `json:"targets"`
}

func TestReproducibleOCIRepositoryContract(t *testing.T) {
	t.Parallel()

	contractBytes, err := os.ReadFile("configs/build/reproducible-oci.json")
	if err != nil {
		t.Fatal(err)
	}
	var contract repositoryOCIContract
	if err = json.Unmarshal(contractBytes, &contract); err != nil {
		t.Fatal(err)
	}
	workflowBytes, err := os.ReadFile(".github/workflows/reproducible-daemon.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(workflowBytes)
	dockerfiles := map[string]string{}
	for _, path := range []string{"Dockerfile", "client-web/Dockerfile"} {
		contents, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		dockerfiles[path] = string(contents)
	}
	if violations := reproducibleOCIRepositoryViolations(contract, workflow, dockerfiles); len(violations) != 0 {
		t.Fatalf("reproducible OCI repository contract violations:\n- %s", strings.Join(violations, "\n- "))
	}

	for _, script := range []string{"scripts/test-reproducible-oci.sh", "scripts/verify-reproducible-oci.sh"} {
		info, statErr := os.Stat(script)
		if statErr != nil {
			t.Fatal(statErr)
		}
		if info.Mode()&0111 == 0 {
			t.Fatalf("%s must be executable", script)
		}
	}

	t.Run("rejects missing no-cache", func(t *testing.T) {
		mutated := contract
		mutated.BuildKit.NoCache = false
		if len(reproducibleOCIRepositoryViolations(mutated, workflow, dockerfiles)) == 0 {
			t.Fatal("contract accepted cache-enabled builds")
		}
	})
	t.Run("rejects target platform drift", func(t *testing.T) {
		mutated := contract
		mutated.Targets = append([]struct {
			ID         string            `json:"id"`
			Dockerfile string            `json:"dockerfile"`
			Context    string            `json:"context"`
			Platform   string            `json:"platform"`
			Runner     string            `json:"runner"`
			RunnerArch string            `json:"runner_arch"`
			BuildArgs  map[string]string `json:"build_args"`
			BaseImages []string          `json:"base_images"`
		}{}, contract.Targets...)
		mutated.Targets[0].Platform = "linux/s390x"
		if len(reproducibleOCIRepositoryViolations(mutated, workflow, dockerfiles)) == 0 {
			t.Fatal("contract accepted platform drift")
		}
	})
	t.Run("rejects workflow push", func(t *testing.T) {
		mutated := workflow + "\n      - run: docker buildx build --push .\n"
		if len(reproducibleOCIRepositoryViolations(contract, mutated, dockerfiles)) == 0 {
			t.Fatal("workflow accepted an image push")
		}
	})
}

func reproducibleOCIRepositoryViolations(contract repositoryOCIContract, workflow string, dockerfiles map[string]string) []string {
	var violations []string
	if contract.Schema != "truerepublic.oci-build/v1" || contract.SourceRef.Kind != "git-commit" || contract.SourceRef.Pattern != "^[0-9a-f]{40}$" || contract.Repetitions != 2 {
		violations = append(violations, "contract identity mismatch")
	}
	settings := contract.BuildKit
	if settings.Driver != "docker-container" || settings.BuilderIdentity != "runner-provided-unpinned" || !settings.NoCache || !settings.Pull || settings.Provenance || settings.SBOM || !settings.RewriteTimestamp || settings.SourceDateEpoch != "git-commit" || settings.Output != "oci" {
		violations = append(violations, "deterministic BuildKit settings mismatch")
	}
	expected := []struct {
		id, dockerfile, context, platform, runner, arch string
		args                                            map[string]string
	}{
		{"daemon-linux-amd64", "Dockerfile", ".", "linux/amd64", "ubuntu-24.04", "x86_64", map[string]string{"VERSION": "{{source_ref}}"}},
		{"client-web-linux-amd64", "client-web/Dockerfile", "client-web", "linux/amd64", "ubuntu-24.04", "x86_64", map[string]string{}},
		{"daemon-linux-arm64", "Dockerfile", ".", "linux/arm64", "ubuntu-24.04-arm", "aarch64", map[string]string{"VERSION": "{{source_ref}}"}},
		{"client-web-linux-arm64", "client-web/Dockerfile", "client-web", "linux/arm64", "ubuntu-24.04-arm", "aarch64", map[string]string{}},
	}
	if len(contract.Targets) != len(expected) {
		violations = append(violations, "target count mismatch")
	} else {
		allBases := []string{}
		for index, want := range expected {
			got := contract.Targets[index]
			if got.ID != want.id || got.Dockerfile != want.dockerfile || got.Context != want.context || got.Platform != want.platform || got.Runner != want.runner || got.RunnerArch != want.arch || !reflect.DeepEqual(got.BuildArgs, want.args) || len(got.BaseImages) != 2 {
				violations = append(violations, "target contract mismatch: "+want.id)
				continue
			}
			contents, ok := dockerfiles[got.Dockerfile]
			if !ok {
				violations = append(violations, "Dockerfile unavailable: "+got.Dockerfile)
				continue
			}
			for _, image := range got.BaseImages {
				if !strings.Contains(contents, "FROM "+image) {
					violations = append(violations, "base image is not active in "+got.Dockerfile)
				}
			}
			if index < 2 {
				allBases = append(allBases, got.BaseImages...)
			}
		}
		if len(allBases) == 4 {
			if binding := dockerBaseBindingViolations(allBases, dockerfiles); len(binding) != 0 {
				violations = append(violations, binding...)
			}
		}
	}

	requiredWorkflow := []string{
		"reproducible-oci:",
		"platform: linux/amd64",
		"runner: ubuntu-24.04",
		"platform: linux/arm64",
		"runner: ubuntu-24.04-arm",
		"for repetition in 1 2",
		"docker buildx create",
		"--driver docker-container",
		"docker buildx build",
		"--no-cache",
		"--pull",
		"--provenance=false",
		"--sbom=false",
		"SOURCE_DATE_EPOCH=",
		"type=oci,dest=",
		"rewrite-timestamp=true",
		"verify-reproducible-oci.sh",
	}
	for _, required := range requiredWorkflow {
		if !strings.Contains(workflow, required) {
			violations = append(violations, "workflow lacks "+required)
		}
	}
	for _, trigger := range []string{"configs/build/reproducible-oci.json", "ocievidence/**", "cmd/oci-evidence/**", "scripts/test-reproducible-oci.sh", "scripts/verify-reproducible-oci.sh", "reproducible_oci_repository_test.go", ".dockerignore", "scripts/docker-entrypoint.sh", "client-web/**"} {
		if strings.Count(workflow, "'"+trigger+"'") != 2 {
			violations = append(violations, "workflow trigger mismatch: "+trigger)
		}
	}
	daemon := dockerfiles["Dockerfile"]
	for _, required := range []string{
		"GOFLAGS=-mod=readonly go build",
		"-trimpath",
		"-buildvcs=false",
		"-buildid=",
		"-extldflags=-Wl,--build-id=none",
		"rm -f /var/log/dpkg.log /var/log/apt/*.log /var/log/alternatives.log",
		"/^truerepublic:/ s/^([^:]*:[^:]*):[0-9]+:/\\1::/",
	} {
		if !strings.Contains(daemon, required) {
			violations = append(violations, "daemon Dockerfile lacks reproducibility control: "+required)
		}
	}
	jobStart := strings.Index(workflow, "  reproducible-oci:")
	if jobStart < 0 {
		violations = append(violations, "OCI job is missing")
	} else {
		job := workflow[jobStart:]
		for _, forbidden := range []string{"docker login", "--push", "type=registry", "cosign", "sigstore", "docker attest", "kubectl", "docker compose"} {
			if strings.Contains(job, forbidden) {
				violations = append(violations, "OCI job contains forbidden operation: "+forbidden)
			}
		}
	}
	return violations
}
