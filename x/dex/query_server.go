package dex

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

type QueryPoolRequest struct {
	AssetDenom string `protobuf:"bytes,1,opt,name=asset_denom,json=assetDenom,proto3" json:"asset_denom"`
}

func (*QueryPoolRequest) ProtoMessage()  {}
func (*QueryPoolRequest) Reset()         {}
func (*QueryPoolRequest) String() string { return "QueryPoolRequest" }

type QueryPoolResponse struct {
	Result []byte `protobuf:"bytes,1,opt,name=result,proto3" json:"result"`
}

func (*QueryPoolResponse) ProtoMessage()  {}
func (*QueryPoolResponse) Reset()         {}
func (*QueryPoolResponse) String() string { return "QueryPoolResponse" }

type QueryPoolsRequest struct{}

func (*QueryPoolsRequest) ProtoMessage()  {}
func (*QueryPoolsRequest) Reset()         {}
func (*QueryPoolsRequest) String() string { return "QueryPoolsRequest" }

type QueryPoolsResponse struct {
	Result []byte `protobuf:"bytes,1,opt,name=result,proto3" json:"result"`
}

func (*QueryPoolsResponse) ProtoMessage()  {}
func (*QueryPoolsResponse) Reset()         {}
func (*QueryPoolsResponse) String() string { return "QueryPoolsResponse" }

// ---------------------------------------------------------------------------
// Register query types with gogoproto
// ---------------------------------------------------------------------------

func init() {
	gogoproto.RegisterType((*QueryPoolRequest)(nil), "dex.QueryPoolRequest")
	gogoproto.RegisterType((*QueryPoolResponse)(nil), "dex.QueryPoolResponse")
	gogoproto.RegisterType((*QueryPoolsRequest)(nil), "dex.QueryPoolsRequest")
	gogoproto.RegisterType((*QueryPoolsResponse)(nil), "dex.QueryPoolsResponse")
}

// ---------------------------------------------------------------------------
// QueryServer interface
// ---------------------------------------------------------------------------

type QueryServer interface {
	Pool(context.Context, *QueryPoolRequest) (*QueryPoolResponse, error)
	Pools(context.Context, *QueryPoolsRequest) (*QueryPoolsResponse, error)
}

var _ QueryServer = Keeper{}

// ---------------------------------------------------------------------------
// QueryServer implementation (on Keeper)
// ---------------------------------------------------------------------------

func (k Keeper) Pool(goCtx context.Context, req *QueryPoolRequest) (*QueryPoolResponse, error) {
	if req == nil || req.AssetDenom == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "asset denom is required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	pool, found := k.GetPool(ctx, req.AssetDenom)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "pool for %s not found", req.AssetDenom)
	}
	bz, err := json.Marshal(pool)
	if err != nil {
		return nil, err
	}
	return &QueryPoolResponse{Result: bz}, nil
}

func (k Keeper) Pools(goCtx context.Context, req *QueryPoolsRequest) (*QueryPoolsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var pools []Pool
	k.IteratePools(ctx, func(p Pool) bool {
		pools = append(pools, p)
		return false
	})
	if pools == nil {
		pools = []Pool{}
	}
	bz, err := json.Marshal(pools)
	if err != nil {
		return nil, err
	}
	return &QueryPoolsResponse{Result: bz}, nil
}

// ---------------------------------------------------------------------------
// gRPC method handlers
// ---------------------------------------------------------------------------

func _Query_Pool_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryPoolRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Pool(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/dex.Query/Pool"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Pool(ctx, req.(*QueryPoolRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Pools_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryPoolsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Pools(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/dex.Query/Pools"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Pools(ctx, req.(*QueryPoolsRequest))
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
	ServiceName: "dex.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Pool", Handler: _Query_Pool_Handler},
		{MethodName: "Pools", Handler: _Query_Pools_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "dex/query.proto",
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

func (c *queryClient) Pool(ctx context.Context, in *QueryPoolRequest) (*QueryPoolResponse, error) {
	out := new(QueryPoolResponse)
	err := c.cc.Invoke(ctx, "/dex.Query/Pool", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Pools(ctx context.Context, in *QueryPoolsRequest) (*QueryPoolsResponse, error) {
	out := new(QueryPoolsResponse)
	err := c.cc.Invoke(ctx, "/dex.Query/Pools", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
