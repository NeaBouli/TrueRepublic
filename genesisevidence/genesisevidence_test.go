package genesisevidence

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/bech32"
)

var testOperator = func() string {
	value, err := bech32.ConvertAndEncode("truerepublic", bytes.Repeat([]byte{0x42}, 20))
	if err != nil {
		panic(err)
	}
	return value
}()

func fixture(t testing.TB) ([]byte, []byte) {
	t.Helper()
	key := base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))
	keyHash := sha256.Sum256([]byte("12345678901234567890123456789012"))
	consensusAddress := strings.ToUpper(hex.EncodeToString(keyHash[:20]))
	genesis := []byte(`{
  "genesis_time": "2026-08-23T00:00:00Z",
  "chain_id": "truerepublic-rollout-1",
  "initial_height": "1",
  "consensus_params": {},
  "consensus": {"validators": [{"address":"` + consensusAddress + `","pub_key":{"type":"tendermint/PubKeyEd25519","value":"` + key + `"},"power":"1","name":"validator-1"}]},
  "app_hash": "",
	  "app_state": {
	    "auth": {"params":{},"accounts":[
	      {"@type":"/cosmos.auth.v1beta1.BaseAccount","address":"` + testOperator + `","pub_key":null,"account_number":"0","sequence":"0"},
	      {"@type":"/cosmos.auth.v1beta1.ModuleAccount","base_account":{"address":"` + trueDemocracyAddress + `"},"name":"truedemocracy","permissions":["minter","burner"]},
      {"@type":"/cosmos.auth.v1beta1.ModuleAccount","base_account":{"address":"` + dexAddress + `"},"name":"dex","permissions":["burner"]},
      {"@type":"/cosmos.auth.v1beta1.ModuleAccount","base_account":{"address":"` + feeCollectorAddress + `"},"name":"fee_collector","permissions":[]},
      {"@type":"/cosmos.auth.v1beta1.ModuleAccount","base_account":{"address":"` + wasmAddress + `"},"name":"wasm","permissions":["burner"]},
      {"@type":"/cosmos.auth.v1beta1.ModuleAccount","base_account":{"address":"` + transferAddress + `"},"name":"transfer","permissions":["minter","burner"]}
    ]},
    "bank": {"params":{},"balances":[{"address":"` + trueDemocracyAddress + `","coins":[{"denom":"upnyx","amount":"100000000000"}]}],"supply":[{"denom":"upnyx","amount":"100000000000"}],"denom_metadata":[],"send_enabled":[]},
    "truedemocracy": {"domains":[{"name":"Bootstrap","admin":"` + testOperator + `","members":["` + testOperator + `"],"treasury":[]}],"validators":[{"operator_addr":"` + testOperator + `","pub_key":"` + key + `","stake":100000000000,"domain":"Bootstrap"}],"revoked_validator_keys":[],"pending_validator_rotations":[],"pending_validator_removals":[],"used_nullifiers":[]},
    "dex": {"pools":[],"registered_assets":[],"lp_positions":[]}
  }
}`)
	digest := sha256.Sum256(genesis)
	manifest := Manifest{
		Schema: ManifestSchema, SourceCommit: strings.Repeat("a", 40), DaemonVersion: strings.Repeat("a", 40),
		ChainID: "truerepublic-rollout-1", GenesisSHA256: hex.EncodeToString(digest[:]), MaxValidatorPower: 4,
		Validators:       []Validator{{OperatorAddress: testOperator, ConsensusPubKey: key, StakeUPNYX: "100000000000", Power: 1}},
		Allocations:      []Allocation{{Address: trueDemocracyAddress, Coins: []Coin{{Denom: "upnyx", Amount: "100000000000"}}}},
		TotalSupplyUPNYX: "100000000000", GovernanceEscrowUPNYX: "100000000000", DEXCustody: []Coin{},
	}
	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return manifestJSON, genesis
}

func TestVerifyAcceptsPrettyExactGenesis(t *testing.T) {
	manifest, genesis := fixture(t)
	evidence := Verify(manifest, genesis)
	if !evidence.Valid {
		t.Fatalf("evidence = %+v", evidence)
	}
	if len(evidence.Checks) != len(checkNames) {
		t.Fatalf("checks = %d", len(evidence.Checks))
	}
	first, _ := MarshalJSON(evidence)
	second, _ := MarshalJSON(Verify(manifest, genesis))
	if string(first) != string(second) {
		t.Fatal("evidence is not stable")
	}
}

