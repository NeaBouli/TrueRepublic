package genesisevidence

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"regexp"
	"sort"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/bech32"
)

var (
	hex40        = regexp.MustCompile(`^[0-9a-f]{40}$`)
	hex64        = regexp.MustCompile(`^[0-9a-f]{64}$`)
	chainPattern = regexp.MustCompile(`^[0-9A-Za-z][0-9A-Za-z._-]{0,49}$`)
	denomPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9/:._-]{2,127}$`)
)

var checkNames = []string{
	"manifest", "genesis-binding", "chain-identity", "consensus-validators",
	"application-validators", "bank-supply", "governance-escrow",
	"dex-custody", "module-isolation",
}

type state struct {
	evidence Evidence
	byName   map[string]int
}

func Verify(manifestBytes, genesisBytes []byte) Evidence {
	mh := sha256.Sum256(manifestBytes)
	gh := sha256.Sum256(genesisBytes)
	s := &state{evidence: Evidence{Schema: EvidenceSchema, ManifestSHA256: hex.EncodeToString(mh[:]), GenesisSHA256: hex.EncodeToString(gh[:])}, byName: map[string]int{}}
	for _, name := range checkNames {
		s.evidence.Checks = append(s.evidence.Checks, Check{Name: name, Pass: true, Violations: []string{}})
		s.byName[name] = len(s.evidence.Checks) - 1
	}
	manifest, err := parseManifest(manifestBytes)
	if err != nil {
		s.fail("manifest", "invalid-manifest")
		s.skipAfter("manifest")
		return s.finish()
	}
	s.validateManifest(manifest)
	if !s.evidence.Checks[s.byName["manifest"]].Pass {
		s.skipAfter("manifest")
		return s.finish()
	}
	_, root, err := strictJSON(genesisBytes, MaxGenesisBytes)
	if err != nil {
		s.fail("genesis-binding", "invalid-genesis-json")
		s.skipAfter("genesis-binding")
		return s.finish()
	}
	if manifest.GenesisSHA256 != s.evidence.GenesisSHA256 {
		s.fail("genesis-binding", "raw-genesis-digest-mismatch")
	}
	s.verifyGenesis(manifest, root)
	return s.finish()
}

func MarshalJSON(e Evidence) ([]byte, error) { return json.MarshalIndent(e, "", "  ") }

func (s *state) finish() Evidence {
	s.evidence.Valid = true
	for i := range s.evidence.Checks {
		if !s.evidence.Checks[i].Pass {
			s.evidence.Valid = false
		}
	}
	return s.evidence
}
func (s *state) fail(name, violation string) {
	c := &s.evidence.Checks[s.byName[name]]
	c.Pass = false
	c.Violations = append(c.Violations, violation)
}
func (s *state) skipAfter(name string) {
	found := false
	for _, n := range checkNames {
		if found {
			s.fail(n, "not-evaluated")
		}
		if n == name {
			found = true
		}
	}
}

func (s *state) validateManifest(m Manifest) {
	if m.Schema != ManifestSchema {
		s.fail("manifest", "unsupported-schema")
	}
	if !hex40.MatchString(m.SourceCommit) {
		s.fail("manifest", "invalid-source-commit")
	}
	if !hex40.MatchString(m.DaemonVersion) || m.DaemonVersion != m.SourceCommit {
		s.fail("manifest", "invalid-daemon-version")
	}
	if !chainPattern.MatchString(m.ChainID) {
		s.fail("manifest", "invalid-chain-id")
	}
	if !hex64.MatchString(m.GenesisSHA256) {
		s.fail("manifest", "invalid-genesis-digest")
	}
	if m.MaxValidatorPower <= 0 || m.MaxValidatorPower > MaxPowerLimit {
		s.fail("manifest", "invalid-power-limit")
	}
	if len(m.Validators) == 0 {
		s.fail("manifest", "empty-validator-set")
	} else if len(m.Validators) > 64 {
		s.fail("manifest", "validator-set-too-large")
	}
	seenOperators := map[string]bool{}
	seenKeys := map[string]bool{}
	derivedAuthorities := make([][]byte, 0, len(m.Validators))
	for _, validator := range m.Validators {
		key, err := base64.StdEncoding.Strict().DecodeString(validator.ConsensusPubKey)
		if err == nil && len(key) == 32 {
			derived := sha256.Sum256(key)
			derivedAuthorities = append(derivedAuthorities, derived[:20])
		}
	}
	for _, v := range m.Validators {
		operator, operatorOK := accountAddress(v.OperatorAddress)
		if !operatorOK || isModuleAddress(v.OperatorAddress) || seenOperators[v.OperatorAddress] {
			s.fail("manifest", "invalid-or-duplicate-operator")
		}
		seenOperators[v.OperatorAddress] = true
		key, err := base64.StdEncoding.Strict().DecodeString(v.ConsensusPubKey)
		if err != nil || len(key) != 32 || seenKeys[v.ConsensusPubKey] {
			s.fail("manifest", "invalid-or-duplicate-consensus-key")
		}
		if operatorOK {
			for _, derived := range derivedAuthorities {
				if bytes.Equal(operator, derived) {
					s.fail("manifest", "operator-consensus-authority-collision")
					break
				}
			}
		}
		seenKeys[v.ConsensusPubKey] = true
		stake, ok := amount(v.StakeUPNYX)
		if !ok || stake.Sign() <= 0 || !stake.IsInt64() {
			s.fail("manifest", "invalid-validator-stake")
		}
		if v.Power <= 0 || v.Power > m.MaxValidatorPower {
			s.fail("manifest", "invalid-validator-power")
		}
		if ok && stake.IsInt64() && (stake.Int64()%StakeUnit != 0 || stake.Int64()/StakeUnit != v.Power) {
			s.fail("manifest", "stake-power-mismatch")
		}
	}
	lastAddress := ""
	for _, allocation := range m.Allocations {
		if _, ok := accountAddress(allocation.Address); !ok || allocation.Address <= lastAddress || len(allocation.Coins) == 0 {
			s.fail("manifest", "invalid-or-unsorted-allocation")
		}
		if _, ok := coinMap(allocation.Coins, true); !ok {
			s.fail("manifest", "invalid-allocation-coins")
		}
		lastAddress = allocation.Address
	}
	if _, ok := boundedAmount(m.TotalSupplyUPNYX, MaxPNYXSupply); !ok {
		s.fail("manifest", "invalid-total-supply")
	}
	if _, ok := amount(m.GovernanceEscrowUPNYX); !ok {
		s.fail("manifest", "invalid-governance-escrow")
	}
	last := ""
	for _, c := range m.DEXCustody {
		if !denomPattern.MatchString(c.Denom) || c.Denom <= last {
			s.fail("manifest", "invalid-or-unsorted-dex-custody")
		}
		if _, ok := amount(c.Amount); !ok {
			s.fail("manifest", "invalid-dex-custody-amount")
		} else if mustAmount(c.Amount).Sign() <= 0 {
			s.fail("manifest", "invalid-dex-custody-amount")
		}
		last = c.Denom
	}
}

type consensusValidator struct {
	Address string `json:"address"`
	PubKey  struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"pub_key"`
	Power any    `json:"power"`
	Name  string `json:"name"`
}
type appValidator struct {
	Operator string      `json:"operator_addr"`
	PubKey   string      `json:"pub_key"`
	Stake    json.Number `json:"stake"`
	Power    int64       `json:"power"`
	Active   *bool       `json:"active"`
	Jailed   bool        `json:"jailed"`
	Domain   string      `json:"domain"`
	Domains  []string    `json:"domains"`
}
type genesisRoot struct {
	ChainID   string `json:"chain_id"`
	Consensus *struct {
		Validators []consensusValidator `json:"validators"`
	} `json:"consensus"`
	AppState map[string]any `json:"app_state"`
}

