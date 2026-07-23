package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"truerepublic/token"
	"truerepublic/x/truedemocracy"
)

const anyDeliverTxFailure = ^uint32(0)

// TestMultiValidatorConsensusKeyRotation proves the complete GH-56 operator
// ceremony against real CometBFT processes. In particular, it keeps the old
// signer offline, activates a pre-synced replacement key through an
// authenticated transaction, and verifies both the H -> H+2 transition and
// permanent old-key revocation across export/import.
func TestMultiValidatorConsensusKeyRotation(t *testing.T) {
	if os.Getenv(multiValidatorSmokeEnv) != "1" {
		t.Skipf("set %s=1 to run the consensus-key rotation process harness", multiValidatorSmokeEnv)
	}
	ctx := t.Context()

	binary := filepath.Join(t.TempDir(), "truerepublicd")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build daemon: %v\n%s", err, output)
	}

	const chainID = "truerepublic-validator-key-rotation-1"
	validators := make([]*smokeValidator, 4)
	operators := make([]smokeAccount, len(validators))
	for i := range validators {
		validator := &smokeValidator{
			name:    fmt.Sprintf("rotation-validator-%d", i+1),
			home:    filepath.Join(t.TempDir(), fmt.Sprintf("rotation-node-%d", i+1)),
			rpcPort: freeTCPPort(t),
			p2pPort: freeTCPPort(t),
			logPath: filepath.Join(t.TempDir(), fmt.Sprintf("rotation-validator-%d.log", i+1)),
		}
		operator := addSmokeKey(t, ctx, binary, validator.home, validator.name+"-operator", uint64(i), 100_000*token.WholeTokenBaseUnits)
		validator.operatorAddr = operator.address
		initSmokeValidator(t, ctx, binary, chainID, validator)
		validators[i] = validator
		operators[i] = operator
	}

	replacement := &smokeValidator{
		name:    "rotation-replacement",
		home:    filepath.Join(t.TempDir(), "rotation-replacement"),
		rpcPort: freeTCPPort(t),
		p2pPort: freeTCPPort(t),
		logPath: filepath.Join(t.TempDir(), "rotation-replacement.log"),
	}
	// The replacement's temporary single-node genesis is overwritten below;
	// reuse a daemon-derived chain-prefix address solely to satisfy fail-closed
	// init without creating another transaction identity.
	replacement.operatorAddr = operators[0].address
	initSmokeValidator(t, ctx, binary, chainID, replacement)

	sharedGenesis := buildSharedSmokeGenesis(t, chainID, validators, operators...)
	allNodes := append(append([]*smokeValidator{}, validators...), replacement)
	for _, node := range allNodes {
		if err := atomicWriteFile(filepath.Join(node.home, "config", "genesis.json"), sharedGenesis, 0o600); err != nil {
			t.Fatalf("write %s shared genesis: %v", node.name, err)
		}
	}

	t.Cleanup(func() {
		for _, node := range allNodes {
			_ = node.stop(false)
		}
		if t.Failed() {
			for _, node := range allNodes {
				node.logContents(t)
			}
		}
	})

	for _, validator := range validators {
		if err := validator.start(ctx, binary, persistentPeers(validator, allNodes)); err != nil {
			t.Fatalf("start %s: %v", validator.name, err)
		}
	}
	waitForSmokeHeight(t, validators, 2, 90*time.Second)
	assertCommonAppHash(t, validators, 2)

	if err := replacement.start(ctx, binary, persistentPeers(replacement, allNodes)); err != nil {
		t.Fatalf("start pre-synced replacement full node: %v", err)
	}
	waitForSmokeHeight(t, []*smokeValidator{replacement}, smokeHeight(t, validators[0]), 90*time.Second)
	if power, found, err := querySmokeValidatorPowerAtHeight(ctx, validators[0], replacement.pubKey, 0); err != nil {
		t.Fatalf("query replacement before rotation: %v", err)
	} else if found || power != "" {
		t.Fatalf("replacement key was active before rotation with power %q", power)
	}

	targetIndex := len(validators) - 1
	target := validators[targetIndex]
	oldPubKey := append([]byte(nil), target.pubKey...)
	oldOperator := target.operatorAddr
	oldValidator := querySmokeApplicationValidator(t, ctx, binary, validators[0], oldOperator)
	if oldValidator.Power != 1 {
		t.Fatalf("target application power = %d, want 1", oldValidator.Power)
	}

	if err := target.stop(true); err != nil {
		t.Fatalf("stop old signer %s: %v", target.name, err)
	}
	oldStoppedState := readValidatorIdentityState(t, target)
	newPreRotationState := readValidatorIdentityState(t, replacement)
	if _, err := querySmokeHeight(ctx, target); err == nil {
		t.Fatalf("old signer %s RPC remained reachable after shutdown", target.name)
	}

	rotationHeight, rawLog := runSmokeTxWithExpectedCode(t, ctx, binary, validators[0], &operators[targetIndex], chainID, 0,
		"rotate-validator-key", hex.EncodeToString(oldPubKey), hex.EncodeToString(replacement.pubKey))
	if rawLog != "" && rawLog != "[]" {
		t.Logf("rotation DeliverTx log at height %d: %s", rotationHeight, rawLog)
	}
	assertRotationValidatorUpdates(t, ctx, validators[0], rotationHeight, oldPubKey, replacement.pubKey, "1")

	oldPower, oldFound, err := querySmokeValidatorPowerAtHeight(ctx, validators[0], oldPubKey, rotationHeight+1)
	if err != nil {
		t.Fatalf("query old validator at H+1: %v", err)
	}
	if !oldFound || oldPower != "1" {
		t.Fatalf("old validator at H+1 had found/power %t/%q, want true/1", oldFound, oldPower)
	}

	waitForSmokeHeight(t, append(append([]*smokeValidator{}, validators[:targetIndex]...), replacement), rotationHeight+3, 120*time.Second)
	oldPower, oldFound, err = querySmokeValidatorPowerAtHeight(ctx, validators[0], oldPubKey, rotationHeight+2)
	if err != nil {
		t.Fatalf("query old validator at H+2: %v", err)
	}
	if oldFound {
		t.Fatalf("old validator remained in the H+2 set with power %q", oldPower)
	}
	newPower, newFound, err := querySmokeValidatorPowerAtHeight(ctx, validators[0], replacement.pubKey, rotationHeight+2)
	if err != nil {
		t.Fatalf("query new validator at H+2: %v", err)
	}
	if !newFound || newPower != "1" {
		t.Fatalf("new validator at H+2 had found/power %t/%q, want true/1", newFound, newPower)
	}
	assertCommitSignedBy(t, ctx, validators[0], rotationHeight+2, replacement.pubKey)
	assertSigningStateAdvanced(t, replacement.name, readValidatorIdentityState(t, replacement).lastSign, newPreRotationState.lastSign)

	if target.command != nil {
		t.Fatalf("old signer %s restarted during rotation", target.name)
	}
	if _, err := querySmokeHeight(ctx, target); err == nil {
		t.Fatalf("old signer %s RPC became reachable during rotation", target.name)
	}
	oldFinalState := readValidatorIdentityState(t, target)
	if !bytes.Equal(oldFinalState.lastSignRaw, oldStoppedState.lastSignRaw) {
		t.Fatalf("old signer state changed after shutdown: before %s, after %s", oldStoppedState.lastSignRaw, oldFinalState.lastSignRaw)
	}

	rotatedValidator := querySmokeApplicationValidator(t, ctx, binary, validators[0], oldOperator)
	if !bytes.Equal(rotatedValidator.PubKey, replacement.pubKey) {
		t.Fatalf("rotated application key = %x, want %x", rotatedValidator.PubKey, replacement.pubKey)
	}
	if rotatedValidator.Stake.String() != oldValidator.Stake.String() || rotatedValidator.Power != oldValidator.Power ||
		rotatedValidator.Jailed != oldValidator.Jailed || !equalStrings(rotatedValidator.Domains, oldValidator.Domains) {
		t.Fatalf("rotation changed validator claims:\n got %#v\nwant %#v", rotatedValidator, oldValidator)
	}

	failedHeight, failureLog := runSmokeTxWithExpectedCode(t, ctx, binary, validators[0], &operators[targetIndex], chainID, anyDeliverTxFailure,
		"rotate-validator-key", hex.EncodeToString(replacement.pubKey), hex.EncodeToString(oldPubKey))
	if failedHeight < rotationHeight+2 {
		t.Fatalf("revoked-key rejection height = %d, want at least %d", failedHeight, rotationHeight+2)
	}
	if !strings.Contains(strings.ToLower(failureLog), "revoked") {
		t.Fatalf("rotate-back failure did not identify revocation: %s", failureLog)
	}

	liveNodes := append(append([]*smokeValidator{}, validators[:targetIndex]...), replacement)
	convergenceHeight := smokeHeight(t, validators[0]) + 1
	waitForSmokeHeight(t, liveNodes, convergenceHeight, 90*time.Second)
	assertCommonAppHash(t, liveNodes, convergenceHeight)
	for _, node := range liveNodes {
		if err := node.stop(true); err != nil {
			t.Fatalf("stop %s after key rotation: %v", node.name, err)
		}
	}

	exported := exportSmokeGenesis(t, ctx, binary, replacement, convergenceHeight)
	exportApp := newGenesisTestApp(t)
	if err := validateLedgerGenesis(exportApp.appCodec, exported.AppState); err != nil {
		t.Fatalf("key-rotation export is not exactly bank-backed: %v", err)
	}
	var democracyGenesis truedemocracy.GenesisState
	if err := json.Unmarshal(exported.AppState[truedemocracy.ModuleName], &democracyGenesis); err != nil {
		t.Fatalf("decode key-rotation export: %v", err)
	}
	assertExportedRotation(t, democracyGenesis, oldOperator, oldPubKey, replacement.pubKey, oldValidator)

	importApp := newGenesisTestApp(t)
	if err := initGenesisApp(importApp, exported.AppState); err != nil {
		t.Fatalf("re-import key-rotation export: %v", err)
	}

	reuseState := cloneSmokeAppState(t, exported.AppState)
	for i := range democracyGenesis.Validators {
		if democracyGenesis.Validators[i].OperatorAddr == oldOperator {
			democracyGenesis.Validators[i].PubKey = append([]byte(nil), oldPubKey...)
		}
	}
	reuseJSON, err := json.Marshal(democracyGenesis)
	if err != nil {
		t.Fatal(err)
	}
	reuseState[truedemocracy.ModuleName] = reuseJSON
	reuseApp := newGenesisTestApp(t)
	if err := initGenesisApp(reuseApp, reuseState); err == nil {
		t.Fatal("re-import accepted a permanently revoked consensus key")
	}
}

