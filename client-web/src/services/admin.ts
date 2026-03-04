import { SigningStargateClient, GasPrice } from '@cosmjs/stargate';
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

export class AdminService {
  private config: ChainConfig;

  constructor(config: ChainConfig) {
    this.config = config;
  }

  /**
   * Check if address is admin of domain.
   * Go Domain has a single Admin field (not an admins list).
   */
  async isAdmin(domainName: string, address: string): Promise<boolean> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/truedemocracy/domain/${domainName}`
      );

      if (!response.ok) return false;

      const data = await response.json();
      const domain = data.domain || data;

      return domain.admin === address;
    } catch {
      return false;
    }
  }

  /**
   * Get domain members with their identity/permission status.
   * Derived from the Domain query (members, identity_commits, permission_reg).
   */
  async getDomainMembers(domainName: string): Promise<DomainMember[]> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/truedemocracy/domain/${domainName}`
      );

      if (!response.ok) return [];

      const data = await response.json();
      const domain = data.domain || data;

      const members: string[] = domain.members || [];
      const identityCommits: string[] = domain.identity_commits || [];
      const permissionReg: string[] = domain.permission_reg || [];

      return members.map((address) => ({
        address,
        hasIdentityCommitment: identityCommits.length > 0,
        inPermissionReg: permissionReg.length > 0,
      }));
    } catch {
      return [];
    }
  }

  /**
   * Compute domain statistics from available query data.
   * No dedicated domain_stats endpoint exists — derived from domain query.
   */
  async getDomainStats(domainName: string): Promise<DomainStats | null> {
    try {
      const response = await fetch(
        `${this.config.rest}/truerepublic/truedemocracy/domain/${domainName}`
      );

      if (!response.ok) return null;

      const data = await response.json();
      const domain = data.domain || data;

      const members: string[] = domain.members || [];
      const issues: unknown[] = domain.issues || [];
      const identityCommits: string[] = domain.identity_commits || [];
      const permissionReg: string[] = domain.permission_reg || [];
      const treasury = domain.treasury || [];

      // Count suggestions across all issues
      let totalSuggestions = 0;
      for (const issue of issues) {
        const iss = issue as { suggestions?: unknown[] };
        totalSuggestions += iss.suggestions?.length || 0;
      }

      // Get PNYX balance from treasury coins
      const pnyxCoin = treasury.find(
        (c: { denom: string; amount: string }) => c.denom === 'upnyx'
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
    } catch {
      return null;
    }
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

    const client = await SigningStargateClient.connectWithSigner(
      this.config.rpc,
      wallet,
      { gasPrice: GasPrice.fromString(this.config.gasPrice) }
    );

    try {
      const msg = {
        typeUrl: '/truerepublic.truedemocracy.MsgApproveOnboarding',
        value: {
          sender: account.address,
          domain_name: params.domain_name,
          requester_addr: params.requester_addr,
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
        throw new Error(result.rawLog || 'Approve onboarding failed');
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
          err instanceof Error ? err.message : 'Approve onboarding failed',
      };
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

    const client = await SigningStargateClient.connectWithSigner(
      this.config.rpc,
      wallet,
      { gasPrice: GasPrice.fromString(this.config.gasPrice) }
    );

    try {
      const msg = {
        typeUrl: '/truerepublic.truedemocracy.MsgAddMember',
        value: {
          sender: account.address,
          domain_name: params.domain_name,
          new_member: params.new_member,
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
        throw new Error(result.rawLog || 'Add member failed');
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
        error: err instanceof Error ? err.message : 'Add member failed',
      };
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

    const client = await SigningStargateClient.connectWithSigner(
      this.config.rpc,
      wallet,
      { gasPrice: GasPrice.fromString(this.config.gasPrice) }
    );

    try {
      const msg = {
        typeUrl: '/truerepublic.truedemocracy.MsgCreateDomain',
        value: {
          name: params.name,
          admin: account.address,
          initial_coins: [
            {
              denom: this.config.coinMinimalDenom,
              amount: params.initial_coins,
            },
          ],
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
        throw new Error(result.rawLog || 'Domain creation failed');
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
        error: err instanceof Error ? err.message : 'Domain creation failed',
      };
    }
  }
}
