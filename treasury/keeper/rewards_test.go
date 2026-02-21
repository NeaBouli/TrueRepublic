package keeper

import (
	"testing"

	"cosmossdk.io/math"
)

// Whitepaper §8.1.4 worked example: treasury of 2000× fee, c_earn = 1000,
// so each reward = treasure/1000. The compensation point (reward == fee)
// is reached after 694 payouts when the treasury is at 50 %.

func TestCalcReward(t *testing.T) {
	tests := []struct {
		name     string
		treasure int64
		want     int64
	}{
		{"standard treasury", 500_000, 500},
		{"minimum for nonzero reward", 1000, 1},
		{"below minimum", 999, 0},
		{"zero treasury", 0, 0},
		{"negative treasury", -1, 0},
		{"large treasury", 21_000_000, 21_000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcReward(math.NewInt(tt.treasure))
			if !got.Equal(math.NewInt(tt.want)) {
				t.Errorf("CalcReward(%d) = %s, want %d", tt.treasure, got, tt.want)
			}
		})
	}
}

func TestCalcPutPrice(t *testing.T) {
	tests := []struct {
		name     string
		treasure int64
		nUser    int64
		want     int64
	}{
		// p_rew = 500000/1000 = 500
		{"small domain, capped by nUser", 500_000, 10, 5_000},   // 500 * 10
		{"large domain, capped by cPut", 500_000, 20, 7_500},    // 500 * 15
		{"exactly cPut users", 500_000, 15, 7_500},               // 500 * 15 == 500 * 15
		{"single user", 500_000, 1, 500},                          // 500 * 1
		{"zero users", 500_000, 0, 0},
		{"empty treasury", 0, 10, 0},
		{"treasury below c_earn", 999, 10, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcPutPrice(math.NewInt(tt.treasure), tt.nUser)
			if !got.Equal(math.NewInt(tt.want)) {
				t.Errorf("CalcPutPrice(%d, %d) = %s, want %d", tt.treasure, tt.nUser, got, tt.want)
			}
		})
	}
}

func TestCalcDomainCost(t *testing.T) {
	tests := []struct {
		name string
		fee  int64
		want int64
	}{
		// p_dom = fee * 2 * 1000
		{"unit fee", 1, 2_000},
		{"typical fee", 10, 20_000},
		{"zero fee", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcDomainCost(math.NewInt(tt.fee))
			if !got.Equal(math.NewInt(tt.want)) {
				t.Errorf("CalcDomainCost(%d) = %s, want %d", tt.fee, got, tt.want)
			}
		})
	}
}

func TestCalcNodeReward(t *testing.T) {
	oneDay := int64(86400)
	oneYear := int64(SecondsPerYear)

	tests := []struct {
		name    string
		stake   int64
		release int64
		elapsed int64
		want    int64
	}{
		// Full year, zero release: 100000 * 0.1 * 1.0 * 1.0 = 10000
		{"full year, no release", 100_000, 0, oneYear, 10_000},

		// Full year, 1M released: 100000 * 0.1 * 1.0 * (1 - 1M/21M)
		// = 10000 * 20/21 = 9523.80... → 9523
		{"full year, 1M released", 100_000, 1_000_000, oneYear, 9523},

		// One day, zero release: 100000 * 0.1 * (86400/31557600) * 1.0
		// = 10000 * 0.002737... = 27.37... → 27
		{"one day, no release", 100_000, 0, oneDay, 27},

		// All coins released: decay = 0 → reward = 0
		{"all coins released", 100_000, SupplyMax, oneYear, 0},

		// Zero stake
		{"zero stake", 0, 0, oneYear, 0},

		// Zero elapsed
		{"zero elapsed", 100_000, 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcNodeReward(math.NewInt(tt.stake), math.NewInt(tt.release), tt.elapsed)
			if !got.Equal(math.NewInt(tt.want)) {
				t.Errorf("CalcNodeReward(stake=%d, release=%d, elapsed=%d) = %s, want %d",
					tt.stake, tt.release, tt.elapsed, got, tt.want)
			}
		})
	}
}

