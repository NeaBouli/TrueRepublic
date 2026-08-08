package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cosmossdk.io/math"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/tmhash"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"truerepublic/migration"
	"truerepublic/token"
	rewards "truerepublic/treasury/keeper"
	"truerepublic/x/truedemocracy"
)

func TestLegacyAuthorityMigrationCommandTransformsAndReimports(t *testing.T) {
	app := newGenesisTestApp(t)
	appState := defaultGenesisForApp(app)
	consensusKey := ed25519.GenPrivKey()
	oldOperator := sdk.AccAddress(tmhash.SumTruncated(consensusKey.PubKey().Bytes())).String()
	freshKey := secp256k1.GenPrivKey()
	newOperator := sdk.AccAddress(freshKey.PubKey().Address()).String()

	democracyState := truedemocracy.GenesisState{
		Domains: []truedemocracy.Domain{{
			Name:          "legacy",
			Admin:         sdk.MustAccAddressFromBech32(oldOperator),
			Members:       []string{oldOperator},
			Treasury:      sdk.NewCoins(),
			Issues:        []truedemocracy.Issue{},
			Options:       truedemocracy.DomainOptions{AdminElectable: true},
			PermissionReg: []string{},
		}},
		Validators: []truedemocracy.GenesisValidator{{
			OperatorAddr: oldOperator,
			PubKey:       consensusKey.PubKey().Bytes(),
			Stake:        rewards.StakeMin,
			Domain:       "legacy",
		}},
	}
	setJSONGenesis(t, appState, truedemocracy.ModuleName, democracyState)

	legacyAccount := authtypes.NewBaseAccount(
		sdk.MustAccAddressFromBech32(oldOperator),
		&ed25519.PubKey{Key: consensusKey.PubKey().Bytes()},
		7,
		3,
	)
	packedAccounts, err := authtypes.PackAccounts(authtypes.GenesisAccounts{legacyAccount})
	if err != nil {
		t.Fatal(err)
	}
	authState := authtypes.GenesisState{Params: authtypes.DefaultParams(), Accounts: packedAccounts}
	appState[authtypes.ModuleName], err = app.appCodec.MarshalJSON(&authState)
	if err != nil {
		t.Fatal(err)
	}
	stake := sdk.NewCoins(sdk.NewCoin(token.BaseDenom, math.NewInt(rewards.StakeMin)))
	setBankGenesis(t, app, appState, []banktypes.Balance{{
		Address: authtypes.NewModuleAddress(truedemocracy.ModuleName).String(),
		Coins:   stake,
	}})

	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		t.Fatal(err)
	}
	cometKey := cmted25519.PubKey(consensusKey.PubKey().Bytes())
	consensusJSON, err := json.Marshal(&genutiltypes.ConsensusGenesis{
		Validators: []cmttypes.GenesisValidator{{
			Address: cometKey.Address(),
			PubKey:  cometKey,
			Power:   1,
			Name:    "legacy-validator",
		}},
		Params: cmttypes.DefaultConsensusParams(),
	})
	if err != nil {
		t.Fatal(err)
	}
	sourceGenesis, err := json.Marshal(map[string]json.RawMessage{
		"app_hash":       json.RawMessage(`null`),
		"app_state":      appStateJSON,
		"chain_id":       json.RawMessage(`"legacy-chain"`),
		"consensus":      consensusJSON,
		"initial_height": json.RawMessage(`101`),
	})
	if err != nil {
		t.Fatal(err)
	}

	sourceAppHash := bytes.Repeat([]byte{0x73}, 32)
	sourceGenesisSHA256 := sha256.Sum256(sourceGenesis)
	descriptor := &migration.Descriptor{
		Version:             migration.DescriptorVersion,
		SourceChainID:       "legacy-chain",
		TargetChainID:       "recovered-chain",
		HaltHeight:          100,
		SourceAppHash:       sourceAppHash,
		SourceGenesisSHA256: append([]byte(nil), sourceGenesisSHA256[:]...),
		TransformID:         "gh-61-cli-round-trip",
		Mappings: []migration.OperatorMapping{{
			OldOperator: oldOperator,
			NewOperator: newOperator,
			PubKeyType:  migration.PubKeyTypeSecp256k1,
			PubKey:      freshKey.PubKey().Bytes(),
		}},
	}
	signingBytes, err := migration.SigningBytes(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	descriptor.Mappings[0].Signature, err = freshKey.Sign(signingBytes)
	if err != nil {
		t.Fatal(err)
	}
	descriptorJSON, err := json.Marshal(descriptor)
	if err != nil {
		t.Fatal(err)
	}

	directory := t.TempDir()
	descriptorPath := filepath.Join(directory, "descriptor.json")
	genesisPath := filepath.Join(directory, "export.json")
	outputPath := filepath.Join(directory, "target.json")
	if err := os.WriteFile(descriptorPath, descriptorJSON, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(genesisPath, sourceGenesis, 0o600); err != nil {
		t.Fatal(err)
	}
	command := newMigrationCmd(app.appCodec)
	var stdout bytes.Buffer
	command.SetOut(&stdout)
	command.SetErr(io.Discard)
	command.SetArgs([]string{
		"legacy-authority",
		"--descriptor", descriptorPath,
		"--genesis", genesisPath,
		"--output", outputPath,
		"--source-app-hash", hex.EncodeToString(sourceAppHash),
	})
	if err := command.Execute(); err != nil {
		t.Fatalf("execute migration command: %v", err)
	}
	if !strings.Contains(stdout.String(), outputPath) ||
		!strings.Contains(stdout.String(), descriptor.TransformID) ||
		strings.Contains(stdout.String(), oldOperator) ||
		strings.Contains(stdout.String(), newOperator) {
		t.Fatalf("unsafe or incomplete command output: %q", stdout.String())
	}

	transformed, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	var target struct {
		ChainID  string                     `json:"chain_id"`
		AppState map[string]json.RawMessage `json:"app_state"`
	}
	if err := json.Unmarshal(transformed, &target); err != nil {
		t.Fatal(err)
	}
	if target.ChainID != descriptor.TargetChainID {
		t.Fatalf("target chain ID = %q", target.ChainID)
	}
	if bytes.Contains(transformed, []byte(oldOperator)) {
		t.Fatal("transformed genesis still contains the legacy operator")
	}
	if err := initGenesisApp(newGenesisTestApp(t), target.AppState); err != nil {
		t.Fatalf("re-import transformed genesis: %v", err)
	}
}

func TestLegacyAuthorityMigrationCommandRequiresEveryFlag(t *testing.T) {
	command := newMigrationCmd(testMigrationCodec())
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	command.SetArgs([]string{"legacy-authority"})

	err := command.Execute()
	if err == nil {
		t.Fatal("command accepted missing required flags")
	}
	for _, flag := range []string{"--descriptor", "--genesis", "--output", "--source-app-hash"} {
		if !strings.Contains(err.Error(), flag) {
			t.Fatalf("missing-flags error omits %s: %v", flag, err)
		}
	}
}

func TestParseTrustedAppHash(t *testing.T) {
	want := bytes.Repeat([]byte{0xab}, 32)
	for _, input := range []string{
		hex.EncodeToString(want),
		strings.ToUpper(hex.EncodeToString(want)),
		"0x" + hex.EncodeToString(want),
		"  0X" + strings.ToUpper(hex.EncodeToString(want)) + "  ",
	} {
		got, err := parseTrustedAppHash(input)
		if err != nil {
			t.Fatalf("parse %q: %v", input, err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("parse %q produced %x", input, got)
		}
	}

	for _, input := range []string{"", "abc", strings.Repeat("00", 31), strings.Repeat("00", 33), "not-hex"} {
		if _, err := parseTrustedAppHash(input); err == nil {
			t.Fatalf("invalid hash %q was accepted", input)
		}
	}
}

func TestDecodeMigrationDescriptorStrict(t *testing.T) {
	raw := testDescriptorJSON(t, bytes.Repeat([]byte{0x42}, 32), []byte("test-export"))
	descriptor, err := decodeMigrationDescriptorStrict(raw)
	if err != nil {
		t.Fatal(err)
	}
	if descriptor.TransformID != "test-transform" {
		t.Fatalf("transform ID = %q", descriptor.TransformID)
	}

	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil {
		t.Fatal(err)
	}
	object["unknown"] = json.RawMessage(`true`)
	withUnknown, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeMigrationDescriptorStrict(withUnknown); err == nil {
		t.Fatal("unknown descriptor field was accepted")
	}
	if _, err := decodeMigrationDescriptorStrict(append(raw, []byte(` {}`)...)); err == nil {
		t.Fatal("trailing descriptor JSON was accepted")
	}
	if _, err := decodeMigrationDescriptorStrict([]byte(`{"version":`)); err == nil {
		t.Fatal("malformed descriptor was accepted")
	}
}

func TestReadRegularFileBounded(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "input.json")
	if err := os.WriteFile(path, []byte("1234"), 0o600); err != nil {
		t.Fatal(err)
	}
	data, _, err := readRegularFileBounded(path, 4)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "1234" {
		t.Fatalf("read data = %q", data)
	}
	if _, _, err := readRegularFileBounded(path, 3); err == nil {
		t.Fatal("oversized input was accepted")
	}
	if _, _, err := readRegularFileBounded(directory, 100); err == nil {
		t.Fatal("directory input was accepted")
	}

	symlink := filepath.Join(directory, "input-link")
	if err := os.Symlink(path, symlink); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	if _, _, err := readRegularFileBounded(symlink, 100); err == nil {
		t.Fatal("symlink input was accepted")
	}
}

func TestWriteNewFileAtomicCreatesPrivateOutput(t *testing.T) {
	directory := t.TempDir()
	output := filepath.Join(directory, "output.json")
	if err := writeNewFileAtomic(output, []byte("complete"), 0o600); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "complete" {
		t.Fatalf("output = %q", data)
	}
	info, err := os.Stat(output)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("output permissions = %o", got)
	}
	assertNoMigrationTempFiles(t, directory)
}

