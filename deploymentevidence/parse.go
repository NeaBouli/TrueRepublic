package deploymentevidence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// LoadManifest reads one strict manifest from an explicit local path. All
// errors are generic and never reflect paths, field names, or input values.
func LoadManifest(path string) (Manifest, error) {
	var manifest Manifest
	err := loadStrict(path, &manifest)
	return manifest, err
}

// ParseManifest reads one strict manifest from a reader. All errors are
// generic and never reflect field names or input values.
func ParseManifest(reader io.Reader) (Manifest, error) {
	var manifest Manifest
	err := parseStrict(reader, &manifest)
	return manifest, err
}

func loadStrict(path string, target any) (err error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("open deployment manifest: file does not exist")
		}
		return fmt.Errorf("open deployment manifest: file is unavailable")
	}
	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close deployment manifest: file is unavailable")
		}
	}()
	return parseStrict(file, target)
}

func parseStrict(reader io.Reader, target any) error {
	data, err := io.ReadAll(io.LimitReader(reader, MaxManifestBytes+1))
	if err != nil {
		return fmt.Errorf("read deployment manifest: input is unavailable")
	}
	if len(data) > MaxManifestBytes {
		return fmt.Errorf("deployment manifest exceeds %d bytes", MaxManifestBytes)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return fmt.Errorf("deployment manifest is empty")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	value, err := decodeStrictValue(decoder, 0)
	if err != nil {
		return fmt.Errorf("invalid deployment manifest JSON: %s", safeParseCategory(err))
	}
	if token, err := decoder.Token(); err != io.EOF {
		if err == nil {
			_ = token
			return fmt.Errorf("invalid deployment manifest JSON: trailing value is forbidden")
		}
		return fmt.Errorf("invalid deployment manifest JSON: trailing data is forbidden")
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("normalize deployment manifest: unavailable")
	}
	strict := json.NewDecoder(bytes.NewReader(canonical))
	strict.DisallowUnknownFields()
	if err := strict.Decode(target); err != nil {
		return fmt.Errorf("invalid deployment manifest schema")
	}
	return nil
}

func decodeStrictValue(decoder *json.Decoder, depth int) (any, error) {
	if depth > maxJSONDepth {
		return nil, fmt.Errorf("maximum JSON depth exceeded")
	}
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	delim, ok := token.(json.Delim)
	if !ok {
		return token, nil
	}
	switch delim {
	case '{':
		object := make(map[string]any)
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			key, ok := keyToken.(string)
			if !ok {
				return nil, fmt.Errorf("object key is invalid")
			}
			if _, exists := object[key]; exists {
				return nil, fmt.Errorf("duplicate object key")
			}
			value, err := decodeStrictValue(decoder, depth+1)
			if err != nil {
				return nil, err
			}
			object[key] = value
		}
		if _, err := decoder.Token(); err != nil {
			return nil, err
		}
		return object, nil
	case '[':
		array := make([]any, 0)
		for decoder.More() {
			value, err := decodeStrictValue(decoder, depth+1)
			if err != nil {
				return nil, err
			}
			array = append(array, value)
		}
		if _, err := decoder.Token(); err != nil {
			return nil, err
		}
		return array, nil
	default:
		return nil, fmt.Errorf("unexpected delimiter")
	}
}

// safeParseCategory collapses parser internals into fixed generic categories
// so rejected keys, delimiters, or values are never reflected.
func safeParseCategory(err error) string {
	message := err.Error()
	if strings.Contains(message, "maximum JSON depth") {
		return "maximum JSON depth exceeded"
	}
	if strings.Contains(message, "duplicate object key") {
		return "duplicate object key"
	}
	return "malformed input"
}
