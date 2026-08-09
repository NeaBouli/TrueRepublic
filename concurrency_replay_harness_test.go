package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signing "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"truerepublic/token"
	"truerepublic/x/truedemocracy"
)

// concurrencyReplaySmokeEnv gates the GH-172 shared-state concurrency, exact
// transaction replay, and same-home restart persistence process harness. It is
// intentionally separate from the multi-validator gate so this bounded
// single-validator drill stays opt-in and normal `go test ./...` skips it
// before any setup runs.
const concurrencyReplaySmokeEnv = "TRUEREPUBLIC_CONCURRENCY_REPLAY_SMOKE"

// TestConcurrentSharedStateReplayRestart proves on a real temporary localhost
// chain that:
//  1. concurrent authenticated transactions from distinct accounts contending
//     for the same canonical domain identity produce exactly one committed
//     mutation and exactly one escrow effect, with the loser failing
//     explicitly;
//  2. broadcasting the exact same signed transaction bytes concurrently,
//     repeatedly after commit, and again after a same-home restart never
//     produces a second state transition;
//  3. the committed governance state, balances, module custody, ledger
//     invariants, and historical app hash survive a deterministic same-home
//     restart with no duplicate object.
func TestConcurrentSharedStateReplayRestart(t *testing.T) {
	if os.Getenv(concurrencyReplaySmokeEnv) != "1" {
		t.Skipf("set %s=1 to run the concurrency/replay/restart process harness", concurrencyReplaySmokeEnv)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Minute)
	defer cancel()

	binary := filepath.Join(t.TempDir(), "truerepublicd")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build daemon: %v\n%s", err, output)
	}

	// The in-process app instance supplies the exact codec, interface registry,
	// and custom signer resolution the node itself uses, so the offline-signed
	// transaction bytes are canonical.
	app := newGenesisTestApp(t)

	const chainID = "truerepublic-concurrency-replay-1"
	validator := &smokeValidator{
		name:    "validator-1",
		home:    filepath.Join(t.TempDir(), "node-1"),
		rpcPort: freeTCPPort(t),
		p2pPort: freeTCPPort(t),
		logPath: filepath.Join(t.TempDir(), "validator-1.log"),
	}
	initSmokeValidator(t, ctx, binary, chainID, validator)
	validators := []*smokeValidator{validator}

	// One validator means the bootstrap operator owns auth account 0, so the
	// disposable accounts are numbered from 1 (matching the existing harness
	// convention of len(validators)+index).
	alpha := addSmokeKey(t, ctx, binary, validator.home, "contender-alpha", 1, 1_000_000*token.WholeTokenBaseUnits)
	beta := addSmokeKey(t, ctx, binary, validator.home, "contender-beta", 2, 1_000_000*token.WholeTokenBaseUnits)

	// The replay proposer key lives in a disposable in-memory test keyring so
	// the exact signed bytes can be produced in-process; nothing is persisted
	// or logged.
	proposerKeyring := keyring.NewInMemory(app.appCodec)
	proposerRecord, _, err := proposerKeyring.NewMnemonic(
		"replay-proposer", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	if err != nil {
		t.Fatalf("generate disposable proposer key: %v", err)
	}
	proposerAddress, err := proposerRecord.GetAddress()
	if err != nil {
		t.Fatalf("derive disposable proposer address: %v", err)
	}
	proposer := smokeAccount{
		name:          "replay-proposer",
		address:       proposerAddress.String(),
		balance:       1_000_000 * token.WholeTokenBaseUnits,
		accountNumber: 3,
	}

	sharedGenesis := buildSharedSmokeGenesis(t, chainID, validators, alpha, beta, proposer)
	if err := atomicWriteFile(filepath.Join(validator.home, "config", "genesis.json"), sharedGenesis, 0o600); err != nil {
		t.Fatalf("write %s shared genesis: %v", validator.name, err)
	}

	t.Cleanup(func() {
		_ = validator.stop(false)
		if t.Failed() {
			validator.logContents(t)
		}
	})

	if err := validator.start(ctx, binary, persistentPeers(validator, validators)); err != nil {
		t.Fatalf("start %s: %v", validator.name, err)
	}
	waitForSmokeHeight(t, validators, 2, 90*time.Second)

	const (
		contentionEscrow = 500_000 * token.WholeTokenBaseUnits
		replayEscrow     = 250_000 * token.WholeTokenBaseUnits
		funded           = 1_000_000 * token.WholeTokenBaseUnits
	)
	moduleAddress := sdk.MustBech32ifyAddressBytes("truerepublic", authtypes.NewModuleAddress(truedemocracy.ModuleName))

	// Phase 1: two distinct authenticated accounts concurrently create the
	// same canonical domain identity. Both pass CheckTx, but only one can
	// commit; the other must fail explicitly with the duplicate-domain error.
	contenders := []*smokeAccount{&alpha, &beta}
	submissions := make([]contentionTxSubmission, len(contenders))
	var submissionGroup sync.WaitGroup
	for index := range contenders {
		submissionGroup.Add(1)
		go func(i int) {
			defer submissionGroup.Done()
			submissions[i] = submitContentionTx(ctx, binary, validator, contenders[i], chainID,
				"create-domain", "Contended", fmt.Sprintf("%d%s", contentionEscrow, token.BaseDenom))
		}(index)
	}
	submissionGroup.Wait()
	deliveries := make([]contentionTxDelivery, len(submissions))
	for index := range submissions {
		if submissions[index].err != nil {
			t.Fatalf("contention broadcast for %s failed before delivery: %v", contenders[index].name, submissions[index].err)
		}
		// Keep fatal polling on the test goroutine. The submissions above are
		// the concurrent behavior under test; indexing both hashes sequentially
		// avoids calling testing.T.FailNow from a helper goroutine.
		deliveries[index].code, deliveries[index].rawLog = waitContentionTxIndexed(t, ctx, validator, submissions[index].hash)
	}

	winnerIndex := -1
	for index, delivery := range deliveries {
		if delivery.code == 0 {
			if winnerIndex >= 0 {
				t.Fatalf("both contending create-domain transactions committed; want exactly one canonical mutation")
			}
			winnerIndex = index
			continue
		}
		if !strings.Contains(delivery.rawLog, "already exists") {
			t.Fatalf("losing contention transaction failed for an unexpected reason (code %d): %s", delivery.code, delivery.rawLog)
		}
	}
	if winnerIndex < 0 {
		t.Fatalf("no contending create-domain transaction committed: %+v", deliveries)
	}
	winner := contenders[winnerIndex]
	loser := contenders[1-winnerIndex]
	t.Logf("contention winner=%s loser=%s loser_code=%d", winner.name, loser.name, deliveries[1-winnerIndex].code)

	// Exactly one canonical mutation and exactly one economic effect.
	contended := querySmokeDomain(t, ctx, binary, validator, "Contended")
	if contended.Admin.String() != winner.address {
		t.Fatalf("Contended domain admin = %s, want contention winner %s", contended.Admin.String(), winner.address)
	}
	if got := contended.Treasury.AmountOf(token.BaseDenom).Int64(); got != contentionEscrow {
		t.Fatalf("Contended treasury = %d, want exactly one escrow of %d", got, contentionEscrow)
	}
	if len(contended.Members) != 1 || contended.Members[0] != winner.address {
		t.Fatalf("Contended members = %v, want only the winner %s", contended.Members, winner.address)
	}
	if got := querySmokeBankBalance(t, ctx, app, validator, winner.address); got != funded-contentionEscrow {
		t.Fatalf("winner balance = %d, want %d (exactly one escrow charged)", got, funded-contentionEscrow)
	}
	if got := querySmokeBankBalance(t, ctx, app, validator, loser.address); got != funded {
		t.Fatalf("loser balance = %d, want unchanged %d (losing tx had no economic effect)", got, funded)
	}
	wantModuleBalance := token.StakeMinBaseUnits + contentionEscrow
	if got := querySmokeBankBalance(t, ctx, app, validator, moduleAddress); got != wantModuleBalance {
		t.Fatalf("module custody = %d, want %d (bootstrap stake + exactly one escrow)", got, wantModuleBalance)
	}
	assertSmokeDomainCount(t, ctx, binary, validator, "Contended", 1)

	// Phase 2: sign one create-domain transaction offline in-process, keep the
	// exact bytes, and broadcast them three times concurrently. Exactly one
	// broadcast may be accepted; the duplicates must be rejected explicitly,
	// and the committed hash must be unique.
	signedTxBytes, signedTxHash := buildSignedDomainTx(t, app, proposerKeyring, proposer, chainID, "ReplayDomain", replayEscrow)
	attempts := make([]replayBroadcastAttempt, 3)
	var broadcastGroup sync.WaitGroup
	for index := range attempts {
		broadcastGroup.Add(1)
		go func(i int) {
			defer broadcastGroup.Done()
			attempts[i] = broadcastSmokeTxBytes(ctx, validator, signedTxBytes, signedTxHash)
		}(index)
	}
	broadcastGroup.Wait()
	acceptedBroadcasts := 0
	for index, attempt := range attempts {
		t.Logf("concurrent duplicate broadcast %d: accepted=%t evidence=%s", index+1, attempt.accepted, attempt.evidence)
		if attempt.accepted {
			acceptedBroadcasts++
			continue
		}
		if attempt.evidence == "" {
			t.Fatalf("duplicate broadcast %d was silently ignored instead of explicitly rejected", index+1)
		}
	}
	if acceptedBroadcasts != 1 {
		t.Fatalf("accepted concurrent duplicate broadcasts = %d, want exactly 1", acceptedBroadcasts)
	}
	commitCode, commitLog := waitContentionTxIndexed(t, ctx, validator, signedTxHash)
	if commitCode != 0 {
		t.Fatalf("saved signed transaction failed at delivery with code %d: %s", commitCode, commitLog)
	}

	// Repeating the exact same bytes after the commit must also be rejected.
	postCommit := broadcastSmokeTxBytes(ctx, validator, signedTxBytes, signedTxHash)
	t.Logf("post-commit replay broadcast: accepted=%t evidence=%s", postCommit.accepted, postCommit.evidence)
	if postCommit.accepted || postCommit.evidence == "" {
		t.Fatalf("post-commit replay of the exact signed transaction was not explicitly rejected")
	}

	postCommitHeight := smokeHeight(t, validator)
	waitForSmokeHeight(t, validators, postCommitHeight+2, 90*time.Second)
	if code, _, _, err := querySmokeTx(ctx, validator, signedTxHash); err != nil || code != 0 {
		t.Fatalf("committed replay-domain transaction changed after duplicate broadcasts: code=%d err=%v", code, err)
	}

	replayDomain := querySmokeDomain(t, ctx, binary, validator, "ReplayDomain")
	if replayDomain.Admin.String() != proposer.address {
		t.Fatalf("ReplayDomain admin = %s, want proposer %s", replayDomain.Admin.String(), proposer.address)
	}
	if got := replayDomain.Treasury.AmountOf(token.BaseDenom).Int64(); got != replayEscrow {
		t.Fatalf("ReplayDomain treasury = %d, want exactly one escrow of %d", got, replayEscrow)
	}
	if got := querySmokeBankBalance(t, ctx, app, validator, proposer.address); got != funded-replayEscrow {
		t.Fatalf("proposer balance = %d, want %d (exactly one escrow despite duplicate broadcasts)", got, funded-replayEscrow)
	}
	wantModuleBalance = token.StakeMinBaseUnits + contentionEscrow + replayEscrow
	if got := querySmokeBankBalance(t, ctx, app, validator, moduleAddress); got != wantModuleBalance {
		t.Fatalf("module custody = %d, want %d after duplicate broadcasts", got, wantModuleBalance)
	}
	assertSmokeDomainCount(t, ctx, binary, validator, "ReplayDomain", 1)

	preStopHeight := smokeHeight(t, validator)
	preStopAppHash := smokeAppHash(t, validator, preStopHeight)
	if preStopAppHash == "" {
		t.Fatalf("empty app hash at pre-stop height %d", preStopHeight)
	}

	// Phase 3: deterministic same-home restart. The exact saved transaction
	// must stay rejected, the historical app hash must be identical, and no
	// duplicate object may appear.
	if err := validator.stop(true); err != nil {
		t.Fatalf("stop %s for same-home restart: %v", validator.name, err)
	}
	if err := validator.start(ctx, binary, persistentPeers(validator, validators)); err != nil {
		t.Fatalf("restart %s from the same home: %v", validator.name, err)
	}
	waitForSmokeHeight(t, validators, preStopHeight+2, 90*time.Second)
	if got := smokeAppHash(t, validator, preStopHeight); got != preStopAppHash {
		t.Fatalf("app hash at height %d changed across same-home restart: %s != %s", preStopHeight, got, preStopAppHash)
	}

	postRestart := broadcastSmokeTxBytes(ctx, validator, signedTxBytes, signedTxHash)
	t.Logf("post-restart replay broadcast: accepted=%t evidence=%s", postRestart.accepted, postRestart.evidence)
	if postRestart.accepted || postRestart.evidence == "" {
		t.Fatalf("post-restart replay of the exact signed transaction was not explicitly rejected")
	}
	if !strings.Contains(postRestart.evidence, "code 32") ||
		!strings.Contains(postRestart.evidence, "account sequence mismatch") {
		t.Fatalf("post-restart replay was not rejected by the expected sequence guard: %s", postRestart.evidence)
	}

	restartHeight := smokeHeight(t, validator)
	waitForSmokeHeight(t, validators, restartHeight+2, 90*time.Second)
	if code, _, _, err := querySmokeTx(ctx, validator, signedTxHash); err != nil || code != 0 {
		t.Fatalf("original committed transaction is not intact after restart and replay attempt: code=%d err=%v", code, err)
	}
	replayedDomain := querySmokeDomain(t, ctx, binary, validator, "ReplayDomain")
	if got := replayedDomain.Treasury.AmountOf(token.BaseDenom).Int64(); got != replayEscrow {
		t.Fatalf("ReplayDomain treasury after restart = %d, want unchanged %d", got, replayEscrow)
	}
	if got := querySmokeBankBalance(t, ctx, app, validator, proposer.address); got != funded-replayEscrow {
		t.Fatalf("proposer balance after restart = %d, want unchanged %d (replay had no economic effect)", got, funded-replayEscrow)
	}
	if got := querySmokeBankBalance(t, ctx, app, validator, moduleAddress); got != wantModuleBalance {
		t.Fatalf("module custody after restart = %d, want unchanged %d", got, wantModuleBalance)
	}
	assertSmokeDomainCount(t, ctx, binary, validator, "Contended", 1)
	assertSmokeDomainCount(t, ctx, binary, validator, "ReplayDomain", 1)

	// Phase 4: stop cleanly, then prove the persisted ledger exports exactly
	// one canonical object per identity, stays bank-backed under the
	// repository invariant reconciliation, and re-imports without drift.
	finalHeight := smokeHeight(t, validator)
	if err := validator.stop(true); err != nil {
		t.Fatalf("stop %s after replay/restart drill: %v", validator.name, err)
	}
	exported := exportSmokeGenesis(t, ctx, binary, validator, preStopHeight)
	exportApp := newGenesisTestApp(t)
	if err := validateLedgerGenesis(exportApp.appCodec, exported.AppState); err != nil {
		t.Fatalf("post-restart export is not exactly bank-backed: %v", err)
	}
	var democracyGenesis truedemocracy.GenesisState
	if err := json.Unmarshal(exported.AppState[truedemocracy.ModuleName], &democracyGenesis); err != nil {
		t.Fatalf("decode post-restart PoD genesis: %v", err)
	}
	domainCounts := make(map[string]int)
	for _, domain := range democracyGenesis.Domains {
		domainCounts[domain.Name]++
	}
	for _, name := range []string{"Bootstrap", "Contended", "ReplayDomain"} {
		if domainCounts[name] != 1 {
			t.Fatalf("export contains %d copies of domain %s, want exactly 1 (all domains: %v)", domainCounts[name], name, domainCounts)
		}
	}
	if len(democracyGenesis.Domains) != 3 {
		t.Fatalf("export contains %d domains, want exactly 3: %v", len(democracyGenesis.Domains), domainCounts)
	}
	var bankGenesis banktypes.GenesisState
	if err := exportApp.appCodec.UnmarshalJSON(exported.AppState[banktypes.ModuleName], &bankGenesis); err != nil {
		t.Fatalf("decode post-restart bank genesis: %v", err)
	}
	if supply := token.CanonicalSupply(bankGenesis); supply.GT(token.MaxSupply()) {
		t.Fatalf("post-restart supply %s exceeds the 21M PNYX cap", supply)
	}
	importApp := newGenesisTestApp(t)
	if err := initGenesisApp(importApp, exported.AppState); err != nil {
		t.Fatalf("re-import post-restart export: %v", err)
	}
	t.Logf("final committed height=%d pre-stop app hash=%s replay tx=%s", finalHeight, preStopAppHash, signedTxHash)
}

