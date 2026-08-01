package capacitypolicy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

const maxJSONDepth = 32

func LoadContract(path string) (Contract, error) {
	var contract Contract
	err := loadStrict(path, &contract, "capacity contract")
	return contract, err
}

func LoadEvidence(path string) (Evidence, error) {
	var evidence Evidence
	err := loadStrict(path, &evidence, "capacity evidence")
	return evidence, err
}

func ParseContract(reader io.Reader) (Contract, error) {
	var contract Contract
	err := parseStrict(reader, &contract, "capacity contract")
	return contract, err
}

func ParseEvidence(reader io.Reader) (Evidence, error) {
	var evidence Evidence
	err := parseStrict(reader, &evidence, "capacity evidence")
	return evidence, err
}

func loadStrict(path string, target any, label string) (err error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("open %s: file does not exist", label)
		}
		return fmt.Errorf("open %s: file is unavailable", label)
	}
	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close %s: file is unavailable", label)
		}
	}()
	return parseStrict(file, target, label)
}

func parseStrict(reader io.Reader, target any, label string) error {
	data, err := io.ReadAll(io.LimitReader(reader, MaxDocumentBytes+1))
	if err != nil {
		return fmt.Errorf("read %s: input is unavailable", label)
	}
	if len(data) > MaxDocumentBytes {
		return fmt.Errorf("%s exceeds %d bytes", label, MaxDocumentBytes)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return fmt.Errorf("%s is empty", label)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	value, err := decodeStrictValue(decoder, 0)
	if err != nil {
		return fmt.Errorf("invalid %s JSON: %s", label, safeParseCategory(err))
	}
	if token, err := decoder.Token(); err != io.EOF {
		if err == nil {
			_ = token
			return fmt.Errorf("invalid %s JSON: trailing value is forbidden", label)
		}
		return fmt.Errorf("invalid %s JSON: trailing data is forbidden", label)
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("normalize %s: unavailable", label)
	}
	strict := json.NewDecoder(bytes.NewReader(canonical))
	strict.DisallowUnknownFields()
	if err := strict.Decode(target); err != nil {
		return fmt.Errorf("invalid %s schema", label)
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
