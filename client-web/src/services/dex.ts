import type { ChainConfig } from '@/types/chain';
import type {
  Pool,
  RegisteredAsset,
  PoolStats,
  SwapEstimate,
  LPPosition,
  SpotPrice,
} from '@/types/dex';
import {
  expectQueryArray,
  expectQueryBoolean,
  expectQueryNumber,
  expectQueryRecord,
  expectQueryString,
  expectQueryStringArray,
  ModuleQueryClient,
  QUERY_PATHS,
} from './moduleQuery';

function expectPool(path: string, value: unknown): Pool {
  const pool = expectQueryRecord(path, value);
  for (const field of [
    'pnyx_reserve',
    'asset_reserve',
    'asset_denom',
    'total_shares',
    'total_burned',
    'total_volume_pnyx',
  ]) {
    expectQueryString(path, field, pool[field]);
  }
  if (pool.asset_symbol !== undefined) {
    expectQueryString(path, 'asset_symbol', pool.asset_symbol);
  }
  expectQueryNumber(path, 'swap_count', pool.swap_count);
  return pool as unknown as Pool;
}

function expectRegisteredAsset(path: string, value: unknown): RegisteredAsset {
  const asset = expectQueryRecord(path, value);
  for (const field of [
    'ibc_denom',
    'symbol',
    'name',
    'origin_chain',
    'ibc_channel',
    'registered_by',
  ]) {
    expectQueryString(path, field, asset[field]);
  }
  expectQueryNumber(path, 'decimals', asset.decimals);
  expectQueryBoolean(path, 'trading_enabled', asset.trading_enabled);
  expectQueryNumber(path, 'registered_height', asset.registered_height);
  return asset as unknown as RegisteredAsset;
}

function expectPoolStats(path: string, value: unknown): PoolStats {
  const stats = expectQueryRecord(path, value);
  for (const field of [
    'asset_denom',
    'asset_symbol',
    'total_volume_pnyx',
    'total_fees_earned',
    'total_burned',
    'pnyx_reserve',
    'asset_reserve',
    'spot_price_per_million',
    'total_shares',
  ]) {
    expectQueryString(path, field, stats[field]);
  }
  expectQueryNumber(path, 'swap_count', stats.swap_count);
  return stats as unknown as PoolStats;
}

function expectSwapEstimate(path: string, value: unknown): SwapEstimate {
  const estimate = expectQueryRecord(path, value);
  expectQueryString(path, 'expected_output', estimate.expected_output);
  expectQueryStringArray(path, 'route', estimate.route);
  expectQueryStringArray(path, 'route_symbols', estimate.route_symbols);
  expectQueryNumber(path, 'hops', estimate.hops);
  return estimate as unknown as SwapEstimate;
}

function expectLPPosition(path: string, value: unknown): LPPosition {
  const position = expectQueryRecord(path, value);
  for (const field of ['asset_denom', 'shares', 'pnyx_value', 'asset_value']) {
    expectQueryString(path, field, position[field]);
  }
  expectQueryNumber(path, 'share_of_pool_bps', position.share_of_pool_bps);
  return position as unknown as LPPosition;
}

function expectSpotPrice(path: string, value: unknown): SpotPrice {
  const price = expectQueryRecord(path, value);
  for (const field of [
    'input_denom',
    'output_denom',
    'price_per_million',
    'input_symbol',
    'output_symbol',
  ]) {
    expectQueryString(path, field, price[field]);
  }
  expectQueryStringArray(path, 'route', price.route);
  return price as unknown as SpotPrice;
}

export class DEXService {
  private readonly queries: ModuleQueryClient;

  constructor(config: ChainConfig, queries = new ModuleQueryClient(config)) {
    this.queries = queries;
  }

  async listPools(): Promise<Pool[]> {
    return expectQueryArray<unknown>(
      QUERY_PATHS.dex.pools,
      await this.queries.query<unknown>(QUERY_PATHS.dex.pools)
    ).map((pool) => expectPool(QUERY_PATHS.dex.pools, pool));
  }

  async getPool(assetDenom: string): Promise<Pool> {
    const value = await this.queries.query<unknown>(QUERY_PATHS.dex.pool, [
      { number: 1, type: 'string', value: assetDenom },
    ]);
    return expectPool(QUERY_PATHS.dex.pool, value);
  }

  async listAssets(): Promise<RegisteredAsset[]> {
    return expectQueryArray<unknown>(
      QUERY_PATHS.dex.registeredAssets,
      await this.queries.query<unknown>(QUERY_PATHS.dex.registeredAssets)
    ).map((asset) => expectRegisteredAsset(QUERY_PATHS.dex.registeredAssets, asset));
  }

  async getPoolStats(assetDenom: string): Promise<PoolStats> {
    const value = await this.queries.query<unknown>(QUERY_PATHS.dex.poolStats, [
      { number: 1, type: 'string', value: assetDenom },
    ]);
    return expectPoolStats(QUERY_PATHS.dex.poolStats, value);
  }

  async estimateSwap(
    inputDenom: string,
    inputAmount: string,
    outputDenom: string
  ): Promise<SwapEstimate> {
    const value = await this.queries.query<unknown>(QUERY_PATHS.dex.estimateSwap, [
      { number: 1, type: 'string', value: inputDenom },
      { number: 2, type: 'int64', value: inputAmount },
      { number: 3, type: 'string', value: outputDenom },
    ]);
    return expectSwapEstimate(QUERY_PATHS.dex.estimateSwap, value);
  }

  async getLPPosition(assetDenom: string, shares: string): Promise<LPPosition> {
    const value = await this.queries.query<unknown>(QUERY_PATHS.dex.lpPosition, [
      { number: 1, type: 'string', value: assetDenom },
      { number: 2, type: 'int64', value: shares },
    ]);
    return expectLPPosition(QUERY_PATHS.dex.lpPosition, value);
  }

  async getSpotPrice(inputDenom: string, outputDenom: string): Promise<SpotPrice> {
    const value = await this.queries.query<unknown>(QUERY_PATHS.dex.spotPrice, [
      { number: 1, type: 'string', value: inputDenom },
      { number: 2, type: 'string', value: outputDenom },
    ]);
    return expectSpotPrice(QUERY_PATHS.dex.spotPrice, value);
  }
}
