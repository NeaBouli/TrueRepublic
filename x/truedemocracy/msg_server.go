package truedemocracy

import (
	"context"
	"encoding/hex"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	gogoproto "github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"
)

// ---------------------------------------------------------------------------
// Response types
// ---------------------------------------------------------------------------

type MsgCreateDomainResponse struct{}

func (*MsgCreateDomainResponse) ProtoMessage()             {}
func (*MsgCreateDomainResponse) Reset()                    {}
func (*MsgCreateDomainResponse) String() string            { return "MsgCreateDomainResponse" }

type MsgSubmitProposalResponse struct{}

func (*MsgSubmitProposalResponse) ProtoMessage()           {}
func (*MsgSubmitProposalResponse) Reset()                  {}
func (*MsgSubmitProposalResponse) String() string          { return "MsgSubmitProposalResponse" }

type MsgRegisterValidatorResponse struct{}

func (*MsgRegisterValidatorResponse) ProtoMessage()        {}
func (*MsgRegisterValidatorResponse) Reset()               {}
func (*MsgRegisterValidatorResponse) String() string       { return "MsgRegisterValidatorResponse" }

type MsgWithdrawStakeResponse struct{}

func (*MsgWithdrawStakeResponse) ProtoMessage()            {}
func (*MsgWithdrawStakeResponse) Reset()                   {}
func (*MsgWithdrawStakeResponse) String() string           { return "MsgWithdrawStakeResponse" }

type MsgRemoveValidatorResponse struct{}

func (*MsgRemoveValidatorResponse) ProtoMessage()          {}
func (*MsgRemoveValidatorResponse) Reset()                 {}
func (*MsgRemoveValidatorResponse) String() string         { return "MsgRemoveValidatorResponse" }

type MsgUnjailResponse struct{}

func (*MsgUnjailResponse) ProtoMessage()                   {}
func (*MsgUnjailResponse) Reset()                          {}
func (*MsgUnjailResponse) String() string                  { return "MsgUnjailResponse" }

type MsgJoinPermissionRegisterResponse struct{}

func (*MsgJoinPermissionRegisterResponse) ProtoMessage()   {}
func (*MsgJoinPermissionRegisterResponse) Reset()          {}
func (*MsgJoinPermissionRegisterResponse) String() string  { return "MsgJoinPermissionRegisterResponse" }

type MsgPurgePermissionRegisterResponse struct{}

func (*MsgPurgePermissionRegisterResponse) ProtoMessage()  {}
func (*MsgPurgePermissionRegisterResponse) Reset()         {}
func (*MsgPurgePermissionRegisterResponse) String() string { return "MsgPurgePermissionRegisterResponse" }

type MsgPlaceStoneOnIssueResponse struct{}

func (*MsgPlaceStoneOnIssueResponse) ProtoMessage()        {}
func (*MsgPlaceStoneOnIssueResponse) Reset()               {}
func (*MsgPlaceStoneOnIssueResponse) String() string       { return "MsgPlaceStoneOnIssueResponse" }

type MsgPlaceStoneOnSuggestionResponse struct{}

func (*MsgPlaceStoneOnSuggestionResponse) ProtoMessage()   {}
func (*MsgPlaceStoneOnSuggestionResponse) Reset()          {}
func (*MsgPlaceStoneOnSuggestionResponse) String() string  { return "MsgPlaceStoneOnSuggestionResponse" }

type MsgPlaceStoneOnMemberResponse struct{}

func (*MsgPlaceStoneOnMemberResponse) ProtoMessage()       {}
func (*MsgPlaceStoneOnMemberResponse) Reset()              {}
func (*MsgPlaceStoneOnMemberResponse) String() string      { return "MsgPlaceStoneOnMemberResponse" }

type MsgVoteToExcludeResponse struct{}

func (*MsgVoteToExcludeResponse) ProtoMessage()            {}
func (*MsgVoteToExcludeResponse) Reset()                   {}
func (*MsgVoteToExcludeResponse) String() string           { return "MsgVoteToExcludeResponse" }

type MsgVoteToDeleteResponse struct{}

