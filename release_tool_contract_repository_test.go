package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type releaseToolPlatformContract struct {
	Schema     string            `json:"schema"`
	Tools      map[string]string `json:"tools"`
	Platforms  []releasePlatform `json:"platforms"`
	BaseImages []string          `json:"base_images"`
}

type releasePlatform struct {
	ID     string `json:"id"`
	Runner string `json:"runner"`
	Arch   string `json:"arch"`
}

func TestReleaseToolContractConsistency(t *testing.T) {
	t.Parallel()

	platformBytes, err := os.ReadFile("configs/release/tool-platform.json")
	if err != nil {
		t.Fatal(err)
	}
	var platform releaseToolPlatformContract
	if err := json.Unmarshal(platformBytes, &platform); err != nil {
		t.Fatalf("parse release tool contract: %v", err)
	}

	gatesBytes, err := os.ReadFile("configs/security/gates.json")
	if err != nil {
		t.Fatal(err)
	}
	var gates securityGateContract
	if err := json.Unmarshal(gatesBytes, &gates); err != nil {
		t.Fatalf("parse security gate contract: %v", err)
	}

	buildBytes, err := os.ReadFile("configs/build/deterministic-linux-daemon.json")
	if err != nil {
		t.Fatal(err)
	}
	var build struct {
		GoVersion string `json:"go_version"`
	}
	if err := json.Unmarshal(buildBytes, &build); err != nil {
		t.Fatalf("parse deterministic build contract: %v", err)
	}

	workflowBytes, err := os.ReadFile(".github/workflows/reproducible-daemon.yml")
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(workflowBytes)

	if violations := releaseToolConsistencyViolations(platform, gates, build.GoVersion, workflow); len(violations) != 0 {
		t.Fatalf("release tool pin drift:\n- %s", strings.Join(violations, "\n- "))
	}

	t.Run("rejects security gate tool drift", func(t *testing.T) {
		mutated := gates
		mutated.Tools = map[string]string{}
		for name, version := range gates.Tools {
			mutated.Tools[name] = version
		}
		mutated.Tools["cyclonedx_gomod"] = "v0.0.0"
		if violations := releaseToolConsistencyViolations(platform, mutated, build.GoVersion, workflow); len(violations) == 0 {
			t.Fatal("security gate cyclonedx-gomod drift was not detected")
		}
	})

	t.Run("rejects node toolchain drift", func(t *testing.T) {
		mutated := gates
		mutated.Toolchains = map[string]string{}
		for name, version := range gates.Toolchains {
			mutated.Toolchains[name] = version
		}
		mutated.Toolchains["node"] = "0.0.0"
		if violations := releaseToolConsistencyViolations(platform, mutated, build.GoVersion, workflow); len(violations) == 0 {
			t.Fatal("node toolchain drift was not detected")
		}
	})

	t.Run("rejects hardcoded workflow tool install", func(t *testing.T) {
		mutated := strings.Replace(
			workflow,
			"jq -er '.tools.cyclonedx_gomod' configs/security/gates.json",
			"echo v1.10.0", 1,
		)
		if violations := releaseToolConsistencyViolations(platform, gates, build.GoVersion, mutated); len(violations) == 0 {
			t.Fatal("a hardcoded workflow tool install was not detected")
		}
	})
}

func releaseToolConsistencyViolations(platform releaseToolPlatformContract, gates securityGateContract, buildGoVersion, workflow string) []string {
	var violations []string
	if platform.Schema != "truerepublic.release-tool-platform/v1" {
		violations = append(violations, "release tool contract schema mismatch")
	}
	for _, tool := range []string{"cyclonedx_gomod", "cyclonedx_npm", "npm"} {
		if platform.Tools[tool] == "" || gates.Tools[tool] != platform.Tools[tool] {
			violations = append(violations, "release/security "+tool+" pin mismatch")
		}
	}
	if platform.Tools["node"] == "" || gates.Toolchains["node"] != platform.Tools["node"] {
		violations = append(violations, "release/security node toolchain mismatch")
	}
	if gates.Toolchains["go"] == "" || gates.Toolchains["go"] != buildGoVersion {
		violations = append(violations, "security/build Go toolchain mismatch")
	}
	for _, tool := range []string{"cyclonedx_gomod", "cyclonedx_npm", "npm"} {
		if !strings.Contains(workflow, "jq -er '.tools."+tool+"' configs/security/gates.json") {
			violations = append(violations, "reproducible-daemon workflow must install "+tool+" from the security gate contract")
		}
	}
	if !strings.Contains(workflow, "go-version: '"+gates.Toolchains["go"]+"'") {
		violations = append(violations, "reproducible-daemon workflow Go toolchain does not match the security gate contract")
	}
	if !strings.Contains(workflow, "node-version: '"+platform.Tools["node"]+"'") {
		violations = append(violations, "reproducible-daemon workflow Node toolchain does not match the release tool contract")
	}
	return violations
}