func (s *state) verifyGenesis(m Manifest, raw map[string]any) {
	var g genesisRoot
	if err := decodeInto(raw, &g); err != nil {
		s.fail("genesis-binding", "invalid-genesis-structure")
		s.skipAfter("genesis-binding")
		return
	}
	if g.ChainID != m.ChainID || !chainPattern.MatchString(g.ChainID) {
		s.fail("chain-identity", "chain-id-mismatch")
	}
	if _, ambiguous := raw["validators"]; ambiguous {
		s.fail("consensus-validators", "ambiguous-top-level-validator-set")
	}
	if g.Consensus == nil {
		s.fail("consensus-validators", "missing-consensus-state")
		s.skipAfter("consensus-validators")
		return
	}
	want := map[string]Validator{}
	for _, v := range m.Validators {
		want[v.ConsensusPubKey] = v
	}
	consensus := map[string]int64{}
	for _, v := range g.Consensus.Validators {
		key, err := base64.StdEncoding.Strict().DecodeString(v.PubKey.Value)
		key64 := base64.StdEncoding.EncodeToString(key)
		power, ok := int64Value(v.Power)
		derived := sha256.Sum256(key)
		wantAddress := strings.ToUpper(hex.EncodeToString(derived[:20]))
		if err != nil || len(key) != 32 || (v.PubKey.Type != "tendermint/PubKeyEd25519" && v.PubKey.Type != "cometbft/PubKeyEd25519") || !ok || power <= 0 || v.Address != wantAddress || strings.TrimSpace(v.Name) == "" {
			s.fail("consensus-validators", "invalid-consensus-validator")
			continue
		}
		if _, dup := consensus[key64]; dup {
			s.fail("consensus-validators", "duplicate-consensus-key")
		}
		consensus[key64] = power
		mv, found := want[key64]
		if !found || mv.Power != power || power > m.MaxValidatorPower {
			s.fail("consensus-validators", "consensus-manifest-mismatch")
		}
	}
	if len(consensus) != len(want) {
		s.fail("consensus-validators", "validator-count-mismatch")
	}
	td, ok := g.AppState["truedemocracy"].(map[string]any)
	if !ok {
		s.fail("application-validators", "missing-truedemocracy-state")
	} else {
		domains, domainsOK := s.verifyDomains(td)
		s.verifyAppValidators(m, td, want, domains, domainsOK)
	}
	bank, ok := g.AppState["bank"].(map[string]any)
	if !ok {
		s.fail("bank-supply", "missing-bank-state")
		s.fail("governance-escrow", "not-evaluated")
		s.fail("dex-custody", "not-evaluated")
		s.fail("module-isolation", "not-evaluated")
		return
	}
	balances, supply, valid := s.verifyBank(m, bank)
	if !valid {
		s.fail("governance-escrow", "not-evaluated")
		s.fail("dex-custody", "not-evaluated")
		s.fail("module-isolation", "not-evaluated")
		return
	}
	govClaims, govOK := governanceClaims(td)
	if !govOK {
		s.fail("governance-escrow", "invalid-governance-claims")
	} else {
		s.compareGovernance(m, balances, govClaims)
	}
	dexState, ok := g.AppState["dex"].(map[string]any)
	if !ok {
		s.fail("dex-custody", "missing-dex-state")
	} else {
		claims, valid := dexClaims(dexState)
		if !valid {
			s.fail("dex-custody", "invalid-dex-claims")
		} else {
			s.compareDEX(m, balances, claims)
		}
	}
	s.verifyModuleIsolation(m, g.AppState, balances, supply)
}

