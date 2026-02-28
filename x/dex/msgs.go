package dex

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// --- MsgCreatePool ---

type MsgCreatePool struct {
	Sender     sdk.AccAddress `protobuf:"bytes,1,opt,name=sender,proto3,casttype=github.com/cosmos/cosmos-sdk/types.AccAddress" json:"sender"`
	AssetDenom string         `protobuf:"bytes,2,opt,name=asset_denom,json=assetDenom,proto3" json:"asset_denom"`
	PnyxAmt    int64          `protobuf:"varint,3,opt,name=pnyx_amt,json=pnyxAmt,proto3" json:"pnyx_amt"`
	AssetAmt   int64          `protobuf:"varint,4,opt,name=asset_amt,json=assetAmt,proto3" json:"asset_amt"`
}

func (m *MsgCreatePool) ProtoMessage()               {}
func (m *MsgCreatePool) Reset()                      { *m = MsgCreatePool{} }
func (m *MsgCreatePool) String() string              { b, _ := json.Marshal(m); return string(b) }
func (m MsgCreatePool) Route() string                { return ModuleName }
func (m MsgCreatePool) Type() string                 { return "create_pool" }
func (m MsgCreatePool) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
func (m MsgCreatePool) ValidateBasic() error {
	if m.AssetDenom == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("asset_denom is required")
	}
	if m.PnyxAmt <= 0 || m.AssetAmt <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("both amounts must be positive")
	}
	return nil
}

// --- MsgSwap ---

type MsgSwap struct {
	Sender      sdk.AccAddress `protobuf:"bytes,1,opt,name=sender,proto3,casttype=github.com/cosmos/cosmos-sdk/types.AccAddress" json:"sender"`
	InputDenom  string         `protobuf:"bytes,2,opt,name=input_denom,json=inputDenom,proto3" json:"input_denom"`
	InputAmt    int64          `protobuf:"varint,3,opt,name=input_amt,json=inputAmt,proto3" json:"input_amt"`
	OutputDenom string         `protobuf:"bytes,4,opt,name=output_denom,json=outputDenom,proto3" json:"output_denom"`
}

func (m *MsgSwap) ProtoMessage()               {}
func (m *MsgSwap) Reset()                      { *m = MsgSwap{} }
func (m *MsgSwap) String() string              { b, _ := json.Marshal(m); return string(b) }
func (m MsgSwap) Route() string                { return ModuleName }
func (m MsgSwap) Type() string                 { return "swap" }
func (m MsgSwap) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
func (m MsgSwap) ValidateBasic() error {
	if m.InputDenom == "" || m.OutputDenom == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("input_denom and output_denom are required")
	}
	if m.InputAmt <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("input_amt must be positive")
	}
	return nil
}

// --- MsgAddLiquidity ---

type MsgAddLiquidity struct {
	Sender     sdk.AccAddress `protobuf:"bytes,1,opt,name=sender,proto3,casttype=github.com/cosmos/cosmos-sdk/types.AccAddress" json:"sender"`
	AssetDenom string         `protobuf:"bytes,2,opt,name=asset_denom,json=assetDenom,proto3" json:"asset_denom"`
	PnyxAmt    int64          `protobuf:"varint,3,opt,name=pnyx_amt,json=pnyxAmt,proto3" json:"pnyx_amt"`
	AssetAmt   int64          `protobuf:"varint,4,opt,name=asset_amt,json=assetAmt,proto3" json:"asset_amt"`
}

func (m *MsgAddLiquidity) ProtoMessage()               {}
func (m *MsgAddLiquidity) Reset()                      { *m = MsgAddLiquidity{} }
func (m *MsgAddLiquidity) String() string              { b, _ := json.Marshal(m); return string(b) }
func (m MsgAddLiquidity) Route() string                { return ModuleName }
func (m MsgAddLiquidity) Type() string                 { return "add_liquidity" }
func (m MsgAddLiquidity) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
func (m MsgAddLiquidity) ValidateBasic() error {
	if m.AssetDenom == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("asset_denom is required")
	}
	if m.PnyxAmt <= 0 || m.AssetAmt <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("both amounts must be positive")
	}
	return nil
}

// --- MsgRemoveLiquidity ---

