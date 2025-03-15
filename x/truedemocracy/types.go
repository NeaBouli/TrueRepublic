package truedemocracy

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    "time"
)

type Domain struct {
    Name          string
    Admin         sdk.AccAddress
    Members       []Member
    Treasury      sdk.Coins
    Issues        []Issue
    Options       DomainOptions
    PermissionReg []string
    GlobalKeys    map[string]string
}

type DomainOptions struct {
    AdminElectable   bool
    AnyoneCanJoin    bool
    OnlyAdminIssues  bool
    OnlyAdminSuggest bool
    CoinBurnRequired bool
    BurnAmount       sdk.Int
    MaxGreenZone     int
    DwellTime        time.Duration
}

type Issue struct {
    Name         string
    Suggestions  []Suggestion
    Stones       sdk.Int
    Consensus    string
    Expiry       time.Time
    Created      time.Time
    ShortDesc    string
    ExternalLink string
}

type Suggestion struct {
    Name         string
    Ratings      []Rating
    Stones       sdk.Int
    Zone         string
    Created      time.Time
    ShortDesc    string
    ExternalLink string
}

type Rating struct {
    VoterKey string
    Value    int
}

type Node struct {
    Address      string
    PubKey       string
    Staked       sdk.Coins
    DomainOrigin string
    LastActive   time.Time
}

type Member struct {
    Avatar    string
    IsActive  bool
    VotesCast int
    Weight    sdk.Dec
}
