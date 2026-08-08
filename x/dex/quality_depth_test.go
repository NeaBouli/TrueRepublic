package dex

import (
	"testing"

	"cosmossdk.io/math"
)

// swapPropertyBound keeps every intermediate product far below the 256-bit
// math.Int overflow limit while covering realistic reserve magnitudes.
const swapPropertyBound = uint64(1) << 40

func positiveBounded(raw uint64) int64 {
	return int64(raw%swapPropertyBound) + 1
}

// checkSwapOutputInvariants asserts the bounding, burn, constant-product, and
// determinism properties every computeSwapOutput result must satisfy.
// Inputs must be positive; the denominator can then never be zero.
func checkSwapOutputInvariants(tb testing.TB, inReserve, outReserve, inputAmt math.Int, outputIsPnyx bool) {
	tb.Helper()

	outAmt, burnAmt := computeSwapOutput(inReserve, outReserve, inputAmt, outputIsPnyx)

	if outAmt.IsNegative() || burnAmt.IsNegative() {
		tb.Fatalf("negative result: out=%s burn=%s", outAmt, burnAmt)
	}
	gross := outAmt.Add(burnAmt)
	// Output is strictly bounded by the out reserve while reserves are positive.
	if !gross.LT(outReserve) {
		tb.Fatalf("gross output %s not strictly below out reserve %s", gross, outReserve)
	}
	// Output must be positive once the input is comparable to the in reserve
	// and the out reserve is non-trivial.
	if inputAmt.GTE(inReserve) && outReserve.GTE(math.NewInt(3)) && !outAmt.IsPositive() {
		tb.Fatalf("non-positive net output for input %s, in reserve %s", inputAmt, inReserve)
	}
	if !outputIsPnyx {
		if !burnAmt.IsZero() {
			tb.Fatalf("non-PNYX output burned %s", burnAmt)
		}
	} else {
		// The burn is exactly floor(gross * BurnBps / 10000) and bounded by gross.
		wantBurn := gross.Mul(math.NewInt(BurnBps)).Quo(math.NewInt(10000))
		if !burnAmt.Equal(wantBurn) || burnAmt.GT(gross) {
			tb.Fatalf("burn %s, want %s bounded by gross %s", burnAmt, wantBurn, gross)
		}
	}
	// The constant product must not decrease: the retained fee keeps
	// (in + input) * (out - gross) >= in * out for the gross output.
	before := inReserve.Mul(outReserve)
	after := inReserve.Add(inputAmt).Mul(outReserve.Sub(gross))
	if after.LT(before) {
		tb.Fatalf("constant product decreased: %s < %s (gross %s)", after, before, gross)
	}
	// The computation is pure and must be deterministic.
	outAgain, burnAgain := computeSwapOutput(inReserve, outReserve, inputAmt, outputIsPnyx)
	if !outAgain.Equal(outAmt) || !burnAgain.Equal(burnAmt) {
		tb.Fatalf("non-deterministic result: (%s,%s) then (%s,%s)", outAmt, burnAmt, outAgain, burnAgain)
	}
}

// TestComputeSwapOutputDeterministicProperties sweeps deterministic bounded
// inputs and asserts bounding, burn, constant-product, monotonicity, and
// determinism properties of computeSwapOutput.
func TestComputeSwapOutputDeterministicProperties(t *testing.T) {
	// Deterministic LCG sweep; every step re-checks all invariants.
	seed := uint64(0x9e3779b97f4a7c15)
	next := func() uint64 {
		seed = seed*6364136223846793005 + 1442695040888963407
		return seed >> 16
	}
	for i := 0; i < 2000; i++ {
		inReserve := math.NewInt(positiveBounded(next()))
		outReserve := math.NewInt(positiveBounded(next()))
		inputAmt := math.NewInt(positiveBounded(next()))
		checkSwapOutputInvariants(t, inReserve, outReserve, inputAmt, i%2 == 0)
	}

	// Monotonicity: net and gross output must be non-decreasing in the input
	// amount for fixed reserves, for both burn and non-burn directions.
	for _, outputIsPnyx := range []bool{false, true} {
		inReserve := math.NewInt(5_000_000)
		outReserve := math.NewInt(7_500_000)
		prevNet := math.ZeroInt()
		prevGross := math.ZeroInt()
		for step := 0; step <= 40; step++ {
			inputAmt := math.NewInt(int64(1) << uint(step))
			outAmt, burnAmt := computeSwapOutput(inReserve, outReserve, inputAmt, outputIsPnyx)
			gross := outAmt.Add(burnAmt)
			if outAmt.LT(prevNet) || gross.LT(prevGross) {
				t.Fatalf("output not monotonic at input %s (outputIsPnyx=%t): net %s<%s gross %s<%s",
					inputAmt, outputIsPnyx, outAmt, prevNet, gross, prevGross)
			}
			prevNet = outAmt
			prevGross = gross
		}
	}

	// Output must not increase when the input-side reserve grows and must not
	// decrease when the output-side reserve grows.
	outReserve := math.NewInt(7_500_000)
	inputAmt := math.NewInt(50_000)
	prevOut := math.NewInt(1 << 62)
	for step := 0; step <= 30; step++ {
		inReserve := math.NewInt(int64(1_000_000 + step*250_000))
		outAmt, burnAmt := computeSwapOutput(inReserve, outReserve, inputAmt, true)
		if outAmt.Add(burnAmt).GT(prevOut) {
			t.Fatalf("output increased with larger in reserve %s: gross %s > %s", inReserve, outAmt.Add(burnAmt), prevOut)
		}
		prevOut = outAmt.Add(burnAmt)
	}
	inReserve := math.NewInt(5_000_000)
	prevOut = math.ZeroInt()
	for step := 0; step <= 30; step++ {
		outReserve := math.NewInt(int64(1_000_000 + step*250_000))
		outAmt, burnAmt := computeSwapOutput(inReserve, outReserve, inputAmt, true)
		if outAmt.Add(burnAmt).LT(prevOut) {
			t.Fatalf("output decreased with larger out reserve %s: gross %s < %s", outReserve, outAmt.Add(burnAmt), prevOut)
		}
		prevOut = outAmt.Add(burnAmt)
	}
}

// FuzzComputeSwapOutput fuzzes the pure constant-product computation with
// bounded positive inputs so the denominator is never zero and no 256-bit
// overflow is reachable.
func FuzzComputeSwapOutput(f *testing.F) {
	f.Add(uint64(1), uint64(1), uint64(1), true)
	f.Add(uint64(1_000_000), uint64(1_000_000), uint64(10_000), false)
	f.Add(uint64(1_000_000), uint64(1_000_000), uint64(10_000), true)
	f.Add(swapPropertyBound-1, uint64(1), swapPropertyBound-1, true)
	f.Add(uint64(1), swapPropertyBound-1, swapPropertyBound-1, false)
	f.Add(uint64(3), uint64(3), uint64(2), true)
	f.Fuzz(func(t *testing.T, rawIn, rawOut, rawAmt uint64, outputIsPnyx bool) {
		inReserve := math.NewInt(positiveBounded(rawIn))
		outReserve := math.NewInt(positiveBounded(rawOut))
		inputAmt := math.NewInt(positiveBounded(rawAmt))
		checkSwapOutputInvariants(t, inReserve, outReserve, inputAmt, outputIsPnyx)
	})
}
