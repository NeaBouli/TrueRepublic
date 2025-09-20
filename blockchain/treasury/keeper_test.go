package treasury_test

import (
    "testing"

    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/stretchr/testify/require"

    "github.com/NeaBouli/TrueRepublic/blockchain/treasury"
)

func TestDeposit_HappyPath(t *testing.T) {
    ctx, keeper := treasury.SetupKeeper(t)
    err := keeper.Deposit(ctx, sdk.NewInt64Coin("PNYX", 1000))
    require.NoError(t, err)
}

func TestWithdraw_InsufficientFails(t *testing.T) {
    ctx, keeper := treasury.SetupKeeper(t)
    err := keeper.Withdraw(ctx, sdk.NewInt64Coin("PNYX", 9999))
    require.Error(t, err)
}
