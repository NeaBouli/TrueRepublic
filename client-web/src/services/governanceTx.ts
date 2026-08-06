import { fromBech32 } from '@cosmjs/encoding';
import type { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import type { ChainConfig } from '@/types/chain';
import type { PayToPutCalculation } from '@/types/governance';
import type { TransactionResult } from '@/types/transaction';
import { connectSigningClient, deliverMessages } from './signingClient';

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

    const client = await connectSigningClient(this.config, wallet);

    try {
      const msg = {
        typeUrl: '/truedemocracy.MsgSubmitProposal',
        value: {
          sender: fromBech32(account.address).data,
          domainName,
          issueName,
          suggestionName,
          creator: account.address,
          fee,
          externalLink: externalLink || '',
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
        error: err instanceof Error ? err.message : 'Suggestion creation failed',
      };
    } finally {
      client.disconnect();
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

    const client = await connectSigningClient(this.config, wallet);

    try {
      const msg = {
        typeUrl: '/truedemocracy.MsgPlaceStoneOnSuggestion',
        value: {
          sender: fromBech32(account.address).data,
          domainName,
          issueName,
          suggestionName,
          memberAddr: account.address,
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
        error: err instanceof Error ? err.message : 'Stone placement failed',
      };
    } finally {
      client.disconnect();
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

    const client = await connectSigningClient(this.config, wallet);

    try {
      const msg = {
        typeUrl: '/truedemocracy.MsgPlaceStoneOnIssue',
        value: {
          sender: fromBech32(account.address).data,
          domainName,
          issueName,
          memberAddr: account.address,
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
        error: err instanceof Error ? err.message : 'Stone placement failed',
      };
    } finally {
      client.disconnect();
    }
  }
}
