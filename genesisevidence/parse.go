package genesisevidence

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var errDuplicateKey = errors.New("duplicate object key")

func strictJSON(data []byte, limit int) ([]byte, map[string]any, error) {
	if len(data) == 0 || len(bytes.TrimSpace(data)) == 0 {
		return nil, nil, errors.New("empty input")
	}
	if len(data) > limit {
		return nil, nil, errors.New("input exceeds size limit")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	value, err := decodeValue(decoder, 0)
	if err != nil {
		return nil, nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return nil, nil, errors.New("trailing value")
		}
		return nil, nil, errors.New("trailing data")
	}
	root, ok := value.(map[string]any)
	if !ok {
		return nil, nil, errors.New("root must be an object")
	}
	normalized, err := json.Marshal(value)
	return normalized, root, err
}

func decodeValue(decoder *json.Decoder, depth int) (any, error) {
	if depth > 64 {
		return nil, errors.New("maximum JSON depth exceeded")
	}
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	delim, isDelim := token.(json.Delim)
	if !isDelim {
		return token, nil
	}
	switch delim {
	case '{':
		result := map[string]any{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			key, ok := keyToken.(string)
			if !ok {
				return nil, errors.New("invalid object key")
			}
			if _, exists := result[key]; exists {
				return nil, errDuplicateKey
			}
			value, err := decodeValue(decoder, depth+1)
			if err != nil {
				return nil, err
			}
			result[key] = value
		}
		if _, err := decoder.Token(); err != nil {
			return nil, err
		}
		return result, nil
	case '[':
		result := []any{}
		for decoder.More() {
			value, err := decodeValue(decoder, depth+1)
			if err != nil {
				return nil, err
			}
			result = append(result, value)
		}
		if _, err := decoder.Token(); err != nil {
			return nil, err
		}
		return result, nil
	default:
		return nil, errors.New("invalid delimiter")
	}
}

func parseManifest(data []byte) (Manifest, error) {
	normalized, root, err := strictJSON(data, MaxManifestBytes)
	if err != nil {
		return Manifest{}, err
	}
	required := []string{"schema", "source_commit", "daemon_version", "chain_id", "genesis_sha256", "max_validator_power", "validators", "allocations", "total_supply_upnyx", "governance_escrow_upnyx", "dex_custody"}
	if len(root) != len(required) {
		return Manifest{}, errors.New("manifest fields are not exact")
	}
	for _, key := range required {
		if _, ok := root[key]; !ok {
			return Manifest{}, errors.New("manifest field is missing")
		}
	}
	var manifest Manifest
	decoder := json.NewDecoder(bytes.NewReader(normalized))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("manifest schema: %w", err)
	}
	if manifest.Validators == nil || manifest.Allocations == nil || manifest.DEXCustody == nil {
		return Manifest{}, errors.New("manifest arrays are required")
	}
	return manifest, nil
}

func decodeInto(value any, target any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	return decoder.Decode(target)
}
