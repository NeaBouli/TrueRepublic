package topologypolicy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const maxJSONDepth = 32

func Load(path string) (Contract, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Contract{}, fmt.Errorf("open topology contract: file does not exist")
		}
		return Contract{}, fmt.Errorf("open topology contract: file is unavailable")
	}
	defer file.Close()
	return Parse(file)
}

// Parse reads one strict JSON contract. It rejects duplicate keys before
// decoding into typed fields because encoding/json otherwise accepts the last
// duplicate value.
func Parse(reader io.Reader) (Contract, error) {
	var contract Contract
	data, err := io.ReadAll(io.LimitReader(reader, MaxContractBytes+1))
	if err != nil {
		return contract, fmt.Errorf("read topology contract: %w", err)
	}
	if len(data) > MaxContractBytes {
		return contract, fmt.Errorf("topology contract exceeds %d bytes", MaxContractBytes)
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return contract, fmt.Errorf("topology contract is empty")
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	value, err := decodeStrictValue(decoder, 0)
	if err != nil {
		return contract, fmt.Errorf("invalid topology JSON: %w", err)
	}
	if token, err := decoder.Token(); err != io.EOF {
		if err == nil {
			_ = token
			return contract, fmt.Errorf("invalid topology JSON: trailing value is forbidden")
		}
		return contract, fmt.Errorf("invalid topology JSON: trailing data: %w", err)
	}

	canonical, err := json.Marshal(value)
	if err != nil {
		return contract, fmt.Errorf("normalize topology JSON: %w", err)
	}
	strict := json.NewDecoder(bytes.NewReader(canonical))
	strict.DisallowUnknownFields()
	if err := strict.Decode(&contract); err != nil {
		return Contract{}, fmt.Errorf("invalid topology schema: %w", err)
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
				return nil, fmt.Errorf("duplicate object key %q", key)
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