func TestCalcDomainInterest(t *testing.T) {
	oneYear := int64(SecondsPerYear)

	tests := []struct {
		name     string
		treasure int64
		payout   int64
		release  int64
		elapsed  int64
		want     int64
	}{
		// Full year, zero release: 500000 * 0.25 * 1.0 * 1.0 = 125000
		// payout = 200000 → min(125000, 200000) = 125000
		{"full year, uncapped", 500_000, 200_000, 0, oneYear, 125_000},

		// Same but payout < interest → capped at payout
		{"full year, capped by payout", 500_000, 50_000, 0, oneYear, 50_000},

		// Full year, 1M released: 500000 * 0.25 * 1.0 * (20/21)
		// = 125000 * 20/21 = 119047.61... → 119047
		{"full year, 1M released", 500_000, 200_000, 1_000_000, oneYear, 119_047},

		// All coins released: decay = 0
		{"all released", 500_000, 200_000, SupplyMax, oneYear, 0},

		// Zero payout → 0 (inactive domain)
		{"no activity", 500_000, 0, 0, oneYear, 0},

		// Zero treasury
		{"empty treasury", 0, 1000, 0, oneYear, 0},

		// Zero elapsed
		{"zero elapsed", 500_000, 1000, 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalcDomainInterest(
				math.NewInt(tt.treasure), math.NewInt(tt.payout),
				math.NewInt(tt.release), tt.elapsed,
			)
			if !got.Equal(math.NewInt(tt.want)) {
				t.Errorf("CalcDomainInterest(treasure=%d, payout=%d, release=%d, elapsed=%d) = %s, want %d",
					tt.treasure, tt.payout, tt.release, tt.elapsed, got, tt.want)
			}
		})
	}
}

// Verify the release decay links inflation to demand: as f_release grows
// from 0 → 1, both node and domain interest approach zero.
func TestReleaseDecayReducesRewards(t *testing.T) {
	stake := math.NewInt(100_000)
	oneYear := int64(SecondsPerYear)

	prev := CalcNodeReward(stake, math.ZeroInt(), oneYear)
	for pct := int64(10); pct <= 100; pct += 10 {
		released := math.NewInt(SupplyMax * pct / 100)
		cur := CalcNodeReward(stake, released, oneYear)
		if cur.GT(prev) {
			t.Errorf("node reward increased at %d%% release: %s > %s", pct, cur, prev)
		}
		prev = cur
	}
	// At 100 % release, reward must be zero
	if !prev.IsZero() {
		t.Errorf("node reward at 100%% release = %s, want 0", prev)
	}
}

// Verify the whitepaper drainage claim: starting at initial treasury T,
// after n payouts each draining treasure/c_earn, the treasury reaches 50 %
// at n = 694 (for c_earn = 1000).
func TestTreasuryDrainage(t *testing.T) {
	initial := math.NewInt(2_000_000) // large enough to avoid rounding to zero
	treasury := initial

	for i := 0; i < 694; i++ {
		reward := CalcReward(treasury)
		if reward.IsZero() {
			t.Fatalf("reward hit zero at payout %d, treasury %s", i, treasury)
		}
		treasury = treasury.Sub(reward)
	}

	// After 694 payouts treasury should be ≈ 50 % of initial.
	// Integer division introduces rounding so allow 49-51 %.
	half := initial.Quo(math.NewInt(2))
	lower := half.Mul(math.NewInt(98)).Quo(math.NewInt(100)) // 49 %
	upper := half.Mul(math.NewInt(102)).Quo(math.NewInt(100)) // 51 %

	if treasury.LT(lower) || treasury.GT(upper) {
		pct := treasury.Mul(math.NewInt(100)).Quo(initial)
		t.Errorf("after 694 payouts treasury = %s (%s%%), want ≈50%%", treasury, pct)
	}
}
