import { fromBech32 } from '@cosmjs/encoding';
import type { SigningStargateClient } from '@cosmjs/stargate';
import type { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import type { ChainConfig } from '@/types/chain';
import type {
  DomainMember,
  DomainStats,
  CreateDomainParams,
  ApproveOnboardingParams,
  AddMemberParams,
} from '@/types/admin';
import type { TransactionResult } from '@/types/transaction';
import { connectSigningClient, deliverMessages } from './signingClient';
import type { ChainDomain } from '@/types/chainData';
import {
  expectChainDomain,
  ModuleQueryClient,
  QUERY_PATHS,
} from './moduleQuery';

export class AdminService {
  private readonly queries: ModuleQueryClient;

  constructor(
    private readonly config: ChainConfig,
    queries = new ModuleQueryClient(config)
  ) {
    this.queries = queries;
  }

  /**
   * Check if address is admin of domain.
   * Go Domain has a single Admin field (not an admins list).
   */
  async isAdmin(domainName: string, address: string): Promise<boolean> {
    const domain = await this.getDomain(domainName);
    return domain.admin === address;
  }

  /**
   * Get domain member addresses without attributing anonymous identity or
   * permission-register entries to individual members.
   */
  async getDomainMembers(domainName: string): Promise<DomainMember[]> {
    const domain = await this.getDomain(domainName);
    const members = domain.members ?? [];
    return members.map((address) => ({
      address,
      hasIdentityCommitment: null,
      inPermissionReg: null,
    }));
  }

  /**
   * Compute domain statistics from available query data.
   * No dedicated domain_stats endpoint exists — derived from domain query.
   */
  async getDomainStats(domainName: string): Promise<DomainStats> {
    const domain = await this.getDomain(domainName);
    const members = domain.members ?? [];
    const issues = domain.issues ?? [];
    const identityCommits = domain.identity_commits ?? [];
    const permissionReg = domain.permission_reg ?? [];
    const treasury = domain.treasury ?? [];

      // Count suggestions across all issues
    let totalSuggestions = 0;
    for (const issue of issues) {
      totalSuggestions += issue.suggestions?.length ?? 0;
    }

      // Get PNYX balance from treasury coins
    const pnyxCoin = treasury.find(
      (coin) => coin.denom === this.config.coinMinimalDenom
    );

    return {
      domainName: domain.name,
      totalMembers: members.length,
      totalIssues: issues.length,
      totalSuggestions,
      treasuryBalance: pnyxCoin?.amount || '0',
      identityCommitments: identityCommits.length,
      permissionRegCount: permissionReg.length,
      merkleRoot: domain.merkle_root || '',
    };
  }

  private getDomain(domainName: string): Promise<ChainDomain> {
    return this.queries
      .query<unknown>(QUERY_PATHS.truedemocracy.domain, [
        { number: 1, type: 'string', value: domainName },
      ])
      .then((value) =>
        expectChainDomain(QUERY_PATHS.truedemocracy.domain, value)
      );
  }

  /**
   * Approve an onboarding request.
   * Go: MsgApproveOnboarding { sender, domain_name, requester_addr }
   */
  async approveOnboarding(
    wallet: DirectSecp256k1HdWallet,
    params: ApproveOnboardingParams
  ): Promise<TransactionResult> {
    const [account] = await wallet.getAccounts();

    let client: SigningStargateClient | undefined;

    try {
      client = await connectSigningClient(this.config, wallet);
      const msg = {
        typeUrl: '/truedemocracy.MsgApproveOnboarding',
        value: {
          sender: fromBech32(account.address).data,
          domainName: params.domain_name,
          requesterAddr: params.requester_addr,
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
          err instanceof Error ? err.message : 'Approve onboarding failed',
      };
    } finally {
      client?.disconnect();
    }
  }

  /**
   * Add a member to a domain.
   * Go: MsgAddMember { sender, domain_name, new_member }
   */
  async addMember(
    wallet: DirectSecp256k1HdWallet,
    params: AddMemberParams
  ): Promise<TransactionResult> {
    const [account] = await wallet.getAccounts();

    let client: SigningStargateClient | undefined;

    try {
      client = await connectSigningClient(this.config, wallet);
      const msg = {
        typeUrl: '/truedemocracy.MsgAddMember',
        value: {
          sender: fromBech32(account.address).data,
          domainName: params.domain_name,
          newMember: params.new_member,
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
        error: err instanceof Error ? err.message : 'Add member failed',
      };
    } finally {
      client?.disconnect();
    }
  }

  /**
   * Create a new domain.
   * Go: MsgCreateDomain { name, admin, initial_coins }
   */
  async createDomain(
    wallet: DirectSecp256k1HdWallet,
    params: CreateDomainParams
  ): Promise<TransactionResult> {
    const [account] = await wallet.getAccounts();

    let client: SigningStargateClient | undefined;

    try {
      client = await connectSigningClient(this.config, wallet);
      const msg = {
        typeUrl: '/truedemocracy.MsgCreateDomain',
        value: {
          name: params.name,
          admin: fromBech32(account.address).data,
          initialCoins: [
            {
              denom: this.config.coinMinimalDenom,
              amount: params.initial_coins,
            },
          ],
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
        error: err instanceof Error ? err.message : 'Domain creation failed',
      };
    } finally {
      client?.disconnect();
    }
  }
}
