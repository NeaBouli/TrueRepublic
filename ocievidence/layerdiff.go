package ocievidence

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	layerMediaTar     = "application/vnd.oci.image.layer.v1.tar"
	layerMediaTarGzip = "application/vnd.oci.image.layer.v1.tar+gzip"
)

// Variables let tests exercise the fail-closed bounds with small fixtures.
// Verification is single-threaded, and production code never mutates them.
var (
	maxLayerStreamBytes   int64 = 1 << 30
	maxLayerEntries             = 1 << 16
	maxLayerDiffs               = 256
	maxLayerPathBytes           = 4096
	maxLayerLinkBytes           = 4096
	maxLayerMetadataBytes       = 1 << 16
)

type layerEntry struct {
	Type           string
	Mode           string
	UID            int
	GID            int
	Size           int64
	LinkTarget     string
	ContentSHA256  string
	MetadataSHA256 string
}

func diagnoseRepeatedLayers(target, firstArchive, secondArchive string, first, second imageIdentity) ([]LayerDiff, []string) {
	count := len(first.Layers)
	if len(second.Layers) < count {
		count = len(second.Layers)
	}
	reports := make([]LayerDiff, 0)
	problems := make([]string, 0)
	for index := 0; index < count; index++ {
		if first.Layers[index] == second.Layers[index] {
			continue
		}
		if index >= len(first.LayerDescriptors) || index >= len(second.LayerDescriptors) {
			problems = append(problems, fmt.Sprintf("OCI layer diagnostic failed for %s layer %d: descriptor unavailable", target, index))
			continue
		}
		diff, err := compareLayerEntries(firstArchive, secondArchive, first.LayerDescriptors[index], second.LayerDescriptors[index])
		if err != nil {
			problems = append(problems, fmt.Sprintf("OCI layer diagnostic failed for %s layer %d: %s", target, index, err))
			continue
		}
		diff.Target = target
		diff.LayerIndex = index
		reports = append(reports, diff)
	}
	if len(first.Layers) != len(second.Layers) {
		problems = append(problems, "OCI layer count differs for "+target)
	}
	return reports, problems
}

func compareLayerEntries(firstArchive, secondArchive string, firstDescriptor, secondDescriptor Descriptor) (LayerDiff, error) {
	first, err := extractLayerEntries(firstArchive, firstDescriptor)
	if err != nil {
		return LayerDiff{}, fmt.Errorf("first layer: %w", err)
	}
	second, err := extractLayerEntries(secondArchive, secondDescriptor)
	if err != nil {
		return LayerDiff{}, fmt.Errorf("second layer: %w", err)
	}
	paths := make([]string, 0)
	for name, entry := range first {
		other, exists := second[name]
		if !exists || !sameLayerEntry(entry, other) {
			paths = append(paths, name)
		}
	}
	for name := range second {
		if _, exists := first[name]; !exists {
			paths = append(paths, name)
		}
	}
	sort.Strings(paths)
	truncated := len(paths) > maxLayerDiffs
	if truncated {
		paths = paths[:maxLayerDiffs]
	}
	entries := make([]LayerEntryDiff, 0, len(paths))
	for _, name := range paths {
		firstEntry, inFirst := first[name]
		secondEntry, inSecond := second[name]
		diff := LayerEntryDiff{Path: name}
		switch {
		case inFirst && inSecond:
			diff.Change = "modified"
			diff.First = layerEntryInfo(firstEntry)
			diff.Second = layerEntryInfo(secondEntry)
		case inFirst:
			diff.Change = "removed"
			diff.First = layerEntryInfo(firstEntry)
		default:
			diff.Change = "added"
			diff.Second = layerEntryInfo(secondEntry)
		}
		entries = append(entries, diff)
	}
	return LayerDiff{Entries: entries, Truncated: truncated}, nil
}

func sameLayerEntry(first, second layerEntry) bool {
	return first.MetadataSHA256 == second.MetadataSHA256 && first.ContentSHA256 == second.ContentSHA256
}

func layerEntryInfo(entry layerEntry) *LayerEntryInfo {
	return &LayerEntryInfo{
		Type:           entry.Type,
		Mode:           entry.Mode,
		UID:            entry.UID,
		GID:            entry.GID,
		Size:           entry.Size,
		LinkTarget:     entry.LinkTarget,
		ContentSHA256:  entry.ContentSHA256,
		MetadataSHA256: entry.MetadataSHA256,
	}
}