func (*MsgVoteToDeleteResponse) ProtoMessage()             {}
func (*MsgVoteToDeleteResponse) Reset()                    {}
func (*MsgVoteToDeleteResponse) String() string            { return "MsgVoteToDeleteResponse" }

type MsgRateProposalResponse struct{}

func (*MsgRateProposalResponse) ProtoMessage()             {}
func (*MsgRateProposalResponse) Reset()                    {}
func (*MsgRateProposalResponse) String() string            { return "MsgRateProposalResponse" }

type MsgAddMemberResponse struct{}

func (*MsgAddMemberResponse) ProtoMessage()             {}
func (*MsgAddMemberResponse) Reset()                    {}
func (*MsgAddMemberResponse) String() string            { return "MsgAddMemberResponse" }

type MsgOnboardToDomainResponse struct{}

func (*MsgOnboardToDomainResponse) ProtoMessage()       {}
func (*MsgOnboardToDomainResponse) Reset()              {}
func (*MsgOnboardToDomainResponse) String() string      { return "MsgOnboardToDomainResponse" }

type MsgApproveOnboardingResponse struct{}

func (*MsgApproveOnboardingResponse) ProtoMessage()     {}
func (*MsgApproveOnboardingResponse) Reset()            {}
func (*MsgApproveOnboardingResponse) String() string    { return "MsgApproveOnboardingResponse" }

type MsgRegisterIdentityResponse struct{}

func (*MsgRegisterIdentityResponse) ProtoMessage()      {}
func (*MsgRegisterIdentityResponse) Reset()              {}
func (*MsgRegisterIdentityResponse) String() string      { return "MsgRegisterIdentityResponse" }

type MsgRejectOnboardingResponse struct{}

func (*MsgRejectOnboardingResponse) ProtoMessage()      {}
func (*MsgRejectOnboardingResponse) Reset()             {}
func (*MsgRejectOnboardingResponse) String() string     { return "MsgRejectOnboardingResponse" }

type MsgCastElectionVoteResponse struct{}

func (*MsgCastElectionVoteResponse) ProtoMessage()         {}
func (*MsgCastElectionVoteResponse) Reset()                {}
func (*MsgCastElectionVoteResponse) String() string        { return "MsgCastElectionVoteResponse" }

// ---------------------------------------------------------------------------
// Register response types with gogoproto
// ---------------------------------------------------------------------------

