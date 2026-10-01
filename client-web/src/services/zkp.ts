/**
 * ZKP Service - Client-side proof generation.
 *
 * PLACEHOLDER for Week 4. Actual gnark-wasm integration requires:
 * 1. Compiling Go MembershipCircuit to WASM (gnark + tinygo)
 * 2. Loading proving key + verifying key artifacts
 * 3. WASM bridge for Groth16 prove/verify
 *
 * Mock identity helpers remain for UI development, but proof generation fails
 * closed and no anonymous transaction can be submitted from this client.
 *
 * Go backend reference:
 * - Circuit: x/truedemocracy/zkp.go (MembershipCircuit)
 * - Merkle:  x/truedemocracy/merkle.go (MiMC, depth=20)
 * - Nullifier: MiMC(identitySecret, externalNullifier)
 * - Commitment: MiMC(identitySecret)
 */

import type {
  Identity,
  ProofInputs,
  GeneratedProof,
  ProofGenerationStatus,
  MerkleProof,
  Groth16Prover,
} from '@/types/zkp';
import type { ChainConfig } from '@/types/chain';
import {
  expectChainMerkleProof,
  expectQueryBoolean,
  expectQueryRecord,
  ModuleQueryClient,
  QUERY_PATHS,
} from './moduleQuery';
import { previewMockIdentityHash } from './previewIdentityHash';
import { bytesToHex, computeVoteNullifierScope, hexToBytes, mimcBn254 } from './zkpEncoding';

const FIELD_HEX = /^[0-9a-f]{64}$/u;

export class ZKPService {
  private wasmLoaded = false;
  private statusCallback?: (status: ProofGenerationStatus) => void;
  private readonly queries: ModuleQueryClient;
  private readonly prover?: Groth16Prover;
  private readonly chainId: string;

  /**
   * @param prover explicitly injected, reviewed Groth16 prover (test-only today).
   *   Without it every proof path fails closed; nothing is constructed implicitly.
   */
  constructor(
    config: ChainConfig,
    queries = new ModuleQueryClient(config),
    prover?: Groth16Prover
  ) {
    this.queries = queries;
    this.prover = prover;
    this.chainId = config.chainId;
  }

  /**
   * Initialize WASM module (loads gnark proving artifacts).
   */
  async initialize(
    onStatus?: (status: ProofGenerationStatus) => void
  ): Promise<void> {
    this.statusCallback = onStatus;
    this.wasmLoaded = false;
    if (!this.prover) {
      const message =
        'Anonymous voting is preview-only: a compatible real Groth16 prover is not installed.';
      this.updateStatus('error', 0, 'ZKP submission unavailable', message);
      throw new Error(message);
    }
    try {
      await this.prover.initialize?.();
    } catch {
      const message = 'The injected Groth16 prover failed to initialize.';
      this.updateStatus('error', 0, 'ZKP submission unavailable', message);
      throw new Error(message);
    }
    this.wasmLoaded = true;
    this.updateStatus('complete', 100, 'Test-only Groth16 prover ready; submission remains disabled');
  }

  get isReady(): boolean {
    return this.wasmLoaded;
  }

  get isSubmittable(): boolean {
    return false;
  }

  /**
   * Generate a new ZKP identity.
   * Identity secret is 32 random bytes. Commitment = MiMC(secret).
   */
  generateIdentity(): Identity {
    const secret = this.randomHex(32);
    const commitment = this.mockMiMCHash(secret);
    const nullifier = this.mockMiMCHash(secret + '00');

    return {
      secret,
      commitment,
      nullifier,
      createdAt: Date.now(),
    };
  }