func (s *state) verifyAppValidators(m Manifest, td map[string]any, want map[string]Validator, domains map[string]map[string]bool, domainsOK bool) {
	values, ok := td["validators"].([]any)
	if !ok {
		s.fail("application-validators", "missing-application-validators")
		return
	}
	seenK := map[string]bool{}
	seenO := map[string]bool{}
	for _, raw := range values {
		var v appValidator
		if decodeInto(raw, &v) != nil {
			s.fail("application-validators", "invalid-application-validator")
			continue
		}
		keyBytes, err := base64.StdEncoding.Strict().DecodeString(v.PubKey)
		key := base64.StdEncoding.EncodeToString(keyBytes)
		stake, stakeOK := amount(v.Stake.String())
		mv, found := want[key]
		if err != nil || len(keyBytes) != 32 || seenK[key] || seenO[v.Operator] {
			s.fail("application-validators", "duplicate-or-invalid-identity")
		}
		seenK[key] = true
		seenO[v.Operator] = true
		if !stakeOK || !stake.IsInt64() || stake.Int64()%StakeUnit != 0 {
			s.fail("application-validators", "application-manifest-mismatch")
			continue
		}
		active := !v.Jailed
		if v.Active != nil {
			active = *v.Active
		}
		storedPower := v.Power
		if v.Active == nil {
			if v.Power != 0 || v.Domains != nil {
				s.fail("application-validators", "application-manifest-mismatch")
			}
			storedPower = stake.Int64() / StakeUnit
		} else if (len(v.Domains) == 0) != (v.Domain == "") || (len(v.Domains) > 0 && v.Domains[0] != v.Domain) {
			s.fail("application-validators", "validator-domain-representation-mismatch")
		}
		declaredDomains := v.Domains
		if v.Active == nil {
			declaredDomains = []string{v.Domain}
		}
		if domainsOK {
			if len(declaredDomains) == 0 {
				s.fail("application-validators", "validator-domain-authority-mismatch")
			}
			seenDomains := map[string]bool{}
			for _, domain := range declaredDomains {
				members, exists := domains[domain]
				if domain == "" || !exists || !members[v.Operator] || seenDomains[domain] {
					s.fail("application-validators", "validator-domain-authority-mismatch")
				}
				seenDomains[domain] = true
			}
		}
		if !found || !stakeOK || !stake.IsInt64() || mv.OperatorAddress != v.Operator || mv.StakeUPNYX != v.Stake.String() || mv.Power != storedPower || !active || v.Jailed || storedPower != stake.Int64()/StakeUnit {
			s.fail("application-validators", "application-manifest-mismatch")
		}
	}
	if len(values) != len(want) {
		s.fail("application-validators", "validator-count-mismatch")
	}
}

