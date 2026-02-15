package truedemocracy

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// All message types implement sdk.Msg (proto.Message) via stubs,
// plus ValidateBasic and GetSigners for legacy amino tx support.

// --- MsgCreateDomain ---

type MsgCreateDomain struct {
	Name         string         `json:"name"`
	Admin        sdk.AccAddress `json:"admin"`
	InitialCoins sdk.Coins      `json:"initial_coins"`
}

func (m *MsgCreateDomain) ProtoMessage()             {}
func (m *MsgCreateDomain) Reset()                    { *m = MsgCreateDomain{} }
func (m *MsgCreateDomain) String() string            { b, _ := json.Marshal(m); return string(b) }
func (m MsgCreateDomain) Route() string              { return ModuleName }
func (m MsgCreateDomain) Type() string               { return "create_domain" }
func (m MsgCreateDomain) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Admin} }
func (m MsgCreateDomain) ValidateBasic() error {
	if m.Name == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("name is required")
	}
	if m.Admin.Empty() {
		return sdkerrors.ErrInvalidAddress.Wrap("admin address is required")
	}
	return nil
}

// --- MsgSubmitProposal ---

type MsgSubmitProposal struct {
	Sender         sdk.AccAddress `json:"sender"`
	DomainName     string         `json:"domain_name"`
	IssueName      string         `json:"issue_name"`
	SuggestionName string         `json:"suggestion_name"`
	Creator        string         `json:"creator"`
	Fee            sdk.Coins      `json:"fee"`
	ExternalLink   string         `json:"external_link"`
}

func (m *MsgSubmitProposal) ProtoMessage()             {}
func (m *MsgSubmitProposal) Reset()                    { *m = MsgSubmitProposal{} }
func (m *MsgSubmitProposal) String() string            { b, _ := json.Marshal(m); return string(b) }
func (m MsgSubmitProposal) Route() string              { return ModuleName }
func (m MsgSubmitProposal) Type() string               { return "submit_proposal" }
func (m MsgSubmitProposal) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
func (m MsgSubmitProposal) ValidateBasic() error {
	if m.DomainName == "" || m.IssueName == "" || m.SuggestionName == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("domain, issue, and suggestion names are required")
	}
	return nil
}

// --- MsgRegisterValidator ---

type MsgRegisterValidator struct {
	Sender       sdk.AccAddress `json:"sender"`
	OperatorAddr string         `json:"operator_addr"`
	PubKey       string         `json:"pub_key"` // hex-encoded 32 bytes
	Stake        sdk.Coins      `json:"stake"`
	DomainName   string         `json:"domain_name"`
}

func (m *MsgRegisterValidator) ProtoMessage()             {}
func (m *MsgRegisterValidator) Reset()                    { *m = MsgRegisterValidator{} }
func (m *MsgRegisterValidator) String() string            { b, _ := json.Marshal(m); return string(b) }
func (m MsgRegisterValidator) Route() string              { return ModuleName }
func (m MsgRegisterValidator) Type() string               { return "register_validator" }
func (m MsgRegisterValidator) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
func (m MsgRegisterValidator) ValidateBasic() error {
	if m.OperatorAddr == "" || m.PubKey == "" || m.DomainName == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("operator_addr, pub_key, and domain_name are required")
	}
	return nil
}

// --- MsgWithdrawStake ---

type MsgWithdrawStake struct {
	Sender       sdk.AccAddress `json:"sender"`
	OperatorAddr string         `json:"operator_addr"`
	Amount       int64          `json:"amount"`
}

func (m *MsgWithdrawStake) ProtoMessage()             {}
func (m *MsgWithdrawStake) Reset()                    { *m = MsgWithdrawStake{} }
func (m *MsgWithdrawStake) String() string            { b, _ := json.Marshal(m); return string(b) }
func (m MsgWithdrawStake) Route() string              { return ModuleName }
func (m MsgWithdrawStake) Type() string               { return "withdraw_stake" }
func (m MsgWithdrawStake) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
func (m MsgWithdrawStake) ValidateBasic() error {
	if m.OperatorAddr == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("operator_addr is required")
	}
	if m.Amount <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("amount must be positive")
	}
	return nil
}

// --- MsgRemoveValidator ---

type MsgRemoveValidator struct {
	Sender       sdk.AccAddress `json:"sender"`
	OperatorAddr string         `json:"operator_addr"`
}

