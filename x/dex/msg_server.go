package dex

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	gogoproto "github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"
)

// ---------------------------------------------------------------------------
// Response types
// ---------------------------------------------------------------------------

type MsgCreatePoolResponse struct{}

func (*MsgCreatePoolResponse) ProtoMessage()  {}
func (*MsgCreatePoolResponse) Reset()         {}
func (*MsgCreatePoolResponse) String() string { return "MsgCreatePoolResponse" }

type MsgSwapResponse struct{}

func (*MsgSwapResponse) ProtoMessage()  {}
func (*MsgSwapResponse) Reset()         {}
func (*MsgSwapResponse) String() string { return "MsgSwapResponse" }

type MsgAddLiquidityResponse struct{}

func (*MsgAddLiquidityResponse) ProtoMessage()  {}
func (*MsgAddLiquidityResponse) Reset()         {}
func (*MsgAddLiquidityResponse) String() string { return "MsgAddLiquidityResponse" }

type MsgRemoveLiquidityResponse struct{}

func (*MsgRemoveLiquidityResponse) ProtoMessage()  {}
func (*MsgRemoveLiquidityResponse) Reset()         {}
func (*MsgRemoveLiquidityResponse) String() string { return "MsgRemoveLiquidityResponse" }

// ---------------------------------------------------------------------------
// Register all types with gogoproto
// ---------------------------------------------------------------------------

func init() {
	// Msg types.
	gogoproto.RegisterType((*MsgCreatePool)(nil), "dex.MsgCreatePool")
	gogoproto.RegisterType((*MsgSwap)(nil), "dex.MsgSwap")
	gogoproto.RegisterType((*MsgAddLiquidity)(nil), "dex.MsgAddLiquidity")
	gogoproto.RegisterType((*MsgRemoveLiquidity)(nil), "dex.MsgRemoveLiquidity")

	// Response types.
	gogoproto.RegisterType((*MsgCreatePoolResponse)(nil), "dex.MsgCreatePoolResponse")
	gogoproto.RegisterType((*MsgSwapResponse)(nil), "dex.MsgSwapResponse")
	gogoproto.RegisterType((*MsgAddLiquidityResponse)(nil), "dex.MsgAddLiquidityResponse")
	gogoproto.RegisterType((*MsgRemoveLiquidityResponse)(nil), "dex.MsgRemoveLiquidityResponse")
}

// ---------------------------------------------------------------------------
// MsgServer implementation
// ---------------------------------------------------------------------------

// MsgServer defines the message handling interface for the dex module.
type MsgServer interface {
	CreatePool(context.Context, *MsgCreatePool) (*MsgCreatePoolResponse, error)
	Swap(context.Context, *MsgSwap) (*MsgSwapResponse, error)
	AddLiquidity(context.Context, *MsgAddLiquidity) (*MsgAddLiquidityResponse, error)
	RemoveLiquidity(context.Context, *MsgRemoveLiquidity) (*MsgRemoveLiquidityResponse, error)
}

type msgServer struct {
	Keeper Keeper
}

// NewMsgServer returns an implementation of the MsgServer interface for the dex module.
func NewMsgServer(keeper Keeper) msgServer {
	return msgServer{Keeper: keeper}
}

var _ MsgServer = msgServer{}

// ---------------------------------------------------------------------------
// Handler methods
// ---------------------------------------------------------------------------

func (m msgServer) CreatePool(goCtx context.Context, msg *MsgCreatePool) (*MsgCreatePoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := m.Keeper.CreatePool(ctx, msg.AssetDenom, math.NewInt(msg.PnyxAmt), math.NewInt(msg.AssetAmt))
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"create_pool",
		sdk.NewAttribute("asset_denom", msg.AssetDenom),
		sdk.NewAttribute("pnyx_amount", fmt.Sprintf("%d", msg.PnyxAmt)),
		sdk.NewAttribute("asset_amount", fmt.Sprintf("%d", msg.AssetAmt)),
	))

	return &MsgCreatePoolResponse{}, nil
}

func (m msgServer) Swap(goCtx context.Context, msg *MsgSwap) (*MsgSwapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	output, err := m.Keeper.Swap(ctx, msg.InputDenom, math.NewInt(msg.InputAmt), msg.OutputDenom)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"swap",
		sdk.NewAttribute("input_denom", msg.InputDenom),
		sdk.NewAttribute("input_amount", fmt.Sprintf("%d", msg.InputAmt)),
		sdk.NewAttribute("output_denom", msg.OutputDenom),
		sdk.NewAttribute("output_amount", output.String()),
	))

	return &MsgSwapResponse{}, nil
}

func (m msgServer) AddLiquidity(goCtx context.Context, msg *MsgAddLiquidity) (*MsgAddLiquidityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	shares, err := m.Keeper.AddLiquidity(ctx, msg.AssetDenom, math.NewInt(msg.PnyxAmt), math.NewInt(msg.AssetAmt))
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"add_liquidity",
		sdk.NewAttribute("asset_denom", msg.AssetDenom),
		sdk.NewAttribute("shares_minted", shares.String()),
	))

	return &MsgAddLiquidityResponse{}, nil
}

func (m msgServer) RemoveLiquidity(goCtx context.Context, msg *MsgRemoveLiquidity) (*MsgRemoveLiquidityResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	pnyxOut, assetOut, err := m.Keeper.RemoveLiquidity(ctx, msg.AssetDenom, math.NewInt(msg.Shares))
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		"remove_liquidity",
		sdk.NewAttribute("asset_denom", msg.AssetDenom),
		sdk.NewAttribute("pnyx_returned", pnyxOut.String()),
		sdk.NewAttribute("asset_returned", assetOut.String()),
	))

	return &MsgRemoveLiquidityResponse{}, nil
}

// ---------------------------------------------------------------------------
// gRPC method handlers
// ---------------------------------------------------------------------------

func _Msg_CreatePool_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgCreatePool)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).CreatePool(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/dex.Msg/CreatePool"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).CreatePool(ctx, req.(*MsgCreatePool))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_Swap_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSwap)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).Swap(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/dex.Msg/Swap"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).Swap(ctx, req.(*MsgSwap))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_AddLiquidity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgAddLiquidity)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).AddLiquidity(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/dex.Msg/AddLiquidity"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).AddLiquidity(ctx, req.(*MsgAddLiquidity))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_RemoveLiquidity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgRemoveLiquidity)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).RemoveLiquidity(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/dex.Msg/RemoveLiquidity"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).RemoveLiquidity(ctx, req.(*MsgRemoveLiquidity))
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
	ServiceName: "dex.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "CreatePool", Handler: _Msg_CreatePool_Handler},
		{MethodName: "Swap", Handler: _Msg_Swap_Handler},
		{MethodName: "AddLiquidity", Handler: _Msg_AddLiquidity_Handler},
		{MethodName: "RemoveLiquidity", Handler: _Msg_RemoveLiquidity_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "dex/tx.proto",
}
