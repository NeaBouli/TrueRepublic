import type { GeneratedProof, Groth16Prover, ProofInputs } from '@/types/zkp';
import { MERKLE_TREE_DEPTH } from '@/types/zkp';
import {
  bytesToHex,
  computeVoteNullifierScope,
  computeVoteSignal,
  hexToBytes,
  mimcBn254,
} from './zkpEncoding';

export const ZKP_PROVER_REQUEST_SCHEMA =
  'truerepublic/zkp-prover-request/v1';
export const ZKP_PROVER_RESULT_SCHEMA = 'truerepublic/zkp-prover-result/v1';
export const MEMBERSHIP_CIRCUIT_ID =
  'truerepublic/membership-vote/v2-bn254-mimc-depth20';

export interface TestOnlyZKPArtifacts {
  constraintSystem: Uint8Array;
  provingKey: Uint8Array;
  verifyingKey: Uint8Array;
}

export interface TestOnlyZKPRuntime {
  prove(
    requestJSON: string,
    artifacts: TestOnlyZKPArtifacts
  ): Promise<string>;
}

export interface TestOnlyGoWasmResponse {
  ok: boolean;
  result: string;
  error: string;
}

export type TestOnlyGoWasmProveFunction = (
  requestJSON: string,
  constraintSystem: Uint8Array,
  provingKey: Uint8Array,
  verifyingKey: Uint8Array
) => TestOnlyGoWasmResponse;

/** Strict adapter around the global function registered by the test-only Go WASM command. */
export class TestOnlyGoWasmRuntime implements TestOnlyZKPRuntime {
  constructor(private readonly rawProve: TestOnlyGoWasmProveFunction) {}

  async prove(
    requestJSON: string,
    artifacts: TestOnlyZKPArtifacts
  ): Promise<string> {
    let response: TestOnlyGoWasmResponse;
    try {
      response = this.rawProve(
        requestJSON,
        artifacts.constraintSystem,
        artifacts.provingKey,
        artifacts.verifyingKey
      );
    } catch (error: unknown) {
      throw new Error(
        `test-only Go WASM runtime failed: ${error instanceof Error ? error.message : 'unknown error'}`
      );
    }
    if (
      !isRecord(response) ||
      JSON.stringify(Object.keys(response).sort()) !==
        JSON.stringify(['error', 'ok', 'result']) ||
      typeof response.ok !== 'boolean' ||
      typeof response.result !== 'string' ||
      typeof response.error !== 'string'
    ) {
      throw new Error('test-only Go WASM runtime returned a malformed envelope');
    }
    if (!response.ok) {
      throw new Error(response.error || 'test-only Go WASM prover rejected request');
    }
    if (response.error !== '' || response.result === '') {
      throw new Error('test-only Go WASM runtime returned an inconsistent envelope');
    }
    return response.result;
  }
}

interface ProverResult {
  schema: string;
  circuit_id: string;
  synthetic_and_test_only: boolean;
  proof_hex: string;
  nullifier_hash_hex: string;
  merkle_root_hex: string;
  public_signals_hex: string[];
}

/**
 * Real Groth16 compatibility prover for the pinned synthetic fixture only.
 * It deliberately has no transaction, wallet, RPC, or submission capability.
 */
export class TestOnlyGroth16WasmProver implements Groth16Prover {
  constructor(
    private readonly runtime: TestOnlyZKPRuntime,
    private readonly artifacts: TestOnlyZKPArtifacts
  ) {}