func TestVerifyRejectsAdversarialInputs(t *testing.T) {
	tests := []struct {
		name   string
		mutate func([]byte, []byte) ([]byte, []byte)
		check  string
	}{
		{"raw binding", func(m, g []byte) ([]byte, []byte) { return m, append(g, '\n') }, "genesis-binding"},
		{"duplicate key", func(m, g []byte) ([]byte, []byte) {
			return m, []byte(strings.Replace(string(g), `"chain_id": "truerepublic-rollout-1",`, `"chain_id":"other","chain_id": "truerepublic-rollout-1",`, 1))
		}, "genesis-binding"},
		{"trailing data", func(m, g []byte) ([]byte, []byte) { return m, append(g, []byte(` {}`)...) }, "genesis-binding"},
		{"unknown manifest", func(m, g []byte) ([]byte, []byte) {
			return []byte(strings.Replace(string(m), `"schema":`, `"unknown":true,"schema":`, 1)), g
		}, "manifest"},
		{"chain mismatch", func(m, g []byte) ([]byte, []byte) {
			g = []byte(strings.Replace(string(g), "truerepublic-rollout-1", "truerepublic-rollout-2", 1))
			return bind(t, m, g), g
		}, "chain-identity"},
		{"consensus power", func(m, g []byte) ([]byte, []byte) {
			g = []byte(strings.Replace(string(g), `"power":"1"`, `"power":"2"`, 1))
			return bind(t, m, g), g
		}, "consensus-validators"},
		{"app key duplicate", func(m, g []byte) ([]byte, []byte) {
			needle := `"validators":[{"operator_addr"`
			g = []byte(strings.Replace(string(g), needle, `"validators":[{"operator_addr":"`+testOperator+`","pub_key":"MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=","stake":100000000000,"domain":"Bootstrap"},{"operator_addr"`, 1))
			return bind(t, m, g), g
		}, "application-validators"},
		{"supply mismatch", func(m, g []byte) ([]byte, []byte) {
			g = []byte(strings.Replace(string(g), `"supply":[{"denom":"upnyx","amount":"100000000000"}]`, `"supply":[{"denom":"upnyx","amount":"100000000001"}]`, 1))
			return bind(t, m, g), g
		}, "bank-supply"},
		{"escrow mismatch", func(m, g []byte) ([]byte, []byte) {
			g = []byte(strings.Replace(string(g), `"stake":100000000000`, `"stake":200000000000`, 1))
			return bind(t, m, g), g
		}, "governance-escrow"},
		{"forbidden custody", func(m, g []byte) ([]byte, []byte) {
			g = []byte(strings.Replace(string(g), `"balances":[`, `"balances":[{"address":"`+feeCollectorAddress+`","coins":[{"denom":"upnyx","amount":"1"}]},`, 1))
			g = []byte(strings.Replace(string(g), `"amount":"100000000000"}],"denom_metadata"`, `"amount":"100000000001"}],"denom_metadata"`, 1))
			return bind(t, m, g), g
		}, "module-isolation"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, g := fixture(t)
			m, g = tt.mutate(m, g)
			e := Verify(m, g)
			if e.Valid {
				t.Fatal("accepted mutation")
			}
			if check(e, tt.check).Pass {
				t.Fatalf("expected %s failure: %+v", tt.check, e)
			}
		})
	}
}

func TestManifestRejectsNonCanonicalAndUnsafeValues(t *testing.T) {
	m, g := fixture(t)
	cases := []string{`"100000000000"`, `"01"`, `"max_validator_power": 4`, `"max_validator_power": 1048577`}
	for i := 0; i < len(cases); i += 2 {
		mutated := []byte(strings.Replace(string(m), cases[i], cases[i+1], 1))
		if Verify(mutated, g).Valid {
			t.Fatalf("accepted replacement %s", cases[i+1])
		}
	}
}

func bind(t testing.TB, manifest, genesis []byte) []byte {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(manifest, &m); err != nil {
		t.Fatal(err)
	}
	d := sha256.Sum256(genesis)
	m["genesis_sha256"] = hex.EncodeToString(d[:])
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return out
}
func check(e Evidence, name string) Check {
	for _, c := range e.Checks {
		if c.Name == name {
			return c
		}
	}
	return Check{}
}

func FuzzVerifyNeverPanics(f *testing.F) {
	m, g := fixture(f)
	f.Add(m, g)
	f.Fuzz(func(t *testing.T, m, g []byte) { _ = Verify(m, g) })
}
