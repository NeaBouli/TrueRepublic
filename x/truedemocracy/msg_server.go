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

// ---------------------------------------------------------------------------
// gRPC service registration
// ---------------------------------------------------------------------------

// RegisterMsgServer registers the MsgServer implementation with the gRPC
// service registrar (typically the Cosmos SDK's BaseApp).
func RegisterMsgServer(s gogogrpc.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
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
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "truedemocracy/tx.proto",
}

// Ensure unused imports are referenced.
var _ = errorsmod.Wrap
var _ = sdkerrors.ErrInvalidRequest