func extractLayerEntries(archivePath string, descriptor Descriptor) (map[string]layerEntry, error) {
	if descriptor.MediaType != layerMediaTar && descriptor.MediaType != layerMediaTarGzip {
		return nil, errors.New("unsupported layer media type")
	}
	digest, size, err := layerDescriptorIdentity(descriptor)
	if err != nil {
		return nil, err
	}
	if size > maxLayerStreamBytes {
		return nil, errors.New("compressed layer byte bound")
	}
	info, err := os.Lstat(archivePath)
	if err != nil || !info.Mode().IsRegular() {
		return nil, errors.New("archive is unavailable")
	}
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, errors.New("archive is unavailable")
	}
	defer func() { _ = file.Close() }()

	reader := tar.NewReader(io.LimitReader(file, maxArchiveBytes+1))
	wanted := "blobs/sha256/" + digest
	for scanned := 1; ; scanned++ {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			return nil, errors.New("layer blob is missing")
		}
		if err != nil {
			return nil, errors.New("archive tar stream")
		}
		if scanned > maxArchiveEntries {
			return nil, errors.New("archive entry bound")
		}
		name, err := normalizedTarPath(header.Name, header.Typeflag == tar.TypeDir)
		if err != nil {
			return nil, errors.New("unsafe archive path")
		}
		if name != wanted {
			continue
		}
		if header.Typeflag != tar.TypeReg || header.Size != size {
			return nil, errors.New("layer blob descriptor mismatch")
		}
		return readLayerBlob(reader, size, digest, descriptor.MediaType == layerMediaTarGzip)
	}
}

func layerDescriptorIdentity(descriptor Descriptor) (string, int64, error) {
	if !strings.HasPrefix(descriptor.Digest, "sha256:") {
		return "", 0, errors.New("layer digest algorithm")
	}
	digest := strings.TrimPrefix(descriptor.Digest, "sha256:")
	if !hexDigestRE.MatchString(digest) {
		return "", 0, errors.New("layer digest shape")
	}
	size, err := strconv.ParseInt(descriptor.Size.String(), 10, 64)
	if err != nil || size < 0 || size > maxArchiveBytes {
		return "", 0, errors.New("layer descriptor size")
	}
	return digest, size, nil
}

func readLayerBlob(reader io.Reader, size int64, digest string, compressed bool) (map[string]layerEntry, error) {
	limited := &io.LimitedReader{R: reader, N: size}
	hasher := sha256.New()
	compressedStream := io.TeeReader(limited, hasher)
	var layerStream io.Reader = compressedStream
	var gzipReader *gzip.Reader
	if compressed {
		var err error
		gzipReader, err = gzip.NewReader(compressedStream)
		if err != nil {
			return nil, errors.New("invalid gzip layer")
		}
		layerStream = gzipReader
	}
	bounded := &layerBoundedReader{reader: layerStream, limit: maxLayerStreamBytes}
	entries, err := readLayerTar(bounded)
	if err != nil {
		return nil, err
	}
	if gzipReader != nil {
		if _, err = io.Copy(io.Discard, bounded); err != nil {
			return nil, errors.New("gzip layer drain")
		}
		if err = gzipReader.Close(); err != nil {
			return nil, errors.New("gzip layer close")
		}
	}
	if _, err = io.Copy(io.Discard, compressedStream); err != nil {
		return nil, errors.New("layer blob drain")
	}
	if limited.N != 0 {
		return nil, errors.New("layer blob size mismatch")
	}
	if hex.EncodeToString(hasher.Sum(nil)) != digest {
		return nil, errors.New("layer blob digest mismatch")
	}
	return entries, nil
}

type layerBoundedReader struct {
	reader io.Reader
	limit  int64
	total  int64
}

func (reader *layerBoundedReader) Read(buffer []byte) (int, error) {
	remaining := reader.limit + 1 - reader.total
	if remaining <= 0 {
		return 0, errors.New("layer stream byte bound")
	}
	if int64(len(buffer)) > remaining {
		buffer = buffer[:remaining]
	}
	count, err := reader.reader.Read(buffer)
	reader.total += int64(count)
	if reader.total > reader.limit {
		return count, errors.New("layer stream byte bound")
	}
	return count, err
}

