package installlifecycle

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
	b, err := io.ReadAll(io.LimitReader(f, MaxJSONBytes+1))
	if err != nil || len(b) == 0 || len(b) > MaxJSONBytes {
		return errors.New("JSON input is unavailable or exceeds byte limit")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	v, err := strictValue(d, 0)
	if err != nil {
		return errors.New("JSON input is malformed")
	}
	if _, err = d.Token(); err != io.EOF {
		return errors.New("JSON input has trailing data")
	}
	canonical, err := json.Marshal(v)
	if err != nil {
		return errors.New("JSON input cannot be normalized")
	}
	d = json.NewDecoder(bytes.NewReader(canonical))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return errors.New("JSON schema is invalid")
	}
	return nil
}

func strictValue(d *json.Decoder, depth int) (any, error) {
	if depth > 32 {
		return nil, errors.New("JSON nesting is too deep")
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
			keyToken, err := d.Token()
			if err != nil {
				return nil, err
			}
			key, ok := keyToken.(string)
			if !ok {
				return nil, errors.New("object key is invalid")
			}
			if _, duplicate := m[key]; duplicate {
				return nil, errors.New("duplicate object key")
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
		var a []any
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
		return nil, fmt.Errorf("unexpected delimiter")
	}
}

func marshal(v any) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(bytes.TrimSpace(b), '\n'), nil
}
