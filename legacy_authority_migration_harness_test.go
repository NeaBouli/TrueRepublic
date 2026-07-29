package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/math"
	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cometbft/cometbft/privval"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"truerepublic/migration"
	"truerepublic/token"
	rewards "truerepublic/treasury/keeper"
	"truerepublic/x/truedemocracy"
)

const legacyAuthoritySourceRevision = "0e51a05b008f395e3f7391358e117f0817d4eb39"

func TestMultiValidatorLegacyAuthorityMigrationRollback(t *testing.T) {
	if os.Getenv(multiValidatorSmokeEnv) != "1" {
		t.Skipf("set %s=1 to run the multi-validator process harness", multiValidatorSmokeEnv)
	}
	ctx := t.Context()
	legacyBinary := buildHistoricalDaemon(t, ctx, legacyAuthoritySourceRevision)
	currentBinary := filepath.Join(t.TempDir(), "truerepublicd-current")
	buildCurrent := exec.CommandContext(ctx, "go", "build", "-o", currentBinary, ".")
	if output, err := buildCurrent.CombinedOutput(); err != nil {
		t.Fatalf("build current daemon: %v\n%s", err, output)
	}

	const sourceChainID = "truerepublic-gh61-legacy-1"
	const targetChainID = "truerepublic-gh61-recovered-1"
	source := make([]*smokeValidator, 4)
	for i := range source {
		validator := &smokeValidator{
			name:    fmt.Sprintf("legacy-validator-%d", i+1),
			home:    filepath.Join(t.TempDir(), fmt.Sprintf("legacy-node-%d", i+1)),
			rpcPort: freeTCPPort(t),
			p2pPort: freeTCPPort(t),
			logPath: filepath.Join(t.TempDir(), fmt.Sprintf("legacy-validator-%d.log", i+1)),
		}
		initLegacySmokeValidator(t, ctx, legacyBinary, sourceChainID, validator)
		source[i] = validator
	}
	sharedSourceGenesis := buildLegacyCoupledGenesis(t, sourceChainID, source)
	for _, validator := range source {
		if err := atomicWriteFile(
			filepath.Join(validator.home, "config", "genesis.json"),
			sharedSourceGenesis,
			0o600,
		); err != nil {
			t.Fatalf("write legacy genesis for %s: %v", validator.name, err)
		}
	}

	target := make([]*smokeValidator, len(source))
	t.Cleanup(func() {
		for _, validator := range append(append([]*smokeValidator{}, source...), target...) {
			if validator != nil {
				_ = validator.stop(false)
			}
		}
		if t.Failed() {
			for _, validator := range append(append([]*smokeValidator{}, source...), target...) {
				if validator != nil {
					validator.logContents(t)
				}
			}
		}
	})

	for _, validator := range source {
		if err := validator.start(ctx, legacyBinary, persistentPeers(validator, source)); err != nil {
			t.Fatalf("start legacy %s: %v", validator.name, err)
		}
	}
	waitForSmokeHeight(t, source, 3, 120*time.Second)
	assertCommonAppHash(t, source, 3)

	// Remove quorum before recording the exact halt height. The two remaining
	// processes cannot commit another block, so the header hash and export are
	// bound to one stable source state.
	for _, validator := range source[2:] {
		if err := validator.stop(true); err != nil {
			t.Fatalf("stop legacy %s before halt: %v", validator.name, err)
		}
	}
	haltHeight, sourceAppHashHex := waitForStableSmokeHeightAndHash(t, source[:2], 30*time.Second)
	for _, validator := range source[:2] {
		if err := validator.stop(true); err != nil {
			t.Fatalf("stop legacy %s at halt: %v", validator.name, err)
		}
	}

	exportCommand := exec.CommandContext(ctx, legacyBinary, "export", "--home", source[0].home)
	var exportStderr bytes.Buffer
	exportCommand.Stderr = &exportStderr
	sourceExport, err := exportCommand.Output()
	if err != nil {
		t.Fatalf("export halted legacy chain: %v\nstderr: %s", err, exportStderr.String())
	}
	var exported struct {
		ChainID       string `json:"chain_id"`
		InitialHeight int64  `json:"initial_height"`
	}
	if err := json.Unmarshal(sourceExport, &exported); err != nil {
		t.Fatalf("decode halted legacy export: %v", err)
	}
	if exported.ChainID != sourceChainID || exported.InitialHeight != haltHeight+1 {
		t.Fatalf(
			"legacy export boundary = chain %q height %d, want %q/%d",
			exported.ChainID,
			exported.InitialHeight,
			sourceChainID,
			haltHeight+1,
		)
	}
	sourceAppHash, err := hex.DecodeString(sourceAppHashHex)
	if err != nil || len(sourceAppHash) != 32 {
		t.Fatalf("decode trusted source app hash %q: %v", sourceAppHashHex, err)
	}
	sourceGenesisSHA256 := sha256.Sum256(sourceExport)

	descriptor, freshKeys := buildLegacyAuthorityDescriptor(
		t,
		source,
		sourceChainID,
		targetChainID,
		haltHeight,
		sourceAppHash,
		sourceGenesisSHA256[:],
	)
	descriptorJSON, err := json.Marshal(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	artifacts := t.TempDir()
	descriptorPath := filepath.Join(artifacts, "descriptor.json")
	exportPath := filepath.Join(artifacts, "legacy-export.json")
	targetGenesisPath := filepath.Join(artifacts, "target-genesis.json")
	if err := os.WriteFile(descriptorPath, descriptorJSON, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exportPath, sourceExport, 0o600); err != nil {
		t.Fatal(err)
	}
	transformCommand := exec.CommandContext(
		ctx,
		currentBinary,
		"migration",
		"legacy-authority",
		"--descriptor", descriptorPath,
		"--genesis", exportPath,
		"--output", targetGenesisPath,
		"--source-app-hash", sourceAppHashHex,
	)
	if output, err := transformCommand.CombinedOutput(); err != nil {
		t.Fatalf("transform halted legacy export: %v\n%s", err, output)
	}
	targetGenesis, err := os.ReadFile(targetGenesisPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, sourceValidator := range source {
		if bytes.Contains(targetGenesis, []byte(sourceValidator.operatorAddr)) {
			t.Fatalf("target genesis retains legacy operator address %s", sourceValidator.operatorAddr)
		}
	}

	for i, sourceValidator := range source {
		freshOperator := sdk.AccAddress(freshKeys[i].PubKey().Address()).String()
		validator := &smokeValidator{
			name:         fmt.Sprintf("target-validator-%d", i+1),
			home:         filepath.Join(t.TempDir(), fmt.Sprintf("target-node-%d", i+1)),
			operatorAddr: freshOperator,
			pubKey:       append([]byte(nil), sourceValidator.pubKey...),
			rpcPort:      freeTCPPort(t),
			p2pPort:      freeTCPPort(t),
			logPath:      filepath.Join(t.TempDir(), fmt.Sprintf("target-validator-%d.log", i+1)),
		}
		initSmokeValidator(t, ctx, currentBinary, targetChainID, validator)
		copyPrivateValidatorKey(t, sourceValidator.home, validator.home)
		if err := atomicWriteFile(
			filepath.Join(validator.home, "config", "genesis.json"),
			targetGenesis,
			0o600,
		); err != nil {
			t.Fatalf("write target genesis for %s: %v", validator.name, err)
		}
		target[i] = validator
	}

	for _, validator := range target {
		if err := validator.start(ctx, currentBinary, persistentPeers(validator, target)); err != nil {
			t.Fatalf("start target %s: %v", validator.name, err)
		}
	}
	targetProofHeight := haltHeight + 3
	waitForSmokeHeight(t, target, targetProofHeight, 120*time.Second)
	assertCommonAppHash(t, target, targetProofHeight)
	for i, validator := range target {
		waitForPositiveSmokeValidatorPower(t, validator, source[i].pubKey, 30*time.Second)
	}
	for _, validator := range target {
		if err := validator.stop(true); err != nil {
			t.Fatalf("stop target %s before rollback: %v", validator.name, err)
		}
	}

	recovered := exportSmokeGenesis(t, ctx, currentBinary, target[0], targetProofHeight)
	recoveredApp := newGenesisTestApp(t)
	if err := validateLedgerGenesis(recoveredApp.appCodec, recovered.AppState); err != nil {
		t.Fatalf("target export ledger invalid: %v", err)
	}
	if err := initGenesisApp(newGenesisTestApp(t), recovered.AppState); err != nil {
		t.Fatalf("re-import target export: %v", err)
	}

	// Rollback is permitted only with every target signer stopped. Source homes
	// were never modified; restarting them proves a clean return to the exact
	// legacy chain without running either consensus identity twice.
	for _, validator := range source {
		if err := validator.start(ctx, legacyBinary, persistentPeers(validator, source)); err != nil {
			t.Fatalf("restart source %s for rollback: %v", validator.name, err)
		}
	}
	rollbackHeight := haltHeight + 2
	waitForSmokeHeight(t, source, rollbackHeight, 120*time.Second)
	assertCommonAppHash(t, source, rollbackHeight)
	for _, validator := range source {
		if err := validator.stop(true); err != nil {
			t.Fatalf("stop rolled-back source %s: %v", validator.name, err)
		}
	}
}

func waitForPositiveSmokeValidatorPower(
	t *testing.T,
	validator *smokeValidator,
	pubKey []byte,
	timeout time.Duration,
) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		power, err := querySmokeValidatorPower(t.Context(), validator, pubKey)
		parsed, parseErr := strconv.ParseInt(power, 10, 64)
		if err == nil && parseErr == nil && parsed > 0 {
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf(
		"%s validator set did not include pubkey %x with positive power within %s",
		validator.name,
		pubKey,
		timeout,
	)
}

func buildHistoricalDaemon(t *testing.T, ctx context.Context, revision string) string {
	t.Helper()
	sourceDir := filepath.Join(t.TempDir(), "historical-source")
	if err := os.MkdirAll(sourceDir, 0o700); err != nil {
		t.Fatal(err)
	}
	archivePath := filepath.Join(t.TempDir(), "historical-source.tar")
	archive := exec.CommandContext(ctx, "git", "archive", "--format=tar", "--output", archivePath, revision)
	if output, err := archive.CombinedOutput(); err != nil {
		t.Fatalf("archive historical revision %s: %v\n%s", revision, err, output)
	}
	extract := exec.CommandContext(ctx, "tar", "-xf", archivePath, "-C", sourceDir)
	if output, err := extract.CombinedOutput(); err != nil {
		t.Fatalf("extract historical revision %s: %v\n%s", revision, err, output)
	}
	binary := filepath.Join(t.TempDir(), "truerepublicd-legacy")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	build.Dir = sourceDir
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build historical revision %s: %v\n%s", revision, err, output)
	}
	return binary
}

func initLegacySmokeValidator(
	t *testing.T,
	ctx context.Context,
	binary string,
	chainID string,
	validator *smokeValidator,
) {
	t.Helper()
	initCommand := exec.CommandContext(
		ctx,
		binary,
		"init",
		validator.name,
		"--chain-id", chainID,
		"--home", validator.home,
	)
	if output, err := initCommand.CombinedOutput(); err != nil {
		t.Fatalf("init legacy %s: %v\n%s", validator.name, err, output)
	}
	configureLocalhostSmokeP2P(t, filepath.Join(validator.home, "config", "config.toml"))
	filePV := privval.LoadFilePV(
		filepath.Join(validator.home, "config", "priv_validator_key.json"),
		filepath.Join(validator.home, "data", "priv_validator_state.json"),
	)
	pubKey, err := filePV.GetPubKey()
	if err != nil {
		t.Fatalf("read legacy %s consensus key: %v", validator.name, err)
	}
	validator.pubKey = append([]byte(nil), pubKey.Bytes()...)
	validator.operatorAddr = sdk.MustBech32ifyAddressBytes(
		"truerepublic",
		tmhash.SumTruncated(validator.pubKey),
	)
	nodeIDCommand := exec.CommandContext(ctx, binary, "comet", "show-node-id", "--home", validator.home)
	var nodeIDStderr bytes.Buffer
	nodeIDCommand.Stderr = &nodeIDStderr
	nodeID, err := nodeIDCommand.Output()
	if err != nil {
		t.Fatalf("read legacy %s node ID: %v\nstderr: %s", validator.name, err, nodeIDStderr.String())
	}
	validator.nodeID = strings.TrimSpace(string(nodeID))
}

func buildLegacyCoupledGenesis(
	t *testing.T,
	chainID string,
	validators []*smokeValidator,
) []byte {
	t.Helper()
	app := newGenesisTestApp(t)
	appState := defaultGenesisForApp(app)
	members := make([]string, len(validators))
	genesisValidators := make([]truedemocracy.GenesisValidator, len(validators))
	genesisAccounts := make(authtypes.GenesisAccounts, len(validators))
	consensusValidators := make([]cmttypes.GenesisValidator, len(validators))
	for i, validator := range validators {
		members[i] = validator.operatorAddr
		genesisValidators[i] = truedemocracy.GenesisValidator{
			OperatorAddr: validator.operatorAddr,
			PubKey:       append([]byte(nil), validator.pubKey...),
			Stake:        rewards.StakeMin,
			Domain:       "Bootstrap",
		}
		genesisAccounts[i] = authtypes.NewBaseAccount(
			sdk.MustAccAddressFromBech32(validator.operatorAddr),
			&ed25519.PubKey{Key: append([]byte(nil), validator.pubKey...)},
			uint64(i),
			0,
		)
		key := cmted25519.PubKey(append([]byte(nil), validator.pubKey...))
		consensusValidators[i] = cmttypes.GenesisValidator{
			Address: key.Address(),
			PubKey:  key,
			Power:   1,
			Name:    validator.name,
		}
	}
	setJSONGenesis(t, appState, truedemocracy.ModuleName, truedemocracy.GenesisState{
		Domains: []truedemocracy.Domain{{
			Name:          "Bootstrap",
			Admin:         sdk.MustAccAddressFromBech32(members[0]),
			Members:       members,
			Treasury:      sdk.NewCoins(),
			Issues:        []truedemocracy.Issue{},
			Options:       truedemocracy.DomainOptions{AdminElectable: true},
			PermissionReg: []string{},
		}},
		Validators: genesisValidators,
	})
	packedAccounts, err := authtypes.PackAccounts(genesisAccounts)
	if err != nil {
		t.Fatal(err)
	}
	authState := authtypes.GenesisState{Params: authtypes.DefaultParams(), Accounts: packedAccounts}
	appState[authtypes.ModuleName], err = app.appCodec.MarshalJSON(&authState)
	if err != nil {
		t.Fatal(err)
	}
	totalStake := math.NewInt(rewards.StakeMin).MulRaw(int64(len(validators)))
	setBankGenesis(t, app, appState, []banktypes.Balance{{
		Address: authtypes.NewModuleAddress(truedemocracy.ModuleName).String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(token.BaseDenom, totalStake)),
	}})
	appStateJSON, err := json.Marshal(appState)
	if err != nil {
		t.Fatal(err)
	}
	genesis := &genutiltypes.AppGenesis{
		AppName:       "truerepublicd",
		AppVersion:    "legacy-gh61-fixture",
		GenesisTime:   time.Unix(1, 0).UTC(),
		ChainID:       chainID,
		InitialHeight: 1,
		AppState:      appStateJSON,
		Consensus: &genutiltypes.ConsensusGenesis{
			Validators: consensusValidators,
			Params:     cmttypes.DefaultConsensusParams(),
		},
	}
	if err := genesis.ValidateAndComplete(); err != nil {
		t.Fatalf("validate shared legacy genesis: %v", err)
	}
	output, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(output, []byte(`"priv_key"`)) {
		t.Fatal("shared legacy genesis contains private validator material")
	}
	return output
}

