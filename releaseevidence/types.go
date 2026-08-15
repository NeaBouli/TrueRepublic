// Package releaseevidence verifies offline release bundles against the exact
// deterministic daemon-build and release tool/platform contracts.
package releaseevidence

import "errors"

const (
	Schema              = "truerepublic.release-evidence/v1"
	BuildSchema         = "truerepublic.daemon-build/v1"
	BuildEvidenceSchema = "truerepublic.daemon-build-evidence/v1"
	ToolSchema          = "truerepublic.release-tool-platform/v1"
	ProvenanceSchema    = "truerepublic.unsigned-provenance/v1"
	MaxJSONBytes        = 1024 * 1024
	maxJSONDepth        = 32
)

var targetIDs = []string{"linux-amd64", "linux-arm64"}

type Bundle struct {
	Schema              string     `json:"schema"`
	SourceRef           string     `json:"source_ref"`
	BuildContractSHA256 string     `json:"build_contract_sha256"`
	ToolContractSHA256  string     `json:"tool_contract_sha256"`
	Tools               Tools      `json:"tools"`
	Claims              Claims     `json:"claims"`
	Provenance          FileDigest `json:"provenance"`
	Targets             []Target   `json:"targets"`
	SBOMs               []SBOM     `json:"sboms"`
}

type Tools struct {
	CycloneDXGoMod string `json:"cyclonedx_gomod"`
	CycloneDXNPM   string `json:"cyclonedx_npm"`
	Node           string `json:"node"`
	NPM            string `json:"npm"`
}

type Claims struct {
	Signed     bool `json:"signed"`
	Published  bool `json:"published"`
	Production bool `json:"production"`
	present    bool
}

func (c *Claims) UnmarshalJSON(data []byte) error {
	var raw struct {
		Signed     *bool `json:"signed"`
		Published  *bool `json:"published"`
		Production *bool `json:"production"`
	}
	if err := parseBytes(data, &raw); err != nil {
		return err
	}
	if raw.Signed == nil || raw.Published == nil || raw.Production == nil {
		return errors.New("release status claims are incomplete")
	}
	c.Signed, c.Published, c.Production, c.present = *raw.Signed, *raw.Published, *raw.Production, true
	return nil
}

func (c Claims) explicitFalse() bool {
	return c.present && !c.Signed && !c.Published && !c.Production
}

type FileDigest struct {
	File   string `json:"file"`
	SHA256 string `json:"sha256"`
}

type Target struct {
	ID            string `json:"id"`
	Artifact      string `json:"artifact"`
	SHA256        string `json:"sha256"`
	ChecksumsFile string `json:"checksums_file"`
	MetadataFile  string `json:"metadata_file"`
}

type SBOM struct {
	Component string `json:"component"`
	File      string `json:"file"`
	SHA256    string `json:"sha256"`
}

type ToolContract struct {
	Schema     string     `json:"schema"`
	Tools      Tools      `json:"tools"`
	Platforms  []Platform `json:"platforms"`
	BaseImages []string   `json:"base_images"`
}

type Platform struct {
	ID     string `json:"id"`
	Runner string `json:"runner"`
	Arch   string `json:"arch"`
}
type Provenance struct {
	Schema              string            `json:"schema"`
	SourceRef           string            `json:"source_ref"`
	BuildContractSHA256 string            `json:"build_contract_sha256"`
	ToolContractSHA256  string            `json:"tool_contract_sha256"`
	Claims              Claims            `json:"claims"`
	Targets             []ProvenanceEntry `json:"targets"`
	SBOMs               []ProvenanceEntry `json:"sboms"`
}
type ProvenanceEntry struct {
	ID     string `json:"id"`
	SHA256 string `json:"sha256"`
}

type BuildContract struct {
	Schema      string                         `json:"schema"`
	Binary      string                         `json:"binary"`
	MainPackage string                         `json:"main_package"`
	GoVersion   string                         `json:"go_version"`
	CGOEnabled  string                         `json:"cgo_enabled"`
	SourceRef   struct{ Kind, Pattern string } `json:"source_ref"`
	BuildFlags  struct {
		Trimpath bool     `json:"trimpath"`
		BuildVCS bool     `json:"buildvcs"`
		Mod      string   `json:"mod"`
		LDFlags  []string `json:"ldflags"`
	} `json:"build_flags"`
	Targets []BuildTarget `json:"targets"`
}

type BuildTarget struct {
	ID, GOOS, GOARCH, CIRunner, RunnerArch, Artifact string
}

func (t *BuildTarget) UnmarshalJSON(data []byte) error {
	type raw struct {
		ID         string `json:"id"`
		GOOS       string `json:"goos"`
		GOARCH     string `json:"goarch"`
		CIRunner   string `json:"ci_runner"`
		RunnerArch string `json:"runner_arch"`
		Artifact   string `json:"artifact"`
	}
	var r raw
	if err := parseBytes(data, &r); err != nil {
		return err
	}
	*t = BuildTarget(r)
	return nil
}

type Report struct {
	Schema     string   `json:"schema"`
	Valid      bool     `json:"valid"`
	Targets    int      `json:"targets"`
	SBOMs      int      `json:"sboms"`
	Violations []string `json:"violations"`
}