  async generate(inputs: ProofInputs): Promise<GeneratedProof> {
    validateInputs(inputs);
    const external = bytesToHex(
      computeVoteNullifierScope(
        inputs.chainId,
        inputs.domainName,
        inputs.issueName,
        inputs.suggestionName
      )
    );
    if (inputs.externalNullifier !== external) {
      throw new Error('external nullifier does not match the canonical vote scope');
    }
    const signal = bytesToHex(
      computeVoteSignal(
        inputs.chainId,
        inputs.domainName,
        inputs.issueName,
        inputs.suggestionName,
        inputs.rating
      )
    );
    const request = JSON.stringify({
      schema: ZKP_PROVER_REQUEST_SCHEMA,
      circuit_id: MEMBERSHIP_CIRCUIT_ID,
      synthetic_and_test_only: true,
      identity_secret_hex: inputs.identitySecret,
      merkle_root_hex: inputs.merkleRoot,
      siblings_hex: inputs.merkleProof.pathElements,
      path_indices: inputs.merkleProof.pathIndices,
      external_nullifier_hex: external,
      signal_hash_hex: signal,
    });
    const result = parseResult(await this.runtime.prove(request, this.artifacts));
    const nullifier = bytesToHex(
      mimcBn254([hexToBytes(inputs.identitySecret), hexToBytes(external)])
    );
    const expectedSignals = [inputs.merkleRoot, nullifier, external, signal];
    if (
      result.merkle_root_hex !== inputs.merkleRoot ||
      result.nullifier_hash_hex !== nullifier ||
      JSON.stringify(result.public_signals_hex) !== JSON.stringify(expectedSignals)
    ) {
      throw new Error('prover result public signals do not match the request');
    }
    return {
      proof: result.proof_hex,
      nullifierHash: result.nullifier_hash_hex,
      merkleRoot: result.merkle_root_hex,
      publicSignals: result.public_signals_hex,
    };
  }
}

function validateInputs(inputs: ProofInputs): void {
  if (
    inputs.merkleRoot !== inputs.merkleProof.root ||
    inputs.merkleProof.pathElements.length !== MERKLE_TREE_DEPTH ||
    inputs.merkleProof.pathIndices.length !== MERKLE_TREE_DEPTH ||
    !inputs.merkleProof.pathIndices.every((value) => value === 0 || value === 1)
  ) {
    throw new Error('invalid Merkle proof shape or root binding');
  }
  for (const [label, value] of [
    ['identity secret', inputs.identitySecret],
    ['merkle root', inputs.merkleRoot],
    ['external nullifier', inputs.externalNullifier],
    ['merkle leaf', inputs.merkleProof.leaf],
    ...inputs.merkleProof.pathElements.map(
      (value, index) => [`sibling ${index}`, value] as const
    ),
  ] as const) {
    const decoded = hexToBytes(value);
    if (decoded.length !== 32) throw new Error(`${label} must be exactly 32 bytes`);
  }
  const expectedLeaf = bytesToHex(mimcBn254([hexToBytes(inputs.identitySecret)]));
  if (inputs.merkleProof.leaf !== expectedLeaf) {
    throw new Error('Merkle proof leaf does not match the identity commitment');
  }
}

function parseResult(encoded: string): ProverResult {
  const value: unknown = JSON.parse(encoded);
  if (!isRecord(value)) throw new Error('prover result must be an object');
  const expectedKeys = [
    'circuit_id',
    'merkle_root_hex',
    'nullifier_hash_hex',
    'proof_hex',
    'public_signals_hex',
    'schema',
    'synthetic_and_test_only',
  ];
  if (JSON.stringify(Object.keys(value).sort()) !== JSON.stringify(expectedKeys)) {
    throw new Error('prover result contains missing or unknown fields');
  }
  if (
    value.schema !== ZKP_PROVER_RESULT_SCHEMA ||
    value.circuit_id !== MEMBERSHIP_CIRCUIT_ID ||
    value.synthetic_and_test_only !== true ||
    !isCanonicalHex(value.proof_hex) ||
    value.proof_hex.length !== 328 ||
    !isCanonicalField(value.nullifier_hash_hex) ||
    !isCanonicalField(value.merkle_root_hex) ||
    !Array.isArray(value.public_signals_hex) ||
    value.public_signals_hex.length !== 4 ||
    !value.public_signals_hex.every(isCanonicalField)
  ) {
    throw new Error('unsafe or malformed prover result');
  }
  return value as unknown as ProverResult;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function isCanonicalHex(value: unknown): value is string {
  return typeof value === 'string' && /^(?:[0-9a-f]{2})+$/u.test(value);
}

function isCanonicalField(value: unknown): value is string {
  return isCanonicalHex(value) && value.length === 64;
}
