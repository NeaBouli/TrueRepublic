package observability

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	sdklog "cosmossdk.io/log"
	"github.com/rs/zerolog"
)

const (
	passwordSecret   = "correct-horse-battery-staple"
	mnemonicSecret   = "abandon ability able about above absent absorb abstract absurd abuse access accident"
	privateKeySecret = "private-key-material-123456"
	rawTxSecret      = "0A8FDEADBEEFCAFEBABE"
	signatureSecret  = "MEUCIQDX-secret-signature"
)

func TestWrappedLoggerSanitizesEveryLevelAndContext(t *testing.T) {
	var output bytes.Buffer
	base := sdklog.NewLogger(
		&output,
		sdklog.OutputJSONOption(),
		sdklog.LevelOption(zerolog.DebugLevel),
		sdklog.ColorOption(false),
	)
	logger := Wrap(base).With(
		"module", "consensus",
		"password", passwordSecret,
		"public_hash", strings.Repeat("a", 64),
	)

	logger.Info("password="+passwordSecret, "height", int64(42))
	logger.Warn("mnemonic: "+mnemonicSecret, "peer_count", 3)
	logger.Error("raw_tx="+rawTxSecret, "duration", "250ms")
	logger.Debug("signature="+signatureSecret, "public_key", "A1B2C3")

	lines := nonEmptyLines(output.String())
	if len(lines) != 4 {
		t.Fatalf("log lines = %d, want 4\n%s", len(lines), output.String())
	}
	for index, line := range lines {
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("line %d is not JSON: %v\n%s", index, err, line)
		}
		if event["password"] != Redacted {
			t.Fatalf("line %d password = %#v", index, event["password"])
		}
		if event["module"] != "consensus" || event["public_hash"] != strings.Repeat("a", 64) {
			t.Fatalf("line %d lost safe context: %#v", index, event)
		}
	}
	assertAbsent(t, output.String(), passwordSecret, mnemonicSecret, rawTxSecret, signatureSecret)
}

func TestWrappedLoggerPreservesConfiguredLevel(t *testing.T) {
	var output bytes.Buffer
	logger := Wrap(sdklog.NewLogger(
		&output,
		sdklog.OutputJSONOption(),
		sdklog.LevelOption(zerolog.WarnLevel),
		sdklog.ColorOption(false),
	))
	logger.Info("hidden info")
	logger.Warn("visible warning")
	if strings.Contains(output.String(), "hidden info") ||
		!strings.Contains(output.String(), "visible warning") {
		t.Fatalf("wrapped logger changed level filtering: %s", output.String())
	}
}

func TestWrappedLoggerRedactsNestedPrivateDataWithoutMutation(t *testing.T) {
	type nested struct {
		Height       int               `json:"height"`
		PrivateKey   string            `json:"private_key"`
		PublicKey    []byte            `json:"public_key"`
		Transactions []string          `json:"transactions"`
		Metadata     map[string]string `json:"metadata"`
	}
	publicKey := []byte{1, 2, 3, 4}
	input := nested{
		Height:       77,
		PrivateKey:   privateKeySecret,
		PublicKey:    publicKey,
		Transactions: []string{rawTxSecret},
		Metadata: map[string]string{
			"safe":      "kept",
			"api_token": passwordSecret,
		},
	}

	var output bytes.Buffer
	logger := Wrap(sdklog.NewLogger(&output, sdklog.OutputJSONOption(), sdklog.ColorOption(false)))
	logger.Info("nested", "state", input, "opaque", []byte(rawTxSecret))

	assertAbsent(t, output.String(), privateKeySecret, rawTxSecret, passwordSecret)
	if !strings.Contains(output.String(), `"height":77`) ||
		!strings.Contains(output.String(), `"safe":"kept"`) ||
		!strings.Contains(output.String(), `"public_key":"AQIDBA=="`) {
		t.Fatalf("safe nested fields were not preserved: %s", output.String())
	}
	if input.PrivateKey != privateKeySecret ||
		input.Transactions[0] != rawTxSecret ||
		input.Metadata["api_token"] != passwordSecret ||
		!bytes.Equal(input.PublicKey, publicKey) {
		t.Fatal("sanitization mutated caller-owned data")
	}
}

func TestSanitizeAdversarialText(t *testing.T) {
	privatePEM := "-----BEGIN PRIVATE KEY-----\n" + privateKeySecret + "\n-----END PRIVATE KEY-----"
	tests := map[string]string{
		"authorization":  "Authorization: Bearer abcdefghijklmnop",
		"bare bearer":    "Bearer abcdefghijklmnop",
		"bare basic":     "Basic YWxhZGRpbjpvcGVuc2VzYW1l",
		"URL userinfo":   "https://operator:credential-value@127.0.0.1",
		"password":       "password=" + passwordSecret,
		"password words": "password=correct horse battery staple",
		"mnemonic":       "mnemonic: " + mnemonicSecret,
		"private key":    "private_key: " + privateKeySecret,
		"raw tx":         "raw tx=" + rawTxSecret,
		"raw tx spaced":  "raw transaction " + rawTxSecret,
		"proof":          "proof: proof-secret",
		"signature":      "signature=" + signatureSecret,
		"private PEM":    privatePEM,
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			output := Sanitize(input)
			if !strings.Contains(output, Redacted) {
				t.Fatalf("Sanitize(%q) = %q, want marker", input, output)
			}
			if output == input {
				t.Fatalf("Sanitize(%q) did not change sensitive input", input)
			}
			assertAbsent(
				t,
				output,
				passwordSecret,
				mnemonicSecret,
				privateKeySecret,
				rawTxSecret,
				signatureSecret,
				"abcdefghijklmnop",
				"YWxhZGRpbjpvcGVuc2VzYW1l",
				"credential-value",
				"proof-secret",
				"correct horse battery staple",
			)
		})
	}

	safeHash := strings.Repeat("b", 64)
	safe := "committed block height=42 hash=" + safeHash + " duration=250ms"
	if got := Sanitize(safe); got != safe {
		t.Fatalf("safe message changed:\n got: %q\nwant: %q", got, safe)
	}
}