func TestWriteNewFileAtomicRefusesExistingOutputAndCleansTemp(t *testing.T) {
	directory := t.TempDir()
	output := filepath.Join(directory, "output.json")
	if err := os.WriteFile(output, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := writeNewFileAtomic(output, []byte("replacement"), 0o600); err == nil {
		t.Fatal("existing output was overwritten")
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "original" {
		t.Fatalf("existing output changed to %q", data)
	}
	assertNoMigrationTempFiles(t, directory)
}

func TestRunLegacyAuthorityMigrationFailsWithoutOutput(t *testing.T) {
	hash := bytes.Repeat([]byte{0x31}, 32)
	genesis := []byte(`not-json`)
	directory := t.TempDir()
	descriptorPath := filepath.Join(directory, "descriptor.json")
	genesisPath := filepath.Join(directory, "genesis.json")
	if err := os.WriteFile(descriptorPath, testDescriptorJSON(t, hash, genesis), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(genesisPath, genesis, 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		options legacyAuthorityMigrationOptions
	}{
		{
			name: "trusted hash mismatch",
			options: legacyAuthorityMigrationOptions{
				descriptorPath: descriptorPath,
				genesisPath:    genesisPath,
				outputPath:     filepath.Join(directory, "mismatch.json"),
				sourceAppHash:  hex.EncodeToString(bytes.Repeat([]byte{0x32}, 32)),
			},
		},
		{
			name: "transform failure",
			options: legacyAuthorityMigrationOptions{
				descriptorPath: descriptorPath,
				genesisPath:    genesisPath,
				outputPath:     filepath.Join(directory, "invalid.json"),
				sourceAppHash:  hex.EncodeToString(hash),
			},
		},
		{
			name: "output aliases descriptor",
			options: legacyAuthorityMigrationOptions{
				descriptorPath: descriptorPath,
				genesisPath:    genesisPath,
				outputPath:     descriptorPath,
				sourceAppHash:  hex.EncodeToString(hash),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := runLegacyAuthorityMigration(testMigrationCodec(), test.options)
			if err == nil {
				t.Fatal("unsafe migration input was accepted")
			}
			if test.options.outputPath != descriptorPath {
				if _, statErr := os.Lstat(test.options.outputPath); !errors.Is(statErr, os.ErrNotExist) {
					t.Fatalf("failure created output: %v", statErr)
				}
			}
		})
	}
	assertNoMigrationTempFiles(t, directory)
}

func TestRunLegacyAuthorityMigrationRejectsChangedExportWithoutOutput(t *testing.T) {
	appHash := bytes.Repeat([]byte{0x41}, 32)
	signedExport := []byte(`{"chain_id":"source-chain","app_state":{}}`)
	changedExport := append(append([]byte(nil), signedExport...), '\n')
	directory := t.TempDir()
	descriptorPath := filepath.Join(directory, "descriptor.json")
	genesisPath := filepath.Join(directory, "genesis.json")
	outputPath := filepath.Join(directory, "output.json")

	if err := os.WriteFile(
		descriptorPath,
		testDescriptorJSON(t, appHash, signedExport),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(genesisPath, changedExport, 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := runLegacyAuthorityMigration(testMigrationCodec(), legacyAuthorityMigrationOptions{
		descriptorPath: descriptorPath,
		genesisPath:    genesisPath,
		outputPath:     outputPath,
		sourceAppHash:  hex.EncodeToString(appHash),
	})
	if err == nil || !strings.Contains(err.Error(), "source genesis SHA-256") {
		t.Fatalf("changed-export error = %v", err)
	}
	if _, statErr := os.Lstat(outputPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("changed-export failure created output: %v", statErr)
	}
	assertNoMigrationTempFiles(t, directory)
}

func TestRunLegacyAuthorityMigrationRejectsSameInputIdentity(t *testing.T) {
	hash := bytes.Repeat([]byte{0x51}, 32)
	directory := t.TempDir()
	descriptorPath := filepath.Join(directory, "descriptor.json")
	genesisPath := filepath.Join(directory, "genesis.json")
	if err := os.WriteFile(descriptorPath, testDescriptorJSON(t, hash, nil), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(descriptorPath, genesisPath); err != nil {
		t.Skipf("hard links unsupported: %v", err)
	}
	output := filepath.Join(directory, "output.json")
	_, err := runLegacyAuthorityMigration(testMigrationCodec(), legacyAuthorityMigrationOptions{
		descriptorPath: descriptorPath,
		genesisPath:    genesisPath,
		outputPath:     output,
		sourceAppHash:  hex.EncodeToString(hash),
	})
	if err == nil || !strings.Contains(err.Error(), "different files") {
		t.Fatalf("same input identity error = %v", err)
	}
	if _, statErr := os.Lstat(output); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("same-input failure created output: %v", statErr)
	}
}

func TestRemoveAtomicOutputAfterErrorRollsBackOutput(t *testing.T) {
	directory := t.TempDir()
	output := filepath.Join(directory, "output.json")
	if err := os.WriteFile(output, []byte("incomplete"), 0o600); err != nil {
		t.Fatal(err)
	}
	cause := errors.New("directory sync failed")
	err := removeAtomicOutputAfterError(output, directory, cause)
	if !errors.Is(err, cause) {
		t.Fatalf("rollback error lost the original cause: %v", err)
	}
	if _, statErr := os.Lstat(output); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("incomplete output survived rollback: %v", statErr)
	}
	assertNoMigrationTempFiles(t, directory)
}

func TestRemoveAtomicOutputAfterErrorJoinsRemovalFailure(t *testing.T) {
	directory := t.TempDir()
	missing := filepath.Join(directory, "already-removed.json")
	cause := errors.New("directory sync failed")
	err := removeAtomicOutputAfterError(missing, directory, cause)
	if err == nil {
		t.Fatal("rollback of a missing output succeeded")
	}
	if !errors.Is(err, cause) {
		t.Fatalf("joined error lost the original cause: %v", err)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("joined error lost the removal failure: %v", err)
	}
	if !strings.Contains(err.Error(), "remove incomplete output") {
		t.Fatalf("joined error lacks removal context: %v", err)
	}
}

func TestRemoveAtomicOutputAfterErrorJoinsDirectoryOpenFailure(t *testing.T) {
	directory := t.TempDir()
	output := filepath.Join(directory, "output.json")
	if err := os.WriteFile(output, []byte("incomplete"), 0o600); err != nil {
		t.Fatal(err)
	}
	missingDirectory := filepath.Join(directory, "no-such-directory")
	cause := errors.New("directory close failed")
	err := removeAtomicOutputAfterError(output, missingDirectory, cause)
	if err == nil {
		t.Fatal("rollback with an unopenable directory succeeded")
	}
	if !errors.Is(err, cause) {
		t.Fatalf("joined error lost the original cause: %v", err)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("joined error lost the directory open failure: %v", err)
	}
	if !strings.Contains(err.Error(), "open output directory after rollback") {
		t.Fatalf("joined error lacks directory context: %v", err)
	}
	// The incomplete output is still removed before the directory open is
	// attempted.
	if _, statErr := os.Lstat(output); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("incomplete output survived rollback: %v", statErr)
	}
}

func testDescriptorJSON(t *testing.T, hash, genesis []byte) []byte {
	t.Helper()
	sourceGenesisSHA256 := sha256.Sum256(genesis)
	raw, err := json.Marshal(migration.Descriptor{
		Version:             migration.DescriptorVersion,
		SourceChainID:       "source-chain",
		TargetChainID:       "target-chain",
		HaltHeight:          100,
		SourceAppHash:       hash,
		SourceGenesisSHA256: append([]byte(nil), sourceGenesisSHA256[:]...),
		TransformID:         "test-transform",
		Mappings:            []migration.OperatorMapping{},
	})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func testMigrationCodec() codec.Codec {
	return codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
}

func assertNoMigrationTempFiles(t *testing.T, directory string) {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".legacy-authority-") {
			t.Fatalf("temporary file remains: %s", entry.Name())
		}
	}
}
