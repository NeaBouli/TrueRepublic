package dex

import (
	"context"
	"encoding/json"

	"cosmossdk.io/math"
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

// --- Asset registry query types ---

type QueryRegisteredAssetsRequest struct{}

func (*QueryRegisteredAssetsRequest) ProtoMessage()  {}
func (*QueryRegisteredAssetsRequest) Reset()         {}
func (*QueryRegisteredAssetsRequest) String() string { return "QueryRegisteredAssetsRequest" }

type QueryRegisteredAssetsResponse struct {
	Result []byte `protobuf:"bytes,1,opt,name=result,proto3" json:"result"`
}

func (*QueryRegisteredAssetsResponse) ProtoMessage()  {}
func (*QueryRegisteredAssetsResponse) Reset()         {}
func (*QueryRegisteredAssetsResponse) String() string { return "QueryRegisteredAssetsResponse" }

type QueryAssetByDenomRequest struct {
	IBCDenom string `protobuf:"bytes,1,opt,name=ibc_denom,json=ibcDenom,proto3" json:"ibc_denom"`
}

func (*QueryAssetByDenomRequest) ProtoMessage()  {}
func (*QueryAssetByDenomRequest) Reset()         {}
func (*QueryAssetByDenomRequest) String() string { return "QueryAssetByDenomRequest" }

type QueryAssetByDenomResponse struct {
	Result []byte `protobuf:"bytes,1,opt,name=result,proto3" json:"result"`
}

func (*QueryAssetByDenomResponse) ProtoMessage()  {}
func (*QueryAssetByDenomResponse) Reset()         {}
func (*QueryAssetByDenomResponse) String() string { return "QueryAssetByDenomResponse" }

type QueryAssetBySymbolRequest struct {
	Symbol string `protobuf:"bytes,1,opt,name=symbol,proto3" json:"symbol"`
}

func (*QueryAssetBySymbolRequest) ProtoMessage()  {}
func (*QueryAssetBySymbolRequest) Reset()         {}
func (*QueryAssetBySymbolRequest) String() string { return "QueryAssetBySymbolRequest" }

type QueryAssetBySymbolResponse struct {
	Result []byte `protobuf:"bytes,1,opt,name=result,proto3" json:"result"`
}

func (*QueryAssetBySymbolResponse) ProtoMessage()  {}
func (*QueryAssetBySymbolResponse) Reset()         {}
func (*QueryAssetBySymbolResponse) String() string { return "QueryAssetBySymbolResponse" }

// --- Estimate swap query types ---

type QueryEstimateSwapRequest struct {
	InputDenom  string `protobuf:"bytes,1,opt,name=input_denom,json=inputDenom,proto3" json:"input_denom"`
	InputAmt    int64  `protobuf:"varint,2,opt,name=input_amt,json=inputAmt,proto3" json:"input_amt"`
	OutputDenom string `protobuf:"bytes,3,opt,name=output_denom,json=outputDenom,proto3" json:"output_denom"`
}

func (*QueryEstimateSwapRequest) ProtoMessage()  {}
func (*QueryEstimateSwapRequest) Reset()         {}
func (*QueryEstimateSwapRequest) String() string { return "QueryEstimateSwapRequest" }

type QueryEstimateSwapResponse struct {
	Result []byte `protobuf:"bytes,1,opt,name=result,proto3" json:"result"`
}

func (*QueryEstimateSwapResponse) ProtoMessage()  {}
func (*QueryEstimateSwapResponse) Reset()         {}
func (*QueryEstimateSwapResponse) String() string { return "QueryEstimateSwapResponse" }

// --- Pool analytics query types ---

type QueryPoolStatsRequest struct {
	AssetDenom string `protobuf:"bytes,1,opt,name=asset_denom,json=assetDenom,proto3" json:"asset_denom"`
}

func (*QueryPoolStatsRequest) ProtoMessage()  {}
func (*QueryPoolStatsRequest) Reset()         {}
func (*QueryPoolStatsRequest) String() string { return "QueryPoolStatsRequest" }

type QueryPoolStatsResponse struct {
	Result []byte `protobuf:"bytes,1,opt,name=result,proto3" json:"result"`
}

func (*QueryPoolStatsResponse) ProtoMessage()  {}
func (*QueryPoolStatsResponse) Reset()         {}
func (*QueryPoolStatsResponse) String() string { return "QueryPoolStatsResponse" }

type QuerySpotPriceRequest struct {
	InputDenom  string `protobuf:"bytes,1,opt,name=input_denom,json=inputDenom,proto3" json:"input_denom"`
	OutputDenom string `protobuf:"bytes,2,opt,name=output_denom,json=outputDenom,proto3" json:"output_denom"`
}

func (*QuerySpotPriceRequest) ProtoMessage()  {}
func (*QuerySpotPriceRequest) Reset()         {}
func (*QuerySpotPriceRequest) String() string { return "QuerySpotPriceRequest" }

type QuerySpotPriceResponse struct {
	Result []byte `protobuf:"bytes,1,opt,name=result,proto3" json:"result"`
}

func (*QuerySpotPriceResponse) ProtoMessage()  {}
func (*QuerySpotPriceResponse) Reset()         {}
func (*QuerySpotPriceResponse) String() string { return "QuerySpotPriceResponse" }

type QueryLiquidityDepthRequest struct {
	InputDenom  string `protobuf:"bytes,1,opt,name=input_denom,json=inputDenom,proto3" json:"input_denom"`
	OutputDenom string `protobuf:"bytes,2,opt,name=output_denom,json=outputDenom,proto3" json:"output_denom"`
}

func (*QueryLiquidityDepthRequest) ProtoMessage()  {}
func (*QueryLiquidityDepthRequest) Reset()         {}
func (*QueryLiquidityDepthRequest) String() string { return "QueryLiquidityDepthRequest" }

type QueryLiquidityDepthResponse struct {
	Result []byte `protobuf:"bytes,1,opt,name=result,proto3" json:"result"`
}

func (*QueryLiquidityDepthResponse) ProtoMessage()  {}
func (*QueryLiquidityDepthResponse) Reset()         {}
func (*QueryLiquidityDepthResponse) String() string { return "QueryLiquidityDepthResponse" }

type QueryLPPositionRequest struct {
	AssetDenom string `protobuf:"bytes,1,opt,name=asset_denom,json=assetDenom,proto3" json:"asset_denom"`
	Shares     int64  `protobuf:"varint,2,opt,name=shares,proto3" json:"shares"`
}

func (*QueryLPPositionRequest) ProtoMessage()  {}
func (*QueryLPPositionRequest) Reset()         {}
func (*QueryLPPositionRequest) String() string { return "QueryLPPositionRequest" }

type QueryLPPositionResponse struct {
	Result []byte `protobuf:"bytes,1,opt,name=result,proto3" json:"result"`
}

func (*QueryLPPositionResponse) ProtoMessage()  {}
func (*QueryLPPositionResponse) Reset()         {}
func (*QueryLPPositionResponse) String() string { return "QueryLPPositionResponse" }

// ---------------------------------------------------------------------------
// Register query types with gogoproto
// ---------------------------------------------------------------------------

func init() {
	gogoproto.RegisterType((*QueryPoolRequest)(nil), "dex.QueryPoolRequest")
	gogoproto.RegisterType((*QueryPoolResponse)(nil), "dex.QueryPoolResponse")
	gogoproto.RegisterType((*QueryPoolsRequest)(nil), "dex.QueryPoolsRequest")
	gogoproto.RegisterType((*QueryPoolsResponse)(nil), "dex.QueryPoolsResponse")
	gogoproto.RegisterType((*QueryRegisteredAssetsRequest)(nil), "dex.QueryRegisteredAssetsRequest")
	gogoproto.RegisterType((*QueryRegisteredAssetsResponse)(nil), "dex.QueryRegisteredAssetsResponse")
	gogoproto.RegisterType((*QueryAssetByDenomRequest)(nil), "dex.QueryAssetByDenomRequest")
	gogoproto.RegisterType((*QueryAssetByDenomResponse)(nil), "dex.QueryAssetByDenomResponse")
	gogoproto.RegisterType((*QueryAssetBySymbolRequest)(nil), "dex.QueryAssetBySymbolRequest")
	gogoproto.RegisterType((*QueryAssetBySymbolResponse)(nil), "dex.QueryAssetBySymbolResponse")
	gogoproto.RegisterType((*QueryEstimateSwapRequest)(nil), "dex.QueryEstimateSwapRequest")
	gogoproto.RegisterType((*QueryEstimateSwapResponse)(nil), "dex.QueryEstimateSwapResponse")
	gogoproto.RegisterType((*QueryPoolStatsRequest)(nil), "dex.QueryPoolStatsRequest")
	gogoproto.RegisterType((*QueryPoolStatsResponse)(nil), "dex.QueryPoolStatsResponse")
	gogoproto.RegisterType((*QuerySpotPriceRequest)(nil), "dex.QuerySpotPriceRequest")
	gogoproto.RegisterType((*QuerySpotPriceResponse)(nil), "dex.QuerySpotPriceResponse")
	gogoproto.RegisterType((*QueryLiquidityDepthRequest)(nil), "dex.QueryLiquidityDepthRequest")
	gogoproto.RegisterType((*QueryLiquidityDepthResponse)(nil), "dex.QueryLiquidityDepthResponse")
	gogoproto.RegisterType((*QueryLPPositionRequest)(nil), "dex.QueryLPPositionRequest")
	gogoproto.RegisterType((*QueryLPPositionResponse)(nil), "dex.QueryLPPositionResponse")
}

// ---------------------------------------------------------------------------
// QueryServer interface
// ---------------------------------------------------------------------------

type QueryServer interface {
	Pool(context.Context, *QueryPoolRequest) (*QueryPoolResponse, error)
	Pools(context.Context, *QueryPoolsRequest) (*QueryPoolsResponse, error)
	RegisteredAssets(context.Context, *QueryRegisteredAssetsRequest) (*QueryRegisteredAssetsResponse, error)
	AssetByDenom(context.Context, *QueryAssetByDenomRequest) (*QueryAssetByDenomResponse, error)
	AssetBySymbol(context.Context, *QueryAssetBySymbolRequest) (*QueryAssetBySymbolResponse, error)
	EstimateSwap(context.Context, *QueryEstimateSwapRequest) (*QueryEstimateSwapResponse, error)
	PoolStats(context.Context, *QueryPoolStatsRequest) (*QueryPoolStatsResponse, error)
	SpotPrice(context.Context, *QuerySpotPriceRequest) (*QuerySpotPriceResponse, error)
	LiquidityDepth(context.Context, *QueryLiquidityDepthRequest) (*QueryLiquidityDepthResponse, error)
	LPPosition(context.Context, *QueryLPPositionRequest) (*QueryLPPositionResponse, error)
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
	pool.AssetSymbol = k.GetSymbolForDenom(ctx, pool.AssetDenom)
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
		p.AssetSymbol = k.GetSymbolForDenom(ctx, p.AssetDenom)
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

func (k Keeper) RegisteredAssets(goCtx context.Context, req *QueryRegisteredAssetsRequest) (*QueryRegisteredAssetsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	assets := k.GetAllAssets(ctx)
	if assets == nil {
		assets = []RegisteredAsset{}
	}
	bz, err := json.Marshal(assets)
	if err != nil {
		return nil, err
	}
	return &QueryRegisteredAssetsResponse{Result: bz}, nil
}

func (k Keeper) AssetByDenom(goCtx context.Context, req *QueryAssetByDenomRequest) (*QueryAssetByDenomResponse, error) {
	if req == nil || req.IBCDenom == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "ibc_denom is required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	asset, found := k.GetAssetByDenom(ctx, req.IBCDenom)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "asset not found: %s", req.IBCDenom)
	}
	bz, err := json.Marshal(asset)
	if err != nil {
		return nil, err
	}
	return &QueryAssetByDenomResponse{Result: bz}, nil
}

func (k Keeper) AssetBySymbol(goCtx context.Context, req *QueryAssetBySymbolRequest) (*QueryAssetBySymbolResponse, error) {
	if req == nil || req.Symbol == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "symbol is required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	asset, found := k.GetAssetBySymbol(ctx, req.Symbol)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "asset not found: %s", req.Symbol)
	}
	bz, err := json.Marshal(asset)
	if err != nil {
		return nil, err
	}
	return &QueryAssetBySymbolResponse{Result: bz}, nil
}

