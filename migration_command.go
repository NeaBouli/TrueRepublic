package main

import (
	"bytes"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"

	"truerepublic/migration"
)

const (
	maxMigrationDescriptorBytes int64 = 4 << 20
	maxMigrationGenesisBytes    int64 = 1 << 30
)

type legacyAuthorityMigrationOptions struct {
	descriptorPath string
	genesisPath    string
	outputPath     string
	sourceAppHash  string
}

func newMigrationCmd(cdc codec.Codec) *cobra.Command {
	command := &cobra.Command{
		Use:   "migration",
		Short: "Offline consensus-state migration tools",
		Args:  cobra.NoArgs,
	}
	command.AddCommand(newLegacyAuthorityMigrationCmd(cdc))
	return command
}

func newLegacyAuthorityMigrationCmd(cdc codec.Codec) *cobra.Command {
	options := legacyAuthorityMigrationOptions{}
	command := &cobra.Command{
		Use:          "legacy-authority",
		Short:        "Rewrite legacy consensus-key-coupled operator authorities",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := validateLegacyAuthorityMigrationOptions(options); err != nil {
				return err
			}
			transformID, err := runLegacyAuthorityMigration(cdc, options)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(
				command.OutOrStdout(),
				"wrote transformed genesis to %q with transform %q\n",
				options.outputPath,
				transformID,
			)
			return err
		},
	}
	command.Flags().StringVar(&options.descriptorPath, "descriptor", "", "path to the signed migration descriptor JSON")
	command.Flags().StringVar(&options.genesisPath, "genesis", "", "path to the exported full genesis JSON")
	command.Flags().StringVar(&options.outputPath, "output", "", "new path for the transformed genesis JSON")
	command.Flags().StringVar(
		&options.sourceAppHash,
		"source-app-hash",
		"",
		"trusted 32-byte source app hash at the signed halt height (hex)",
	)
	return command
}

func validateLegacyAuthorityMigrationOptions(options legacyAuthorityMigrationOptions) error {
	missing := make([]string, 0, 4)
	for flag, value := range map[string]string{
		"--descriptor":      options.descriptorPath,
		"--genesis":         options.genesisPath,
		"--output":          options.outputPath,
		"--source-app-hash": options.sourceAppHash,
	} {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, flag)
		}
	}
	if len(missing) != 0 {
		return fmt.Errorf("migration: required flag(s) missing: %s", strings.Join(missing, ", "))
	}
	return nil
}

func runLegacyAuthorityMigration(cdc codec.Codec, options legacyAuthorityMigrationOptions) (string, error) {
	if cdc == nil {
		return "", errors.New("migration: nil application codec")
	}
	if err := rejectMigrationPathCollisions(options); err != nil {
		return "", err
	}

	descriptorJSON, descriptorInfo, err := readRegularFileBounded(
		options.descriptorPath,
		maxMigrationDescriptorBytes,
	)
	if err != nil {
		return "", fmt.Errorf("migration: read descriptor: %w", err)
	}
	descriptor, err := decodeMigrationDescriptorStrict(descriptorJSON)
	if err != nil {
		return "", err
	}
	trustedHash, err := parseTrustedAppHash(options.sourceAppHash)
	if err != nil {
		return "", err
	}
	if len(descriptor.SourceAppHash) != len(trustedHash) ||
		subtle.ConstantTimeCompare(descriptor.SourceAppHash, trustedHash) != 1 {
		return "", errors.New("migration: trusted source app hash does not match the descriptor")
	}
	genesisJSON, genesisInfo, err := readRegularFileBounded(
		options.genesisPath,
		maxMigrationGenesisBytes,
	)
	if err != nil {
		return "", fmt.Errorf("migration: read exported genesis: %w", err)
	}
	if os.SameFile(descriptorInfo, genesisInfo) {
		return "", errors.New("migration: descriptor and exported genesis must be different files")
	}

	transformed, err := migration.TransformApplicationGenesis(descriptor, genesisJSON, cdc)
	if err != nil {
		return "", err
	}
	if err := validateMigratedLedger(cdc, transformed); err != nil {
		return "", err
	}
	if err := writeNewFileAtomic(options.outputPath, transformed, 0o600); err != nil {
		return "", fmt.Errorf("migration: write transformed genesis: %w", err)
	}
	return descriptor.TransformID, nil
}

func validateMigratedLedger(cdc codec.Codec, raw json.RawMessage) error {
	var root struct {
		AppState map[string]json.RawMessage `json:"app_state"`
	}
	if err := json.Unmarshal(raw, &root); err != nil {
		return fmt.Errorf("migration: decode transformed genesis for ledger validation: %w", err)
	}
	if root.AppState == nil {
		return errors.New("migration: transformed genesis has no application state")
	}
	if err := validateLedgerGenesis(cdc, root.AppState); err != nil {
		return fmt.Errorf("migration: transformed ledger validation failed: %w", err)
	}
	return nil
}

