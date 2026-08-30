package candidateevidence

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

func parseFile(path string, target any) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.New("input file is unavailable")
	}
	defer func() { _ = f.Close() }()
	return parseReader(f, target)
}

func parseBytes(data []byte, target any) error { return parseReader(bytes.NewReader(data), target) }

func parseReader(r io.Reader, target any) error {
	v, err := parseValue(r)
	if err != nil {
		return err
	}
	canonical, err := json.Marshal(v)
	if err != nil {
		return errors.New("JSON input cannot be normalized")
	}
	d := json.NewDecoder(bytes.NewReader(canonical))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return errors.New("JSON schema is invalid")
	}
	return nil
}

func parseValue(r io.Reader) (any, error) {
	data, err := io.ReadAll(io.LimitReader(r, MaxJSONBytes+1))
	if err != nil || len(data) == 0 {
		return nil, errors.New("JSON input is unavailable")
	}
	if len(data) > MaxJSONBytes {
		return nil, errors.New("JSON input exceeds byte limit")
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	v, err := strictValue(d, 0)
	if err != nil {
		return nil, errors.New("JSON input is malformed")
	}
	if _, err = d.Token(); err != io.EOF {
		return nil, errors.New("JSON input has trailing data")
	}
	return v, nil
}

func strictValue(d *json.Decoder, depth int) (any, error) {
	if depth > maxJSONDepth {
		return nil, errors.New("depth")
	}
	tok, err := d.Token()
	if err != nil {
		return nil, err
	}
	delim, ok := tok.(json.Delim)
	if !ok {
		return tok, nil
	}
	switch delim {
	case '{':
		m := map[string]any{}
		for d.More() {
			k, err := d.Token()
			if err != nil {
				return nil, err
			}
			key, ok := k.(string)
			if !ok {
				return nil, errors.New("key")
			}
			if _, exists := m[key]; exists {
				return nil, errors.New("duplicate")
			}
			value, err := strictValue(d, depth+1)
			if err != nil {
				return nil, err
			}
			m[key] = value
		}
		_, err = d.Token()
		return m, err
	case '[':
		a := []any{}
		for d.More() {
			value, err := strictValue(d, depth+1)
			if err != nil {
				return nil, err
			}
			a = append(a, value)
		}
		_, err = d.Token()
		return a, err
	default:
		return nil, fmt.Errorf("delimiter")
	}
}
