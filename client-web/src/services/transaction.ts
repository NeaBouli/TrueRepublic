import { SigningStargateClient } from '@cosmjs/stargate';
import type { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import type { ChainConfig } from '@/types/chain';
import type { SendParams, TransactionResult, Transaction } from '@/types/transaction';
import { connectSigningClient, deliverMessages } from './signingClient';

export class TransactionService {
  private config: ChainConfig;

  constructor(config: ChainConfig) {
    this.config = config;
  }

  /**
   * Send tokens via the standard bank message retained through the canonical
   * registry's default types.
   */
  async send(
    wallet: DirectSecp256k1HdWallet,
    params: SendParams
  ): Promise<TransactionResult> {
    const [account] = await wallet.getAccounts();

    const client = await connectSigningClient(this.config, wallet);

    try {
      const msg = {
        typeUrl: '/cosmos.bank.v1beta1.MsgSend',
        value: {
          fromAddress: account.address,
          toAddress: params.to,
          amount: [
            {
              denom: params.denom,
              amount: params.amount,
            },
          ],
        },
      };

      return await deliverMessages(
        client,
        account.address,
        [msg],
        this.config.gasPrice,
        params.memo || ''
      );
    } catch (error: unknown) {
      const message =
        error instanceof Error ? error.message : 'Transaction failed';
      return {
        hash: '',
        height: 0,
        success: false,
        error: message,
      };
    } finally {
      client.disconnect();
    }
  }

  /**
   * Get transaction by hash (read-only connection).
   */
  async getTransaction(hash: string): Promise<Transaction | null> {
    const client = await SigningStargateClient.connect(this.config.rpc);

    try {
      const tx = await client.getTx(hash);

      if (!tx) return null;

      return {
        hash: tx.hash,
        height: tx.height,
        timestamp: new Date().toISOString(),
        type: 'cosmos.bank.v1beta1.MsgSend',
        from: '',
        fee: { denom: this.config.coinMinimalDenom, amount: '0' },
        status: tx.code === 0 ? 'success' : 'failed',
      };
    } catch {
      return null;
    } finally {
      client.disconnect();
    }
  }
}
