package dex

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// --- MsgCreatePool ---

type MsgCreatePool struct {
	Sender     sdk.AccAddress `json:"sender"`
	AssetDenom string         `json:"asset_denom"`
	PnyxAmt    int64          `json:"pnyx_amt"`
	AssetAmt   int64          `json:"asset_amt"`
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
	Sender      sdk.AccAddress `json:"sender"`
	InputDenom  string         `json:"input_denom"`
	InputAmt    int64          `json:"input_amt"`
	OutputDenom string         `json:"output_denom"`
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
	Sender     sdk.AccAddress `json:"sender"`
	AssetDenom string         `json:"asset_denom"`
	PnyxAmt    int64          `json:"pnyx_amt"`
	AssetAmt   int64          `json:"asset_amt"`
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
	Sender     sdk.AccAddress `json:"sender"`
	AssetDenom string         `json:"asset_denom"`
	Shares     int64          `json:"shares"`
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
