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

export class ZKPService {
  private wasmLoaded = false;
  private statusCallback?: (status: ProofGenerationStatus) => void;
  private readonly queries: ModuleQueryClient;
  private readonly testProver?: Groth16Prover;

  constructor(
    config: ChainConfig,
    queries = new ModuleQueryClient(config),
    testProver?: Groth16Prover
  ) {
    this.queries = queries;
    this.testProver = testProver;
  }

  /**
   * Initialize WASM module (loads gnark proving artifacts).
   */
  async initialize(
    onStatus?: (status: ProofGenerationStatus) => void
  ): Promise<void> {
    this.statusCallback = onStatus;
    const message =
      'Anonymous voting is preview-only: a compatible real Groth16 prover is not installed.';
    this.wasmLoaded = false;
    this.updateStatus('error', 0, 'ZKP submission unavailable', message);
    throw new Error(message);
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
   * Compute the external nullifier for a specific vote context.
   * externalNullifier = MiMC(domainName + issueName + suggestionName)
   * This matches Go ComputeExternalNullifier().
   */
  computeExternalNullifier(
    domainName: string,
    issueName: string,
    suggestionName: string
  ): string {
    return this.mockMiMCHash(domainName + ':' + issueName + ':' + suggestionName);
  }

  /**
   * Compute the nullifier hash for double-voting prevention.
   * nullifierHash = MiMC(identitySecret, externalNullifier)
   * This matches Go ComputeNullifier().
   */
  computeNullifierHash(
    identitySecret: string,
    externalNullifier: string
  ): string {
    return this.mockMiMCHash(identitySecret + externalNullifier);
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
    if (this.testProver) return this.testProver.generate(inputs);
    const message =
      'Mock proofs are not chain-compatible; real Groth16 proof generation is unavailable.';
    this.updateStatus('error', 0, 'ZKP submission unavailable', message);
    throw new Error(message);
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
    // Sync mock: simple deterministic hash for testing.
    // Real MiMC operates on field elements; this is a placeholder.
    let hash = 0x811c9dc5;
    for (let i = 0; i < input.length; i++) {
      hash ^= input.charCodeAt(i);
      hash = Math.imul(hash, 0x01000193);
    }
    const hex = (hash >>> 0).toString(16).padStart(8, '0');
    // Pad to 64 chars (32 bytes) for consistency with real hashes
    return (hex + hex + hex + hex + hex + hex + hex + hex).slice(0, 64);
  }

}
