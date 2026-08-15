package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

var dockerFromLine = regexp.MustCompile(`^\s*FROM\s+`)
var dockerFromReference = regexp.MustCompile(`^\s*FROM\s+(\S+)(?:\s+AS\s+\S+)?\s*$`)
var pinnedDockerBase = regexp.MustCompile(`^\s*FROM\s+[^\s]+@sha256:[0-9a-f]{64}(?:\s+AS\s+\S+)?\s*$`)

func TestClientDockerBuilderIncludesBuildScripts(t *testing.T) {
	t.Parallel()

	dockerfile, err := os.ReadFile("client-web/Dockerfile")
	if err != nil {
		t.Fatal(err)
	}
	contents := string(dockerfile)
	copyScripts := strings.Index(contents, "COPY scripts ./scripts")
	build := strings.Index(contents, "RUN npm run build")
	if copyScripts < 0 {
		t.Fatal("client Docker builder omits the scripts required by npm run build")
	}
	if build < 0 {
		t.Fatal("client Docker builder omits npm run build")
	}
	if copyScripts > build {
		t.Fatal("client Docker builder copies build scripts after npm run build")
	}

	packageJSON, err := os.ReadFile("client-web/package.json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(packageJSON), "node scripts/bundle-budget.mjs") {
		t.Fatal("client build no longer invokes the maintained bundle budget")
	}
}

func TestMaintainedDockerBasesAreDigestPinned(t *testing.T) {
	t.Parallel()

	for _, path := range []string{"Dockerfile", "client-web/Dockerfile"} {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		fromLines := 0
		for _, line := range strings.Split(string(contents), "\n") {
			if !dockerFromLine.MatchString(line) {
				continue
			}
			fromLines++
			if !pinnedDockerBase.MatchString(line) {
				t.Errorf("%s contains an unpinned or malformed base image: %q", path, line)
			}
		}
		if fromLines != 2 {
			t.Errorf("%s has %d FROM lines; want exactly 2 maintained stages", path, fromLines)
		}
	}
}

func TestReleaseToolContractBindsMaintainedDockerBases(t *testing.T) {
	t.Parallel()

	contractBytes, err := os.ReadFile("configs/release/tool-platform.json")
	if err != nil {
		t.Fatal(err)
	}
	var contract struct {
		BaseImages []string `json:"base_images"`
	}
	if err := json.Unmarshal(contractBytes, &contract); err != nil {
		t.Fatal(err)
	}
	if len(contract.BaseImages) != 4 {
		t.Fatalf("release contract has %d base images; want 4", len(contract.BaseImages))
	}

	dockerfiles := map[string]string{}
	for _, path := range []string{"Dockerfile", "client-web/Dockerfile"} {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		dockerfiles[path] = string(contents)
	}
	if violations := dockerBaseBindingViolations(contract.BaseImages, dockerfiles); len(violations) != 0 {
		t.Fatalf("release contract does not bind the maintained Docker bases:\n- %s", strings.Join(violations, "\n- "))
	}

	t.Run("rejects comment-only base reference", func(t *testing.T) {
		mutated := cloneDockerfiles(dockerfiles)
		mutated["client-web/Dockerfile"] = strings.Replace(
			mutated["client-web/Dockerfile"],
			"FROM node:22.22.2-alpine@sha256:", "# FROM node:22.22.2-alpine@sha256:", 1,
		)
		if violations := dockerBaseBindingViolations(contract.BaseImages, mutated); len(violations) == 0 {
			t.Fatal("comment-only base reference satisfied the release contract binding")
		}
	})

	t.Run("rejects base digest drift", func(t *testing.T) {
		mutated := cloneDockerfiles(dockerfiles)
		mutated["Dockerfile"] = strings.Replace(
			mutated["Dockerfile"],
			"golang:1.26.6-bookworm@sha256:116d58cb", "golang:1.26.6-bookworm@sha256:016d58cb", 1,
		)
		if violations := dockerBaseBindingViolations(contract.BaseImages, mutated); len(violations) == 0 {
			t.Fatal("base image digest drift satisfied the release contract binding")
		}
	})

	t.Run("rejects extra active stage", func(t *testing.T) {
		mutated := cloneDockerfiles(dockerfiles)
		mutated["client-web/Dockerfile"] += "\nFROM " + contract.BaseImages[3] + "\n"
		if violations := dockerBaseBindingViolations(contract.BaseImages, mutated); len(violations) == 0 {
			t.Fatal("an extra active FROM stage satisfied the release contract binding")
		}
	})
}

func cloneDockerfiles(dockerfiles map[string]string) map[string]string {
	clone := make(map[string]string, len(dockerfiles))
	for path, contents := range dockerfiles {
		clone[path] = contents
	}
	return clone
}

func dockerBaseBindingViolations(contractImages []string, dockerfiles map[string]string) []string {
	parsed := make([]string, 0, len(contractImages))
	for _, path := range []string{"Dockerfile", "client-web/Dockerfile"} {
		contents, exists := dockerfiles[path]
		if !exists {
			return []string{path + " is unavailable"}
		}
		for _, line := range strings.Split(contents, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") || !dockerFromLine.MatchString(line) {
				continue
			}
			match := dockerFromReference.FindStringSubmatch(line)
			if match == nil {
				return []string{fmt.Sprintf("%s has a malformed active FROM line: %q", path, line)}
			}
			parsed = append(parsed, match[1])
		}
	}
	if len(parsed) != len(contractImages) {
		return []string{fmt.Sprintf("maintained Dockerfiles declare %d active FROM images; the release contract pins %d", len(parsed), len(contractImages))}
	}
	remaining := make(map[string]int, len(contractImages))
	for _, image := range contractImages {
		remaining[image]++
	}
	for _, image := range parsed {
		remaining[image]--
	}
	for image, count := range remaining {
		if count > 0 {
			return []string{"release contract base image is not an active Dockerfile FROM reference: " + image}
		}
		if count < 0 {
			return []string{"active Dockerfile FROM reference is not pinned by the release contract: " + image}
		}
	}
	return nil
}
