import { SigningStargateClient, GasPrice } from '@cosmjs/stargate';
import type { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import type { ChainConfig } from '@/types/chain';
import type { AddLiquidityParams, RemoveLiquidityParams } from '@/types/dex';
import type { TransactionResult } from '@/types/transaction';

export class DEXTxService {
  private config: ChainConfig;

  constructor(config: ChainConfig) {
    this.config = config;
  }

  /**
   * Add liquidity to a PNYX-paired pool.
   * Go: MsgAddLiquidity { sender, asset_denom, pnyx_amt, asset_amt }
   */
  async addLiquidity(
    wallet: DirectSecp256k1HdWallet,
    params: AddLiquidityParams
  ): Promise<TransactionResult> {
    const [account] = await wallet.getAccounts();

    const client = await SigningStargateClient.connectWithSigner(
      this.config.rpc,
      wallet,
      { gasPrice: GasPrice.fromString(this.config.gasPrice) }
    );

    try {
      const msg = {
        typeUrl: '/dex.MsgAddLiquidity',
        value: {
          sender: account.address,
          asset_denom: params.asset_denom,
          pnyx_amt: parseInt(params.pnyx_amt, 10),
          asset_amt: parseInt(params.asset_amt, 10),
        },
      };

      const gasEstimate = await client.simulate(account.address, [msg], '');
      const gas = Math.ceil(gasEstimate * 1.3);

      const result = await client.signAndBroadcast(
        account.address,
        [msg],
        {
          amount: [
            {
              denom: this.config.coinMinimalDenom,
              amount: '5000',
            },
          ],
          gas: gas.toString(),
        },
        ''
      );

      client.disconnect();

      if (result.code !== 0) {
        throw new Error(result.rawLog || 'Add liquidity failed');
      }

      return {
        hash: result.transactionHash,
        height: result.height,
        success: true,
      };
    } catch (err: unknown) {
      client.disconnect();
      return {
        hash: '',
        height: 0,
        success: false,
        error: err instanceof Error ? err.message : 'Add liquidity failed',
      };
    }
  }

  /**
   * Remove liquidity from a PNYX-paired pool.
   * Go: MsgRemoveLiquidity { sender, asset_denom, shares }
   */
  async removeLiquidity(
    wallet: DirectSecp256k1HdWallet,
    params: RemoveLiquidityParams
  ): Promise<TransactionResult> {
    const [account] = await wallet.getAccounts();

    const client = await SigningStargateClient.connectWithSigner(
      this.config.rpc,
      wallet,
      { gasPrice: GasPrice.fromString(this.config.gasPrice) }
    );

    try {
      const msg = {
        typeUrl: '/dex.MsgRemoveLiquidity',
        value: {
          sender: account.address,
          asset_denom: params.asset_denom,
          shares: parseInt(params.shares, 10),
        },
      };

      const gasEstimate = await client.simulate(account.address, [msg], '');
      const gas = Math.ceil(gasEstimate * 1.3);

      const result = await client.signAndBroadcast(
        account.address,
        [msg],
        {
          amount: [
            {
              denom: this.config.coinMinimalDenom,
              amount: '5000',
            },
          ],
          gas: gas.toString(),
        },
        ''
      );

      client.disconnect();

      if (result.code !== 0) {
        throw new Error(result.rawLog || 'Remove liquidity failed');
      }

      return {
        hash: result.transactionHash,
        height: result.height,
        success: true,
      };
    } catch (err: unknown) {
      client.disconnect();
      return {
        hash: '',
        height: 0,
        success: false,
        error:
          err instanceof Error ? err.message : 'Remove liquidity failed',
      };
    }
  }
}