func rejectMigrationPathCollisions(options legacyAuthorityMigrationOptions) error {
	descriptorPath, err := filepath.Abs(options.descriptorPath)
	if err != nil {
		return fmt.Errorf("migration: resolve descriptor path: %w", err)
	}
	genesisPath, err := filepath.Abs(options.genesisPath)
	if err != nil {
		return fmt.Errorf("migration: resolve genesis path: %w", err)
	}
	outputPath, err := filepath.Abs(options.outputPath)
	if err != nil {
		return fmt.Errorf("migration: resolve output path: %w", err)
	}
	if filepath.Clean(descriptorPath) == filepath.Clean(outputPath) ||
		filepath.Clean(genesisPath) == filepath.Clean(outputPath) {
		return errors.New("migration: output path must differ from every input path")
	}
	if _, err := os.Lstat(options.outputPath); err == nil {
		return errors.New("migration: output path already exists")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("migration: inspect output path: %w", err)
	}
	return nil
}

func decodeMigrationDescriptorStrict(raw []byte) (*migration.Descriptor, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var descriptor migration.Descriptor
	if err := decoder.Decode(&descriptor); err != nil {
		return nil, fmt.Errorf("migration: decode descriptor: %w", err)
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, errors.New("migration: descriptor contains trailing JSON")
		}
		return nil, fmt.Errorf("migration: descriptor contains invalid trailing data: %w", err)
	}
	return &descriptor, nil
}

func parseTrustedAppHash(value string) ([]byte, error) {
	normalized := strings.TrimSpace(value)
	if strings.HasPrefix(normalized, "0x") || strings.HasPrefix(normalized, "0X") {
		normalized = normalized[2:]
	}
	hash, err := hex.DecodeString(normalized)
	if err != nil {
		return nil, errors.New("migration: trusted source app hash is not valid hex")
	}
	if len(hash) != 32 {
		return nil, fmt.Errorf("migration: trusted source app hash must be exactly 32 bytes, got %d", len(hash))
	}
	return hash, nil
}

func readRegularFileBounded(path string, maximum int64) (_ []byte, info os.FileInfo, err error) {
	pathInfo, err := os.Lstat(path)
	if err != nil {
		return nil, nil, err
	}
	if !pathInfo.Mode().IsRegular() {
		return nil, nil, fmt.Errorf("%q is not a regular file", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		closeErr := file.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	info, err = file.Stat()
	if err != nil {
		return nil, nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, nil, fmt.Errorf("%q is not a regular file", path)
	}
	if !os.SameFile(pathInfo, info) {
		return nil, nil, fmt.Errorf("%q changed while opening", path)
	}
	if info.Size() > maximum {
		return nil, nil, fmt.Errorf("%q exceeds the %d-byte limit", path, maximum)
	}
	data, err := io.ReadAll(io.LimitReader(file, maximum+1))
	if err != nil {
		return nil, nil, err
	}
	if int64(len(data)) > maximum {
		return nil, nil, fmt.Errorf("%q grew beyond the %d-byte limit while reading", path, maximum)
	}
	return data, info, nil
}

func writeNewFileAtomic(path string, data []byte, mode os.FileMode) (err error) {
	directory := filepath.Dir(path)
	temp, err := os.CreateTemp(directory, ".legacy-authority-*.tmp")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	tempClosed := false
	tempRemoved := false
	defer func() {
		if !tempClosed {
			if closeErr := temp.Close(); err == nil && closeErr != nil {
				err = closeErr
			}
		}
		if !tempRemoved {
			if removeErr := os.Remove(tempPath); err == nil &&
				removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				err = removeErr
			}
		}
	}()

	if err = temp.Chmod(mode); err != nil {
		return err
	}
	if _, err = temp.Write(data); err != nil {
		return err
	}
	if err = temp.Sync(); err != nil {
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	tempClosed = true
	if err = os.Link(tempPath, path); err != nil {
		return err
	}
	if err = os.Remove(tempPath); err != nil {
		// The output link is already complete and synced through the same inode.
		// Treat cleanup failure as an error only after removing the new output.
		removeOutputErr := os.Remove(path)
		if removeOutputErr != nil {
			return errors.Join(err, removeOutputErr)
		}
		return err
	}
	tempRemoved = true

	directoryHandle, openErr := os.Open(directory)
	if openErr != nil {
		return removeAtomicOutputAfterError(path, directory, openErr)
	}
	syncErr := directoryHandle.Sync()
	closeErr := directoryHandle.Close()
	if syncErr != nil {
		return removeAtomicOutputAfterError(path, directory, syncErr)
	}
	if closeErr != nil {
		return removeAtomicOutputAfterError(path, directory, closeErr)
	}
	return nil
}

func removeAtomicOutputAfterError(path, directory string, cause error) error {
	if err := os.Remove(path); err != nil {
		return errors.Join(cause, fmt.Errorf("remove incomplete output: %w", err))
	}
	directoryHandle, err := os.Open(directory)
	if err != nil {
		return errors.Join(cause, fmt.Errorf("open output directory after rollback: %w", err))
	}
	syncErr := directoryHandle.Sync()
	closeErr := directoryHandle.Close()
	return errors.Join(cause, syncErr, closeErr)
}
