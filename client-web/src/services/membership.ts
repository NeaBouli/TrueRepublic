import { fromBech32 } from '@cosmjs/encoding';
import type { SigningStargateClient } from '@cosmjs/stargate';
import type { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import type { ChainConfig } from '@/types/chain';
import type { DomainInvite, MembershipStatus } from '@/types/membership';
import type { TransactionResult } from '@/types/transaction';
import { connectSigningClient, deliverMessages } from './signingClient';
import {
  expectChainDomain,
  expectChainMerkleProof,
  ModuleQueryClient,
  QUERY_PATHS,
} from './moduleQuery';

export class MembershipService {
  private readonly queries: ModuleQueryClient;

  constructor(
    private readonly config: ChainConfig,
    queries = new ModuleQueryClient(config)
  ) {
    this.queries = queries;
  }

  /**
   * Parse invite link.
   * Format: truerepublic://join/domain/{domain_id}?invite={signature}
   */
  parseInviteLink(link: string): DomainInvite | null {
    try {
      const url = new URL(link.replace('truerepublic://', 'https://'));

      if (!url.pathname.startsWith('/join/domain/')) {
        return null;
      }

      const domainId = url.pathname.split('/').pop();
      const signature = url.searchParams.get('invite');

      if (!domainId) return null;

      return {
        domainId,
        domainName: domainId,
        inviter: '',
        signature: signature || undefined,
      };
    } catch {
      return null;
    }
  }

  /**
   * Get membership status by querying the domain and checking members list.
   * The Go Domain struct has: members[]string, identity_commits[]string.
   */
  async getMembershipStatus(
    domainId: string,
    address: string,
    locallyOwnedCommitment?: string
  ): Promise<MembershipStatus> {
    const value = await this.queries.query<unknown>(
      QUERY_PATHS.truedemocracy.domain,
      [{ number: 1, type: 'string', value: domainId }]
    );
    const domain = expectChainDomain(
      QUERY_PATHS.truedemocracy.domain,
      value
    );
    const members = domain.members ?? [];
    const identityCommits = domain.identity_commits ?? [];
    const isMember = members.includes(address);
    const canonicalCommitment = locallyOwnedCommitment?.toLowerCase();
    const hasIdentityCommitment = canonicalCommitment
      ? identityCommits.includes(canonicalCommitment)
      : null;
    return {
      domainId,
      address,
      isMember,
      hasIdentityCommitment,
      inMerkleTree:
        hasIdentityCommitment === null
          ? null
          : hasIdentityCommitment && domain.merkle_root !== '',
      step1Complete: isMember,
      step2Complete: isMember,
    };
  }

  /**
   * Submit onboarding request (Step 1).
   * Go: MsgOnboardToDomain { sender, domain_name, domain_pub_key_hex,
   *   global_pub_key_hex, signature_hex }
   */
  async submitOnboarding(
    wallet: DirectSecp256k1HdWallet,
    domainName: string,
    domainPubKeyHex: string,
    globalPubKeyHex: string,
    signatureHex: string
  ): Promise<TransactionResult> {
    const [account] = await wallet.getAccounts();

    let client: SigningStargateClient | undefined;

    try {
      client = await connectSigningClient(this.config, wallet);
      const msg = {
        typeUrl: '/truedemocracy.MsgOnboardToDomain',
        value: {
          sender: fromBech32(account.address).data,
          domainName,
          domainPubKeyHex,
          globalPubKeyHex,
          signatureHex,
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
        error: err instanceof Error ? err.message : 'Onboarding failed',
      };
    } finally {
      client?.disconnect();
    }
  }

  /**
   * Register ZKP identity commitment (after becoming a member).
   * Go: MsgRegisterIdentity { sender, domain_name, commitment }
   * Commitment must be 64 hex chars (32 bytes MiMC hash).
   */
  async registerIdentity(
    wallet: DirectSecp256k1HdWallet,
    domainName: string,
    commitment: string
  ): Promise<TransactionResult> {
    const [account] = await wallet.getAccounts();

    let client: SigningStargateClient | undefined;

    try {
      client = await connectSigningClient(this.config, wallet);
      const msg = {
        typeUrl: '/truedemocracy.MsgRegisterIdentity',
        value: {
          sender: fromBech32(account.address).data,
          domainName,
          commitment,
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
          err instanceof Error
            ? err.message
            : 'Identity registration failed',
      };
    } finally {
      client?.disconnect();
    }
  }

  /**
   * Check if admin has approved onboarding (Step 2 complete).
   * In practice, if the address appears in domain.members, both steps are done.
   */
  async checkStep2Complete(
    domainId: string,
    address: string
  ): Promise<boolean> {
    const status = await this.getMembershipStatus(domainId, address);
    return status.step2Complete;
  }

  /**
   * Fetch Merkle proof for a member's identity commitment.
   */
  async fetchMerkleProof(
    domainId: string,
    identityCommitment: string
  ): Promise<{ root: string; pathIndices: number[]; pathElements: string[] }> {
    const value = await this.queries.query<unknown>(
      QUERY_PATHS.truedemocracy.merkleProof,
      [
        { number: 1, type: 'string', value: domainId },
        { number: 2, type: 'string', value: identityCommitment },
      ]
    );
    const proof = expectChainMerkleProof(
      QUERY_PATHS.truedemocracy.merkleProof,
      value
    );
    return {
      root: proof.root,
      pathIndices: proof.path_indices,
      pathElements: proof.path_elements,
    };
  }
}