func (k Keeper) EstimateSwap(goCtx context.Context, req *QueryEstimateSwapRequest) (*QueryEstimateSwapResponse, error) {
	if req == nil || req.InputDenom == "" || req.OutputDenom == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "input_denom and output_denom are required")
	}
	if req.InputAmt <= 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "input_amt must be positive")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	expectedOutput, route, err := k.EstimateSwapOutput(ctx, req.InputDenom, math.NewInt(req.InputAmt), req.OutputDenom)
	if err != nil {
		return nil, err
	}

	// Build symbols for the route.
	routeSymbols := make([]string, len(route))
	for i, denom := range route {
		routeSymbols[i] = k.GetSymbolForDenom(ctx, denom)
	}

	result := struct {
		ExpectedOutput string   `json:"expected_output"`
		Route          []string `json:"route"`
		RouteSymbols   []string `json:"route_symbols"`
		Hops           int      `json:"hops"`
	}{
		ExpectedOutput: expectedOutput.String(),
		Route:          route,
		RouteSymbols:   routeSymbols,
		Hops:           len(route) - 1,
	}

	bz, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &QueryEstimateSwapResponse{Result: bz}, nil
}

func (k Keeper) PoolStats(goCtx context.Context, req *QueryPoolStatsRequest) (*QueryPoolStatsResponse, error) {
	if req == nil || req.AssetDenom == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "asset denom is required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	pool, found := k.GetPool(ctx, req.AssetDenom)
	if !found {
		return nil, errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "pool for %s not found", req.AssetDenom)
	}

	// Compute derived stats.
	totalFeesEarned := pool.TotalVolumePnyx.Mul(math.NewInt(SwapFeeBps)).Quo(math.NewInt(10000))
	spotPrice, _ := k.ComputeSpotPrice(ctx, pnyxDenom, req.AssetDenom)

	result := struct {
		AssetDenom          string `json:"asset_denom"`
		AssetSymbol         string `json:"asset_symbol"`
		SwapCount           int64  `json:"swap_count"`
		TotalVolumePnyx     string `json:"total_volume_pnyx"`
		TotalFeesEarned     string `json:"total_fees_earned"`
		TotalBurned         string `json:"total_burned"`
		PnyxReserve         string `json:"pnyx_reserve"`
		AssetReserve        string `json:"asset_reserve"`
		SpotPricePerMillion string `json:"spot_price_per_million"`
		TotalShares         string `json:"total_shares"`
	}{
		AssetDenom:          pool.AssetDenom,
		AssetSymbol:         k.GetSymbolForDenom(ctx, pool.AssetDenom),
		SwapCount:           pool.SwapCount,
		TotalVolumePnyx:     pool.TotalVolumePnyx.String(),
		TotalFeesEarned:     totalFeesEarned.String(),
		TotalBurned:         pool.TotalBurned.String(),
		PnyxReserve:         pool.PnyxReserve.String(),
		AssetReserve:        pool.AssetReserve.String(),
		SpotPricePerMillion: spotPrice.String(),
		TotalShares:         pool.TotalShares.String(),
	}

	bz, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &QueryPoolStatsResponse{Result: bz}, nil
}