func (m *MsgRemoveValidator) ProtoMessage()             {}
func (m *MsgRemoveValidator) Reset()                    { *m = MsgRemoveValidator{} }
func (m *MsgRemoveValidator) String() string            { b, _ := json.Marshal(m); return string(b) }
func (m MsgRemoveValidator) Route() string              { return ModuleName }
func (m MsgRemoveValidator) Type() string               { return "remove_validator" }
func (m MsgRemoveValidator) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
func (m MsgRemoveValidator) ValidateBasic() error {
	if m.OperatorAddr == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("operator_addr is required")
	}
	return nil
}

// --- MsgUnjail ---

type MsgUnjail struct {
	Sender       sdk.AccAddress `json:"sender"`
	OperatorAddr string         `json:"operator_addr"`
}

func (m *MsgUnjail) ProtoMessage()             {}
func (m *MsgUnjail) Reset()                    { *m = MsgUnjail{} }
func (m *MsgUnjail) String() string            { b, _ := json.Marshal(m); return string(b) }
func (m MsgUnjail) Route() string              { return ModuleName }
func (m MsgUnjail) Type() string               { return "unjail" }
func (m MsgUnjail) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
func (m MsgUnjail) ValidateBasic() error {
	if m.OperatorAddr == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("operator_addr is required")
	}
	return nil
}

// --- MsgJoinPermissionRegister ---

type MsgJoinPermissionRegister struct {
	Sender       sdk.AccAddress `json:"sender"`
	DomainName   string         `json:"domain_name"`
	MemberAddr   string         `json:"member_addr"`
	DomainPubKey string         `json:"domain_pub_key"` // hex-encoded
}

func (m *MsgJoinPermissionRegister) ProtoMessage()             {}
func (m *MsgJoinPermissionRegister) Reset()                    { *m = MsgJoinPermissionRegister{} }
func (m *MsgJoinPermissionRegister) String() string            { b, _ := json.Marshal(m); return string(b) }
func (m MsgJoinPermissionRegister) Route() string              { return ModuleName }
func (m MsgJoinPermissionRegister) Type() string               { return "join_permission_register" }
func (m MsgJoinPermissionRegister) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
func (m MsgJoinPermissionRegister) ValidateBasic() error {
	if m.DomainName == "" || m.MemberAddr == "" || m.DomainPubKey == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("domain_name, member_addr, and domain_pub_key are required")
	}
	return nil
}

// --- MsgPurgePermissionRegister ---

type MsgPurgePermissionRegister struct {
	Caller     sdk.AccAddress `json:"caller"`
	DomainName string         `json:"domain_name"`
}

func (m *MsgPurgePermissionRegister) ProtoMessage()             {}
func (m *MsgPurgePermissionRegister) Reset()                    { *m = MsgPurgePermissionRegister{} }
func (m *MsgPurgePermissionRegister) String() string            { b, _ := json.Marshal(m); return string(b) }
func (m MsgPurgePermissionRegister) Route() string              { return ModuleName }
func (m MsgPurgePermissionRegister) Type() string               { return "purge_permission_register" }
func (m MsgPurgePermissionRegister) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Caller} }
func (m MsgPurgePermissionRegister) ValidateBasic() error {
	if m.DomainName == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("domain_name is required")
	}
	return nil
}

// --- MsgPlaceStoneOnIssue ---

type MsgPlaceStoneOnIssue struct {
	Sender     sdk.AccAddress `json:"sender"`
	DomainName string         `json:"domain_name"`
	IssueName  string         `json:"issue_name"`
	MemberAddr string         `json:"member_addr"`
}

func (m *MsgPlaceStoneOnIssue) ProtoMessage()             {}
func (m *MsgPlaceStoneOnIssue) Reset()                    { *m = MsgPlaceStoneOnIssue{} }
func (m *MsgPlaceStoneOnIssue) String() string            { b, _ := json.Marshal(m); return string(b) }
func (m MsgPlaceStoneOnIssue) Route() string              { return ModuleName }
func (m MsgPlaceStoneOnIssue) Type() string               { return "place_stone_issue" }
func (m MsgPlaceStoneOnIssue) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
func (m MsgPlaceStoneOnIssue) ValidateBasic() error {
	if m.DomainName == "" || m.IssueName == "" || m.MemberAddr == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("domain_name, issue_name, and member_addr are required")
	}
	return nil
}

