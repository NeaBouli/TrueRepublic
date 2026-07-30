package main

import (
	"testing"

	"cosmossdk.io/log"
	dbm "github.com/cosmos/cosmos-db"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
)

// TestInvariantMetricCoupling guards the two facts that make the
// truerepublic_app_last_successful_invariant_cycle_height metric valid:
//
//  1. crisis is the last EndBlocker in SetOrderEndBlockers, so every custom
//     module settles before invariants are asserted; and
//  2. the crisis invariant check period is one block, so every successful
//     app.mm.EndBlock return proves crisis.AssertInvariants ran at that
//     height without panicking.
//
// If either fact changes, the invariant-cycle signal recorded in EndBlocker
// must be redesigned instead of silently weakening.
func TestInvariantMetricCoupling(t *testing.T) {
	app := NewTrueRepublicApp(log.NewNopLogger(), dbm.NewMemDB(), t.TempDir())

	if period := app.crisisKeeper.InvCheckPeriod(); period != 1 {
		t.Fatalf("crisis invariant check period = %d, want 1", period)
	}
	order := app.mm.OrderEndBlockers
	if len(order) == 0 || order[len(order)-1] != crisistypes.ModuleName {
		t.Fatalf("crisis must be the last EndBlocker, order = %v", order)
	}
}