func (k Keeper) SpotPrice(goCtx context.Context, req *QuerySpotPriceRequest) (*QuerySpotPriceResponse, error) {
	if req == nil || req.InputDenom == "" || req.OutputDenom == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "input_denom and output_denom are required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	price, err := k.ComputeSpotPrice(ctx, req.InputDenom, req.OutputDenom)
	if err != nil {
		return nil, err
	}

	// Determine route.
	var route []string
	if req.InputDenom == pnyxDenom || req.OutputDenom == pnyxDenom {
		route = []string{req.InputDenom, req.OutputDenom}
	} else {
		route = []string{req.InputDenom, pnyxDenom, req.OutputDenom}
	}

	result := struct {
		InputDenom      string   `json:"input_denom"`
		OutputDenom     string   `json:"output_denom"`
		PricePerMillion string   `json:"price_per_million"`
		InputSymbol     string   `json:"input_symbol"`
		OutputSymbol    string   `json:"output_symbol"`
		Route           []string `json:"route"`
	}{
		InputDenom:      req.InputDenom,
		OutputDenom:     req.OutputDenom,
		PricePerMillion: price.String(),
		InputSymbol:     k.GetSymbolForDenom(ctx, req.InputDenom),
		OutputSymbol:    k.GetSymbolForDenom(ctx, req.OutputDenom),
		Route:           route,
	}

	bz, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &QuerySpotPriceResponse{Result: bz}, nil
}

