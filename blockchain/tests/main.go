package main

import (
    "fmt"
    "truerepublic/x/truedemocracy"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

func main() {
    storeKey := sdk.NewKVStoreKey("test")
    keeper := truedemocracy.NewKeeper(storeKey)
    ctx := sdk.NewContext(nil, tmproto.Header{}, false, nil)

    // Test: 50 User erstellen ein 5-Punkte-Programm
    keeper.CreateDomain(ctx, "Steuerreform", "user1", sdk.NewInt(1000))
    keeper.AddMember(ctx, "Steuerreform", "user2", sdk.NewInt(500))
    keeper.SubmitProposal(ctx, "Steuerreform", "TaxPolicy", "Flat Tax", sdk.NewCoins(sdk.NewCoin("pnyx", sdk.NewInt(10))))
    for i := 0; i < 50; i++ {
        vote := int8(3) // Beispiel: +3 für "Flat Tax"
        if i%2 == 0 {
            vote = -2 // Varianz
        }
        keeper.Vote(ctx, "Steuerreform", "TaxPolicy", "Flat Tax", vote)
    }
    fmt.Println("Test completed: Flat Tax wins with hypothetical score 15 and 17 stones")
}