func init() {
	// Register Msg types with gogoproto for MsgServiceRouter resolution.
	gogoproto.RegisterType((*MsgCreateDomain)(nil), "truedemocracy.MsgCreateDomain")
	gogoproto.RegisterType((*MsgSubmitProposal)(nil), "truedemocracy.MsgSubmitProposal")
	gogoproto.RegisterType((*MsgRegisterValidator)(nil), "truedemocracy.MsgRegisterValidator")
	gogoproto.RegisterType((*MsgWithdrawStake)(nil), "truedemocracy.MsgWithdrawStake")
	gogoproto.RegisterType((*MsgRemoveValidator)(nil), "truedemocracy.MsgRemoveValidator")
	gogoproto.RegisterType((*MsgUnjail)(nil), "truedemocracy.MsgUnjail")
	gogoproto.RegisterType((*MsgJoinPermissionRegister)(nil), "truedemocracy.MsgJoinPermissionRegister")
	gogoproto.RegisterType((*MsgPurgePermissionRegister)(nil), "truedemocracy.MsgPurgePermissionRegister")
	gogoproto.RegisterType((*MsgPlaceStoneOnIssue)(nil), "truedemocracy.MsgPlaceStoneOnIssue")
	gogoproto.RegisterType((*MsgPlaceStoneOnSuggestion)(nil), "truedemocracy.MsgPlaceStoneOnSuggestion")
	gogoproto.RegisterType((*MsgPlaceStoneOnMember)(nil), "truedemocracy.MsgPlaceStoneOnMember")
	gogoproto.RegisterType((*MsgVoteToExclude)(nil), "truedemocracy.MsgVoteToExclude")
	gogoproto.RegisterType((*MsgVoteToDelete)(nil), "truedemocracy.MsgVoteToDelete")
	gogoproto.RegisterType((*MsgRateProposal)(nil), "truedemocracy.MsgRateProposal")
	gogoproto.RegisterType((*MsgCastElectionVote)(nil), "truedemocracy.MsgCastElectionVote")
	gogoproto.RegisterType((*MsgAddMember)(nil), "truedemocracy.MsgAddMember")
	gogoproto.RegisterType((*MsgOnboardToDomain)(nil), "truedemocracy.MsgOnboardToDomain")
	gogoproto.RegisterType((*MsgApproveOnboarding)(nil), "truedemocracy.MsgApproveOnboarding")
	gogoproto.RegisterType((*MsgRejectOnboarding)(nil), "truedemocracy.MsgRejectOnboarding")
	gogoproto.RegisterType((*MsgRegisterIdentity)(nil), "truedemocracy.MsgRegisterIdentity")

	// Register response types.
	gogoproto.RegisterType((*MsgCreateDomainResponse)(nil), "truedemocracy.MsgCreateDomainResponse")
	gogoproto.RegisterType((*MsgSubmitProposalResponse)(nil), "truedemocracy.MsgSubmitProposalResponse")
	gogoproto.RegisterType((*MsgRegisterValidatorResponse)(nil), "truedemocracy.MsgRegisterValidatorResponse")
	gogoproto.RegisterType((*MsgWithdrawStakeResponse)(nil), "truedemocracy.MsgWithdrawStakeResponse")
	gogoproto.RegisterType((*MsgRemoveValidatorResponse)(nil), "truedemocracy.MsgRemoveValidatorResponse")
	gogoproto.RegisterType((*MsgUnjailResponse)(nil), "truedemocracy.MsgUnjailResponse")
	gogoproto.RegisterType((*MsgJoinPermissionRegisterResponse)(nil), "truedemocracy.MsgJoinPermissionRegisterResponse")
	gogoproto.RegisterType((*MsgPurgePermissionRegisterResponse)(nil), "truedemocracy.MsgPurgePermissionRegisterResponse")
	gogoproto.RegisterType((*MsgPlaceStoneOnIssueResponse)(nil), "truedemocracy.MsgPlaceStoneOnIssueResponse")
	gogoproto.RegisterType((*MsgPlaceStoneOnSuggestionResponse)(nil), "truedemocracy.MsgPlaceStoneOnSuggestionResponse")
	gogoproto.RegisterType((*MsgPlaceStoneOnMemberResponse)(nil), "truedemocracy.MsgPlaceStoneOnMemberResponse")
	gogoproto.RegisterType((*MsgVoteToExcludeResponse)(nil), "truedemocracy.MsgVoteToExcludeResponse")
	gogoproto.RegisterType((*MsgVoteToDeleteResponse)(nil), "truedemocracy.MsgVoteToDeleteResponse")
	gogoproto.RegisterType((*MsgRateProposalResponse)(nil), "truedemocracy.MsgRateProposalResponse")
	gogoproto.RegisterType((*MsgCastElectionVoteResponse)(nil), "truedemocracy.MsgCastElectionVoteResponse")
	gogoproto.RegisterType((*MsgAddMemberResponse)(nil), "truedemocracy.MsgAddMemberResponse")
	gogoproto.RegisterType((*MsgOnboardToDomainResponse)(nil), "truedemocracy.MsgOnboardToDomainResponse")
	gogoproto.RegisterType((*MsgApproveOnboardingResponse)(nil), "truedemocracy.MsgApproveOnboardingResponse")
	gogoproto.RegisterType((*MsgRejectOnboardingResponse)(nil), "truedemocracy.MsgRejectOnboardingResponse")
	gogoproto.RegisterType((*MsgRegisterIdentityResponse)(nil), "truedemocracy.MsgRegisterIdentityResponse")
}

// ---------------------------------------------------------------------------
// MsgServer implementation
// ---------------------------------------------------------------------------

type msgServer struct {
	Keeper Keeper
}

// NewMsgServer returns an implementation of the MsgServer interface for the
// truedemocracy module.
func NewMsgServer(keeper Keeper) msgServer {
	return msgServer{Keeper: keeper}
}

