use cosmwasm_std::Uint128;

/// Mock constant-product AMM pool matching the Go DEX formula.
/// Fee: 0.3% (SwapFeeBps=30), Burn: 1% on PNYX output (BurnBps=100).
pub struct MockPool {
    pub pnyx_reserve: Uint128,
    pub asset_reserve: Uint128,
    pub asset_denom: String,
}

impl MockPool {
    pub fn new(asset_denom: &str, pnyx: u128, asset: u128) -> Self {
        Self {
            pnyx_reserve: Uint128::new(pnyx),
            asset_reserve: Uint128::new(asset),
            asset_denom: asset_denom.to_string(),
        }
    }

    /// Compute swap output using constant-product AMM formula.
    /// Matches Go: out = outReserve * in * (10000 - fee) / (inReserve * 10000 + in * (10000 - fee))
    pub fn compute_swap_output(
        &self,
        input_denom: &str,
        input_amount: Uint128,
    ) -> Result<(Uint128, Uint128), String> {
        let fee_bps: u128 = 30;
        let burn_bps: u128 = 100;

        let (in_reserve, out_reserve, output_is_pnyx) = if input_denom == "pnyx" {
            (self.pnyx_reserve, self.asset_reserve, false)
        } else if input_denom == self.asset_denom {
            (self.asset_reserve, self.pnyx_reserve, true)
        } else {
            return Err(format!("unknown denom: {}", input_denom));
        };

        let input = input_amount.u128();
        let input_with_fee = input * (10000 - fee_bps);
        let numerator = out_reserve.u128() * input_with_fee;
        let denominator = in_reserve.u128() * 10000 + input_with_fee;
        let mut output = numerator / denominator;

        let mut burn = 0u128;
        if output_is_pnyx {
            burn = output * burn_bps / 10000;
            output -= burn;
        }

        Ok((Uint128::new(output), Uint128::new(burn)))
    }

    /// Compute marginal spot price (output per 1M input units).
    pub fn spot_price(&self, input_denom: &str) -> Result<Uint128, String> {
        let fee_bps: u128 = 30;
        let ref_amt: u128 = 1_000_000;

        let (in_reserve, out_reserve) = if input_denom == "pnyx" {
            (self.pnyx_reserve.u128(), self.asset_reserve.u128())
        } else if input_denom == self.asset_denom {
            (self.asset_reserve.u128(), self.pnyx_reserve.u128())
        } else {
            return Err(format!("unknown denom: {}", input_denom));
        };

        let price = out_reserve * ref_amt * (10000 - fee_bps) / (in_reserve * 10000);
        Ok(Uint128::new(price))
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_equal_pool_swap() {
        let pool = MockPool::new("atom", 1_000_000, 1_000_000);
        let (output, burn) = pool
            .compute_swap_output("pnyx", Uint128::new(1000))
            .unwrap();
        // With 0.3% fee: out = 1M * 1000 * 9970 / (1M * 10000 + 1000 * 9970) = 996
        assert!(output.u128() > 990 && output.u128() < 1000);
        assert_eq!(burn.u128(), 0); // output is atom, no burn
    }

    #[test]
    fn test_swap_to_pnyx_burns() {
        let pool = MockPool::new("atom", 1_000_000, 1_000_000);
        let (output, burn) = pool
            .compute_swap_output("atom", Uint128::new(1000))
            .unwrap();
        assert!(output.u128() > 0);
        assert!(burn.u128() > 0); // output is PNYX, should burn
    }

    #[test]
    fn test_spot_price_equal_pool() {
        let pool = MockPool::new("atom", 1_000_000, 1_000_000);
        let price = pool.spot_price("pnyx").unwrap();
        // price = 1M * 1M * 9970 / (1M * 10000) = 997000
        assert_eq!(price.u128(), 997_000);
    }

    #[test]
    fn test_unknown_denom_error() {
        let pool = MockPool::new("atom", 1_000_000, 1_000_000);
        assert!(pool.compute_swap_output("btc", Uint128::new(1000)).is_err());
    }
}
