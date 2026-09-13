package securityreview

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func ParseScope(data []byte) (Scope, error) {
	var scope Scope
	if err := decodeStrict(data, &scope); err != nil {
		return Scope{}, fmt.Errorf("parse review scope: %w", err)
	}
	return scope, nil
}

func ParseFindings(data []byte) (Findings, error) {
	var findings Findings
	if err := decodeStrict(data, &findings); err != nil {
		return Findings{}, fmt.Errorf("parse review findings: %w", err)
	}
	return findings, nil
}

func ReadScope(path string) (Scope, error) {
	raw, err := readBoundedFile(path, MaxDocumentBytes)
	if err != nil {
		return Scope{}, fmt.Errorf("read review scope: %w", err)
	}
	return ParseScope(raw)
}

func ReadFindings(path string) (Findings, error) {
	raw, err := readBoundedFile(path, MaxDocumentBytes)
	if err != nil {
		return Findings{}, fmt.Errorf("read review findings: %w", err)
	}
	return ParseFindings(raw)
}

func decodeStrict(data []byte, target any) error {
	if len(data) == 0 {
		return fmt.Errorf("document is empty")
	}
	if len(data) > MaxDocumentBytes {
		return fmt.Errorf("document exceeds %d bytes", MaxDocumentBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return fmt.Errorf("trailing JSON: %w", err)
	}
	return nil
}

func readBoundedFile(path string, maximum int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("symlink is forbidden")
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("not a regular file")
	}
	if info.Size() > maximum {
		return nil, fmt.Errorf("file exceeds %d bytes", maximum)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	openedInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !openedInfo.Mode().IsRegular() || !os.SameFile(info, openedInfo) {
		return nil, fmt.Errorf("file changed during open or is not regular")
	}
	if openedInfo.Size() > maximum {
		return nil, fmt.Errorf("file exceeds %d bytes", maximum)
	}
	raw, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > maximum {
		return nil, fmt.Errorf("file exceeds %d bytes", maximum)
	}
	return raw, nil
}
