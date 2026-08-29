package ocievidence

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	maxArchiveBytes      int64 = 2 << 30
	maxBufferedBlobBytes int64 = 64 << 20
	maxArchiveEntries          = 4096
	maxLayers                  = 128
)

var (
	hexDigestRE = regexp.MustCompile(`^[0-9a-f]{64}$`)
	sourceRefRE = regexp.MustCompile(`^[0-9a-f]{40}$`)
	imageRefRE  = regexp.MustCompile(`^[^[:space:]@]+@sha256:[0-9a-f]{64}$`)
	targetIDRE  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)
)

type archiveBlob struct {
	Size int64
	Data []byte
}

func VerifyDirectory(evidenceDir, contractPath string) Report {
	report := Report{Schema: ReportSchema, Images: []ImageReport{}, Violations: []string{}}
	fail := func(message string) { report.Violations = append(report.Violations, message) }

	base, err := filepath.Abs(evidenceDir)
	if err != nil {
		fail("evidence path is invalid")
		return report
	}
	base, err = filepath.EvalSymlinks(base)
	if err != nil {
		fail("evidence directory is unavailable")
		return report
	}
	info, err := os.Stat(base)
	if err != nil || !info.IsDir() {
		fail("evidence directory is unavailable")
		return report
	}

	manifestPath, err := safeMember(base, "oci-evidence.json")
	if err != nil {
		fail("OCI evidence manifest is unsafe")
		return report
	}
	var bundle Bundle
	if err := parseFile(manifestPath, &bundle); err != nil {
		fail("OCI evidence manifest is invalid")
		return report
	}
	var contract Contract
	if err := parseFile(contractPath, &contract); err != nil {
		fail("OCI build contract is invalid")
		return report
	}
	contractHash, err := hashRegularFile(contractPath, MaxJSONBytes)
	if err != nil {
		fail("OCI build contract digest is unavailable")
		return report
	}

	report.Platform = bundle.Platform
	report.Targets = len(bundle.Targets)
	if bundle.Schema != Schema {
		fail("OCI evidence schema mismatch")
	}
	if !sourceRefRE.MatchString(bundle.SourceRef) {
		fail("source ref is malformed")
	}
	if bundle.ContractSHA256 != contractHash {
		fail("OCI build contract digest mismatch")
	}
	if !bundle.Claims.explicitFalse() {
		fail("OCI status claims must remain false")
	}
	if err := validateContract(contract); err != nil {
		fail("OCI build contract mismatch: " + err.Error())
		return report
	}

	expected := targetsForPlatform(contract.Targets, bundle.Platform)
	if len(expected) != 2 {
		fail("evidence platform is unsupported")
		return report
	}
	if len(bundle.Targets) != len(expected) {
		fail("OCI target count mismatch")
	}
	seenFiles := map[string]bool{"oci-evidence.json": true}
	for index, target := range bundle.Targets {
		if index >= len(expected) || target.ID != expected[index].ID {
			fail("OCI target order or set mismatch")
			continue
		}
		if len(target.Builds) != contract.Repetitions {
			fail("OCI build repetition count mismatch for " + target.ID)
			continue
		}
		identities := make([]imageIdentity, 0, len(target.Builds))
		archivePaths := make([]string, 0, len(target.Builds))
		for _, build := range target.Builds {
			if !hexDigestRE.MatchString(build.SHA256) {
				fail("OCI archive digest is malformed for " + target.ID)
				continue
			}
			if seenFiles[build.File] {
				fail("OCI archive file is duplicated")
				continue
			}
			seenFiles[build.File] = true
			archivePath, err := safeMember(base, build.File)
			if err != nil {
				fail("OCI archive path is unsafe for " + target.ID)
				continue
			}
			archiveHash, err := hashRegularFile(archivePath, maxArchiveBytes)
			if err != nil || archiveHash != build.SHA256 {
				fail("OCI archive digest mismatch for " + target.ID)
				continue
			}
			identity, err := inspectOCIArchive(archivePath, bundle.Platform)
			if err != nil {
				fail("OCI archive is invalid for " + target.ID + ": " + err.Error())
				continue
			}
			identities = append(identities, identity)
			archivePaths = append(archivePaths, archivePath)
		}
		if len(identities) == contract.Repetitions && !sameIdentity(identities[0], identities[1]) {
			for repetition, identity := range identities {
				report.Images = append(report.Images, reportIdentity(target.ID, repetition+1, identity))
			}
			fail("repeated OCI image digests differ for " + target.ID)
			diffs, problems := diagnoseRepeatedLayers(target.ID, archivePaths[0], archivePaths[1], identities[0], identities[1])
			report.LayerDiffs = append(report.LayerDiffs, diffs...)
			for _, problem := range problems {
				fail(problem)
			}
		} else if len(identities) == contract.Repetitions {
			report.Images = append(report.Images, reportIdentity(target.ID, 0, identities[0]))
		}
	}
	report.Valid = len(report.Violations) == 0
	return report
}

