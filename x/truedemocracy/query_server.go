package truedemocracy

import (
	"context"
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
	gogoproto "github.com/cosmos/gogoproto/proto"
	gogogrpc "github.com/cosmos/gogoproto/grpc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc"
)

// ---------------------------------------------------------------------------
// Query request/response types
// ---------------------------------------------------------------------------

type QueryDomainRequest struct {
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name"`
}

func (*QueryDomainRequest) ProtoMessage()  {}
func (*QueryDomainRequest) Reset()         {}
func (*QueryDomainRequest) String() string { return "QueryDomainRequest" }

type QueryDomainResponse struct {
	Result []byte `protobuf:"bytes,1,opt,name=result,proto3" json:"result"`
}

func (*QueryDomainResponse) ProtoMessage()  {}
func (*QueryDomainResponse) Reset()         {}
func (*QueryDomainResponse) String() string { return "QueryDomainResponse" }

type QueryDomainsRequest struct{}

func (*QueryDomainsRequest) ProtoMessage()  {}
func (*QueryDomainsRequest) Reset()         {}
func (*QueryDomainsRequest) String() string { return "QueryDomainsRequest" }

type QueryDomainsResponse struct {
	Result []byte `protobuf:"bytes,1,opt,name=result,proto3" json:"result"`
}

func (*QueryDomainsResponse) ProtoMessage()  {}
func (*QueryDomainsResponse) Reset()         {}
func (*QueryDomainsResponse) String() string { return "QueryDomainsResponse" }

type QueryValidatorRequest struct {
	OperatorAddr string `protobuf:"bytes,1,opt,name=operator_addr,json=operatorAddr,proto3" json:"operator_addr"`
}

func (*QueryValidatorRequest) ProtoMessage()  {}
func (*QueryValidatorRequest) Reset()         {}
func (*QueryValidatorRequest) String() string { return "QueryValidatorRequest" }

type QueryValidatorResponse struct {
	Result []byte `protobuf:"bytes,1,opt,name=result,proto3" json:"result"`
}

func (*QueryValidatorResponse) ProtoMessage()  {}
func (*QueryValidatorResponse) Reset()         {}
func (*QueryValidatorResponse) String() string { return "QueryValidatorResponse" }

type QueryValidatorsRequest struct{}

func (*QueryValidatorsRequest) ProtoMessage()  {}
func (*QueryValidatorsRequest) Reset()         {}
func (*QueryValidatorsRequest) String() string { return "QueryValidatorsRequest" }

type QueryValidatorsResponse struct {
	Result []byte `protobuf:"bytes,1,opt,name=result,proto3" json:"result"`
}

func (*QueryValidatorsResponse) ProtoMessage()  {}
func (*QueryValidatorsResponse) Reset()         {}
func (*QueryValidatorsResponse) String() string { return "QueryValidatorsResponse" }

// ---------------------------------------------------------------------------
// Register query types with gogoproto
// ---------------------------------------------------------------------------

func init() {
	gogoproto.RegisterType((*QueryDomainRequest)(nil), "truedemocracy.QueryDomainRequest")
	gogoproto.RegisterType((*QueryDomainResponse)(nil), "truedemocracy.QueryDomainResponse")
	gogoproto.RegisterType((*QueryDomainsRequest)(nil), "truedemocracy.QueryDomainsRequest")
	gogoproto.RegisterType((*QueryDomainsResponse)(nil), "truedemocracy.QueryDomainsResponse")
	gogoproto.RegisterType((*QueryValidatorRequest)(nil), "truedemocracy.QueryValidatorRequest")
	gogoproto.RegisterType((*QueryValidatorResponse)(nil), "truedemocracy.QueryValidatorResponse")
	gogoproto.RegisterType((*QueryValidatorsRequest)(nil), "truedemocracy.QueryValidatorsRequest")
	gogoproto.RegisterType((*QueryValidatorsResponse)(nil), "truedemocracy.QueryValidatorsResponse")
}

// ---------------------------------------------------------------------------
// QueryServer interface
// ---------------------------------------------------------------------------

type QueryServer interface {
	Domain(context.Context, *QueryDomainRequest) (*QueryDomainResponse, error)
	Domains(context.Context, *QueryDomainsRequest) (*QueryDomainsResponse, error)
	Validator(context.Context, *QueryValidatorRequest) (*QueryValidatorResponse, error)
	Validators(context.Context, *QueryValidatorsRequest) (*QueryValidatorsResponse, error)
}

