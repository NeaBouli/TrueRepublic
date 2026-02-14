package truedemocracy

import (
    "cosmossdk.io/math"
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

const ModuleName = "truedemocracy"

// Proof of Domain slashing and reward parameters.
const (
	SlashFractionDowntime   int64 = 1   // 1% of stake slashed for downtime
	SlashFractionDoubleSign int64 = 5   // 5% of stake slashed for equivocation
	DowntimeJailDuration    int64 = 600 // seconds (10 min)
	SignedBlocksWindow      int64 = 100 // blocks tracked for liveness
	MinSignedPerWindow      int64 = 50  // must sign ≥50% of blocks in window
	RewardInterval          int64 = 3600 // distribute rewards every hour (seconds)
)

type Domain struct {
    Name          string         `json:"name"`
    Admin         sdk.AccAddress `json:"admin"`
    Members       []string       `json:"members"`
    Treasury      sdk.Coins      `json:"treasury"`
    Issues        []Issue        `json:"issues"`
    Options       DomainOptions  `json:"options"`
    PermissionReg []string       `json:"permission_reg"`
}

type DomainOptions struct {
    AdminElectable   bool `json:"admin_electable"`
    AnyoneCanJoin    bool `json:"anyone_can_join"`
    OnlyAdminIssues  bool `json:"only_admin_issues"`
    CoinBurnRequired bool `json:"coin_burn_required"`
}

type Issue struct {
    Name        string       `json:"name"`
    Stones      int          `json:"stones"`
    Suggestions []Suggestion `json:"suggestions"`
}

type Suggestion struct {
    Name       string   `json:"name"`
    Creator    string   `json:"creator"`
    Stones     int      `json:"stones"`
    Ratings    []Rating `json:"ratings"`
    Color      string   `json:"color"`
    DwellTime  int64    `json:"dwell_time"`
}

type Rating struct {
    Voter string `json:"voter"`
    Value int    `json:"value"`
}

type DexPool struct {
    PnyxReserve  math.Int `json:"pnyx_reserve"`
    AssetReserve math.Int `json:"asset_reserve"`
    AssetType    string   `json:"asset_type"`
}

// Validator represents an active Proof of Domain validator node.
type Validator struct {
    OperatorAddr string    `json:"operator_addr"`
    PubKey       []byte    `json:"pub_key"`
    Stake        sdk.Coins `json:"stake"`
    Domains      []string  `json:"domains"`
    Power        int64     `json:"power"`
    Jailed       bool      `json:"jailed"`
    JailedUntil  int64     `json:"jailed_until"`
    MissedBlocks int64     `json:"missed_blocks"`
}

// GenesisValidator is the genesis-file representation of a validator.
type GenesisValidator struct {
    OperatorAddr string `json:"operator_addr"`
    PubKey       []byte `json:"pub_key"`
    Stake        int64  `json:"stake"`
    Domain       string `json:"domain"`
}

type GenesisState struct {
    Domains    []Domain           `json:"domains"`
    Validators []GenesisValidator `json:"validators"`
}

func RegisterCodec(cdc *codec.LegacyAmino) {
    cdc.RegisterConcrete(Domain{}, "truedemocracy/Domain", nil)
    cdc.RegisterConcrete(DomainOptions{}, "truedemocracy/DomainOptions", nil)
    cdc.RegisterConcrete(Issue{}, "truedemocracy/Issue", nil)
    cdc.RegisterConcrete(Suggestion{}, "truedemocracy/Suggestion", nil)
    cdc.RegisterConcrete(Rating{}, "truedemocracy/Rating", nil)
    cdc.RegisterConcrete(DexPool{}, "truedemocracy/DexPool", nil)
    cdc.RegisterConcrete(GenesisState{}, "truedemocracy/GenesisState", nil)
    cdc.RegisterConcrete(Validator{}, "truedemocracy/Validator", nil)
    cdc.RegisterConcrete(GenesisValidator{}, "truedemocracy/GenesisValidator", nil)
}

func DefaultGenesisState() GenesisState {
    // Deterministic key for the default test validator.
    privKey := ed25519.GenPrivKeyFromSecret([]byte("test-validator-0"))
    pubKey := privKey.PubKey().Bytes()

    return GenesisState{
        Domains: []Domain{
            {
                Name:          "TestParty",
                Admin:         sdk.AccAddress("admin1"),
                Members:       []string{"user1", "user2", "user3", "validator1"},
                Treasury:      sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500000)),
                Options:       DomainOptions{AdminElectable: true, AnyoneCanJoin: false},
                PermissionReg: []string{},
            },
        },
        Validators: []GenesisValidator{
            {
                OperatorAddr: "validator1",
                PubKey:       pubKey,
                Stake:        100_000,
                Domain:       "TestParty",
            },
        },
    }
}
