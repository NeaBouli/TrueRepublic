package genesisevidence

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cmted25519 "github.com/cometbft/cometbft/crypto/ed25519"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

func objects(t testing.TB) (Manifest, map[string]any) {
	t.Helper()
	m, g := fixture(t)
	var manifest Manifest
	var genesis map[string]any
	if json.Unmarshal(m, &manifest) != nil || json.Unmarshal(g, &genesis) != nil {
		t.Fatal("decode fixture")
	}
	return manifest, genesis
}

func encoded(t testing.TB, manifest Manifest, genesis map[string]any) ([]byte, []byte) {
	t.Helper()
	g, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(g)
	manifest.GenesisSHA256 = hex.EncodeToString(digest[:])
	m, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return m, g
}

func TestRealConsensusShapeAndIdentityBoundaries(t *testing.T) {
	t.Run("missing consensus", func(t *testing.T) {
		m, g := objects(t)
		delete(g, "consensus")
		mb, gb := encoded(t, m, g)
		e := Verify(mb, gb)
		if check(e, "consensus-validators").Pass {
			t.Fatal("accepted missing consensus")
		}
		for _, name := range []string{"application-validators", "bank-supply", "governance-escrow", "dex-custody", "module-isolation"} {
			if check(e, name).Pass {
				t.Fatalf("%s incorrectly passed without evaluation", name)
			}
		}
	})
	t.Run("ambiguous top level", func(t *testing.T) {
		m, g := objects(t)
		g["validators"] = []any{}
		mb, gb := encoded(t, m, g)
		e := Verify(mb, gb)
		if check(e, "consensus-validators").Pass {
			t.Fatal("accepted ambiguous validators")
		}
	})
	t.Run("missing app validator", func(t *testing.T) {
		m, g := objects(t)
		app := g["app_state"].(map[string]any)
		app["truedemocracy"].(map[string]any)["validators"] = []any{}
		mb, gb := encoded(t, m, g)
		e := Verify(mb, gb)
		if check(e, "application-validators").Pass {
			t.Fatal("accepted missing app validator")
		}
	})
	t.Run("legacy validator mixes explicit domains", func(t *testing.T) {
		m, g := objects(t)
		validator := g["app_state"].(map[string]any)["truedemocracy"].(map[string]any)["validators"].([]any)[0].(map[string]any)
		validator["domains"] = []any{}
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "application-validators").Pass {
			t.Fatal("accepted mixed legacy/explicit domain representation")
		}
	})
	t.Run("explicit primary domain mismatch", func(t *testing.T) {
		m, g := objects(t)
		validator := g["app_state"].(map[string]any)["truedemocracy"].(map[string]any)["validators"].([]any)[0].(map[string]any)
		validator["active"] = true
		validator["power"] = 1
		validator["domains"] = []any{"Bootstrap"}
		validator["domain"] = "Other"
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "application-validators").Pass {
			t.Fatal("accepted contradictory primary domain")
		}
	})
	t.Run("extra consensus validator", func(t *testing.T) {
		m, g := objects(t)
		c := g["consensus"].(map[string]any)
		vals := c["validators"].([]any)
		copyVal := map[string]any{}
		for k, v := range vals[0].(map[string]any) {
			copyVal[k] = v
		}
		pub := copyVal["pub_key"].(map[string]any)
		pubCopy := map[string]any{}
		for k, v := range pub {
			pubCopy[k] = v
		}
		pubCopy["value"] = base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))
		copyVal["pub_key"] = pubCopy
		c["validators"] = append(vals, copyVal)
		mb, gb := encoded(t, m, g)
		e := Verify(mb, gb)
		if check(e, "consensus-validators").Pass {
			t.Fatal("accepted extra validator")
		}
	})
	t.Run("manifest duplicate", func(t *testing.T) {
		m, g := objects(t)
		m.Validators = append(m.Validators, m.Validators[0])
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "manifest").Pass {
			t.Fatal("accepted duplicate manifest validator")
		}
	})
	t.Run("stake remainder", func(t *testing.T) {
		m, g := objects(t)
		m.Validators[0].StakeUPNYX = "100000000001"
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "manifest").Pass {
			t.Fatal("accepted non-divisible stake")
		}
	})
	t.Run("invalid checksum", func(t *testing.T) {
		m, g := objects(t)
		m.Validators[0].OperatorAddress = m.Validators[0].OperatorAddress[:len(m.Validators[0].OperatorAddress)-1] + "q"
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "manifest").Pass {
			t.Fatal("accepted bad checksum")
		}
	})
	t.Run("module operator", func(t *testing.T) {
		m, g := objects(t)
		m.Validators[0].OperatorAddress = dexAddress
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "manifest").Pass {
			t.Fatal("accepted module operator")
		}
	})
	t.Run("consensus-derived operator", func(t *testing.T) {
		m, g := objects(t)
		key, _ := base64.StdEncoding.DecodeString(m.Validators[0].ConsensusPubKey)
		sum := sha256.Sum256(key)
		operator, err := encodeAddress(sum[:20])
		if err != nil {
			t.Fatal(err)
		}
		m.Validators[0].OperatorAddress = operator
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "manifest").Pass {
			t.Fatal("accepted coupled authority")
		}
	})
	t.Run("cross-validator-derived operator", func(t *testing.T) {
		m, g := objects(t)
		key2 := bytes.Repeat([]byte{9}, 32)
		sum := sha256.Sum256(key2)
		operator, err := encodeAddress(sum[:20])
		if err != nil {
			t.Fatal(err)
		}
		m.Validators[0].OperatorAddress = operator
		m.Validators = append(m.Validators, Validator{OperatorAddress: testOperator, ConsensusPubKey: base64.StdEncoding.EncodeToString(key2), StakeUPNYX: "100000000000", Power: 1})
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "manifest").Pass {
			t.Fatal("accepted cross-validator authority collision")
		}
	})
}