  /**
   * Fetch Merkle proof for a commitment from the chain.
   */
  async fetchMerkleProof(
    domainName: string,
    commitment: string
  ): Promise<MerkleProof> {
    const value = await this.queries.query<unknown>(
      QUERY_PATHS.truedemocracy.merkleProof,
      [
        { number: 1, type: 'string', value: domainName },
        { number: 2, type: 'string', value: commitment },
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
      leaf: proof.commitment,
    };
  }

  /**
   * Canonical chain-scoped external nullifier (frozen protocol):
   * computeVoteNullifierScope(chainId, domain, issue, suggestion).
   */
  computeExternalNullifier(
    domainName: string,
    issueName: string,
    suggestionName: string
  ): string {
    return bytesToHex(computeVoteNullifierScope(this.chainId, domainName, issueName, suggestionName));
  }

  /**
   * Canonical vote nullifier hash = MiMC_BN254(identitySecret, externalNullifier).
   * Both inputs must be canonical 32-byte BN254 field elements.
   */
  computeNullifierHash(
    identitySecret: string,
    externalNullifier: string
  ): string {
    if (!FIELD_HEX.test(identitySecret) || !FIELD_HEX.test(externalNullifier)) {
      throw new Error('identity secret and external nullifier must be exactly 32 bytes of lowercase hex');
    }
    try {
      return bytesToHex(mimcBn254([hexToBytes(identitySecret), hexToBytes(externalNullifier)]));
    } catch {
      throw new Error('identity secret and external nullifier must be canonical BN254 field elements');
    }
  }

  /**
   * Check if a nullifier has already been used on-chain.
   */
  async isNullifierUsed(
    domainName: string,
    nullifierHash: string
  ): Promise<boolean> {
    const value = await this.queries.query<unknown>(
      QUERY_PATHS.truedemocracy.nullifier,
      [
        { number: 1, type: 'string', value: domainName },
        { number: 2, type: 'string', value: nullifierHash },
      ]
    );
    const result = expectQueryRecord(
      QUERY_PATHS.truedemocracy.nullifier,
      value
    );
    return expectQueryBoolean(
      QUERY_PATHS.truedemocracy.nullifier,
      'used',
      result.used
    );
  }

  /**
   * Generate Groth16 proof for anonymous vote.
   *
   * This is the main entry point for proof generation.
   * Real implementation calls gnark-wasm with the proving key.
   */
  async generateProof(inputs: ProofInputs): Promise<GeneratedProof> {
    if (!this.prover) {
      const message =
        'Mock proofs are not chain-compatible; real Groth16 proof generation is unavailable.';
      this.updateStatus('error', 0, 'ZKP submission unavailable', message);
      throw new Error(message);
    }
    if (!this.wasmLoaded) {
      throw new Error('The injected Groth16 prover has not been initialized.');
    }
    this.assertCanonicalProofInputs(inputs);
    return this.prover.generate(inputs);
  }

  /**
   * Only a canonical identity can enter the prover: the secret must be a BN254
   * field element whose MiMC commitment is the Merkle leaf, and the external
   * nullifier must be this chain's scope for the vote context. Preview (FNV)
   * identities fail here and are never promoted.
   */
  private assertCanonicalProofInputs(inputs: ProofInputs): void {
    if (inputs.chainId !== this.chainId) {
      throw new Error('proof inputs are bound to a different chain');
    }
    const secret = inputs.identitySecret;
    const leaf = inputs.merkleProof?.leaf;
    if (typeof secret !== 'string' || !FIELD_HEX.test(secret) || typeof leaf !== 'string' || !FIELD_HEX.test(leaf)) {
      throw new Error('proof inputs require a 32-byte canonical identity secret and Merkle leaf');
    }
    let commitment: string;
    try {
      commitment = bytesToHex(mimcBn254([hexToBytes(secret)]));
    } catch {
      throw new Error('preview or non-canonical identity secrets cannot generate proofs');
    }
    if (commitment !== leaf) {
      throw new Error('preview or non-canonical identity: commitment is not MiMC(secret)');
    }
    const expectedScope = this.computeExternalNullifier(inputs.domainName, inputs.issueName, inputs.suggestionName);
    if (inputs.externalNullifier !== expectedScope) {
      throw new Error('external nullifier does not match the canonical vote scope');
    }
  }

  // ---------------------------------------------------------------
  // Private helpers
  // ---------------------------------------------------------------

  private updateStatus(
    step: ProofGenerationStatus['step'],
    progress: number,
    message: string,
    error?: string
  ): void {
    this.statusCallback?.({ step, progress, message, error });
  }

  private randomHex(bytes: number): string {
    const array = new Uint8Array(bytes);
    crypto.getRandomValues(array);
    return Array.from(array)
      .map((b) => b.toString(16).padStart(2, '0'))
      .join('');
  }

  /**
   * Mock MiMC hash (SHA-256 truncated to 32 bytes).
   * Real implementation uses MiMC over BN254 scalar field.
   */
  private mockMiMCHash(input: string): string {
    // Preview placeholder, not MiMC: see previewIdentityHash.ts.
    return previewMockIdentityHash(input);
  }

}
