import { fromBech32 } from '@cosmjs/encoding';
import type { SigningStargateClient } from '@cosmjs/stargate';
import type { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import type { ChainConfig } from '@/types/chain';
import type {
  AddLiquidityParams,
  RemoveLiquidityParams,
  SwapExactParams,
} from '@/types/dex';
import type { TransactionResult } from '@/types/transaction';
import { connectSigningClient, deliverMessages } from './signingClient';
import { assertPositiveInt64Decimal } from './txRegistry';

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

    let client: SigningStargateClient | undefined;

    try {
      client = await connectSigningClient(this.config, wallet);
      const msg = {
        typeUrl: '/dex.MsgAddLiquidity',
        value: {
          sender: fromBech32(account.address).data,
          assetDenom: params.asset_denom,
          pnyxAmt: assertPositiveInt64Decimal(params.pnyx_amt, 'pnyx_amt'),
          assetAmt: assertPositiveInt64Decimal(params.asset_amt, 'asset_amt'),
        },
      };

      return await deliverMessages(
        client,
        account.address,
        [msg],
        this.config.gasPrice
      );
    } catch (err: unknown) {
      return {
        hash: '',
        height: 0,
        success: false,
        error: err instanceof Error ? err.message : 'Add liquidity failed',
      };
    } finally {
      client?.disconnect();
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

    let client: SigningStargateClient | undefined;

    try {
      client = await connectSigningClient(this.config, wallet);
      const msg = {
        typeUrl: '/dex.MsgRemoveLiquidity',
        value: {
          sender: fromBech32(account.address).data,
          assetDenom: params.asset_denom,
          shares: assertPositiveInt64Decimal(params.shares, 'shares'),
        },
      };

      return await deliverMessages(
        client,
        account.address,
        [msg],
        this.config.gasPrice
      );
    } catch (err: unknown) {
      return {
        hash: '',
        height: 0,
        success: false,
        error:
          err instanceof Error ? err.message : 'Remove liquidity failed',
      };
    } finally {
      client?.disconnect();
    }
  }

  /**
   * Execute a slippage-protected swap.
   * Go: MsgSwapExact { sender, input_denom, input_amt, output_denom, min_output }
   * min_output is the on-chain slippage bound and must be positive.
   */
  async swapExact(
    wallet: DirectSecp256k1HdWallet,
    params: SwapExactParams
  ): Promise<TransactionResult> {
    const [account] = await wallet.getAccounts();

    let client: SigningStargateClient | undefined;

    try {
      client = await connectSigningClient(this.config, wallet);
      const msg = {
        typeUrl: '/dex.MsgSwapExact',
        value: {
          sender: fromBech32(account.address).data,
          inputDenom: params.input_denom,
          inputAmt: assertPositiveInt64Decimal(
            params.input_amt,
            'input_amt'
          ),
          outputDenom: params.output_denom,
          minOutput: assertPositiveInt64Decimal(
            params.min_output,
            'min_output'
          ),
        },
      };

      return await deliverMessages(
        client,
        account.address,
        [msg],
        this.config.gasPrice
      );
    } catch (err: unknown) {
      return {
        hash: '',
        height: 0,
        success: false,
        error: err instanceof Error ? err.message : 'Swap failed',
      };
    } finally {
      client?.disconnect();
    }
  }
}