var _ QueryServer = Keeper{}

// ---------------------------------------------------------------------------
// QueryServer implementation (on Keeper)
// ---------------------------------------------------------------------------

func (k Keeper) Domain(goCtx context.Context, req *QueryDomainRequest) (*QueryDomainResponse, error) {
	if req == nil || req.Name == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "domain name is required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	domain, found := k.GetDomain(ctx, req.Name)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "domain %s not found", req.Name)
	}
	bz, err := json.Marshal(domain)
	if err != nil {
		return nil, err
	}
	return &QueryDomainResponse{Result: bz}, nil
}

func (k Keeper) Domains(goCtx context.Context, req *QueryDomainsRequest) (*QueryDomainsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var domains []Domain
	k.IterateDomains(ctx, func(d Domain) bool {
		domains = append(domains, d)
		return false
	})
	if domains == nil {
		domains = []Domain{}
	}
	bz, err := json.Marshal(domains)
	if err != nil {
		return nil, err
	}
	return &QueryDomainsResponse{Result: bz}, nil
}

func (k Keeper) Validator(goCtx context.Context, req *QueryValidatorRequest) (*QueryValidatorResponse, error) {
	if req == nil || req.OperatorAddr == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "operator address is required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	val, found := k.GetValidator(ctx, req.OperatorAddr)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "validator %s not found", req.OperatorAddr)
	}
	bz, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	return &QueryValidatorResponse{Result: bz}, nil
}

func (k Keeper) Validators(goCtx context.Context, req *QueryValidatorsRequest) (*QueryValidatorsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var validators []Validator
	k.IterateValidators(ctx, func(v Validator) bool {
		validators = append(validators, v)
		return false
	})
	if validators == nil {
		validators = []Validator{}
	}
	bz, err := json.Marshal(validators)
	if err != nil {
		return nil, err
	}
	return &QueryValidatorsResponse{Result: bz}, nil
}

// ---------------------------------------------------------------------------
// gRPC method handlers
// ---------------------------------------------------------------------------

func _Query_Domain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryDomainRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Domain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/truedemocracy.Query/Domain"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Domain(ctx, req.(*QueryDomainRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Domains_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryDomainsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Domains(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/truedemocracy.Query/Domains"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Domains(ctx, req.(*QueryDomainsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Validator_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryValidatorRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Validator(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/truedemocracy.Query/Validator"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Validator(ctx, req.(*QueryValidatorRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Validators_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryValidatorsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Validators(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/truedemocracy.Query/Validators"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Validators(ctx, req.(*QueryValidatorsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ---------------------------------------------------------------------------
// gRPC service registration
// ---------------------------------------------------------------------------

func RegisterQueryServer(s gogogrpc.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "truedemocracy.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Domain", Handler: _Query_Domain_Handler},
		{MethodName: "Domains", Handler: _Query_Domains_Handler},
		{MethodName: "Validator", Handler: _Query_Validator_Handler},
		{MethodName: "Validators", Handler: _Query_Validators_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "truedemocracy/query.proto",
}

// ---------------------------------------------------------------------------
// gRPC query client (for CLI)
// ---------------------------------------------------------------------------

type queryClient struct {
	cc gogogrpc.ClientConn
}

func NewQueryClient(cc gogogrpc.ClientConn) QueryServer {
	return &queryClient{cc}
}

func (c *queryClient) Domain(ctx context.Context, in *QueryDomainRequest) (*QueryDomainResponse, error) {
	out := new(QueryDomainResponse)
	err := c.cc.Invoke(ctx, "/truedemocracy.Query/Domain", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Domains(ctx context.Context, in *QueryDomainsRequest) (*QueryDomainsResponse, error) {
	out := new(QueryDomainsResponse)
	err := c.cc.Invoke(ctx, "/truedemocracy.Query/Domains", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Validator(ctx context.Context, in *QueryValidatorRequest) (*QueryValidatorResponse, error) {
	out := new(QueryValidatorResponse)
	err := c.cc.Invoke(ctx, "/truedemocracy.Query/Validator", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Validators(ctx context.Context, in *QueryValidatorsRequest) (*QueryValidatorsResponse, error) {
	out := new(QueryValidatorsResponse)
	err := c.cc.Invoke(ctx, "/truedemocracy.Query/Validators", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
