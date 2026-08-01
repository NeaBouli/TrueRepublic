package incidentpolicy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

const maxJSONDepth = 32

func Load(path string) (contract Contract, err error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Contract{}, fmt.Errorf("open incident rehearsal: file does not exist")
		}
		return Contract{}, fmt.Errorf("open incident rehearsal: file is unavailable")
	}
	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			contract = Contract{}
			err = fmt.Errorf("close incident rehearsal: file is unavailable")
		}
	}()
	return Parse(file)
}

// Parse reads exactly one strict JSON contract. Duplicate keys, unknown
// fields, trailing values, excessive depth, and oversized inputs fail closed.
func Parse(reader io.Reader) (Contract, error) {
	var contract Contract
	data, err := io.ReadAll(io.LimitReader(reader, MaxContractBytes+1))
	if err != nil {
		return contract, fmt.Errorf("read incident rehearsal: input is unavailable")
	}
	if len(data) > MaxContractBytes {
		return contract, fmt.Errorf("incident rehearsal exceeds %d bytes", MaxContractBytes)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return contract, fmt.Errorf("incident rehearsal is empty")
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	value, err := decodeStrictValue(decoder, 0)
	if err != nil {
		return contract, fmt.Errorf("invalid incident rehearsal JSON: %s", safeParseCategory(err))
	}
	if token, err := decoder.Token(); err != io.EOF {
		if err == nil {
			_ = token
			return contract, fmt.Errorf("invalid incident rehearsal JSON: trailing value is forbidden")
		}
		return contract, fmt.Errorf("invalid incident rehearsal JSON: trailing data is forbidden")
	}

	canonical, err := json.Marshal(value)
	if err != nil {
		return contract, fmt.Errorf("normalize incident rehearsal JSON: %w", err)
	}
	strict := json.NewDecoder(bytes.NewReader(canonical))
	strict.DisallowUnknownFields()
	if err := strict.Decode(&contract); err != nil {
		return Contract{}, fmt.Errorf("invalid incident rehearsal schema")
	}
	return contract, nil
}

func decodeStrictValue(decoder *json.Decoder, depth int) (any, error) {
	if depth > maxJSONDepth {
		return nil, fmt.Errorf("maximum JSON depth %d exceeded", maxJSONDepth)
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
		object := make(map[string]any)
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return nil, err
			}
			key, ok := keyToken.(string)
			if !ok {
				return nil, fmt.Errorf("object key is not a string")
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
		return nil, fmt.Errorf("unexpected delimiter %q", delim)
	}
}

func safeParseCategory(err error) string {
	message := err.Error()
	if strings.Contains(message, "maximum JSON depth") {
		return fmt.Sprintf("maximum JSON depth %d exceeded", maxJSONDepth)
	}
	if strings.Contains(message, "duplicate object key") {
		return "duplicate object key"
	}
	return "malformed input"
}