type contentionTxSubmission struct {
	hash string
	err  error
}

type contentionTxDelivery struct {
	code   uint32
	rawLog string
}

// submitContentionTx signs and broadcasts one transaction like runSmokeTx but
// reports failures instead of stopping the test, so concurrent contention can
// be asserted explicitly.
func submitContentionTx(ctx context.Context, binary string, node *smokeValidator, from *smokeAccount, chainID string, args ...string) contentionTxSubmission {
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
	command := exec.CommandContext(ctx, binary, commandArgs...)
	output, err := command.CombinedOutput()
	if err != nil {
		return contentionTxSubmission{err: fmt.Errorf("tx %s broadcast command failed: %v: %s", strings.Join(args, " "), err, output)}
	}
	var result struct {
		Code   uint32 `json:"code"`
		RawLog string `json:"raw_log"`
		TxHash string `json:"txhash"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return contentionTxSubmission{err: fmt.Errorf("decode tx %s broadcast response: %v: %s", strings.Join(args, " "), err, output)}
	}
	if result.Code != 0 {
		return contentionTxSubmission{err: fmt.Errorf("tx %s rejected at CheckTx with code %d: %s", strings.Join(args, " "), result.Code, result.RawLog)}
	}
	if result.TxHash == "" {
		return contentionTxSubmission{err: fmt.Errorf("tx %s returned an empty hash: %s", strings.Join(args, " "), output)}
	}
	return contentionTxSubmission{hash: result.TxHash}
}

// waitContentionTxIndexed polls CometBFT until the transaction is indexed and
// returns its delivered code and log so both success and expected failure can
// be asserted.
func waitContentionTxIndexed(t *testing.T, ctx context.Context, node *smokeValidator, txHash string) (uint32, string) {
	t.Helper()
	deadline := time.Now().Add(45 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		code, _, rawLog, err := querySmokeTx(ctx, node, txHash)
		if err == nil {
			return code, rawLog
		}
		lastErr = err
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("tx %s was not indexed within 45s: %v", txHash, lastErr)
	return 0, ""
}

// buildSignedDomainTx produces one canonical offline-signed create-domain
// transaction with the same codec and signer resolution the node uses, and
// returns the exact bytes every replay attempt must reuse byte-for-byte plus
// their CometBFT hash.
func buildSignedDomainTx(t *testing.T, app *TrueRepublicApp, keyringBase keyring.Keyring, from smokeAccount, chainID, domain string, escrow int64) ([]byte, string) {
	t.Helper()
	record, err := keyringBase.Key(from.name)
	if err != nil {
		t.Fatalf("load disposable proposer key: %v", err)
	}
	pubKey, err := record.GetPubKey()
	if err != nil {
		t.Fatalf("derive disposable proposer public key: %v", err)
	}
	address, err := record.GetAddress()
	if err != nil {
		t.Fatalf("derive disposable proposer address: %v", err)
	}
	if address.String() != from.address {
		t.Fatalf("proposer keyring address %s does not match genesis account %s", address.String(), from.address)
	}
	msg := &truedemocracy.MsgCreateDomain{
		Name:         domain,
		Admin:        address,
		InitialCoins: sdk.NewCoins(token.NewCoin(math.NewInt(escrow))),
	}
	if err := msg.ValidateBasic(); err != nil {
		t.Fatalf("offline create-domain message is invalid: %v", err)
	}
	builder := app.txConfig.NewTxBuilder()
	if err := builder.SetMsgs(msg); err != nil {
		t.Fatalf("build offline create-domain transaction: %v", err)
	}
	builder.SetGasLimit(500000)
	builder.SetFeeAmount(sdk.NewCoins(token.NewCoin(math.ZeroInt())))
	signature := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode: signing.SignMode_SIGN_MODE_DIRECT,
		},
		Sequence: from.sequence,
	}
	if err := builder.SetSignatures(signature); err != nil {
		t.Fatalf("attach unsigned signature envelope: %v", err)
	}
	signerData := authsigning.SignerData{
		ChainID:       chainID,
		AccountNumber: from.accountNumber,
		Sequence:      from.sequence,
	}
	signBytes, err := authsigning.GetSignBytesAdapter(
		context.Background(), app.txConfig.SignModeHandler(), signing.SignMode_SIGN_MODE_DIRECT, signerData, builder.GetTx())
	if err != nil {
		t.Fatalf("derive offline sign bytes: %v", err)
	}
	signed, _, err := keyringBase.Sign(from.name, signBytes, signing.SignMode_SIGN_MODE_DIRECT)
	if err != nil {
		t.Fatalf("sign offline create-domain transaction: %v", err)
	}
	signature.Data = &signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: signed,
	}
	if err := builder.SetSignatures(signature); err != nil {
		t.Fatalf("attach offline signature: %v", err)
	}
	txBytes, err := app.txConfig.TxEncoder()(builder.GetTx())
	if err != nil {
		t.Fatalf("encode offline create-domain transaction: %v", err)
	}
	if _, err := app.txConfig.TxDecoder()(txBytes); err != nil {
		t.Fatalf("offline transaction does not round-trip through the canonical decoder: %v", err)
	}
	digest := sha256.Sum256(txBytes)
	return txBytes, strings.ToUpper(hex.EncodeToString(digest[:]))
}

type replayBroadcastAttempt struct {
	accepted bool
	evidence string
}

// broadcastSmokeTxBytes submits the exact signed transaction bytes to
// CometBFT over the bounded localhost RPC and classifies the result.
// Acceptance requires a JSON-RPC success, a code-0 CheckTx, and the expected
// hash; every rejection must carry explicit evidence.
func broadcastSmokeTxBytes(ctx context.Context, node *smokeValidator, txBytes []byte, wantHash string) replayBroadcastAttempt {
	endpoint := fmt.Sprintf("http://127.0.0.1:%d/broadcast_tx_sync?tx=0x%s", node.rpcPort, strings.ToUpper(hex.EncodeToString(txBytes)))
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return replayBroadcastAttempt{evidence: fmt.Sprintf("build broadcast request: %v", err)}
	}
	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return replayBroadcastAttempt{evidence: fmt.Sprintf("broadcast RPC transport: %v", err)}
	}
	defer response.Body.Close()
	var result struct {
		Result *struct {
			Code uint32 `json:"code"`
			Log  string `json:"log"`
			Hash string `json:"hash"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data"`
		} `json:"error"`
	}
	// CometBFT maps JSON-RPC internal errors such as duplicate-tx rejections
	// to HTTP 500 while still returning the error envelope in the body, so the
	// body is decoded regardless of status to keep every rejection explicit.
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return replayBroadcastAttempt{evidence: fmt.Sprintf("decode broadcast response (status %s): %v", response.Status, err)}
	}
	if result.Error != nil {
		return replayBroadcastAttempt{
			evidence: fmt.Sprintf("broadcast rejected by the node: RPC error %d %s: %s", result.Error.Code, result.Error.Message, result.Error.Data),
		}
	}
	if result.Result == nil {
		return replayBroadcastAttempt{evidence: fmt.Sprintf("broadcast response carried neither result nor error (status %s)", response.Status)}
	}
	if result.Result.Code != 0 {
		return replayBroadcastAttempt{
			evidence: fmt.Sprintf("broadcast rejected at CheckTx with code %d: %s", result.Result.Code, result.Result.Log),
		}
	}
	if !strings.EqualFold(result.Result.Hash, wantHash) {
		return replayBroadcastAttempt{
			evidence: fmt.Sprintf("broadcast hash = %s, want the exact saved transaction hash %s", result.Result.Hash, wantHash),
		}
	}
	return replayBroadcastAttempt{accepted: true, evidence: fmt.Sprintf("accepted with hash %s", result.Result.Hash)}
}

