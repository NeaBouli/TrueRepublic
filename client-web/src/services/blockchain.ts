import { StargateClient } from '@cosmjs/stargate';
import type { ChainConfig } from '@/types/chain';
import type { Balance } from '@/types/wallet';

export class BlockchainService {
  private client: StargateClient | null = null;
  private config: ChainConfig;

  constructor(config: ChainConfig) {
    this.config = config;
  }

  /**
   * Connect to blockchain
   */
  async connect(): Promise<void> {
    this.client = await StargateClient.connect(this.config.rpc);
  }

  /**
   * Get client (connect if needed)
   */
  private async getClient(): Promise<StargateClient> {
    if (!this.client) {
      await this.connect();
    }
    return this.client!;
  }

  /**
   * Get account balance
   */
  async getBalance(address: string): Promise<Balance[]> {
    const client = await this.getClient();
    const balances = await client.getAllBalances(address);

    return balances.map((b) => ({
      denom: b.denom,
      amount: b.amount,
    }));
  }

  /**
   * Get PNYX balance
   */
  async getPnyxBalance(address: string): Promise<string> {
    const balances = await this.getBalance(address);
    const pnyx = balances.find((b) => b.denom === this.config.coinMinimalDenom);
    return pnyx?.amount || '0';
  }

  /**
   * Get account info
   */
  async getAccount(address: string) {
    const client = await this.getClient();
    return client.getAccount(address);
  }

  /**
   * Get block height
   */
  async getHeight(): Promise<number> {
    const client = await this.getClient();
    return client.getHeight();
  }

  /**
   * Disconnect
   */
  async disconnect(): Promise<void> {
    if (this.client) {
      this.client.disconnect();
      this.client = null;
    }
  }
}
