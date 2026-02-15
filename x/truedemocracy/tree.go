package truedemocracy

import (
    "fmt"
    "sync"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
)

type Node struct {
    ID        string
    Children  []*Node
    Parent    *Node
    Domain    string
    Cache     map[string]interface{}
    Mu        sync.Mutex
    Stake     sdk.Coins
    PubKey    []byte // validator ed25519 public key (32 bytes)
    Operator  string // validator operator address
}

func BuildTree() []*Node {
    nodes := make([]*Node, 7)
    for i := 0; i < 7; i++ {
        nodes[i] = &Node{
            ID:     fmt.Sprintf("Node%d", i),
            Domain: "TestParty",
            Cache:  make(map[string]interface{}),
            Stake:  sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100000)),
        }
    }
    nodes[0].Children = []*Node{nodes[1], nodes[2]}
    nodes[1].Children = []*Node{nodes[3], nodes[4]}
    nodes[2].Children = []*Node{nodes[5], nodes[6]}
    nodes[1].Parent, nodes[2].Parent = nodes[0], nodes[0]
    nodes[3].Parent, nodes[4].Parent = nodes[1], nodes[1]
    nodes[5].Parent, nodes[6].Parent = nodes[2], nodes[2]
    return nodes
}

func (n *Node) PropagateAsync(ctx sdk.Context, k Keeper, domainName, issueName, suggestionName string, rating int, domainPrivKey *ed25519.PrivKey) {
    go func() {
        n.Mu.Lock()
        reward, cache, _ := k.RateProposal(ctx, domainName, issueName, suggestionName, rating, domainPrivKey)
        n.Cache["reward"] = reward
        n.Cache["avg_rating"] = cache["avg_rating"]
        n.Cache["stones"] = cache["stones"]
        n.Cache["treasury"] = cache["treasury"]
        if n.Parent != nil {
            n.Parent.PropagateAsync(ctx, k, domainName, issueName, suggestionName, rating, domainPrivKey)
        }
        n.Mu.Unlock()
    }()
}