// MsgServer defines the message handling interface for the truedemocracy module.
type MsgServer interface {
	CreateDomain(context.Context, *MsgCreateDomain) (*MsgCreateDomainResponse, error)
	SubmitProposal(context.Context, *MsgSubmitProposal) (*MsgSubmitProposalResponse, error)
	RegisterValidator(context.Context, *MsgRegisterValidator) (*MsgRegisterValidatorResponse, error)
	WithdrawStake(context.Context, *MsgWithdrawStake) (*MsgWithdrawStakeResponse, error)
	RemoveValidator(context.Context, *MsgRemoveValidator) (*MsgRemoveValidatorResponse, error)
	Unjail(context.Context, *MsgUnjail) (*MsgUnjailResponse, error)
	JoinPermissionRegister(context.Context, *MsgJoinPermissionRegister) (*MsgJoinPermissionRegisterResponse, error)
	PurgePermissionRegister(context.Context, *MsgPurgePermissionRegister) (*MsgPurgePermissionRegisterResponse, error)
	PlaceStoneOnIssue(context.Context, *MsgPlaceStoneOnIssue) (*MsgPlaceStoneOnIssueResponse, error)
	PlaceStoneOnSuggestion(context.Context, *MsgPlaceStoneOnSuggestion) (*MsgPlaceStoneOnSuggestionResponse, error)
	PlaceStoneOnMember(context.Context, *MsgPlaceStoneOnMember) (*MsgPlaceStoneOnMemberResponse, error)
	VoteToExclude(context.Context, *MsgVoteToExclude) (*MsgVoteToExcludeResponse, error)
	VoteToDelete(context.Context, *MsgVoteToDelete) (*MsgVoteToDeleteResponse, error)
	RateProposal(context.Context, *MsgRateProposal) (*MsgRateProposalResponse, error)
	CastElectionVote(context.Context, *MsgCastElectionVote) (*MsgCastElectionVoteResponse, error)
	AddMember(context.Context, *MsgAddMember) (*MsgAddMemberResponse, error)
	OnboardToDomain(context.Context, *MsgOnboardToDomain) (*MsgOnboardToDomainResponse, error)
	ApproveOnboarding(context.Context, *MsgApproveOnboarding) (*MsgApproveOnboardingResponse, error)
	RejectOnboarding(context.Context, *MsgRejectOnboarding) (*MsgRejectOnboardingResponse, error)
	RegisterIdentity(context.Context, *MsgRegisterIdentity) (*MsgRegisterIdentityResponse, error)
}

var _ MsgServer = msgServer{}

// ---------------------------------------------------------------------------
// Handler methods
// ---------------------------------------------------------------------------

func (m msgServer) CreateDomain(goCtx context.Context, msg *MsgCreateDomain) (*MsgCreateDomainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	m.Keeper.CreateDomain(ctx, msg.Name, msg.Admin, msg.InitialCoins)

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"create_domain",
		sdk.NewAttribute("domain", msg.Name),
		sdk.NewAttribute("admin", msg.Admin.String()),
	))

	return &MsgCreateDomainResponse{}, nil
}

func (m msgServer) SubmitProposal(goCtx context.Context, msg *MsgSubmitProposal) (*MsgSubmitProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := m.Keeper.SubmitProposal(ctx, msg.DomainName, msg.IssueName, msg.SuggestionName, msg.Creator, msg.Fee, msg.ExternalLink)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"submit_proposal",
		sdk.NewAttribute("domain", msg.DomainName),
		sdk.NewAttribute("issue", msg.IssueName),
		sdk.NewAttribute("suggestion", msg.SuggestionName),
	))

	return &MsgSubmitProposalResponse{}, nil
}

func (m msgServer) RegisterValidator(goCtx context.Context, msg *MsgRegisterValidator) (*MsgRegisterValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	pubKeyBytes, err := hex.DecodeString(msg.PubKey)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid hex-encoded public key")
	}

	err = m.Keeper.RegisterValidator(ctx, msg.OperatorAddr, pubKeyBytes, msg.Stake, msg.DomainName)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"register_validator",
		sdk.NewAttribute("operator", msg.OperatorAddr),
		sdk.NewAttribute("domain", msg.DomainName),
	))

	return &MsgRegisterValidatorResponse{}, nil
}