// --- MsgPlaceStoneOnSuggestion ---

type MsgPlaceStoneOnSuggestion struct {
	Sender         sdk.AccAddress `json:"sender"`
	DomainName     string         `json:"domain_name"`
	IssueName      string         `json:"issue_name"`
	SuggestionName string         `json:"suggestion_name"`
	MemberAddr     string         `json:"member_addr"`
}

func (m *MsgPlaceStoneOnSuggestion) ProtoMessage()             {}
func (m *MsgPlaceStoneOnSuggestion) Reset()                    { *m = MsgPlaceStoneOnSuggestion{} }
func (m *MsgPlaceStoneOnSuggestion) String() string            { b, _ := json.Marshal(m); return string(b) }
func (m MsgPlaceStoneOnSuggestion) Route() string              { return ModuleName }
func (m MsgPlaceStoneOnSuggestion) Type() string               { return "place_stone_suggestion" }
func (m MsgPlaceStoneOnSuggestion) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
func (m MsgPlaceStoneOnSuggestion) ValidateBasic() error {
	if m.DomainName == "" || m.IssueName == "" || m.SuggestionName == "" || m.MemberAddr == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("domain_name, issue_name, suggestion_name, and member_addr are required")
	}
	return nil
}

// --- MsgPlaceStoneOnMember ---

type MsgPlaceStoneOnMember struct {
	Sender       sdk.AccAddress `json:"sender"`
	DomainName   string         `json:"domain_name"`
	TargetMember string         `json:"target_member"`
	VoterAddr    string         `json:"voter_addr"`
}

func (m *MsgPlaceStoneOnMember) ProtoMessage()             {}
func (m *MsgPlaceStoneOnMember) Reset()                    { *m = MsgPlaceStoneOnMember{} }
func (m *MsgPlaceStoneOnMember) String() string            { b, _ := json.Marshal(m); return string(b) }
func (m MsgPlaceStoneOnMember) Route() string              { return ModuleName }
func (m MsgPlaceStoneOnMember) Type() string               { return "place_stone_member" }
func (m MsgPlaceStoneOnMember) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
func (m MsgPlaceStoneOnMember) ValidateBasic() error {
	if m.DomainName == "" || m.TargetMember == "" || m.VoterAddr == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("domain_name, target_member, and voter_addr are required")
	}
	return nil
}

// --- MsgVoteToExclude ---

type MsgVoteToExclude struct {
	Sender       sdk.AccAddress `json:"sender"`
	DomainName   string         `json:"domain_name"`
	TargetMember string         `json:"target_member"`
	VoterAddr    string         `json:"voter_addr"`
}

func (m *MsgVoteToExclude) ProtoMessage()             {}
func (m *MsgVoteToExclude) Reset()                    { *m = MsgVoteToExclude{} }
func (m *MsgVoteToExclude) String() string            { b, _ := json.Marshal(m); return string(b) }
func (m MsgVoteToExclude) Route() string              { return ModuleName }
func (m MsgVoteToExclude) Type() string               { return "vote_exclude" }
func (m MsgVoteToExclude) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
func (m MsgVoteToExclude) ValidateBasic() error {
	if m.DomainName == "" || m.TargetMember == "" || m.VoterAddr == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("domain_name, target_member, and voter_addr are required")
	}
	return nil
}

// --- MsgVoteToDelete ---

type MsgVoteToDelete struct {
	Sender         sdk.AccAddress `json:"sender"`
	DomainName     string         `json:"domain_name"`
	IssueName      string         `json:"issue_name"`
	SuggestionName string         `json:"suggestion_name"`
	MemberAddr     string         `json:"member_addr"`
}

func (m *MsgVoteToDelete) ProtoMessage()             {}
func (m *MsgVoteToDelete) Reset()                    { *m = MsgVoteToDelete{} }
func (m *MsgVoteToDelete) String() string            { b, _ := json.Marshal(m); return string(b) }
func (m MsgVoteToDelete) Route() string              { return ModuleName }
func (m MsgVoteToDelete) Type() string               { return "vote_delete" }
func (m MsgVoteToDelete) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
func (m MsgVoteToDelete) ValidateBasic() error {
	if m.DomainName == "" || m.IssueName == "" || m.SuggestionName == "" || m.MemberAddr == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("domain_name, issue_name, suggestion_name, and member_addr are required")
	}
	return nil
}