func querySmokeDomain(t *testing.T, ctx context.Context, binary string, node *smokeValidator, name string) truedemocracy.Domain {
	t.Helper()
	command := exec.CommandContext(ctx, binary, "query", truedemocracy.ModuleName, "domain", name,
		"--home", node.home,
		"--node", fmt.Sprintf("tcp://127.0.0.1:%d", node.rpcPort),
		"--output", "json",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("query domain %s: %v\n%s", name, err, output)
	}
	var domain truedemocracy.Domain
	if err := json.Unmarshal(output, &domain); err != nil {
		t.Fatalf("decode domain %s: %v\n%s", name, err, output)
	}
	if domain.Name != name {
		t.Fatalf("query domain %s returned domain %s", name, domain.Name)
	}
	return domain
}

// assertSmokeDomainCount proves through the live query path that exactly one
// canonical object exists for the given identity.
func assertSmokeDomainCount(t *testing.T, ctx context.Context, binary string, node *smokeValidator, name string, want int) {
	t.Helper()
	command := exec.CommandContext(ctx, binary, "query", truedemocracy.ModuleName, "domains",
		"--home", node.home,
		"--node", fmt.Sprintf("tcp://127.0.0.1:%d", node.rpcPort),
		"--output", "json",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("query domains: %v\n%s", err, output)
	}
	var domains []truedemocracy.Domain
	if err := json.Unmarshal(output, &domains); err != nil {
		t.Fatalf("decode domains: %v\n%s", err, output)
	}
	count := 0
	for _, domain := range domains {
		if domain.Name == name {
			count++
		}
	}
	if count != want {
		t.Fatalf("live state contains %d copies of domain %s, want %d", count, name, want)
	}
}