func (m msgServer) WithdrawStake(goCtx context.Context, msg *MsgWithdrawStake) (*MsgWithdrawStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := m.Keeper.WithdrawStake(ctx, msg.OperatorAddr, msg.Amount)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"withdraw_stake",
		sdk.NewAttribute("operator", msg.OperatorAddr),
		sdk.NewAttribute("amount", fmt.Sprintf("%d", msg.Amount)),
	))

	return &MsgWithdrawStakeResponse{}, nil
}

func (m msgServer) RemoveValidator(goCtx context.Context, msg *MsgRemoveValidator) (*MsgRemoveValidatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := m.Keeper.RemoveValidator(ctx, msg.OperatorAddr)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"remove_validator",
		sdk.NewAttribute("operator", msg.OperatorAddr),
	))

	return &MsgRemoveValidatorResponse{}, nil
}

func (m msgServer) Unjail(goCtx context.Context, msg *MsgUnjail) (*MsgUnjailResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := m.Keeper.Unjail(ctx, msg.OperatorAddr)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"unjail",
		sdk.NewAttribute("operator", msg.OperatorAddr),
	))

	return &MsgUnjailResponse{}, nil
}

func (m msgServer) JoinPermissionRegister(goCtx context.Context, msg *MsgJoinPermissionRegister) (*MsgJoinPermissionRegisterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	pubKeyBytes, err := hex.DecodeString(msg.DomainPubKey)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid hex-encoded domain public key")
	}

	err = m.Keeper.JoinPermissionRegister(ctx, msg.DomainName, msg.MemberAddr, pubKeyBytes)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"join_permission_register",
		sdk.NewAttribute("domain", msg.DomainName),
	))

	return &MsgJoinPermissionRegisterResponse{}, nil
}

func (m msgServer) PurgePermissionRegister(goCtx context.Context, msg *MsgPurgePermissionRegister) (*MsgPurgePermissionRegisterResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := m.Keeper.PurgePermissionRegister(ctx, msg.DomainName, msg.Caller)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"purge_permission_register",
		sdk.NewAttribute("domain", msg.DomainName),
	))

	return &MsgPurgePermissionRegisterResponse{}, nil
}

func (m msgServer) PlaceStoneOnIssue(goCtx context.Context, msg *MsgPlaceStoneOnIssue) (*MsgPlaceStoneOnIssueResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	reward, err := m.Keeper.PlaceStoneOnIssue(ctx, msg.DomainName, msg.IssueName, msg.MemberAddr)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"place_stone_issue",
		sdk.NewAttribute("domain", msg.DomainName),
		sdk.NewAttribute("issue", msg.IssueName),
		sdk.NewAttribute("reward", reward.String()),
	))

	return &MsgPlaceStoneOnIssueResponse{}, nil
}

func (m msgServer) PlaceStoneOnSuggestion(goCtx context.Context, msg *MsgPlaceStoneOnSuggestion) (*MsgPlaceStoneOnSuggestionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	reward, err := m.Keeper.PlaceStoneOnSuggestion(ctx, msg.DomainName, msg.IssueName, msg.SuggestionName, msg.MemberAddr)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"place_stone_suggestion",
		sdk.NewAttribute("domain", msg.DomainName),
		sdk.NewAttribute("issue", msg.IssueName),
		sdk.NewAttribute("suggestion", msg.SuggestionName),
		sdk.NewAttribute("reward", reward.String()),
	))

	return &MsgPlaceStoneOnSuggestionResponse{}, nil
}

func (m msgServer) PlaceStoneOnMember(goCtx context.Context, msg *MsgPlaceStoneOnMember) (*MsgPlaceStoneOnMemberResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := m.Keeper.PlaceStoneOnMember(ctx, msg.DomainName, msg.TargetMember, msg.VoterAddr)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"place_stone_member",
		sdk.NewAttribute("domain", msg.DomainName),
		sdk.NewAttribute("target", msg.TargetMember),
	))

	return &MsgPlaceStoneOnMemberResponse{}, nil
}

func (m msgServer) VoteToExclude(goCtx context.Context, msg *MsgVoteToExclude) (*MsgVoteToExcludeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	excluded, err := m.Keeper.VoteToExclude(ctx, msg.DomainName, msg.TargetMember, msg.VoterAddr)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"vote_exclude",
		sdk.NewAttribute("domain", msg.DomainName),
		sdk.NewAttribute("target", msg.TargetMember),
		sdk.NewAttribute("excluded", fmt.Sprintf("%t", excluded)),
	))

	return &MsgVoteToExcludeResponse{}, nil
}

