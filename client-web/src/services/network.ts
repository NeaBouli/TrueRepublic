import type { ChainConfig } from '@/types/chain';
import type {
  NetworkInfo,
  Validator,
  Block,
  IBCChannel,
} from '@/types/network';

export class NetworkService {
  private config: ChainConfig;

  constructor(config: ChainConfig) {
    this.config = config;
  }

  /**
   * Get network info from CometBFT RPC /status.
   * No cosmos/base/tendermint/v1beta1 endpoint registered — use RPC directly.
   */
  async getNetworkInfo(): Promise<NetworkInfo | null> {
    try {
      const response = await fetch(`${this.config.rpc}/status`);
      if (!response.ok) return null;

      const data = await response.json();
      const result = data.result;

      return {
        chainId: result.node_info?.network || '',
        latestBlockHeight: parseInt(
          result.sync_info?.latest_block_height || '0',
          10
        ),
        latestBlockTime: result.sync_info?.latest_block_time || '',
        totalValidators: 0, // Updated separately via loadValidators
        nodeInfo: {
          moniker: result.node_info?.moniker || '',
          version: result.node_info?.version || '',
          network: result.node_info?.network || '',
        },
      };
    } catch {
      return null;
    }
  }

  /**
   * Get validators from truedemocracy module.
   * No cosmos/staking module — validators are Proof-of-Domain.
   * REST path: /truerepublic/truedemocracy/validators
   */
  async getValidators(): Promise<Validator[]> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/truedemocracy/validators`
      );

      if (!response.ok) return [];

      const data = await response.json();
      // gRPC query returns JSON bytes in result field
      const validators: Validator[] = data.result
        ? JSON.parse(atob(data.result))
        : data.validators || data || [];

      return validators.sort((a, b) => b.power - a.power);
    } catch {
      return [];
    }
  }

  /**
   * Get recent blocks from CometBFT RPC.
   * Uses /block?height=N (not cosmos/base/tendermint/v1beta1).
   */
  async getRecentBlocks(limit: number = 10): Promise<Block[]> {
    try {
      // Get latest block height
      const statusResponse = await fetch(`${this.config.rpc}/status`);
      if (!statusResponse.ok) return [];

      const statusData = await statusResponse.json();
      const latestHeight = parseInt(
        statusData.result.sync_info?.latest_block_height || '0',
        10
      );

      if (latestHeight === 0) return [];

      // Fetch blocks in parallel (limit to avoid too many requests)
      const count = Math.min(limit, latestHeight);
      const promises = Array.from({ length: count }, (_, i) =>
        this.getBlock(latestHeight - i)
      );

      const results = await Promise.allSettled(promises);
      return results
        .filter(
          (r): r is PromiseFulfilledResult<Block | null> =>
            r.status === 'fulfilled' && r.value !== null
        )
        .map((r) => r.value!);
    } catch {
      return [];
    }
  }

  /**
   * Get a single block by height from CometBFT RPC.
   */
  private async getBlock(height: number): Promise<Block | null> {
    try {
      const response = await fetch(
        `${this.config.rpc}/block?height=${height}`
      );
      if (!response.ok) return null;

      const data = await response.json();
      const block = data.result?.block;
      const blockId = data.result?.block_id;

      if (!block) return null;

      return {
        height: parseInt(block.header.height, 10),
        hash: blockId?.hash || '',
        time: block.header.time,
        proposer: block.header.proposer_address || '',
        txCount: block.data?.txs?.length || 0,
      };
    } catch {
      return null;
    }
  }

  /**
   * Get IBC channels (ibc-go is registered).
   */
  async getIBCChannels(): Promise<IBCChannel[]> {
    try {
      const response = await fetch(
        `${this.config.rest}/ibc/core/channel/v1/channels`
      );

      if (!response.ok) return [];

      const data = await response.json();
      return data.channels || [];
    } catch {
      return [];
    }
  }
}