func (s *state) verifyDomains(td map[string]any) (map[string]map[string]bool, bool) {
	values, ok := td["domains"].([]any)
	if !ok {
		s.fail("governance-escrow", "invalid-domain-authorities")
		return nil, false
	}
	domains := map[string]map[string]bool{}
	valid := true
	for _, raw := range values {
		o, ok := raw.(map[string]any)
		if !ok {
			valid = false
			continue
		}
		name, _ := o["name"].(string)
		admin, _ := o["admin"].(string)
		membersRaw, membersOK := o["members"].([]any)
		if name == "" || domains[name] != nil || !membersOK {
			valid = false
			continue
		}
		if _, ok := accountAddress(admin); !ok || isModuleAddress(admin) {
			valid = false
		}
		members := map[string]bool{}
		for _, item := range membersRaw {
			member, ok := item.(string)
			if !ok {
				valid = false
				continue
			}
			if _, ok := accountAddress(member); !ok || isModuleAddress(member) || members[member] {
				valid = false
			}
			members[member] = true
		}
		if !members[admin] {
			valid = false
		}
		domains[name] = members
	}
	if !valid {
		s.fail("governance-escrow", "invalid-domain-authorities")
	}
	return domains, valid
}

type bankState struct {
	Balances []struct {
		Address string `json:"address"`
		Coins   []Coin `json:"coins"`
	} `json:"balances"`
	Supply []Coin `json:"supply"`
}

func (s *state) verifyBank(m Manifest, raw map[string]any) (map[string]map[string]*big.Int, map[string]*big.Int, bool) {
	var b bankState
	if decodeInto(raw, &b) != nil {
		s.fail("bank-supply", "invalid-bank-state")
		return nil, nil, false
	}
	balances := map[string]map[string]*big.Int{}
	sums := map[string]*big.Int{}
	for _, bal := range b.Balances {
		if _, ok := accountAddress(bal.Address); !ok || balances[bal.Address] != nil {
			s.fail("bank-supply", "duplicate-or-invalid-balance")
			continue
		}
		coins, ok := coinMap(bal.Coins, true)
		if !ok || len(coins) == 0 {
			s.fail("bank-supply", "invalid-balance-coins")
		}
		balances[bal.Address] = coins
		addCoins(sums, coins)
	}
	manifestBalances := map[string]map[string]*big.Int{}
	for _, allocation := range m.Allocations {
		coins, ok := coinMap(allocation.Coins, true)
		if ok {
			manifestBalances[allocation.Address] = coins
		}
	}
	if len(manifestBalances) != len(balances) {
		s.fail("bank-supply", "manifest-allocation-mismatch")
	} else {
		for address, coins := range balances {
			want, found := manifestBalances[address]
			if !found || !coinMapsEqual(coins, want) {
				s.fail("bank-supply", "manifest-allocation-mismatch")
				break
			}
		}
	}
	supply, ok := coinMap(b.Supply, true)
	if !ok {
		s.fail("bank-supply", "invalid-supply-coins")
	}
	if !coinMapsEqual(sums, supply) {
		s.fail("bank-supply", "supply-balance-mismatch")
	}
	pnyx := zero(supply["upnyx"])
	if pnyx.String() != m.TotalSupplyUPNYX {
		s.fail("bank-supply", "manifest-supply-mismatch")
	}
	if pnyx.Cmp(mustAmount(MaxPNYXSupply)) > 0 {
		s.fail("bank-supply", "pnyx-cap-exceeded")
	}
	return balances, supply, ok
}

