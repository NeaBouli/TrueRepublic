/**
 * DEX types mirroring Go x/dex module JSON responses.
 * Field names use snake_case to match Go json tags exactly.
 */

/** Pool mirrors Go `dex.Pool` */
export interface Pool {
  pnyx_reserve: string;
  asset_reserve: string;
  asset_denom: string;
  total_shares: string;
  total_burned: string;
  asset_symbol?: string;
  swap_count: number;
  total_volume_pnyx: string;
}

/** RegisteredAsset mirrors Go `dex.RegisteredAsset` */
export interface RegisteredAsset {
  ibc_denom: string;
  symbol: string;
  name: string;
  decimals: number;
  origin_chain: string;
  ibc_channel: string;
  trading_enabled: boolean;
  registered_height: number;
  registered_by: string;
}

/** PoolStats from the pool_stats query response */
export interface PoolStats {
  asset_denom: string;
  asset_symbol: string;
  swap_count: number;
  total_volume_pnyx: string;
  total_fees_earned: string;
  total_burned: string;
  pnyx_reserve: string;
  asset_reserve: string;
  spot_price_per_million: string;
  total_shares: string;
}

/** SwapEstimate from the estimate_swap query response */
export interface SwapEstimate {
  expected_output: string;
  route: string[];
  route_symbols: string[];
  hops: number;
}

/** LPPosition from the lp_position query response */
export interface LPPosition {
  asset_denom: string;
  shares: string;
  pnyx_value: string;
  asset_value: string;
  share_of_pool_bps: number;
}

/** SpotPrice from the spot_price query response */
export interface SpotPrice {
  input_denom: string;
  output_denom: string;
  price_per_million: string;
  input_symbol: string;
  output_symbol: string;
  route: string[];
}

/** Params for single-hop swap execution */
export interface SwapParams {
  inputDenom: string;
  inputAmount: string;
  outputDenom: string;
  minOutput?: string;
}

/** Params for multi-hop swap execution (cross-asset via PNYX hub) */
export interface MultiHopSwapParams {
  inputDenom: string;
  inputAmount: string;
  outputDenom: string;
  minOutput?: string;
}

/**
 * Params for MsgAddLiquidity (Go: add_liquidity).
 * All pools are PNYX-paired: sender deposits pnyx_amt + asset_amt.
 */
export interface AddLiquidityParams {
  asset_denom: string;
  pnyx_amt: string;
  asset_amt: string;
}

/**
 * Params for MsgRemoveLiquidity (Go: remove_liquidity).
 * Sender redeems LP shares for proportional pool assets.
 */
export interface RemoveLiquidityParams {
  asset_denom: string;
  shares: string;
}

/** Extended LP position detail for the management UI */
export interface LPPositionDetail {
  asset_denom: string;
  shares: string;
  share_of_pool_bps: number;
  pnyx_value: string;
  asset_value: string;
  total_value_pnyx: string;
}

/** Pool analytics for the pool detail view */
export interface PoolAnalytics {
  asset_denom: string;
  asset_symbol: string;
  total_volume_pnyx: string;
  total_fees_earned: string;
  total_burned: string;
  swap_count: number;
  spot_price_per_million: string;
}
