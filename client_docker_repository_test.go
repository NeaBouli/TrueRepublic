package main

import (
	"os"
	"strings"
	"testing"
)

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
