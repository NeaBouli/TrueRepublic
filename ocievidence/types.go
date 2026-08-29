// Package ocievidence verifies repository-only repeated OCI image builds.
package ocievidence

import (
	"encoding/json"
	"errors"
)

const (
	Schema         = "truerepublic.oci-evidence/v1"
	ContractSchema = "truerepublic.oci-build/v1"
	ReportSchema   = "truerepublic.oci-evidence-report/v1"
	MaxJSONBytes   = 1 << 20
	maxJSONDepth   = 32
)

type Contract struct {
	Schema      string           `json:"schema"`
	SourceRef   SourceRef        `json:"source_ref"`
	Repetitions int              `json:"repetitions"`
	BuildKit    BuildKitSettings `json:"buildkit"`
	Targets     []ContractTarget `json:"targets"`
}

type SourceRef struct {
	Kind    string `json:"kind"`
	Pattern string `json:"pattern"`
}

type BuildKitSettings struct {
	Driver           string `json:"driver"`
	BuilderIdentity  string `json:"builder_identity"`
	NoCache          *bool  `json:"no_cache"`
	Pull             *bool  `json:"pull"`
	Provenance       *bool  `json:"provenance"`
	SBOM             *bool  `json:"sbom"`
	RewriteTimestamp *bool  `json:"rewrite_timestamp"`
	SourceDateEpoch  string `json:"source_date_epoch"`
	Output           string `json:"output"`
}

type ContractTarget struct {
	ID         string            `json:"id"`
	Dockerfile string            `json:"dockerfile"`
	Context    string            `json:"context"`
	Platform   string            `json:"platform"`
	Runner     string            `json:"runner"`
	RunnerArch string            `json:"runner_arch"`
	BuildArgs  map[string]string `json:"build_args"`
	BaseImages []string          `json:"base_images"`
}

type Bundle struct {
	Schema         string           `json:"schema"`
	SourceRef      string           `json:"source_ref"`
	ContractSHA256 string           `json:"contract_sha256"`
	Platform       string           `json:"platform"`
	Claims         Claims           `json:"claims"`
	Targets        []TargetEvidence `json:"targets"`
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
		return errors.New("OCI status claims are incomplete")
	}
	c.Signed, c.Published, c.Production, c.present = *raw.Signed, *raw.Published, *raw.Production, true
	return nil
}

func (c Claims) explicitFalse() bool {
	return c.present && !c.Signed && !c.Published && !c.Production
}

type TargetEvidence struct {
	ID     string          `json:"id"`
	Builds []BuildEvidence `json:"builds"`
}

type BuildEvidence struct {
	File   string `json:"file"`
	SHA256 string `json:"sha256"`
}

type Descriptor struct {
	MediaType    string            `json:"mediaType"`
	Digest       string            `json:"digest"`
	Size         json.Number       `json:"size"`
	URLs         []string          `json:"urls,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
	Data         string            `json:"data,omitempty"`
	Platform     *OCIPlatform      `json:"platform,omitempty"`
	ArtifactType string            `json:"artifactType,omitempty"`
}

type OCIPlatform struct {
	Architecture string   `json:"architecture"`
	OS           string   `json:"os"`
	OSVersion    string   `json:"os.version,omitempty"`
	OSFeatures   []string `json:"os.features,omitempty"`
	Variant      string   `json:"variant,omitempty"`
	Features     []string `json:"features,omitempty"`
}

type OCIIndex struct {
	SchemaVersion int               `json:"schemaVersion"`
	MediaType     string            `json:"mediaType,omitempty"`
	Manifests     []Descriptor      `json:"manifests"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	Subject       *Descriptor       `json:"subject,omitempty"`
	ArtifactType  string            `json:"artifactType,omitempty"`
}

type OCIManifest struct {
	SchemaVersion int               `json:"schemaVersion"`
	MediaType     string            `json:"mediaType,omitempty"`
	ArtifactType  string            `json:"artifactType,omitempty"`
	Config        Descriptor        `json:"config"`
	Layers        []Descriptor      `json:"layers"`
	Subject       *Descriptor       `json:"subject,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
}

type Report struct {
	Schema     string        `json:"schema"`
	Valid      bool          `json:"valid"`
	Platform   string        `json:"platform"`
	Targets    int           `json:"targets"`
	Images     []ImageReport `json:"images"`
	LayerDiffs []LayerDiff   `json:"layer_diffs,omitempty"`
	Violations []string      `json:"violations"`
}

type ImageReport struct {
	ID         string   `json:"id"`
	Repetition int      `json:"repetition,omitempty"`
	Index      string   `json:"index_sha256"`
	Manifest   string   `json:"manifest_sha256"`
	Config     string   `json:"config_sha256"`
	Layers     []string `json:"layer_sha256"`
}

// LayerDiff is a bounded, content-free diagnostic for one unequal OCI layer.
type LayerDiff struct {
	Target     string           `json:"target"`
	LayerIndex int              `json:"layer_index"`
	Entries    []LayerEntryDiff `json:"entries"`
	Truncated  bool             `json:"truncated,omitempty"`
}

// LayerEntryDiff describes how one path differs between repetitions.
type LayerEntryDiff struct {
	Path   string          `json:"path"`
	Change string          `json:"change"`
	First  *LayerEntryInfo `json:"first,omitempty"`
	Second *LayerEntryInfo `json:"second,omitempty"`
}

// LayerEntryInfo exposes only bounded metadata and digests, never file bytes.
type LayerEntryInfo struct {
	Type           string `json:"type"`
	Mode           string `json:"mode"`
	UID            int    `json:"uid"`
	GID            int    `json:"gid"`
	Size           int64  `json:"size"`
	LinkTarget     string `json:"link_target,omitempty"`
	ContentSHA256  string `json:"content_sha256,omitempty"`
	MetadataSHA256 string `json:"metadata_sha256"`
}

type imageIdentity struct {
	Index            string
	Manifest         string
	Config           string
	Layers           []string
	LayerDescriptors []Descriptor
}