func runSmokeTxWithExpectedCode(t *testing.T, ctx context.Context, binary string, node *smokeValidator, from *smokeAccount, chainID string, wantCode uint32, args ...string) (int64, string) {
	t.Helper()
	commandArgs := append([]string{"tx", truedemocracy.ModuleName}, args...)
	commandArgs = append(commandArgs,
		"--home", node.home,
		"--keyring-dir", from.keyringDir,
		"--keyring-backend", "test",
		"--from", from.name,
		"--chain-id", chainID,
		"--node", fmt.Sprintf("tcp://127.0.0.1:%d", node.rpcPort),
		"--offline",
		"--account-number", strconv.FormatUint(from.accountNumber, 10),
		"--sequence", strconv.FormatUint(from.sequence, 10),
		"--broadcast-mode", "sync",
		"--fees", "0"+token.BaseDenom,
		"--gas", "500000",
		"--yes",
		"--output", "json",
	)
	output, err := exec.CommandContext(ctx, binary, commandArgs...).CombinedOutput()
	if err != nil {
		t.Fatalf("broadcast tx %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	var broadcast struct {
		Code   uint32 `json:"code"`
		RawLog string `json:"raw_log"`
		TxHash string `json:"txhash"`
	}
	if err := json.Unmarshal(output, &broadcast); err != nil {
		t.Fatalf("decode tx %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	if broadcast.Code != 0 || broadcast.TxHash == "" {
		t.Fatalf("CheckTx %s failed with code %d: %s", strings.Join(args, " "), broadcast.Code, broadcast.RawLog)
	}

	deadline := time.Now().Add(45 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		code, heightText, rawLog, queryErr := querySmokeTx(ctx, node, broadcast.TxHash)
		if queryErr == nil {
			height, parseErr := strconv.ParseInt(heightText, 10, 64)
			if parseErr != nil {
				t.Fatalf("parse tx %s height %q: %v", strings.Join(args, " "), heightText, parseErr)
			}
			// DeliverTx consumes the sequence on both success and message failure.
			from.sequence++
			if wantCode == anyDeliverTxFailure && code == 0 {
				t.Fatalf("DeliverTx %s succeeded, want an on-chain failure", strings.Join(args, " "))
			}
			if wantCode != anyDeliverTxFailure && code != wantCode {
				t.Fatalf("DeliverTx %s code = %d, want %d: %s", strings.Join(args, " "), code, wantCode, rawLog)
			}
			return height, rawLog
		}
		lastErr = queryErr
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("tx %s (%s) was not indexed within 45s: %v", strings.Join(args, " "), broadcast.TxHash, lastErr)
	return 0, ""
}

func querySmokeValidatorPowerAtHeight(ctx context.Context, node *smokeValidator, pubKey []byte, height int64) (string, bool, error) {
	url := fmt.Sprintf("http://127.0.0.1:%d/validators", node.rpcPort)
	if height > 0 {
		url += "?height=" + strconv.FormatInt(height, 10)
	}
	var response struct {
		Result struct {
			Validators []struct {
				PubKey struct {
					Value string `json:"value"`
				} `json:"pub_key"`
				VotingPower string `json:"voting_power"`
			} `json:"validators"`
		} `json:"result"`
		Error *struct {
			Code int    `json:"code"`
			Data string `json:"data"`
		} `json:"error"`
	}
	if err := getSmokeRPCJSON(ctx, url, &response); err != nil {
		return "", false, err
	}
	if response.Error != nil {
		return "", false, fmt.Errorf("validator query error %d: %s", response.Error.Code, response.Error.Data)
	}
	for _, candidate := range response.Result.Validators {
		decoded, err := decodeCometPubKey(candidate.PubKey.Value)
		if err == nil && bytes.Equal(decoded, pubKey) {
			return candidate.VotingPower, true, nil
		}
	}
	return "", false, nil
}

func assertRotationValidatorUpdates(t *testing.T, ctx context.Context, node *smokeValidator, height int64, oldPubKey, newPubKey []byte, wantPower string) {
	t.Helper()
	var response struct {
		Result struct {
			ValidatorUpdates []json.RawMessage `json:"validator_updates"`
		} `json:"result"`
		Error *struct {
			Code int    `json:"code"`
			Data string `json:"data"`
		} `json:"error"`
	}
	url := fmt.Sprintf("http://127.0.0.1:%d/block_results?height=%d", node.rpcPort, height)
	if err := getSmokeRPCJSON(ctx, url, &response); err != nil {
		t.Fatalf("query block results at height %d: %v", height, err)
	}
	if response.Error != nil {
		t.Fatalf("block results at height %d returned error %d: %s", height, response.Error.Code, response.Error.Data)
	}
	want := map[string]string{hex.EncodeToString(oldPubKey): "0", hex.EncodeToString(newPubKey): wantPower}
	got := make(map[string]string)
	for _, raw := range response.Result.ValidatorUpdates {
		key, power, err := decodeValidatorUpdate(raw)
		if err != nil {
			t.Fatalf("decode validator update at height %d: %v (%s)", height, err, raw)
		}
		got[hex.EncodeToString(key)] = power
	}
	for key, power := range want {
		if got[key] != power {
			t.Fatalf("validator update %s at height %d = %q, want %q (all updates: %v)", key, height, got[key], power, got)
		}
	}
}

func decodeValidatorUpdate(raw json.RawMessage) ([]byte, string, error) {
	var update struct {
		PubKey json.RawMessage `json:"pub_key"`
		Power  json.RawMessage `json:"power"`
	}
	if err := json.Unmarshal(raw, &update); err != nil {
		return nil, "", err
	}
	power := "0" // protobuf JSON omits the zero-valued removal power
	if len(update.Power) > 0 && string(update.Power) != "null" {
		if err := json.Unmarshal(update.Power, &power); err == nil {
			// Power was encoded as a JSON string.
		} else {
			var numeric json.Number
			if err := json.Unmarshal(update.Power, &numeric); err != nil {
				return nil, "", fmt.Errorf("decode power: %w", err)
			}
			power = numeric.String()
		}
	}
	var pubKeyObject any
	if err := json.Unmarshal(update.PubKey, &pubKeyObject); err != nil {
		return nil, "", err
	}
	key := findEncodedEd25519Key(pubKeyObject)
	if len(key) != 32 {
		return nil, "", errors.New("validator update did not contain a 32-byte ed25519 public key")
	}
	return key, power, nil
}

func findEncodedEd25519Key(value any) []byte {
	switch typed := value.(type) {
	case string:
		if decoded, err := base64.StdEncoding.DecodeString(typed); err == nil && len(decoded) == 32 {
			return decoded
		}
		if decoded, err := hex.DecodeString(typed); err == nil && len(decoded) == 32 {
			return decoded
		}
	case []any:
		for _, item := range typed {
			if key := findEncodedEd25519Key(item); len(key) == 32 {
				return key
			}
		}
	case map[string]any:
		for _, item := range typed {
			if key := findEncodedEd25519Key(item); len(key) == 32 {
				return key
			}
		}
	}
	return nil
}

func assertCommitSignedBy(t *testing.T, ctx context.Context, node *smokeValidator, height int64, pubKey []byte) {
	t.Helper()
	digest := sha256.Sum256(pubKey)
	wantAddress := strings.ToUpper(hex.EncodeToString(digest[:20]))
	var response struct {
		Result struct {
			SignedHeader struct {
				Commit struct {
					Signatures []struct {
						ValidatorAddress string `json:"validator_address"`
						Signature        string `json:"signature"`
					} `json:"signatures"`
				} `json:"commit"`
			} `json:"signed_header"`
		} `json:"result"`
	}
	url := fmt.Sprintf("http://127.0.0.1:%d/commit?height=%d", node.rpcPort, height)
	if err := getSmokeRPCJSON(ctx, url, &response); err != nil {
		t.Fatalf("query commit %d: %v", height, err)
	}
	for _, signature := range response.Result.SignedHeader.Commit.Signatures {
		if strings.EqualFold(signature.ValidatorAddress, wantAddress) && signature.Signature != "" {
			return
		}
	}
	t.Fatalf("commit %d was not signed by rotated validator address %s", height, wantAddress)
}

func querySmokeApplicationValidator(t *testing.T, ctx context.Context, binary string, node *smokeValidator, operator string) truedemocracy.Validator {
	t.Helper()
	command := exec.CommandContext(ctx, binary, "query", truedemocracy.ModuleName, "validator", operator,
		"--home", node.home,
		"--node", fmt.Sprintf("tcp://127.0.0.1:%d", node.rpcPort),
		"--output", "json",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("query application validator %s: %v\n%s", operator, err, output)
	}
	var validator truedemocracy.Validator
	if err := json.Unmarshal(output, &validator); err != nil {
		t.Fatalf("decode application validator %s: %v\n%s", operator, err, output)
	}
	return validator
}

func assertExportedRotation(t *testing.T, genesis truedemocracy.GenesisState, operator string, oldPubKey, newPubKey []byte, before truedemocracy.Validator) {
	t.Helper()
	var foundValidator bool
	for _, validator := range genesis.Validators {
		if validator.OperatorAddr != operator {
			continue
		}
		foundValidator = true
		if !bytes.Equal(validator.PubKey, newPubKey) || validator.Stake != before.Stake.AmountOf(token.BaseDenom).Int64() || validator.Domain != before.Domains[0] {
			t.Fatalf("exported rotated validator lost claims: %#v", validator)
		}
	}
	if !foundValidator {
		t.Fatalf("export omitted rotated operator %s", operator)
	}
	for _, record := range genesis.RevokedValidatorKeys {
		if record.OperatorAddr == operator && bytes.Equal(record.PubKey, oldPubKey) {
			return
		}
	}
	t.Fatalf("export omitted old-key revocation for operator %s and key %x", operator, oldPubKey)
}

func cloneSmokeAppState(t *testing.T, state map[string]json.RawMessage) map[string]json.RawMessage {
	t.Helper()
	clone := make(map[string]json.RawMessage, len(state))
	for key, value := range state {
		clone[key] = append(json.RawMessage(nil), value...)
	}
	return clone
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func getSmokeRPCJSON(ctx context.Context, url string, target any) error {
	client := &http.Client{Timeout: 2 * time.Second}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("RPC response = %s", response.Status)
	}
	return json.NewDecoder(response.Body).Decode(target)
}