func governanceClaims(td map[string]any) (*big.Int, bool) {
	if td == nil {
		return nil, false
	}
	total := new(big.Int)
	domains, ok := td["domains"].([]any)
	if !ok {
		return nil, false
	}
	for _, r := range domains {
		o, ok := r.(map[string]any)
		if !ok {
			return nil, false
		}
		coins, ok := coinsFrom(o["treasury"])
		if !ok {
			return nil, false
		}
		if len(coins) > 1 || (len(coins) == 1 && coins["upnyx"] == nil) {
			return nil, false
		}
		total.Add(total, zero(coins["upnyx"]))
	}
	vals, ok := td["validators"].([]any)
	if !ok {
		return nil, false
	}
	for _, r := range vals {
		o, ok := r.(map[string]any)
		if !ok {
			return nil, false
		}
		n, ok := numberAmount(o["stake"])
		if !ok {
			return nil, false
		}
		total.Add(total, n)
	}
	if removals, exists := td["pending_validator_removals"]; exists {
		a, ok := removals.([]any)
		if !ok {
			return nil, false
		}
		for _, r := range a {
			o, ok := r.(map[string]any)
			if !ok {
				return nil, false
			}
			v, ok := o["validator"].(map[string]any)
			if !ok {
				return nil, false
			}
			coins, ok := coinsFrom(v["stake"])
			if !ok {
				return nil, false
			}
			if len(coins) > 1 || (len(coins) == 1 && coins["upnyx"] == nil) {
				return nil, false
			}
			total.Add(total, zero(coins["upnyx"]))
		}
	}
	return total, true
}
func (s *state) compareGovernance(m Manifest, b map[string]map[string]*big.Int, claims *big.Int) {
	if claims.String() != m.GovernanceEscrowUPNYX {
		s.fail("governance-escrow", "manifest-governance-claim-mismatch")
	}
	coins := b[trueDemocracyAddress]
	if len(coins) != (func() int {
		if claims.Sign() > 0 {
			return 1
		}
		return 0
	})() || zero(coins["upnyx"]).Cmp(claims) != 0 {
		s.fail("governance-escrow", "governance-bank-custody-mismatch")
	}
}