// querySmokeBankBalance reads one canonical upnyx balance through the
// registered protobuf gRPC-over-CometBFT-ABCI JSON-RPC transport, using the
// same POST envelope and plain-hex encoding as the maintained client.
func querySmokeBankBalance(t *testing.T, ctx context.Context, app *TrueRepublicApp, node *smokeValidator, address string) int64 {
	t.Helper()
	requestBody, err := app.appCodec.Marshal(&banktypes.QueryBalanceRequest{Address: address, Denom: token.BaseDenom})
	if err != nil {
		t.Fatalf("encode bank balance query for %s: %v", address, err)
	}
	envelope := fmt.Sprintf(
		`{"jsonrpc":"2.0","id":1,"method":"abci_query","params":{"path":"/cosmos.bank.v1beta1.Query/Balance","data":"%s","prove":false}}`,
		hex.EncodeToString(requestBody))
	request, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("http://127.0.0.1:%d/", node.rpcPort), strings.NewReader(envelope))
	if err != nil {
		t.Fatalf("build bank balance query for %s: %v", address, err)
	}
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("bank balance query for %s: %v", address, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("bank balance query for %s: status = %s", address, response.Status)
	}
	var combined struct {
		Result *struct {
			Response struct {
				Code  uint32 `json:"code"`
				Log   string `json:"log"`
				Value string `json:"value"`
			} `json:"response"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    string `json:"data"`
		} `json:"error"`
	}
	if err := json.NewDecoder(response.Body).Decode(&combined); err != nil {
		t.Fatalf("decode bank balance query for %s: %v", address, err)
	}
	if combined.Error != nil {
		t.Fatalf("bank balance query for %s: RPC error %d %s: %s", address, combined.Error.Code, combined.Error.Message, combined.Error.Data)
	}
	if combined.Result == nil {
		t.Fatalf("bank balance query for %s: response carried neither result nor error", address)
	}
	if combined.Result.Response.Code != 0 {
		t.Fatalf("bank balance query for %s failed with code %d: %s", address, combined.Result.Response.Code, combined.Result.Response.Log)
	}
	value, err := base64.StdEncoding.DecodeString(combined.Result.Response.Value)
	if err != nil {
		t.Fatalf("decode bank balance value for %s: %v", address, err)
	}
	var balance banktypes.QueryBalanceResponse
	if err := app.appCodec.Unmarshal(value, &balance); err != nil {
		t.Fatalf("decode bank balance response for %s: %v", address, err)
	}
	if balance.Balance == nil {
		return 0
	}
	return balance.Balance.Amount.Int64()
}
