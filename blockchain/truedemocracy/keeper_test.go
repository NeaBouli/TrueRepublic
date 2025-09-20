package truedemocracy_test

import (
    "testing"

    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/stretchr/testify/require"

    "github.com/NeaBouli/TrueRepublic/blockchain/truedemocracy"
)

func TestCreateProposal_HappyPath(t *testing.T) {
    ctx, keeper := truedemocracy.SetupKeeper(t)
    err := keeper.CreateProposal(ctx, "Test proposal")
    require.NoError(t, err)
}

func TestCreateProposal_EmptyFails(t *testing.T) {
    ctx, keeper := truedemocracy.SetupKeeper(t)
    err := keeper.CreateProposal(ctx, "")
    require.Error(t, err)
}
