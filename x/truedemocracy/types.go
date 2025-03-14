package truedemocracy

import sdk "github.com/cosmos/cosmos-sdk/types"

type Domain struct {
    Name          string
    Admin         sdk.AccAddress
    Members       []string
    Treasury      sdk.Coins
    Issues        []Issue
    Options       DomainOptions
    PermissionReg []string
}

type DomainOptions struct {
    AdminElectable   bool
    AnyoneCanJoin    bool
    OnlyAdminIssues  bool
    CoinBurnRequired bool
}

type Issue struct {
    Name        string
    Suggestions []Suggestion
    Stones      int
}

type Suggestion struct {
    Name    string
    Ratings []Rating
    Stones  int
}

type Rating struct {
    Voter string
    Value int
}

type Node struct {
    Address string
    PubKey  string
}