type MsgRemoveLiquidity struct {
	Sender     sdk.AccAddress `protobuf:"bytes,1,opt,name=sender,proto3,casttype=github.com/cosmos/cosmos-sdk/types.AccAddress" json:"sender"`
	AssetDenom string         `protobuf:"bytes,2,opt,name=asset_denom,json=assetDenom,proto3" json:"asset_denom"`
	Shares     int64          `protobuf:"varint,3,opt,name=shares,proto3" json:"shares"`
}

func (m *MsgRemoveLiquidity) ProtoMessage()               {}
func (m *MsgRemoveLiquidity) Reset()                      { *m = MsgRemoveLiquidity{} }
func (m *MsgRemoveLiquidity) String() string              { b, _ := json.Marshal(m); return string(b) }
func (m MsgRemoveLiquidity) Route() string                { return ModuleName }
func (m MsgRemoveLiquidity) Type() string                 { return "remove_liquidity" }
func (m MsgRemoveLiquidity) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
func (m MsgRemoveLiquidity) ValidateBasic() error {
	if m.AssetDenom == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("asset_denom is required")
	}
	if m.Shares <= 0 {
		return sdkerrors.ErrInvalidRequest.Wrap("shares must be positive")
	}
	return nil
}

// --- MsgRegisterAsset ---

type MsgRegisterAsset struct {
	Sender      sdk.AccAddress `protobuf:"bytes,1,opt,name=sender,proto3,casttype=github.com/cosmos/cosmos-sdk/types.AccAddress" json:"sender"`
	IBCDenom    string         `protobuf:"bytes,2,opt,name=ibc_denom,json=ibcDenom,proto3" json:"ibc_denom"`
	Symbol      string         `protobuf:"bytes,3,opt,name=symbol,proto3" json:"symbol"`
	Name        string         `protobuf:"bytes,4,opt,name=name,proto3" json:"name"`
	Decimals    uint32         `protobuf:"varint,5,opt,name=decimals,proto3" json:"decimals"`
	OriginChain string         `protobuf:"bytes,6,opt,name=origin_chain,json=originChain,proto3" json:"origin_chain"`
	IBCChannel  string         `protobuf:"bytes,7,opt,name=ibc_channel,json=ibcChannel,proto3" json:"ibc_channel"`
}

func (m *MsgRegisterAsset) ProtoMessage()               {}
func (m *MsgRegisterAsset) Reset()                      { *m = MsgRegisterAsset{} }
func (m *MsgRegisterAsset) String() string              { b, _ := json.Marshal(m); return string(b) }
func (m MsgRegisterAsset) Route() string                { return ModuleName }
func (m MsgRegisterAsset) Type() string                 { return "register_asset" }
func (m MsgRegisterAsset) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
func (m MsgRegisterAsset) ValidateBasic() error {
	if m.Sender.Empty() {
		return sdkerrors.ErrInvalidRequest.Wrap("sender is required")
	}
	if m.IBCDenom == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("ibc_denom is required")
	}
	if m.Symbol == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("symbol is required")
	}
	if m.Decimals > 18 {
		return sdkerrors.ErrInvalidRequest.Wrap("decimals cannot exceed 18")
	}
	return nil
}

// --- MsgUpdateAssetStatus ---

type MsgUpdateAssetStatus struct {
	Sender   sdk.AccAddress `protobuf:"bytes,1,opt,name=sender,proto3,casttype=github.com/cosmos/cosmos-sdk/types.AccAddress" json:"sender"`
	IBCDenom string         `protobuf:"bytes,2,opt,name=ibc_denom,json=ibcDenom,proto3" json:"ibc_denom"`
	Enabled  bool           `protobuf:"varint,3,opt,name=enabled,proto3" json:"enabled"`
}

func (m *MsgUpdateAssetStatus) ProtoMessage()               {}
func (m *MsgUpdateAssetStatus) Reset()                      { *m = MsgUpdateAssetStatus{} }
func (m *MsgUpdateAssetStatus) String() string              { b, _ := json.Marshal(m); return string(b) }
func (m MsgUpdateAssetStatus) Route() string                { return ModuleName }
func (m MsgUpdateAssetStatus) Type() string                 { return "update_asset_status" }
func (m MsgUpdateAssetStatus) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{m.Sender} }
func (m MsgUpdateAssetStatus) ValidateBasic() error {
	if m.Sender.Empty() {
		return sdkerrors.ErrInvalidRequest.Wrap("sender is required")
	}
	if m.IBCDenom == "" {
		return sdkerrors.ErrInvalidRequest.Wrap("ibc_denom is required")
	}
	return nil
}
