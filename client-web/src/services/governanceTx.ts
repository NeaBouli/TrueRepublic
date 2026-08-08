import { fromBech32 } from '@cosmjs/encoding';
import type { SigningStargateClient } from '@cosmjs/stargate';
import type { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import type { ChainConfig } from '@/types/chain';
import type { PayToPutCalculation } from '@/types/governance';
import type { TransactionResult } from '@/types/transaction';
import { connectSigningClient, deliverMessages } from './signingClient';
import type { ChainPayToPut } from '@/types/chainData';
import {
  expectQueryNumber,
  expectQueryRecord,
  expectQueryString,
  ModuleQueryError,
  ModuleQueryClient,
  QUERY_PATHS,
} from './moduleQuery';

export class GovernanceTxService {
  private readonly queries: ModuleQueryClient;

  constructor(
    private readonly config: ChainConfig,
    queries = new ModuleQueryClient(config)
  ) {
    this.queries = queries;
  }

  /**
   * Calculate PayToPut cost (eq.3 from whitepaper).
   * Queries the chain for the current put price for a domain.
   */
  async calculatePayToPut(
    domainName: string
  ): Promise<PayToPutCalculation> {
    const value = await this.queries.query<unknown>(
      QUERY_PATHS.truedemocracy.payToPut,
      [{ number: 1, type: 'string', value: domainName }]
    );
    const result = expectQueryRecord(
      QUERY_PATHS.truedemocracy.payToPut,
      value
    );
    const calculation: ChainPayToPut = {
      base_cost: expectQueryString(
        QUERY_PATHS.truedemocracy.payToPut,
        'base_cost',
        result.base_cost
      ),
      domain_multiplier: expectQueryNumber(
        QUERY_PATHS.truedemocracy.payToPut,
        'domain_multiplier',
        result.domain_multiplier
      ),
      final_cost: expectQueryString(
        QUERY_PATHS.truedemocracy.payToPut,
        'final_cost',
        result.final_cost
      ),
      formula: expectQueryString(
        QUERY_PATHS.truedemocracy.payToPut,
        'formula',
        result.formula
      ),
    };
    if (
      !Number.isSafeInteger(calculation.domain_multiplier) ||
      calculation.domain_multiplier < 0 ||
      !/^\d+$/.test(calculation.base_cost) ||
      !/^\d+$/.test(calculation.final_cost)
    ) {
      throw new ModuleQueryError(
        QUERY_PATHS.truedemocracy.payToPut,
        'decode',
        'numeric fields must be unsigned safe integers'
      );
    }
    return {
      baseCost: calculation.base_cost,
      domainMultiplier: calculation.domain_multiplier,
      finalCost: calculation.final_cost,
      formula: calculation.formula,
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

    let client: SigningStargateClient | undefined;

    try {
      client = await connectSigningClient(this.config, wallet);
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
      client?.disconnect();
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

    let client: SigningStargateClient | undefined;

    try {
      client = await connectSigningClient(this.config, wallet);
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
      client?.disconnect();
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

    let client: SigningStargateClient | undefined;

    try {
      client = await connectSigningClient(this.config, wallet);
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
      client?.disconnect();
    }
  }
}
