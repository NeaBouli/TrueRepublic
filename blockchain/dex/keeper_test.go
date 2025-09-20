package dex_test

import (
    "testing"

    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/stretchr/testify/require"

    "github.com/NeaBouli/TrueRepublic/blockchain/dex"
)

func TestSwap_HappyPath(t *testing.T) {
    ctx, keeper := dex.SetupKeeperWithPool(t, "PNYX", "ATOM", 1_000_000, 1_000_000)
    out, err := keeper.Swap(ctx, "trader", sdk.NewInt(1000), "PNYX", "ATOM")
    require.NoError(t, err)
    require.True(t, out.Amount.GT(sdk.NewInt(0)))
}

func TestSwap_InvalidDenomFails(t *testing.T) {
    ctx, keeper := dex.SetupKeeperWithPool(t, "PNYX", "ATOM", 1_000_000, 1_000_000)
    _, err := keeper.Swap(ctx, "trader", sdk.NewInt(1000), "XXX", "ATOM")
    require.Error(t, err)
}