func (k Keeper) LiquidityDepth(goCtx context.Context, req *QueryLiquidityDepthRequest) (*QueryLiquidityDepthResponse, error) {
	if req == nil || req.InputDenom == "" || req.OutputDenom == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "input_denom and output_denom are required")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	levels, err := k.ComputeLiquidityDepth(ctx, req.InputDenom, req.OutputDenom)
	if err != nil {
		return nil, err
	}

	result := struct {
		InputDenom  string       `json:"input_denom"`
		OutputDenom string       `json:"output_denom"`
		Levels      []DepthLevel `json:"levels"`
	}{
		InputDenom:  req.InputDenom,
		OutputDenom: req.OutputDenom,
		Levels:      levels,
	}

	bz, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &QueryLiquidityDepthResponse{Result: bz}, nil
}

func (k Keeper) LPPosition(goCtx context.Context, req *QueryLPPositionRequest) (*QueryLPPositionResponse, error) {
	if req == nil || req.AssetDenom == "" {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "asset_denom is required")
	}
	if req.Shares <= 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "shares must be positive")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	shares := math.NewInt(req.Shares)
	pnyxVal, assetVal, shareBps, err := k.ComputeLPPosition(ctx, req.AssetDenom, shares)
	if err != nil {
		return nil, err
	}

	result := struct {
		AssetDenom     string `json:"asset_denom"`
		Shares         string `json:"shares"`
		PnyxValue      string `json:"pnyx_value"`
		AssetValue     string `json:"asset_value"`
		ShareOfPoolBps int64  `json:"share_of_pool_bps"`
	}{
		AssetDenom:     req.AssetDenom,
		Shares:         shares.String(),
		PnyxValue:      pnyxVal.String(),
		AssetValue:     assetVal.String(),
		ShareOfPoolBps: shareBps,
	}

	bz, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &QueryLPPositionResponse{Result: bz}, nil
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

