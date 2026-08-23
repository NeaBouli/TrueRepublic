// Package genesisevidence verifies a candidate TrueRepublic rollout genesis
// offline. Verification is deterministic and binds the manifest to the exact
// genesis bytes; it does not read keys, contact a node, or mutate state.
package genesisevidence

const (
	ManifestSchema         = "truerepublic.rollout-genesis-manifest/v1"
	EvidenceSchema         = "truerepublic.rollout-genesis-evidence/v1"
	MaxGenesisBytes        = 64 << 20
	MaxManifestBytes       = 1 << 20
	MaxPNYXSupply          = "21000000000000"
	StakeUnit        int64 = 100_000_000_000
	MaxPowerLimit    int64 = 1_048_576
)

type Manifest struct {
	Schema                string       `json:"schema"`
	SourceCommit          string       `json:"source_commit"`
	DaemonVersion         string       `json:"daemon_version"`
	ChainID               string       `json:"chain_id"`
	GenesisSHA256         string       `json:"genesis_sha256"`
	MaxValidatorPower     int64        `json:"max_validator_power"`
	Validators            []Validator  `json:"validators"`
	Allocations           []Allocation `json:"allocations"`
	TotalSupplyUPNYX      string       `json:"total_supply_upnyx"`
	GovernanceEscrowUPNYX string       `json:"governance_escrow_upnyx"`
	DEXCustody            []Coin       `json:"dex_custody"`
}

type Validator struct {
	OperatorAddress string `json:"operator_address"`
	ConsensusPubKey string `json:"consensus_pubkey"`
	StakeUPNYX      string `json:"stake_upnyx"`
	Power           int64  `json:"power"`
}

type Coin struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type Allocation struct {
	Address string `json:"address"`
	Coins   []Coin `json:"coins"`
}

type Evidence struct {
	Schema         string  `json:"schema"`
	Valid          bool    `json:"valid"`
	ManifestSHA256 string  `json:"manifest_sha256"`
	GenesisSHA256  string  `json:"genesis_sha256"`
	Checks         []Check `json:"checks"`
}

type Check struct {
	Name       string   `json:"name"`
	Pass       bool     `json:"pass"`
	Violations []string `json:"violations"`
}

const (
	trueDemocracyAddress = "truerepublic1kre208nn6ucjxjcs9mmq2qkfrn33g6223edr47"
	dexAddress           = "truerepublic1n58mly6f7er0zs6swtetqgfqs36jaarqluymru"
	feeCollectorAddress  = "truerepublic17xpfvakm2amg962yls6f84z3kell8c5l5m3sxr"
	wasmAddress          = "truerepublic1xds4f0m87ajl3a6az6s2enhxrd0wta485yw67p"
	transferAddress      = "truerepublic1yl6hdjhmkf37639730gffanpzndzdpmh2aye6a"
)