func dexClaims(raw map[string]any) (map[string]*big.Int, bool) {
	pools, ok := raw["pools"].([]any)
	if !ok {
		return nil, false
	}
	registered, ok := raw["registered_assets"].([]any)
	if !ok {
		return nil, false
	}
	assets := map[string]bool{}
	for _, r := range registered {
		o, ok := r.(map[string]any)
		if !ok {
			return nil, false
		}
		denom, ok := o["ibc_denom"].(string)
		enabled, enabledOK := o["trading_enabled"].(bool)
		if !ok || !enabledOK || !denomPattern.MatchString(denom) || assets[denom] {
			return nil, false
		}
		assets[denom] = enabled
	}
	claims := map[string]*big.Int{}
	seen := map[string]bool{}
	poolShares := map[string]*big.Int{}
	for _, r := range pools {
		o, ok := r.(map[string]any)
		if !ok {
			return nil, false
		}
		denom, ok := o["asset_denom"].(string)
		if !ok || !denomPattern.MatchString(denom) || denom == "upnyx" || seen[denom] || !assets[denom] {
			return nil, false
		}
		seen[denom] = true
		p, ok := numberAmount(o["pnyx_reserve"])
		if !ok || p.Sign() <= 0 {
			return nil, false
		}
		a, ok := numberAmount(o["asset_reserve"])
		if !ok || a.Sign() <= 0 {
			return nil, false
		}
		totalShares, ok := numberAmount(o["total_shares"])
		if !ok || totalShares.Sign() <= 0 {
			return nil, false
		}
		poolShares[denom] = totalShares
		add(claims, "upnyx", p)
		add(claims, denom, a)
	}
	positions, ok := raw["lp_positions"].([]any)
	if !ok {
		return nil, false
	}
	totals := map[string]*big.Int{}
	seenPositions := map[string]bool{}
	for _, r := range positions {
		o, ok := r.(map[string]any)
		if !ok {
			return nil, false
		}
		denom, ok := o["asset_denom"].(string)
		if !ok || poolShares[denom] == nil {
			return nil, false
		}
		provider, ok := o["provider"].(string)
		if !ok {
			return nil, false
		}
		if _, ok := accountAddress(provider); !ok {
			return nil, false
		}
		shares, ok := numberAmount(o["shares"])
		if !ok || shares.Sign() <= 0 {
			return nil, false
		}
		key := denom + "\x00" + provider
		if seenPositions[key] {
			return nil, false
		}
		seenPositions[key] = true
		add(totals, denom, shares)
	}
	for denom, totalShares := range poolShares {
		if zero(totals[denom]).Cmp(totalShares) != 0 {
			return nil, false
		}
	}
	return claims, true
}
func (s *state) compareDEX(m Manifest, b map[string]map[string]*big.Int, claims map[string]*big.Int) {
	want := map[string]*big.Int{}
	for _, c := range m.DEXCustody {
		n, ok := amount(c.Amount)
		if !ok {
			continue
		}
		want[c.Denom] = n
	}
	if !coinMapsEqual(want, claims) {
		s.fail("dex-custody", "manifest-dex-claim-mismatch")
	}
	if !coinMapsEqual(b[dexAddress], claims) {
		s.fail("dex-custody", "dex-bank-custody-mismatch")
	}
}

type moduleAccountSpec struct {
	address     string
	permissions []string
}

func (s *state) verifyModuleIsolation(m Manifest, app map[string]any, b map[string]map[string]*big.Int, _ map[string]*big.Int) {
	forbidden := []string{feeCollectorAddress, wasmAddress, transferAddress}
	for _, addr := range forbidden {
		if len(b[addr]) > 0 {
			s.fail("module-isolation", "forbidden-funded-module")
		}
	}
	auth, ok := app["auth"].(map[string]any)
	if !ok {
		s.fail("module-isolation", "missing-auth-state")
		return
	}
	accounts, ok := auth["accounts"].([]any)
	if !ok && auth["accounts"] != nil {
		s.fail("module-isolation", "invalid-auth-accounts")
		return
	}
	known := map[string]moduleAccountSpec{
		"truedemocracy": {address: trueDemocracyAddress, permissions: []string{"minter", "burner"}},
		"dex":           {address: dexAddress, permissions: []string{"burner"}},
		"fee_collector": {address: feeCollectorAddress, permissions: []string{}},
		"wasm":          {address: wasmAddress, permissions: []string{"burner"}},
		"transfer":      {address: transferAddress, permissions: []string{"minter", "burner"}},
	}
	seenNames := map[string]bool{}
	seenAddresses := map[string]bool{}
	baseAccounts := map[string]bool{}
	for _, r := range accounts {
		o, ok := r.(map[string]any)
		if !ok {
			s.fail("module-isolation", "invalid-auth-account")
			continue
		}
		name, _ := o["name"].(string)
		if name == "" {
			address, _ := o["address"].(string)
			if _, valid := accountAddress(address); !valid || seenAddresses[address] {
				s.fail("module-isolation", "duplicate-or-invalid-auth-account")
				continue
			}
			seenAddresses[address] = true
			baseAccounts[address] = true
			continue
		}
		base, _ := o["base_account"].(map[string]any)
		addr, _ := base["address"].(string)
		want, recognized := known[name]
		if !recognized {
			s.fail("module-isolation", "unknown-module-account")
			continue
		}
		if seenNames[name] || seenAddresses[addr] {
			s.fail("module-isolation", "duplicate-module-account")
		}
		seenNames[name] = true
		seenAddresses[addr] = true
		if accountType, _ := o["@type"].(string); accountType != "/cosmos.auth.v1beta1.ModuleAccount" || addr != want.address {
			s.fail("module-isolation", "module-address-mismatch")
		}
		if !stringListEqual(o["permissions"], want.permissions) {
			s.fail("module-isolation", "module-permissions-mismatch")
		}
	}
	for name := range known {
		if !seenNames[name] {
			s.fail("module-isolation", "missing-module-account")
		}
	}
	for _, validator := range m.Validators {
		if !baseAccounts[validator.OperatorAddress] {
			s.fail("module-isolation", "missing-validator-operator-account")
		}
	}
}

