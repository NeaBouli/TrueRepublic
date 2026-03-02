import type { ChainConfig } from '@/types/chain';
import type {
  Pool,
  RegisteredAsset,
  PoolStats,
  SwapEstimate,
  LPPosition,
  SpotPrice,
} from '@/types/dex';

export class DEXService {
  private config: ChainConfig;

  constructor(config: ChainConfig) {
    this.config = config;
  }

  /**
   * List all liquidity pools
   */
  async listPools(): Promise<Pool[]> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/dex/pools`
      );

      if (!response.ok) {
        throw new Error('Failed to fetch pools');
      }

      const data = await response.json();
      return data.pools || data || [];
    } catch {
      return [];
    }
  }

  /**
   * Get pool by asset denom
   */
  async getPool(assetDenom: string): Promise<Pool | null> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/dex/pool/${assetDenom}`
      );

      if (!response.ok) return null;

      const data = await response.json();
      return data.pool || data || null;
    } catch {
      return null;
    }
  }

  /**
   * List all registered assets
   */
  async listAssets(): Promise<RegisteredAsset[]> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/dex/registered_assets`
      );

      if (!response.ok) {
        throw new Error('Failed to fetch assets');
      }

      const data = await response.json();
      return data.assets || data || [];
    } catch {
      return [];
    }
  }

  /**
   * Get pool statistics (volume, fees, burn, spot price)
   */
  async getPoolStats(assetDenom: string): Promise<PoolStats | null> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/dex/pool_stats/${assetDenom}`
      );

      if (!response.ok) return null;

      const data = await response.json();
      return data.stats || data || null;
    } catch {
      return null;
    }
  }

  /**
   * Estimate swap output (read-only, no tx execution)
   */
  async estimateSwap(
    inputDenom: string,
    inputAmount: string,
    outputDenom: string
  ): Promise<SwapEstimate | null> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/dex/estimate_swap/${inputDenom}/${inputAmount}/${outputDenom}`
      );

      if (!response.ok) return null;

      const data = await response.json();
      return data.estimate || data || null;
    } catch {
      return null;
    }
  }

  /**
   * Get LP position value for given shares
   */
  async getLPPosition(
    assetDenom: string,
    shares: string
  ): Promise<LPPosition | null> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/dex/lp_position/${assetDenom}/${shares}`
      );

      if (!response.ok) return null;

      const data = await response.json();
      return data.position || data || null;
    } catch {
      return null;
    }
  }

  /**
   * Get spot price between two denoms
   */
  async getSpotPrice(
    inputDenom: string,
    outputDenom: string
  ): Promise<SpotPrice | null> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/dex/spot_price/${inputDenom}/${outputDenom}`
      );

      if (!response.ok) return null;

      const data = await response.json();
      return data.price || data || null;
    } catch {
      return null;
    }
  }
}
