import { SigningStargateClient, GasPrice } from '@cosmjs/stargate';
import type { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import type { ChainConfig } from '@/types/chain';
import type { PayToPutCalculation } from '@/types/governance';
import type { TransactionResult } from '@/types/transaction';

export class GovernanceTxService {
  private config: ChainConfig;

  constructor(config: ChainConfig) {
    this.config = config;
  }

  /**
   * Calculate PayToPut cost (eq.3 from whitepaper).
   * Queries the chain for the current put price for a domain.
   */
  async calculatePayToPut(
    domainName: string
  ): Promise<PayToPutCalculation> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/truedemocracy/paytoput/${domainName}`
      );

      if (!response.ok) {
        return this.defaultPayToPut();
      }

      const data = await response.json();
      return data.calculation || this.defaultPayToPut();
    } catch {
      return this.defaultPayToPut();
    }
  }

  private defaultPayToPut(): PayToPutCalculation {
    return {
      baseCost: '1000000',
      domainMultiplier: 1,
      finalCost: '1000000',
      formula: 'Base: 1 PNYX',
    };
  }

  /**
   * Submit a suggestion via MsgSubmitProposal.
   * Go: sender, domain_name, issue_name, suggestion_name, creator, fee, external_link.
   * If the issue doesn't exist, it will be created automatically.
   */
  async createSuggestion(
    wallet: DirectSecp256k1HdWallet,
    domainName: string,
    issueName: string,
    suggestionName: string,
    fee: { denom: string; amount: string }[],
    externalLink?: string
  ): Promise<TransactionResult> {
    const [account] = await wallet.getAccounts();

    const client = await SigningStargateClient.connectWithSigner(
      this.config.rpc,
      wallet,
      { gasPrice: GasPrice.fromString(this.config.gasPrice) }
    );

    try {
      const msg = {
        typeUrl: '/truerepublic.truedemocracy.MsgSubmitProposal',
        value: {
          sender: account.address,
          domain_name: domainName,
          issue_name: issueName,
          suggestion_name: suggestionName,
          creator: account.address,
          fee,
          external_link: externalLink || '',
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
        throw new Error(result.rawLog || 'Suggestion creation failed');
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
        error: err instanceof Error ? err.message : 'Suggestion creation failed',
      };
    }
  }

  /**
   * Place a stone on a suggestion.
   * Go: MsgPlaceStoneOnSuggestion { sender, domain_name, issue_name, suggestion_name, member_addr }.
   * Stones are endorsement counts — no color in the on-chain message.
   */
  async placeStoneOnSuggestion(
    wallet: DirectSecp256k1HdWallet,
    domainName: string,
    issueName: string,
    suggestionName: string
  ): Promise<TransactionResult> {
    const [account] = await wallet.getAccounts();

    const client = await SigningStargateClient.connectWithSigner(
      this.config.rpc,
      wallet,
      { gasPrice: GasPrice.fromString(this.config.gasPrice) }
    );

    try {
      const msg = {
        typeUrl: '/truerepublic.truedemocracy.MsgPlaceStoneOnSuggestion',
        value: {
          sender: account.address,
          domain_name: domainName,
          issue_name: issueName,
          suggestion_name: suggestionName,
          member_addr: account.address,
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
        throw new Error(result.rawLog || 'Stone placement failed');
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
        error: err instanceof Error ? err.message : 'Stone placement failed',
      };
    }
  }

  /**
   * Place a stone on an issue.
   * Go: MsgPlaceStoneOnIssue { sender, domain_name, issue_name, member_addr }.
   */
  async placeStoneOnIssue(
    wallet: DirectSecp256k1HdWallet,
    domainName: string,
    issueName: string
  ): Promise<TransactionResult> {
    const [account] = await wallet.getAccounts();

    const client = await SigningStargateClient.connectWithSigner(
      this.config.rpc,
      wallet,
      { gasPrice: GasPrice.fromString(this.config.gasPrice) }
    );

    try {
      const msg = {
        typeUrl: '/truerepublic.truedemocracy.MsgPlaceStoneOnIssue',
        value: {
          sender: account.address,
          domain_name: domainName,
          issue_name: issueName,
          member_addr: account.address,
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
        throw new Error(result.rawLog || 'Stone placement failed');
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
        error: err instanceof Error ? err.message : 'Stone placement failed',
      };
    }
  }
}
