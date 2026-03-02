import { SigningStargateClient, GasPrice } from '@cosmjs/stargate';
import type { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import type { ChainConfig } from '@/types/chain';
import type { SendParams, TransactionResult, Transaction } from '@/types/transaction';

export class TransactionService {
  private config: ChainConfig;

  constructor(config: ChainConfig) {
    this.config = config;
  }

  /**
   * Send tokens
   */
  async send(
    wallet: DirectSecp256k1HdWallet,
    params: SendParams
  ): Promise<TransactionResult> {
    const [account] = await wallet.getAccounts();

    const client = await SigningStargateClient.connectWithSigner(
      this.config.rpc,
      wallet,
      {
        gasPrice: GasPrice.fromString(this.config.gasPrice),
      }
    );

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

      const gasEstimate = await client.simulate(
        account.address,
        [msg],
        params.memo
      );
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
        params.memo || ''
      );

      client.disconnect();

      if (result.code !== 0) {
        throw new Error(result.rawLog || 'Transaction failed');
      }

      return {
        hash: result.transactionHash,
        height: result.height,
        success: true,
      };
    } catch (error: unknown) {
      client.disconnect();
      const message =
        error instanceof Error ? error.message : 'Transaction failed';
      return {
        hash: '',
        height: 0,
        success: false,
        error: message,
      };
    }
  }

  /**
   * Get transaction by hash
   */
  async getTransaction(hash: string): Promise<Transaction | null> {
    const client = await SigningStargateClient.connect(this.config.rpc);

    try {
      const tx = await client.getTx(hash);
      client.disconnect();

      if (!tx) return null;

      return {
        hash: tx.hash,
        height: tx.height,
        timestamp: new Date().toISOString(),
        type: 'cosmos.bank.v1beta1.MsgSend',
        from: '',
        fee: { denom: 'pnyx', amount: '0' },
        status: tx.code === 0 ? 'success' : 'failed',
      };
    } catch {
      client.disconnect();
      return null;
    }
  }
}