func TestSDKAppGenesisMarshalShape(t *testing.T) {
	m, original := objects(t)
	appState, err := json.Marshal(original["app_state"])
	if err != nil {
		t.Fatal(err)
	}
	key := cmted25519.PubKey([]byte("12345678901234567890123456789012"))
	addressHash := sha256.Sum256(key)
	genesis := genutiltypes.AppGenesis{
		AppName: "truerepublicd", AppVersion: "v0.4.1", GenesisTime: time.Unix(0, 0).UTC(),
		ChainID: m.ChainID, InitialHeight: 1, AppState: appState,
		Consensus: &genutiltypes.ConsensusGenesis{Validators: []cmttypes.GenesisValidator{{Address: addressHash[:20], PubKey: key, Power: 1, Name: "validator-1"}}},
	}
	genesisJSON, err := json.MarshalIndent(genesis, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(genesisJSON)
	m.GenesisSHA256 = hex.EncodeToString(digest[:])
	manifestJSON, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if e := Verify(manifestJSON, genesisJSON); !e.Valid {
		t.Fatalf("SDK AppGenesis rejected: %+v\n%s", e, genesisJSON)
	}
}

func TestConsensusAddressNameAllocationAndDomainAdversaries(t *testing.T) {
	t.Run("consensus address", func(t *testing.T) {
		m, g := objects(t)
		v := g["consensus"].(map[string]any)["validators"].([]any)[0].(map[string]any)
		v["address"] = "0000000000000000000000000000000000000000"
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "consensus-validators").Pass {
			t.Fatal("accepted wrong consensus address")
		}
	})
	t.Run("consensus name", func(t *testing.T) {
		m, g := objects(t)
		v := g["consensus"].(map[string]any)["validators"].([]any)[0].(map[string]any)
		v["name"] = " "
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "consensus-validators").Pass {
			t.Fatal("accepted empty consensus name")
		}
	})
	t.Run("allocation mismatch", func(t *testing.T) {
		m, g := objects(t)
		m.Allocations[0].Coins[0].Amount = "99999999999"
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "bank-supply").Pass {
			t.Fatal("accepted allocation mismatch")
		}
	})
	t.Run("module admin", func(t *testing.T) {
		m, g := objects(t)
		domain := g["app_state"].(map[string]any)["truedemocracy"].(map[string]any)["domains"].([]any)[0].(map[string]any)
		domain["admin"] = dexAddress
		domain["members"] = []any{dexAddress}
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "governance-escrow").Pass {
			t.Fatal("accepted module domain authority")
		}
	})
	t.Run("admin outside members", func(t *testing.T) {
		m, g := objects(t)
		domain := g["app_state"].(map[string]any)["truedemocracy"].(map[string]any)["domains"].([]any)[0].(map[string]any)
		domain["members"] = []any{}
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "governance-escrow").Pass {
			t.Fatal("accepted missing admin membership")
		}
	})
	t.Run("validator missing domain", func(t *testing.T) {
		m, g := objects(t)
		validator := g["app_state"].(map[string]any)["truedemocracy"].(map[string]any)["validators"].([]any)[0].(map[string]any)
		validator["domain"] = "Missing"
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "application-validators").Pass {
			t.Fatal("accepted missing validator domain")
		}
	})
	t.Run("validator bound", func(t *testing.T) {
		m, g := objects(t)
		for len(m.Validators) < 65 {
			m.Validators = append(m.Validators, m.Validators[0])
		}
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "manifest").Pass {
			t.Fatal("accepted oversized validator set")
		}
	})
}