func _Query_RegisteredAssets_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryRegisteredAssetsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).RegisteredAssets(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/dex.Query/RegisteredAssets"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).RegisteredAssets(ctx, req.(*QueryRegisteredAssetsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_AssetByDenom_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryAssetByDenomRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).AssetByDenom(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/dex.Query/AssetByDenom"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).AssetByDenom(ctx, req.(*QueryAssetByDenomRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_AssetBySymbol_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryAssetBySymbolRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).AssetBySymbol(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/dex.Query/AssetBySymbol"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).AssetBySymbol(ctx, req.(*QueryAssetBySymbolRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_EstimateSwap_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryEstimateSwapRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).EstimateSwap(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/dex.Query/EstimateSwap"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).EstimateSwap(ctx, req.(*QueryEstimateSwapRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_PoolStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryPoolStatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).PoolStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/dex.Query/PoolStats"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).PoolStats(ctx, req.(*QueryPoolStatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_SpotPrice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QuerySpotPriceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).SpotPrice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/dex.Query/SpotPrice"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).SpotPrice(ctx, req.(*QuerySpotPriceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_LiquidityDepth_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryLiquidityDepthRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).LiquidityDepth(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/dex.Query/LiquidityDepth"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).LiquidityDepth(ctx, req.(*QueryLiquidityDepthRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_LPPosition_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryLPPositionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).LPPosition(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/dex.Query/LPPosition"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).LPPosition(ctx, req.(*QueryLPPositionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func RegisterQueryServer(s gogogrpc.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "dex.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "Pool", Handler: _Query_Pool_Handler},
		{MethodName: "Pools", Handler: _Query_Pools_Handler},
		{MethodName: "RegisteredAssets", Handler: _Query_RegisteredAssets_Handler},
		{MethodName: "AssetByDenom", Handler: _Query_AssetByDenom_Handler},
		{MethodName: "AssetBySymbol", Handler: _Query_AssetBySymbol_Handler},
		{MethodName: "EstimateSwap", Handler: _Query_EstimateSwap_Handler},
		{MethodName: "PoolStats", Handler: _Query_PoolStats_Handler},
		{MethodName: "SpotPrice", Handler: _Query_SpotPrice_Handler},
		{MethodName: "LiquidityDepth", Handler: _Query_LiquidityDepth_Handler},
		{MethodName: "LPPosition", Handler: _Query_LPPosition_Handler},
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

func (c *queryClient) RegisteredAssets(ctx context.Context, in *QueryRegisteredAssetsRequest) (*QueryRegisteredAssetsResponse, error) {
	out := new(QueryRegisteredAssetsResponse)
	err := c.cc.Invoke(ctx, "/dex.Query/RegisteredAssets", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) AssetByDenom(ctx context.Context, in *QueryAssetByDenomRequest) (*QueryAssetByDenomResponse, error) {
	out := new(QueryAssetByDenomResponse)
	err := c.cc.Invoke(ctx, "/dex.Query/AssetByDenom", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) AssetBySymbol(ctx context.Context, in *QueryAssetBySymbolRequest) (*QueryAssetBySymbolResponse, error) {
	out := new(QueryAssetBySymbolResponse)
	err := c.cc.Invoke(ctx, "/dex.Query/AssetBySymbol", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) EstimateSwap(ctx context.Context, in *QueryEstimateSwapRequest) (*QueryEstimateSwapResponse, error) {
	out := new(QueryEstimateSwapResponse)
	err := c.cc.Invoke(ctx, "/dex.Query/EstimateSwap", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) PoolStats(ctx context.Context, in *QueryPoolStatsRequest) (*QueryPoolStatsResponse, error) {
	out := new(QueryPoolStatsResponse)
	err := c.cc.Invoke(ctx, "/dex.Query/PoolStats", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) SpotPrice(ctx context.Context, in *QuerySpotPriceRequest) (*QuerySpotPriceResponse, error) {
	out := new(QuerySpotPriceResponse)
	err := c.cc.Invoke(ctx, "/dex.Query/SpotPrice", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) LiquidityDepth(ctx context.Context, in *QueryLiquidityDepthRequest) (*QueryLiquidityDepthResponse, error) {
	out := new(QueryLiquidityDepthResponse)
	err := c.cc.Invoke(ctx, "/dex.Query/LiquidityDepth", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) LPPosition(ctx context.Context, in *QueryLPPositionRequest) (*QueryLPPositionResponse, error) {
	out := new(QueryLPPositionResponse)
	err := c.cc.Invoke(ctx, "/dex.Query/LPPosition", in, out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