type cyclicNode struct {
	Name   string
	Secret string
	Next   *cyclicNode
}

type panicStringer struct{}

func (panicStringer) String() string {
	panic("stringer secret")
}

type panicError struct{}

func (panicError) Error() string {
	panic("error secret")
}

func TestWrappedLoggerMalformedValuesDoNotPanicOrLeak(t *testing.T) {
	node := &cyclicNode{Name: "root", Secret: passwordSecret}
	node.Next = node
	cyclicMap := map[string]any{}
	cyclicMap["self"] = cyclicMap
	cyclicMap["signature"] = signatureSecret

	var output bytes.Buffer
	logger := Wrap(sdklog.NewLogger(&output, sdklog.OutputJSONOption(), sdklog.ColorOption(false)))
	mustNotPanic(t, func() {
		logger.Info(
			"malformed",
			panicStringer{}, panicError{},
			"cycle", node,
			"map", cyclicMap,
			"odd",
		)
	})

	var event map[string]any
	lines := nonEmptyLines(output.String())
	if len(lines) != 1 {
		t.Fatalf("log lines = %d, want 1\n%s", len(lines), output.String())
	}
	if err := json.Unmarshal([]byte(lines[0]), &event); err != nil {
		t.Fatalf("malformed case did not emit valid JSON: %v\n%s", err, lines[0])
	}
	assertAbsent(t, output.String(), passwordSecret, signatureSecret, "stringer secret", "error secret")
}

func TestSensitiveKeyClassificationPreservesExplicitPublicFields(t *testing.T) {
	tests := []struct {
		key       string
		sensitive bool
	}{
		{key: "private_key", sensitive: true},
		{key: "validator-key", sensitive: true},
		{key: "signer_state", sensitive: true},
		{key: "tx_bytes", sensitive: true},
		{key: "proof", sensitive: true},
		{key: "signature", sensitive: true},
		{key: "authorization", sensitive: true},
		{key: "authentication", sensitive: true},
		{key: "token", sensitive: true},
		{key: "data", sensitive: true},
		{key: "transaction_hash"},
		{key: "proof_hash", sensitive: true},
		{key: "public_key"},
		{key: "pubKey"},
		{key: "height"},
		{key: "module"},
		{key: "authority"},
		{key: "token_denom"},
	}
	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			if got := isSensitiveKey(test.key); got != test.sensitive {
				t.Fatalf("isSensitiveKey(%q) = %v, want %v", test.key, got, test.sensitive)
			}
		})
	}
}

func TestSanitizerBoundsUntrustedLogValues(t *testing.T) {
	largeText := strings.Repeat("λ", maxTextBytes)
	sanitizedText := Sanitize(largeText)
	if len(sanitizedText) > maxTextBytes+len(Redacted) || !strings.HasSuffix(sanitizedText, Redacted) {
		t.Fatalf("large text was not bounded: %d bytes", len(sanitizedText))
	}

	largeSlice := make([]int, maxCollectionItems+10)
	var output bytes.Buffer
	Wrap(sdklog.NewLogger(&output, sdklog.OutputJSONOption(), sdklog.ColorOption(false))).
		Info("bounded", "items", largeSlice)
	if !strings.Contains(output.String(), Redacted) {
		t.Fatalf("large collection lacks truncation marker: %s", output.String())
	}
}

func TestImplDoesNotExposeUnderlyingLogger(t *testing.T) {
	base := sdklog.NewNopLogger()
	wrapped := Wrap(base)
	if wrapped.Impl() == base.Impl() {
		t.Fatal("Impl exposed the underlying logger and created a redaction bypass")
	}
	if _, ok := wrapped.Impl().(redactingLogger); !ok {
		t.Fatalf("Impl returned unexpected type %T", wrapped.Impl())
	}
}

func TestNilLoggerIsSafe(t *testing.T) {
	mustNotPanic(t, func() {
		Wrap(nil).Error("password="+passwordSecret, "raw_tx", rawTxSecret)
	})
}

func nonEmptyLines(input string) []string {
	var lines []string
	for _, line := range strings.Split(input, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func assertAbsent(t *testing.T, output string, secrets ...string) {
	t.Helper()
	for _, secret := range secrets {
		if strings.Contains(output, secret) {
			t.Fatalf("output leaked %q:\n%s", secret, output)
		}
	}
}

func mustNotPanic(t *testing.T, function func()) {
	t.Helper()
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("unexpected panic: %v", recovered)
		}
	}()
	function()
}

var (
	_ fmt.Stringer = panicStringer{}
	_ error        = panicError{}
)