func (m msgServer) VoteToDelete(goCtx context.Context, msg *MsgVoteToDelete) (*MsgVoteToDeleteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	deleted, err := m.Keeper.VoteToDelete(ctx, msg.DomainName, msg.IssueName, msg.SuggestionName, msg.MemberAddr)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"vote_delete",
		sdk.NewAttribute("domain", msg.DomainName),
		sdk.NewAttribute("issue", msg.IssueName),
		sdk.NewAttribute("suggestion", msg.SuggestionName),
		sdk.NewAttribute("deleted", fmt.Sprintf("%t", deleted)),
	))

	return &MsgVoteToDeleteResponse{}, nil
}

func (m msgServer) RateProposal(goCtx context.Context, msg *MsgRateProposal) (*MsgRateProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	reward, err := m.Keeper.RateProposalWithSignature(ctx, msg.DomainName, msg.IssueName, msg.SuggestionName, int(msg.Rating), msg.DomainPubKey, msg.Signature)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"rate_proposal",
		sdk.NewAttribute("domain", msg.DomainName),
		sdk.NewAttribute("issue", msg.IssueName),
		sdk.NewAttribute("suggestion", msg.SuggestionName),
		sdk.NewAttribute("rating", fmt.Sprintf("%d", msg.Rating)),
		sdk.NewAttribute("reward", reward.String()),
	))

	return &MsgRateProposalResponse{}, nil
}

func (m msgServer) CastElectionVote(goCtx context.Context, msg *MsgCastElectionVote) (*MsgCastElectionVoteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := m.Keeper.CastElectionVote(ctx, msg.DomainName, msg.IssueName, msg.CandidateName, msg.VoterAddr, VoteChoice(msg.Choice))
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"cast_election_vote",
		sdk.NewAttribute("domain", msg.DomainName),
		sdk.NewAttribute("issue", msg.IssueName),
		sdk.NewAttribute("candidate", msg.CandidateName),
		sdk.NewAttribute("voter", msg.VoterAddr),
		sdk.NewAttribute("choice", fmt.Sprintf("%d", msg.Choice)),
	))

	return &MsgCastElectionVoteResponse{}, nil
}

func (m msgServer) AddMember(goCtx context.Context, msg *MsgAddMember) (*MsgAddMemberResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := m.Keeper.AddMember(ctx, msg.DomainName, msg.NewMember, msg.Sender)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"add_member",
		sdk.NewAttribute("domain", msg.DomainName),
		sdk.NewAttribute("new_member", msg.NewMember),
		sdk.NewAttribute("added_by", msg.Sender.String()),
	))

	return &MsgAddMemberResponse{}, nil
}

func (m msgServer) OnboardToDomain(goCtx context.Context, msg *MsgOnboardToDomain) (*MsgOnboardToDomainResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Parse global public key.
	globalPubKey, err := ParseEd25519PubKeyFromHex(msg.GlobalPubKeyHex)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid global public key: "+err.Error())
	}

	// Verify the onboarding signature (proves sender controls global key).
	sigBytes, err := hex.DecodeString(msg.SignatureHex)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid signature hex")
	}
	if err := VerifyOnboardingSignature(msg.Sender.String(), msg.DomainName, msg.DomainPubKeyHex, globalPubKey, sigBytes); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, err.Error())
	}

	// Verify domain key != global key (anonymity requirement).
	if err := VerifyKeysAreDifferent(msg.GlobalPubKeyHex, msg.DomainPubKeyHex); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	// Decode domain public key.
	domainPubKeyBytes, err := hex.DecodeString(msg.DomainPubKeyHex)
	if err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "invalid domain public key hex")
	}

	// JoinPermissionRegister validates membership, key length, and duplicates.
	if err := m.Keeper.JoinPermissionRegister(ctx, msg.DomainName, msg.Sender.String(), domainPubKeyBytes); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"onboard_to_domain",
		sdk.NewAttribute("domain", msg.DomainName),
		sdk.NewAttribute("member", msg.Sender.String()),
	))

	return &MsgOnboardToDomainResponse{}, nil
}