func TestDEXCustodyAndBankReconciliation(t *testing.T) {
	m, g := objects(t)
	app := g["app_state"].(map[string]any)
	bank := app["bank"].(map[string]any)
	bank["balances"] = []any{
		map[string]any{"address": trueDemocracyAddress, "coins": []any{map[string]any{"denom": "upnyx", "amount": "100000000000"}}},
		map[string]any{"address": dexAddress, "coins": []any{map[string]any{"denom": "atom", "amount": "10"}, map[string]any{"denom": "upnyx", "amount": "50"}}},
	}
	bank["supply"] = []any{map[string]any{"denom": "atom", "amount": "10"}, map[string]any{"denom": "upnyx", "amount": "100000000050"}}
	app["dex"].(map[string]any)["pools"] = []any{map[string]any{"pnyx_reserve": "50", "asset_reserve": "10", "asset_denom": "atom", "total_shares": "1", "total_burned": "0", "swap_count": 0, "total_volume_pnyx": "0"}}
	app["dex"].(map[string]any)["registered_assets"] = []any{map[string]any{"ibc_denom": "atom", "trading_enabled": true}}
	m.TotalSupplyUPNYX = "100000000050"
	m.DEXCustody = []Coin{{Denom: "atom", Amount: "10"}, {Denom: "upnyx", Amount: "50"}}
	m.Allocations = []Allocation{
		{Address: trueDemocracyAddress, Coins: []Coin{{Denom: "upnyx", Amount: "100000000000"}}},
		{Address: dexAddress, Coins: []Coin{{Denom: "atom", Amount: "10"}, {Denom: "upnyx", Amount: "50"}}},
	}
	mb, gb := encoded(t, m, g)
	if e := Verify(mb, gb); !e.Valid {
		t.Fatalf("valid DEX fixture rejected: %+v", e)
	}
	t.Run("custody mismatch", func(t *testing.T) {
		bad := m
		bad.DEXCustody = []Coin{{Denom: "atom", Amount: "11"}, {Denom: "upnyx", Amount: "50"}}
		bm, bg := encoded(t, bad, g)
		if check(Verify(bm, bg), "dex-custody").Pass {
			t.Fatal("accepted custody mismatch")
		}
	})
	t.Run("duplicate allocation", func(t *testing.T) {
		_, badG := objects(t)
		badApp := badG["app_state"].(map[string]any)
		badBank := badApp["bank"].(map[string]any)
		balances := badBank["balances"].([]any)
		badBank["balances"] = append(balances, balances[0])
		bm, bg := encoded(t, func() Manifest { x, _ := objects(t); return x }(), badG)
		if check(Verify(bm, bg), "bank-supply").Pass {
			t.Fatal("accepted duplicate allocation")
		}
	})
	t.Run("malformed pool", func(t *testing.T) {
		badG := cloneMap(t, g)
		badApp := badG["app_state"].(map[string]any)
		badApp["dex"].(map[string]any)["pools"].([]any)[0].(map[string]any)["pnyx_reserve"] = "0"
		bm, bg := encoded(t, m, badG)
		if check(Verify(bm, bg), "dex-custody").Pass {
			t.Fatal("accepted zero reserve")
		}
	})
	t.Run("unregistered pool", func(t *testing.T) {
		badG := cloneMap(t, g)
		badG["app_state"].(map[string]any)["dex"].(map[string]any)["registered_assets"] = []any{}
		bm, bg := encoded(t, m, badG)
		if check(Verify(bm, bg), "dex-custody").Pass {
			t.Fatal("accepted unregistered pool")
		}
	})
	t.Run("duplicate pool", func(t *testing.T) {
		badG := cloneMap(t, g)
		dex := badG["app_state"].(map[string]any)["dex"].(map[string]any)
		pools := dex["pools"].([]any)
		dex["pools"] = append(pools, pools[0])
		bm, bg := encoded(t, m, badG)
		if check(Verify(bm, bg), "dex-custody").Pass {
			t.Fatal("accepted duplicate pool")
		}
	})
}

