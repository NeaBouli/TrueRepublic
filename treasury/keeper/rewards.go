package keeper

import "cosmossdk.io/math"

// Whitepaper appendix 8.1.1 — global constants and default values.
const (
	CDom      int64 = 2          // Factor for initial domain treasury calculation
	CPut      int64 = 15         // Factor for put price (caps user count in price formula)
	CEarn     int64 = 1000       // Factor for rewards (voting, stoning)
	StakeMin  int64 = 100_000    // Required node stake in PNYX
	SupplyMax int64 = 21_000_000 // Fixed maximum PNYX supply

	SecondsPerYear int64 = 31_557_600 // 365.25 * 24 * 60 * 60
)

var (
	ApyDom  = math.LegacyNewDecWithPrec(25, 2) // 0.25 — initial domain interest rate per year
	ApyNode = math.LegacyNewDecWithPrec(1, 1)  // 0.10 — initial node staking reward rate per year
)

// CalcReward computes the reward for a single evaluation (eq.2).
//
//	p_rew = treasure / c_earn
//
// The reward is a fraction of the current treasury. With c_earn = 1000 the
// treasury needs at least 1000 PNYX to pay out a 1 PNYX reward. This ensures
// the treasury drains gradually — after 694 payouts it reaches 50 % of its
// initial value (see whitepaper §8.1.4).
func CalcReward(treasure math.Int) math.Int {
	if !treasure.IsPositive() {
		return math.ZeroInt()
	}
	return treasure.Quo(math.NewInt(CEarn))
}

// CalcPutPrice computes the cost for posting content (eq.3).
//
//	p_put = min(p_rew * c_put, p_rew * n_user)
//
// The poster pays for the evaluation of their idea by the domain members.
// The effective multiplier is min(c_put, n_user), so in small domains the
// price scales with users while c_put = 15 caps it in larger domains.
func CalcPutPrice(treasure math.Int, nUser int64) math.Int {
	pRew := CalcReward(treasure)
	if pRew.IsZero() || nUser <= 0 {
		return math.ZeroInt()
	}

	byCPut := pRew.Mul(math.NewInt(CPut))
	byUsers := pRew.Mul(math.NewInt(nUser))

	if byCPut.LT(byUsers) {
		return byCPut
	}
	return byUsers
}

// CalcDomainCost computes the minimum initial treasury to open a domain (eq.1).
//
//	p_dom = fee * c_dom * c_earn
//
// With c_dom = 2 and c_earn = 1000 the initial treasury must be 2000× the
// base transaction fee. This guarantees that each VoteToEarn reward pays back
// at least twice the fee the voter spent, keeping participation profitable.
func CalcDomainCost(fee math.Int) math.Int {
	if !fee.IsPositive() {
		return math.ZeroInt()
	}
	return fee.Mul(math.NewInt(CDom)).Mul(math.NewInt(CEarn))
}

// releaseDecay returns (1 - f_release) where f_release = release / supply_max.
// The result is clamped to [0, 1]. As more coins enter circulation the decay
// factor shrinks, reducing all interest rates proportionally.
func releaseDecay(release math.Int) math.LegacyDec {
	if !release.IsPositive() {
		return math.LegacyOneDec()
	}
	fRelease := math.LegacyNewDecFromInt(release).Quo(
		math.LegacyNewDecFromInt(math.NewInt(SupplyMax)),
	)
	decay := math.LegacyOneDec().Sub(fRelease)
	if decay.IsNegative() {
		return math.LegacyZeroDec()
	}
	return decay
}

// timeInYears converts an elapsed duration in seconds to a fractional year
// using the whitepaper convention: 1 year = 365.25 * 86400 = 31 557 600 s.
func timeInYears(elapsedSeconds int64) math.LegacyDec {
	if elapsedSeconds <= 0 {
		return math.LegacyZeroDec()
	}
	return math.LegacyNewDec(elapsedSeconds).Quo(
		math.LegacyNewDec(SecondsPerYear),
	)
}

// CalcDomainInterest computes the coin release for an active domain (eq.4).
//
//	i_dom = min(treasure * apy_dom * T_dom * [1 - f_release], payout)
//
// Parameters:
//   - treasure: current coins in the domain treasury
//   - payout:   total domain payouts during the interval (RateToEarn + VoteToEarn)
//   - release:  total coins currently in circulation
//   - elapsedSeconds: length of the payout interval
//
// The interest is capped by actual payouts so only active domains receive
// new coins. If a domain had zero payouts during the interval it earns
// zero interest regardless of its treasury size.
func CalcDomainInterest(treasure, payout, release math.Int, elapsedSeconds int64) math.Int {
	if !treasure.IsPositive() || !payout.IsPositive() {
		return math.ZeroInt()
	}

	interest := math.LegacyNewDecFromInt(treasure).
		Mul(ApyDom).
		Mul(timeInYears(elapsedSeconds)).
		Mul(releaseDecay(release)).
		TruncateInt()

	if interest.GT(payout) {
		return payout
	}
	return interest
}

// CalcNodeReward computes the staking reward for a validator node (eq.5).
//
//	i_node = stake * apy_node * T_node * [1 - f_release]
//
// Parameters:
//   - stake:          coins staked by the node
//   - release:        total coins currently in circulation
//   - elapsedSeconds: length of the payout interval
//
// The reward diminishes as f_release grows, tying inflation to demand:
// when the network is young (few coins released) validators earn more,
// and as the supply approaches supply_max the rate drops toward zero.
func CalcNodeReward(stake, release math.Int, elapsedSeconds int64) math.Int {
	if !stake.IsPositive() {
		return math.ZeroInt()
	}

	return math.LegacyNewDecFromInt(stake).
		Mul(ApyNode).
		Mul(timeInYears(elapsedSeconds)).
		Mul(releaseDecay(release)).
		TruncateInt()
}