func waitForStableSmokeHeightAndHash(
	t *testing.T,
	validators []*smokeValidator,
	timeout time.Duration,
) (int64, string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var previousHeight int64 = -1
	stable := 0
	for time.Now().Before(deadline) {
		height := smokeHeight(t, validators[0])
		if smokeHeight(t, validators[1]) == height {
			firstHash := smokeAppHash(t, validators[0], height)
			secondHash := smokeAppHash(t, validators[1], height)
			if firstHash == secondHash && firstHash != "" {
				if height == previousHeight {
					stable++
				} else {
					previousHeight = height
					stable = 1
				}
				if stable >= 3 {
					return height, firstHash
				}
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("legacy source did not reach a stable common halt height within %s", timeout)
	return 0, ""
}

func buildLegacyAuthorityDescriptor(
	t *testing.T,
	validators []*smokeValidator,
	sourceChainID string,
	targetChainID string,
	haltHeight int64,
	sourceAppHash []byte,
	sourceGenesisSHA256 []byte,
) (*migration.Descriptor, []cryptotypes.PrivKey) {
	t.Helper()
	type entry struct {
		mapping migration.OperatorMapping
		key     cryptotypes.PrivKey
	}
	entries := make([]entry, len(validators))
	for i, validator := range validators {
		fresh := secp256k1.GenPrivKey()
		entries[i] = entry{
			mapping: migration.OperatorMapping{
				OldOperator: validator.operatorAddr,
				NewOperator: sdk.AccAddress(fresh.PubKey().Address()).String(),
				PubKeyType:  migration.PubKeyTypeSecp256k1,
				PubKey:      fresh.PubKey().Bytes(),
			},
			key: fresh,
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		left := sdk.MustAccAddressFromBech32(entries[i].mapping.OldOperator)
		right := sdk.MustAccAddressFromBech32(entries[j].mapping.OldOperator)
		return bytes.Compare(left, right) < 0
	})
	descriptor := &migration.Descriptor{
		Version:             migration.DescriptorVersion,
		SourceChainID:       sourceChainID,
		TargetChainID:       targetChainID,
		HaltHeight:          haltHeight,
		SourceAppHash:       append([]byte(nil), sourceAppHash...),
		SourceGenesisSHA256: append([]byte(nil), sourceGenesisSHA256...),
		TransformID:         "gh-61-live-four-validator-v1",
		Mappings:            make([]migration.OperatorMapping, len(entries)),
	}
	sortedKeys := make([]cryptotypes.PrivKey, len(entries))
	for i, entry := range entries {
		descriptor.Mappings[i] = entry.mapping
		sortedKeys[i] = entry.key
	}
	signingBytes, err := migration.SigningBytes(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	for i, key := range sortedKeys {
		descriptor.Mappings[i].Signature, err = key.Sign(signingBytes)
		if err != nil {
			t.Fatal(err)
		}
	}

	// The target validators are initialized in original validator order, so
	// return the fresh keys in that order rather than canonical descriptor order.
	keysByOld := make(map[string]cryptotypes.PrivKey, len(entries))
	for _, entry := range entries {
		keysByOld[entry.mapping.OldOperator] = entry.key
	}
	originalOrder := make([]cryptotypes.PrivKey, len(validators))
	for i, validator := range validators {
		originalOrder[i] = keysByOld[validator.operatorAddr]
	}
	return descriptor, originalOrder
}

// copyPrivateValidatorKey copies ephemeral consensus key material strictly
// between isolated t.TempDir test homes created by this harness. It exists only
// so the throwaway target validators can continue the throwaway source
// consensus identity; it must never be used with real or production validator
// homes, keys, or signer state.
func copyPrivateValidatorKey(t *testing.T, sourceHome, targetHome string) {
	t.Helper()
	sourcePath := filepath.Join(sourceHome, "config", "priv_validator_key.json")
	targetPath := filepath.Join(targetHome, "config", "priv_validator_key.json")
	privateKey, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("read isolated source consensus key: %v", err)
	}
	if err := atomicWriteFile(targetPath, privateKey, 0o600); err != nil {
		t.Fatalf("install isolated target consensus key: %v", err)
	}
}
