package truedemocracy

import (
    "cosmossdk.io/math"
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

const ModuleName = "truedemocracy"

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

type GenesisState struct {
    Domains []Domain `json:"domains"`
}

func RegisterCodec(cdc *codec.LegacyAmino) {
    cdc.RegisterConcrete(Domain{}, "truedemocracy/Domain", nil)
    cdc.RegisterConcrete(DomainOptions{}, "truedemocracy/DomainOptions", nil)
    cdc.RegisterConcrete(Issue{}, "truedemocracy/Issue", nil)
    cdc.RegisterConcrete(Suggestion{}, "truedemocracy/Suggestion", nil)
    cdc.RegisterConcrete(Rating{}, "truedemocracy/Rating", nil)
    cdc.RegisterConcrete(DexPool{}, "truedemocracy/DexPool", nil)
    cdc.RegisterConcrete(GenesisState{}, "truedemocracy/GenesisState", nil)
}

func DefaultGenesisState() GenesisState {
    return GenesisState{
        Domains: []Domain{
            {
                Name:          "TestParty",
                Admin:         sdk.AccAddress("admin1"),
                Members:       []string{"user1", "user2", "user3"},
                Treasury:      sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500000)),
                Options:       DomainOptions{AdminElectable: true, AnyoneCanJoin: false},
                PermissionReg: []string{},
            },
        },
    }
}
