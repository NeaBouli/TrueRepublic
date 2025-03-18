package ibc

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type Handler struct {
    storeKey sdk.StoreKey
}

func NewHandler(storeKey sdk.StoreKey) Handler {
    return Handler{storeKey: storeKey}
}

func (h Handler) HandlePacket(ctx sdk.Context, packet Packet) error {
    store := ctx.KVStore(h.storeKey)
    store.Set([]byte("packet:"+packet.ID), packet.Data)
    return nil
}

type Packet struct {
    ID   string
    Data []byte
}