func (m msgServer) ApproveOnboarding(goCtx context.Context, msg *MsgApproveOnboarding) (*MsgApproveOnboardingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := m.Keeper.ApproveOnboardingRequest(ctx, msg.DomainName, msg.RequesterAddr, msg.Sender)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"approve_onboarding",
		sdk.NewAttribute("domain", msg.DomainName),
		sdk.NewAttribute("requester", msg.RequesterAddr),
		sdk.NewAttribute("admin", msg.Sender.String()),
	))

	return &MsgApproveOnboardingResponse{}, nil
}

func (m msgServer) RejectOnboarding(goCtx context.Context, msg *MsgRejectOnboarding) (*MsgRejectOnboardingResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := m.Keeper.RejectOnboardingRequest(ctx, msg.DomainName, msg.RequesterAddr, msg.Sender)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"reject_onboarding",
		sdk.NewAttribute("domain", msg.DomainName),
		sdk.NewAttribute("requester", msg.RequesterAddr),
		sdk.NewAttribute("admin", msg.Sender.String()),
	))

	return &MsgRejectOnboardingResponse{}, nil
}

func (m msgServer) RegisterIdentity(goCtx context.Context, msg *MsgRegisterIdentity) (*MsgRegisterIdentityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := m.Keeper.RegisterIdentityCommitment(ctx, msg.DomainName, msg.Sender.String(), msg.Commitment)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"register_identity",
		sdk.NewAttribute("domain", msg.DomainName),
		sdk.NewAttribute("member", msg.Sender.String()),
	))

	return &MsgRegisterIdentityResponse{}, nil
}

// ---------------------------------------------------------------------------
// gRPC method handlers
// ---------------------------------------------------------------------------

func _Msg_CreateDomain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreateDomain)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreateDomain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/CreateDomain",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreateDomain(ctx, req.(*MsgCreateDomain))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_SubmitProposal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSubmitProposal)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SubmitProposal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/SubmitProposal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SubmitProposal(ctx, req.(*MsgSubmitProposal))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RegisterValidator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRegisterValidator)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RegisterValidator(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/RegisterValidator",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RegisterValidator(ctx, req.(*MsgRegisterValidator))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_WithdrawStake_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgWithdrawStake)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).WithdrawStake(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/WithdrawStake",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).WithdrawStake(ctx, req.(*MsgWithdrawStake))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RemoveValidator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRemoveValidator)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RemoveValidator(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/RemoveValidator",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RemoveValidator(ctx, req.(*MsgRemoveValidator))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_Unjail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUnjail)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).Unjail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/Unjail",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).Unjail(ctx, req.(*MsgUnjail))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_JoinPermissionRegister_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgJoinPermissionRegister)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).JoinPermissionRegister(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/JoinPermissionRegister",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).JoinPermissionRegister(ctx, req.(*MsgJoinPermissionRegister))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_PurgePermissionRegister_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgPurgePermissionRegister)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).PurgePermissionRegister(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/PurgePermissionRegister",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).PurgePermissionRegister(ctx, req.(*MsgPurgePermissionRegister))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_PlaceStoneOnIssue_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgPlaceStoneOnIssue)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).PlaceStoneOnIssue(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/PlaceStoneOnIssue",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).PlaceStoneOnIssue(ctx, req.(*MsgPlaceStoneOnIssue))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_PlaceStoneOnSuggestion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgPlaceStoneOnSuggestion)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).PlaceStoneOnSuggestion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/PlaceStoneOnSuggestion",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).PlaceStoneOnSuggestion(ctx, req.(*MsgPlaceStoneOnSuggestion))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_PlaceStoneOnMember_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgPlaceStoneOnMember)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).PlaceStoneOnMember(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/PlaceStoneOnMember",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).PlaceStoneOnMember(ctx, req.(*MsgPlaceStoneOnMember))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_VoteToExclude_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgVoteToExclude)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).VoteToExclude(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/VoteToExclude",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).VoteToExclude(ctx, req.(*MsgVoteToExclude))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_VoteToDelete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgVoteToDelete)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).VoteToDelete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/VoteToDelete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).VoteToDelete(ctx, req.(*MsgVoteToDelete))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RateProposal_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRateProposal)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RateProposal(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/RateProposal",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RateProposal(ctx, req.(*MsgRateProposal))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_CastElectionVote_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCastElectionVote)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CastElectionVote(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/CastElectionVote",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CastElectionVote(ctx, req.(*MsgCastElectionVote))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_AddMember_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgAddMember)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).AddMember(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/AddMember",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).AddMember(ctx, req.(*MsgAddMember))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_OnboardToDomain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgOnboardToDomain)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).OnboardToDomain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/OnboardToDomain",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).OnboardToDomain(ctx, req.(*MsgOnboardToDomain))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_ApproveOnboarding_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgApproveOnboarding)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).ApproveOnboarding(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/ApproveOnboarding",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).ApproveOnboarding(ctx, req.(*MsgApproveOnboarding))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RejectOnboarding_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRejectOnboarding)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RejectOnboarding(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/RejectOnboarding",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RejectOnboarding(ctx, req.(*MsgRejectOnboarding))
	}
	return interceptor(ctx, in, info, handler)
}

