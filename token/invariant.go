package token

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SupplyCapInvariant detects canonical bank supply outside the fixed PNYX cap.
func SupplyCapInvariant(bank IssuanceBankKeeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		if bank == nil {
			return "PNYX supply invariant has no bank keeper", true
		}
		supply := bank.GetSupply(ctx, BaseDenom).Amount
		broken := supply.IsNegative() || supply.GT(MaxSupply())
		return fmt.Sprintf("PNYX supply %s, cap %s", supply, MaxSupply()), broken
	}
}
