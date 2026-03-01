use cosmwasm_std::Uint128;

pub const SWAP_FEE_BPS: u64 = 30;
pub const BURN_BPS: u64 = 100;

/// Compute swap output using constant-product AMM formula.
/// Matches Go computeSwapOutput in x/dex/keeper.go.
pub fn compute_swap_output(
    in_reserve: Uint128,
    out_reserve: Uint128,
    input_amt: Uint128,
    output_is_pnyx: bool,
) -> (Uint128, Uint128) {
    let input = input_amt.u128();
    let fee = SWAP_FEE_BPS as u128;
    let input_with_fee = input * (10000 - fee);
    let numerator = out_reserve.u128() * input_with_fee;
    let denominator = in_reserve.u128() * 10000 + input_with_fee;
    let mut output = numerator / denominator;

    let mut burn = 0u128;
    if output_is_pnyx {
        burn = output * BURN_BPS as u128 / 10000;
        output -= burn;
    }

    (Uint128::new(output), Uint128::new(burn))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_equal_pool() {
        let (out, burn) = compute_swap_output(
            Uint128::new(1_000_000),
            Uint128::new(1_000_000),
            Uint128::new(1000),
            false,
        );
        assert_eq!(out.u128(), 996); // matches Go
        assert_eq!(burn.u128(), 0);
    }

    #[test]
    fn test_output_is_pnyx_burns() {
        let (out, burn) = compute_swap_output(
            Uint128::new(1_000_000),
            Uint128::new(1_000_000),
            Uint128::new(1000),
            true,
        );
        // Without burn: 996, burn = 996 * 100 / 10000 = 9
        assert_eq!(burn.u128(), 9);
        assert_eq!(out.u128(), 987); // 996 - 9
    }

    #[test]
    fn test_large_trade_impact() {
        let (out, _) = compute_swap_output(
            Uint128::new(1_000_000),
            Uint128::new(1_000_000),
            Uint128::new(500_000),
            false,
        );
        // Large trade: significant price impact, output << 500k
        assert!(out.u128() < 400_000);
        assert!(out.u128() > 300_000);
    }
}