func stringListEqual(raw any, want []string) bool {
	values, ok := raw.([]any)
	if !ok || len(values) != len(want) {
		return false
	}
	seen := map[string]bool{}
	for _, value := range values {
		permission, ok := value.(string)
		if !ok || seen[permission] {
			return false
		}
		seen[permission] = true
	}
	for _, permission := range want {
		if !seen[permission] {
			return false
		}
	}
	return true
}

func accountAddress(value string) ([]byte, bool) {
	prefix, decoded, err := bech32.DecodeAndConvert(value)
	return decoded, err == nil && prefix == "truerepublic" && len(decoded) == 20
}

func isModuleAddress(value string) bool {
	switch value {
	case trueDemocracyAddress, dexAddress, feeCollectorAddress, wasmAddress, transferAddress:
		return true
	default:
		return false
	}
}

func amount(v string) (*big.Int, bool) {
	if v == "" || (len(v) > 1 && v[0] == '0') {
		return nil, false
	}
	for _, r := range v {
		if r < '0' || r > '9' {
			return nil, false
		}
	}
	n, ok := new(big.Int).SetString(v, 10)
	return n, ok
}
func boundedAmount(v, max string) (*big.Int, bool) {
	n, ok := amount(v)
	return n, ok && n.Cmp(mustAmount(max)) <= 0
}
func mustAmount(v string) *big.Int { n, _ := new(big.Int).SetString(v, 10); return n }
func numberAmount(v any) (*big.Int, bool) {
	switch x := v.(type) {
	case json.Number:
		return amount(x.String())
	case string:
		return amount(x)
	default:
		return nil, false
	}
}
func int64Value(v any) (int64, bool) {
	n, ok := numberAmount(v)
	if !ok || !n.IsInt64() {
		return 0, false
	}
	return n.Int64(), true
}
func coinMap(coins []Coin, positive bool) (map[string]*big.Int, bool) {
	r := map[string]*big.Int{}
	last := ""
	for _, c := range coins {
		n, ok := amount(c.Amount)
		if !ok || (positive && n.Sign() <= 0) || !denomPattern.MatchString(c.Denom) || c.Denom <= last {
			return r, false
		}
		r[c.Denom] = n
		last = c.Denom
	}
	return r, true
}
func coinsFrom(v any) (map[string]*big.Int, bool) {
	a, ok := v.([]any)
	if !ok {
		return nil, false
	}
	coins := make([]Coin, 0, len(a))
	for _, r := range a {
		var c Coin
		if decodeInto(r, &c) != nil {
			return nil, false
		}
		coins = append(coins, c)
	}
	return coinMap(coins, true)
}
func addCoins(dst, src map[string]*big.Int) {
	keys := make([]string, 0, len(src))
	for k := range src {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		add(dst, k, src[k])
	}
}
func add(m map[string]*big.Int, k string, n *big.Int) {
	if m[k] == nil {
		m[k] = new(big.Int)
	}
	m[k].Add(m[k], n)
}
func zero(n *big.Int) *big.Int {
	if n == nil {
		return new(big.Int)
	}
	return new(big.Int).Set(n)
}
func coinMapsEqual(a, b map[string]*big.Int) bool {
	for k, v := range a {
		if v.Sign() != 0 && zero(b[k]).Cmp(v) != 0 {
			return false
		}
	}
	for k, v := range b {
		if v.Sign() != 0 && zero(a[k]).Cmp(v) != 0 {
			return false
		}
	}
	return true
}

// SortedCoins returns canonical denomination order for callers producing a manifest.
func SortedCoins(coins map[string]string) []Coin {
	keys := make([]string, 0, len(coins))
	for k := range coins {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]Coin, 0, len(keys))
	for _, k := range keys {
		out = append(out, Coin{Denom: k, Amount: coins[k]})
	}
	return out
}