func TestSupplyCapModuleIsolationAndGoldenAddresses(t *testing.T) {
	for name, address := range map[string]string{"truedemocracy": trueDemocracyAddress, "dex": dexAddress, "fee_collector": feeCollectorAddress, "wasm": wasmAddress, "transfer": transferAddress} {
		decoded, ok := accountAddress(address)
		if !ok || len(decoded) != 20 || !bytes.Equal(decoded, authtypes.NewModuleAddress(name)) {
			t.Fatalf("%s golden invalid", name)
		}
	}
	t.Run("above cap", func(t *testing.T) {
		m, g := objects(t)
		m.TotalSupplyUPNYX = "21000000000001"
		app := g["app_state"].(map[string]any)
		bank := app["bank"].(map[string]any)
		bank["balances"].([]any)[0].(map[string]any)["coins"].([]any)[0].(map[string]any)["amount"] = "21000000000001"
		bank["supply"].([]any)[0].(map[string]any)["amount"] = "21000000000001"
		mb, gb := encoded(t, m, g)
		e := Verify(mb, gb)
		if check(e, "manifest").Pass || check(e, "bank-supply").Pass {
			t.Fatal("accepted cap breach")
		}
	})
	t.Run("forbidden module", func(t *testing.T) {
		m, g := objects(t)
		app := g["app_state"].(map[string]any)
		bank := app["bank"].(map[string]any)
		bank["balances"] = append(bank["balances"].([]any), map[string]any{"address": wasmAddress, "coins": []any{map[string]any{"denom": "atom", "amount": "1"}}})
		bank["supply"] = append([]any{map[string]any{"denom": "atom", "amount": "1"}}, bank["supply"].([]any)...)
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "module-isolation").Pass {
			t.Fatal("accepted funded wasm module")
		}
	})
	t.Run("empty auth accounts", func(t *testing.T) {
		m, g := objects(t)
		g["app_state"].(map[string]any)["auth"].(map[string]any)["accounts"] = []any{}
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "module-isolation").Pass {
			t.Fatal("accepted missing module and operator accounts")
		}
	})
	t.Run("module permissions", func(t *testing.T) {
		m, g := objects(t)
		accounts := g["app_state"].(map[string]any)["auth"].(map[string]any)["accounts"].([]any)
		for _, raw := range accounts {
			o := raw.(map[string]any)
			if o["name"] == "truedemocracy" {
				o["permissions"] = []any{"minter"}
			}
		}
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "module-isolation").Pass {
			t.Fatal("accepted incomplete module permissions")
		}
	})
}

func TestGovernanceTreasuryAndPendingRemovalClaims(t *testing.T) {
	m, g := objects(t)
	app := g["app_state"].(map[string]any)
	td := app["truedemocracy"].(map[string]any)
	td["domains"].([]any)[0].(map[string]any)["treasury"] = []any{map[string]any{"denom": "upnyx", "amount": "25"}}
	td["pending_validator_removals"] = []any{map[string]any{"validator": map[string]any{"stake": []any{map[string]any{"denom": "upnyx", "amount": "50"}}}}}
	bank := app["bank"].(map[string]any)
	bank["balances"].([]any)[0].(map[string]any)["coins"].([]any)[0].(map[string]any)["amount"] = "100000000075"
	bank["supply"].([]any)[0].(map[string]any)["amount"] = "100000000075"
	m.TotalSupplyUPNYX = "100000000075"
	m.GovernanceEscrowUPNYX = "100000000075"
	m.Allocations[0].Coins[0].Amount = "100000000075"
	mb, gb := encoded(t, m, g)
	if e := Verify(mb, gb); !e.Valid {
		t.Fatalf("valid aggregate claims rejected: %+v", e)
	}
	td["domains"].([]any)[0].(map[string]any)["treasury"] = []any{map[string]any{"denom": "atom", "amount": "25"}}
	mb, gb = encoded(t, m, g)
	if check(Verify(mb, gb), "governance-escrow").Pass {
		t.Fatal("accepted non-PNYX governance treasury")
	}
}