// ---------------------------------------------------------------------------
// gRPC service registration
// ---------------------------------------------------------------------------

// RegisterMsgServer registers the MsgServer implementation with the gRPC
// service registrar (typically the Cosmos SDK's BaseApp).
func RegisterMsgServer(s gogogrpc.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_RegisterIdentity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRegisterIdentity)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RegisterIdentity(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/truedemocracy.Msg/RegisterIdentity",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RegisterIdentity(ctx, req.(*MsgRegisterIdentity))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "truedemocracy.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateDomain",
			Handler:    _Msg_CreateDomain_Handler,
		},
		{
			MethodName: "SubmitProposal",
			Handler:    _Msg_SubmitProposal_Handler,
		},
		{
			MethodName: "RegisterValidator",
			Handler:    _Msg_RegisterValidator_Handler,
		},
		{
			MethodName: "WithdrawStake",
			Handler:    _Msg_WithdrawStake_Handler,
		},
		{
			MethodName: "RemoveValidator",
			Handler:    _Msg_RemoveValidator_Handler,
		},
		{
			MethodName: "Unjail",
			Handler:    _Msg_Unjail_Handler,
		},
		{
			MethodName: "JoinPermissionRegister",
			Handler:    _Msg_JoinPermissionRegister_Handler,
		},
		{
			MethodName: "PurgePermissionRegister",
			Handler:    _Msg_PurgePermissionRegister_Handler,
		},
		{
			MethodName: "PlaceStoneOnIssue",
			Handler:    _Msg_PlaceStoneOnIssue_Handler,
		},
		{
			MethodName: "PlaceStoneOnSuggestion",
			Handler:    _Msg_PlaceStoneOnSuggestion_Handler,
		},
		{
			MethodName: "PlaceStoneOnMember",
			Handler:    _Msg_PlaceStoneOnMember_Handler,
		},
		{
			MethodName: "VoteToExclude",
			Handler:    _Msg_VoteToExclude_Handler,
		},
		{
			MethodName: "VoteToDelete",
			Handler:    _Msg_VoteToDelete_Handler,
		},
		{
			MethodName: "RateProposal",
			Handler:    _Msg_RateProposal_Handler,
		},
		{
			MethodName: "CastElectionVote",
			Handler:    _Msg_CastElectionVote_Handler,
		},
		{
			MethodName: "AddMember",
			Handler:    _Msg_AddMember_Handler,
		},
		{
			MethodName: "OnboardToDomain",
			Handler:    _Msg_OnboardToDomain_Handler,
		},
		{
			MethodName: "ApproveOnboarding",
			Handler:    _Msg_ApproveOnboarding_Handler,
		},
		{
			MethodName: "RejectOnboarding",
			Handler:    _Msg_RejectOnboarding_Handler,
		},
		{
			MethodName: "RegisterIdentity",
			Handler:    _Msg_RegisterIdentity_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "truedemocracy/tx.proto",
}

// Ensure unused imports are referenced.
var _ = errorsmod.Wrap
var _ = sdkerrors.ErrInvalidRequest