func readLayerTar(stream io.Reader) (map[string]layerEntry, error) {
	reader := tar.NewReader(stream)
	entries := make(map[string]layerEntry)
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			return entries, nil
		}
		if err != nil {
			return nil, errors.New("layer tar stream")
		}
		if len(entries) >= maxLayerEntries {
			return nil, errors.New("layer entry bound")
		}
		if header.Size < 0 {
			return nil, errors.New("layer entry size")
		}
		name, err := normalizedTarPath(header.Name, header.Typeflag == tar.TypeDir)
		if err != nil {
			return nil, errors.New("unsafe layer path")
		}
		if len(name) > maxLayerPathBytes {
			return nil, errors.New("layer path bound")
		}
		if _, exists := entries[name]; exists {
			return nil, errors.New("duplicate layer path")
		}
		entry, err := describeLayerEntry(reader, header)
		if err != nil {
			return nil, err
		}
		entries[name] = entry
	}
}

func describeLayerEntry(reader io.Reader, header *tar.Header) (layerEntry, error) {
	if !boundedLayerMetadata(header) {
		return layerEntry{}, errors.New("layer metadata bound")
	}
	entry := layerEntry{
		Mode:       fmt.Sprintf("%04o", header.Mode&0o7777),
		UID:        header.Uid,
		GID:        header.Gid,
		Size:       header.Size,
		LinkTarget: header.Linkname,
	}
	switch header.Typeflag {
	case tar.TypeReg, tar.TypeRegA:
		entry.Type = "file"
		hasher := sha256.New()
		written, err := io.Copy(hasher, io.LimitReader(reader, header.Size+1))
		if err != nil {
			return layerEntry{}, fmt.Errorf("layer entry read: %w", err)
		}
		if written != header.Size {
			return layerEntry{}, errors.New("layer entry read")
		}
		entry.ContentSHA256 = hex.EncodeToString(hasher.Sum(nil))
	case tar.TypeDir:
		entry.Type = "directory"
	case tar.TypeSymlink:
		entry.Type = "symlink"
	case tar.TypeLink:
		entry.Type = "hardlink"
	case tar.TypeChar:
		entry.Type = "char-device"
	case tar.TypeBlock:
		entry.Type = "block-device"
	case tar.TypeFifo:
		entry.Type = "fifo"
	default:
		return layerEntry{}, errors.New("unsupported layer entry type")
	}
	if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeRegA && header.Size != 0 {
		return layerEntry{}, errors.New("non-regular layer entry has content")
	}
	if len(header.Linkname) > maxLayerLinkBytes {
		return layerEntry{}, errors.New("layer link target bound")
	}
	entry.MetadataSHA256 = layerMetadataHash(header, entry)
	return entry, nil
}

func boundedLayerMetadata(header *tar.Header) bool {
	total := len(header.Uname) + len(header.Gname) + len(header.Linkname)
	for key, value := range header.PAXRecords {
		total += len(key) + len(value)
		if total > maxLayerMetadataBytes {
			return false
		}
	}
	for key, value := range header.Xattrs {
		total += len(key) + len(value)
		if total > maxLayerMetadataBytes {
			return false
		}
	}
	return total <= maxLayerMetadataBytes
}

func layerMetadataHash(header *tar.Header, entry layerEntry) string {
	hasher := sha256.New()
	_, _ = fmt.Fprintf(hasher, "%s\x00%s\x00%d\x00%d\x00%d\x00%s\x00%d\x00%d\x00%d\x00%d\x00%d\x00%s\x00%s",
		entry.Type, entry.Mode, entry.UID, entry.GID, entry.Size, entry.LinkTarget,
		header.ModTime.UnixNano(), header.AccessTime.UnixNano(), header.ChangeTime.UnixNano(),
		header.Devmajor, header.Devminor, header.Uname, header.Gname)
	writeSortedMap(hasher, header.PAXRecords)
	writeSortedMap(hasher, header.Xattrs)
	return hex.EncodeToString(hasher.Sum(nil))
}

func writeSortedMap(writer io.Writer, values map[string]string) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		_, _ = fmt.Fprintf(writer, "\x00%s\x00%s", key, values[key])
	}
}
