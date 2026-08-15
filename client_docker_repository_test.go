package main

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
)

var dockerFromLine = regexp.MustCompile(`^\s*FROM\s+`)
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

	rootDockerfile, err := os.ReadFile("Dockerfile")
	if err != nil {
		t.Fatal(err)
	}
	clientDockerfile, err := os.ReadFile("client-web/Dockerfile")
	if err != nil {
		t.Fatal(err)
	}
	maintainedDockerfiles := string(rootDockerfile) + "\n" + string(clientDockerfile)
	for _, image := range contract.BaseImages {
		if !strings.Contains(maintainedDockerfiles, image) {
			t.Errorf("release contract base image is not bound to a maintained Dockerfile: %s", image)
		}
	}
}