func TestAdditionalStrictAndModuleFailurePaths(t *testing.T) {
	t.Run("source drift", func(t *testing.T) {
		m, g := objects(t)
		m.SourceCommit = strings.Repeat("A", 40)
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "manifest").Pass {
			t.Fatal("accepted source drift")
		}
	})
	t.Run("daemon version drift", func(t *testing.T) {
		m, g := objects(t)
		m.DaemonVersion = strings.Repeat("b", 40)
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "manifest").Pass {
			t.Fatal("accepted daemon/source drift")
		}
	})
	t.Run("unknown funded module", func(t *testing.T) {
		m, g := objects(t)
		app := g["app_state"].(map[string]any)
		auth := app["auth"].(map[string]any)
		auth["accounts"] = append(auth["accounts"].([]any), map[string]any{"base_account": map[string]any{"address": testOperator}, "name": "unknown"})
		bank := app["bank"].(map[string]any)
		bank["balances"] = append(bank["balances"].([]any), map[string]any{"address": testOperator, "coins": []any{map[string]any{"denom": "atom", "amount": "1"}}})
		bank["supply"] = append([]any{map[string]any{"denom": "atom", "amount": "1"}}, bank["supply"].([]any)...)
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "module-isolation").Pass {
			t.Fatal("accepted unknown funded module")
		}
	})
	t.Run("known module wrong address", func(t *testing.T) {
		m, g := objects(t)
		accounts := g["app_state"].(map[string]any)["auth"].(map[string]any)["accounts"].([]any)
		for _, raw := range accounts {
			o := raw.(map[string]any)
			if o["name"] == "dex" {
				o["base_account"].(map[string]any)["address"] = testOperator
			}
		}
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "module-isolation").Pass {
			t.Fatal("accepted wrong module address")
		}
	})
	t.Run("invalid auth accounts", func(t *testing.T) {
		m, g := objects(t)
		g["app_state"].(map[string]any)["auth"].(map[string]any)["accounts"] = map[string]any{}
		mb, gb := encoded(t, m, g)
		if check(Verify(mb, gb), "module-isolation").Pass {
			t.Fatal("accepted invalid auth accounts")
		}
	})
	if got := SortedCoins(map[string]string{"upnyx": "2", "atom": "1"}); len(got) != 2 || got[0].Denom != "atom" {
		t.Fatalf("sorted coins = %+v", got)
	}
	for _, raw := range [][]byte{[]byte(`[]`), []byte(`{"a":1} {}`), []byte(`{"a":1,"a":2}`)} {
		if _, _, err := strictJSON(raw, 1024); err == nil {
			t.Fatalf("accepted strict input %s", raw)
		}
	}
}

func TestCommandJSONTextAndFailure(t *testing.T) {
	m, g := fixture(t)
	dir := t.TempDir()
	mp := filepath.Join(dir, "manifest.json")
	gp := filepath.Join(dir, "genesis.json")
	if os.WriteFile(mp, m, 0600) != nil || os.WriteFile(gp, g, 0600) != nil {
		t.Fatal("write fixtures")
	}
	for _, output := range []string{"json", "text"} {
		cmd := NewCommand()
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetArgs([]string{"verify", "--manifest", mp, "--genesis", gp, "--output", output})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("%s: %v", output, err)
		}
		if out.Len() == 0 {
			t.Fatal("missing output")
		}
	}
	bad := append(append([]byte(nil), g...), '\n')
	if os.WriteFile(gp, bad, 0600) != nil {
		t.Fatal("write bad")
	}
	cmd := NewCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"verify", "--manifest", mp, "--genesis", gp, "--output", "json"})
	if err := cmd.Execute(); err == nil || !strings.Contains(out.String(), `"valid": false`) {
		t.Fatal("invalid evidence was not printed")
	}
	cmd = NewCommand()
	cmd.SetArgs([]string{"verify", "--manifest", filepath.Join(dir, "missing"), "--genesis", gp})
	if cmd.Execute() == nil {
		t.Fatal("accepted missing file")
	}
	cmd = NewCommand()
	cmd.SetArgs([]string{"verify", "extra"})
	if cmd.Execute() == nil {
		t.Fatal("accepted positional argument")
	}
	cmd = NewCommand()
	cmd.SetArgs([]string{"verify", "--manifest", mp, "--genesis", gp, "--output", "yaml"})
	if cmd.Execute() == nil {
		t.Fatal("accepted unknown output")
	}
	if err := os.Symlink(mp, filepath.Join(dir, "manifest-link.json")); err == nil {
		cmd = NewCommand()
		cmd.SetArgs([]string{"verify", "--manifest", filepath.Join(dir, "manifest-link.json"), "--genesis", gp})
		if cmd.Execute() == nil {
			t.Fatal("accepted symlink input")
		}
	}
	cmd = NewCommand()
	cmd.SetArgs([]string{"verify", "--manifest", dir, "--genesis", gp})
	if cmd.Execute() == nil {
		t.Fatal("accepted non-regular input")
	}
}

func cloneMap(t testing.TB, value map[string]any) map[string]any {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if json.Unmarshal(data, &result) != nil {
		t.Fatal("clone")
	}
	return result
}

func encodeAddress(raw []byte) (string, error) { return bech32.ConvertAndEncode("truerepublic", raw) }