func reportIdentity(id string, repetition int, identity imageIdentity) ImageReport {
	return ImageReport{
		ID:         id,
		Repetition: repetition,
		Index:      identity.Index,
		Manifest:   identity.Manifest,
		Config:     identity.Config,
		Layers:     append([]string(nil), identity.Layers...),
	}
}

func validateContract(contract Contract) error {
	if contract.Schema != ContractSchema || contract.SourceRef != (SourceRef{Kind: "git-commit", Pattern: "^[0-9a-f]{40}$"}) {
		return errors.New("schema or source identity")
	}
	settings := contract.BuildKit
	if contract.Repetitions != 2 || settings.NoCache == nil || settings.Pull == nil || settings.Provenance == nil || settings.SBOM == nil || settings.RewriteTimestamp == nil {
		return errors.New("incomplete deterministic settings")
	}
	if settings.Driver != "docker-container" || settings.BuilderIdentity != "runner-provided-unpinned" || !*settings.NoCache || !*settings.Pull || *settings.Provenance || *settings.SBOM || !*settings.RewriteTimestamp || settings.SourceDateEpoch != "git-commit" || settings.Output != "oci" {
		return errors.New("unsafe deterministic settings")
	}
	if len(contract.Targets) != 4 {
		return errors.New("target count")
	}
	seen := map[string]bool{}
	platformCount := map[string]int{}
	expectedTargets := []struct {
		id, dockerfile, context, platform, runner, arch string
		daemon                                          bool
	}{
		{"daemon-linux-amd64", "Dockerfile", ".", "linux/amd64", "ubuntu-24.04", "x86_64", true},
		{"client-web-linux-amd64", "client-web/Dockerfile", "client-web", "linux/amd64", "ubuntu-24.04", "x86_64", false},
		{"daemon-linux-arm64", "Dockerfile", ".", "linux/arm64", "ubuntu-24.04-arm", "aarch64", true},
		{"client-web-linux-arm64", "client-web/Dockerfile", "client-web", "linux/arm64", "ubuntu-24.04-arm", "aarch64", false},
	}
	for index, target := range contract.Targets {
		if !targetIDRE.MatchString(target.ID) || seen[target.ID] {
			return errors.New("target identity")
		}
		seen[target.ID] = true
		want := expectedTargets[index]
		if target.ID != want.id || target.Dockerfile != want.dockerfile || target.Context != want.context || target.Platform != want.platform || target.Runner != want.runner || target.RunnerArch != want.arch {
			return errors.New("target definition")
		}
		if !safeRepositoryPath(target.Dockerfile, false) || !safeRepositoryPath(target.Context, true) {
			return errors.New("target path")
		}
		if target.Platform != "linux/amd64" && target.Platform != "linux/arm64" {
			return errors.New("target platform")
		}
		if target.Platform == "linux/amd64" && (target.Runner != "ubuntu-24.04" || target.RunnerArch != "x86_64") {
			return errors.New("amd64 runner")
		}
		if target.Platform == "linux/arm64" && (target.Runner != "ubuntu-24.04-arm" || target.RunnerArch != "aarch64") {
			return errors.New("arm64 runner")
		}
		platformCount[target.Platform]++
		if len(target.BaseImages) != 2 {
			return errors.New("base image count")
		}
		for _, image := range target.BaseImages {
			if !imageRefRE.MatchString(image) {
				return errors.New("base image pin")
			}
		}
		for name, value := range target.BuildArgs {
			if name == "" || strings.ContainsAny(name, " \t\r\n=") || value == "" {
				return errors.New("build argument")
			}
		}
		if want.daemon {
			if len(target.BuildArgs) != 1 || target.BuildArgs["VERSION"] != "{{source_ref}}" {
				return errors.New("daemon build arguments")
			}
		} else if len(target.BuildArgs) != 0 {
			return errors.New("client build arguments")
		}
	}
	if platformCount["linux/amd64"] != 2 || platformCount["linux/arm64"] != 2 {
		return errors.New("platform target coverage")
	}
	return nil
}

func targetsForPlatform(targets []ContractTarget, platform string) []ContractTarget {
	var selected []ContractTarget
	for _, target := range targets {
		if target.Platform == platform {
			selected = append(selected, target)
		}
	}
	return selected
}

