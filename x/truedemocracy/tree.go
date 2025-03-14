package truedemocracy

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
)

func RateProposal(k Keeper, ctx sdk.Context, domainName, issueName, suggestionName, voter string, rating int, privKey *ed25519.PrivKey) (sdk.Coins, map[string]interface{}, error) {
    return k.RateProposal(ctx, domainName, issueName, suggestionName, voter, rating, privKey)
}