func safeRepositoryPath(value string, allowDot bool) bool {
	if value == "." {
		return allowDot
	}
	return value != "" && value == path.Clean(value) && !path.IsAbs(value) && value != ".." && !strings.HasPrefix(value, "../") && !strings.Contains(value, `\`)
}

func safeMember(base, name string) (string, error) {
	if name == "" || name != filepath.Base(name) || filepath.IsAbs(name) || name == "." || name == ".." || strings.Contains(name, `\`) {
		return "", errors.New("unsafe member")
	}
	member := filepath.Join(base, name)
	info, err := os.Lstat(member)
	if err != nil || !info.Mode().IsRegular() {
		return "", errors.New("member is unavailable or not regular")
	}
	return member, nil
}

func hashRegularFile(filePath string, limit int64) (string, error) {
	info, err := os.Lstat(filePath)
	if err != nil || !info.Mode().IsRegular() || info.Size() < 0 || info.Size() > limit {
		return "", errors.New("file is unavailable, unsafe, or oversized")
	}
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	hash := sha256.New()
	written, err := io.Copy(hash, io.LimitReader(file, limit+1))
	if err != nil || written != info.Size() || written > limit {
		return "", errors.New("file read failed or exceeded limit")
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func inspectOCIArchive(filePath, platform string) (imageIdentity, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return imageIdentity{}, err
	}
	defer func() { _ = file.Close() }()
	reader := tar.NewReader(io.LimitReader(file, maxArchiveBytes+1))
	blobs := map[string]archiveBlob{}
	seen := map[string]bool{}
	var indexRaw, layoutRaw []byte
	var total, buffered int64
	entries := 0
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return imageIdentity{}, errors.New("tar stream")
		}
		entries++
		if entries > maxArchiveEntries || header.Size < 0 {
			return imageIdentity{}, errors.New("tar entry bound")
		}
		name, err := normalizedTarPath(header.Name, header.Typeflag == tar.TypeDir)
		if err != nil {
			return imageIdentity{}, errors.New("unsafe tar path")
		}
		if seen[name] {
			return imageIdentity{}, errors.New("duplicate tar path")
		}
		seen[name] = true
		if header.Typeflag == tar.TypeDir {
			if header.Size != 0 {
				return imageIdentity{}, errors.New("directory tar entry has content")
			}
			continue
		}
		if header.Typeflag != tar.TypeReg {
			return imageIdentity{}, errors.New("non-regular tar entry")
		}
		total += header.Size
		if total > maxArchiveBytes {
			return imageIdentity{}, errors.New("tar content bound")
		}
		contentHash := sha256.New()
		var buffer bytes.Buffer
		writer := io.Writer(contentHash)
		if header.Size <= MaxJSONBytes {
			buffered += header.Size
			if buffered > maxBufferedBlobBytes {
				return imageIdentity{}, errors.New("buffered blob bound")
			}
			writer = io.MultiWriter(contentHash, &buffer)
		}
		written, err := io.Copy(writer, io.LimitReader(reader, header.Size+1))
		if err != nil || written != header.Size {
			return imageIdentity{}, errors.New("tar entry read")
		}
		digest := hex.EncodeToString(contentHash.Sum(nil))
		switch name {
		case "index.json":
			if header.Size > MaxJSONBytes {
				return imageIdentity{}, errors.New("index bound")
			}
			indexRaw = append([]byte(nil), buffer.Bytes()...)
		case "oci-layout":
			if header.Size > MaxJSONBytes {
				return imageIdentity{}, errors.New("layout bound")
			}
			layoutRaw = append([]byte(nil), buffer.Bytes()...)
		default:
			const prefix = "blobs/sha256/"
			if !strings.HasPrefix(name, prefix) || !hexDigestRE.MatchString(strings.TrimPrefix(name, prefix)) || strings.TrimPrefix(name, prefix) != digest {
				return imageIdentity{}, errors.New("unrecognized or mismatched blob")
			}
			blobs[digest] = archiveBlob{Size: header.Size, Data: append([]byte(nil), buffer.Bytes()...)}
		}
	}
	if len(indexRaw) == 0 || len(layoutRaw) == 0 {
		return imageIdentity{}, errors.New("layout or index is missing")
	}
	var layout struct {
		ImageLayoutVersion string `json:"imageLayoutVersion"`
	}
	if parseBytes(layoutRaw, &layout) != nil || layout.ImageLayoutVersion != "1.0.0" {
		return imageIdentity{}, errors.New("OCI layout")
	}
	var index OCIIndex
	if parseBytes(indexRaw, &index) != nil || index.SchemaVersion != 2 || len(index.Manifests) != 1 {
		return imageIdentity{}, errors.New("OCI index")
	}
	manifestDescriptor := index.Manifests[0]
	if manifestDescriptor.MediaType != "application/vnd.oci.image.manifest.v1+json" {
		return imageIdentity{}, errors.New("manifest media type")
	}
	if manifestDescriptor.Platform == nil || manifestDescriptor.Platform.OS != "linux" || "linux/"+manifestDescriptor.Platform.Architecture != platform {
		return imageIdentity{}, errors.New("manifest platform")
	}
	manifestHash, manifestBlob, err := descriptorBlob(manifestDescriptor, blobs)
	if err != nil || len(manifestBlob.Data) == 0 {
		return imageIdentity{}, errors.New("manifest descriptor")
	}
	var manifest OCIManifest
	if parseBytes(manifestBlob.Data, &manifest) != nil || manifest.SchemaVersion != 2 || manifest.MediaType != "application/vnd.oci.image.manifest.v1+json" || len(manifest.Layers) == 0 || len(manifest.Layers) > maxLayers {
		return imageIdentity{}, errors.New("OCI manifest")
	}
	if manifest.Config.MediaType != "application/vnd.oci.image.config.v1+json" {
		return imageIdentity{}, errors.New("config media type")
	}
	configHash, configBlob, err := descriptorBlob(manifest.Config, blobs)
	if err != nil || len(configBlob.Data) == 0 {
		return imageIdentity{}, errors.New("config descriptor")
	}
	if err := validateOCIConfig(configBlob.Data, platform); err != nil {
		return imageIdentity{}, errors.New("config JSON")
	}
	layers := make([]string, 0, len(manifest.Layers))
	layerDescriptors := make([]Descriptor, 0, len(manifest.Layers))
	for _, descriptor := range manifest.Layers {
		if !strings.HasPrefix(descriptor.MediaType, "application/vnd.oci.image.layer.v1.tar") {
			return imageIdentity{}, errors.New("layer media type")
		}
		layerHash, _, err := descriptorBlob(descriptor, blobs)
		if err != nil {
			return imageIdentity{}, errors.New("layer descriptor")
		}
		layers = append(layers, layerHash)
		layerDescriptors = append(layerDescriptors, descriptor)
	}
	indexSum := sha256.Sum256(indexRaw)
	return imageIdentity{Index: hex.EncodeToString(indexSum[:]), Manifest: manifestHash, Config: configHash, Layers: layers, LayerDescriptors: layerDescriptors}, nil
}

func normalizedTarPath(raw string, directory bool) (string, error) {
	name := strings.TrimPrefix(raw, "./")
	if directory {
		name = strings.TrimSuffix(name, "/")
	}
	if name == "" || name != path.Clean(name) || path.IsAbs(name) || name == ".." || strings.HasPrefix(name, "../") || strings.Contains(name, `\`) {
		return "", errors.New("unsafe tar path")
	}
	if directory {
		if raw != name && raw != "./"+name && raw != name+"/" && raw != "./"+name+"/" {
			return "", errors.New("non-canonical directory tar path")
		}
	} else if raw != name && raw != "./"+name {
		return "", errors.New("non-canonical regular tar path")
	}
	return name, nil
}

func validateOCIConfig(data []byte, platform string) error {
	value, err := parseValue(bytes.NewReader(data))
	if err != nil {
		return err
	}
	object, ok := value.(map[string]any)
	if !ok {
		return errors.New("config is not an object")
	}
	osName, osOK := object["os"].(string)
	architecture, archOK := object["architecture"].(string)
	if !osOK || !archOK || osName != "linux" || "linux/"+architecture != platform {
		return errors.New("config platform mismatch")
	}
	return nil
}

func descriptorBlob(descriptor Descriptor, blobs map[string]archiveBlob) (string, archiveBlob, error) {
	if !strings.HasPrefix(descriptor.Digest, "sha256:") {
		return "", archiveBlob{}, errors.New("digest algorithm")
	}
	digest := strings.TrimPrefix(descriptor.Digest, "sha256:")
	if !hexDigestRE.MatchString(digest) {
		return "", archiveBlob{}, errors.New("digest shape")
	}
	size, err := strconv.ParseInt(descriptor.Size.String(), 10, 64)
	if err != nil || size < 0 || size > maxArchiveBytes {
		return "", archiveBlob{}, errors.New("descriptor size")
	}
	blob, ok := blobs[digest]
	if !ok || blob.Size != size {
		return "", archiveBlob{}, errors.New("blob size or presence")
	}
	return digest, blob, nil
}

func sameIdentity(first, second imageIdentity) bool {
	return first.Index == second.Index && first.Manifest == second.Manifest && first.Config == second.Config && strings.Join(first.Layers, "\x00") == strings.Join(second.Layers, "\x00")
}

func formatViolations(report Report) string {
	if report.Valid {
		return fmt.Sprintf("OCI evidence valid for %s (%d targets)", report.Platform, report.Targets)
	}
	return strings.Join(report.Violations, "; ")
}
