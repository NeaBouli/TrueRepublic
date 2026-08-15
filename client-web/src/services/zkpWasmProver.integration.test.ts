import { readFileSync, writeFileSync } from 'node:fs';
import { runInThisContext } from 'node:vm';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';
import type { ProofInputs } from '@/types/zkp';
import {
  TestOnlyGroth16WasmProver,
  TestOnlyGoWasmRuntime,
  type TestOnlyZKPArtifacts,
  type TestOnlyGoWasmResponse,
} from './zkpWasmProver';

interface GoldenVector {
  chain_id: string;
  domain_name: string;
  issue_name: string;
  suggestion_name: string;
  rating: number;
  synthetic_witness_hex: string;
  commitment_hex: string;
  merkle_root_hex: string;
  siblings_hex: string[];
  path_indices: number[];
  external_nullifier_hex: string;
  signal_hash_hex: string;
  nullifier_hash_hex: string;
}

interface CircuitSpecV2 {
  vote_context_v2: {
    vector: {
      reward_recipient: string;
      signal_hash_hex: string;
    };
  };
}

interface GoRuntime {
  importObject: WebAssembly.Imports;
  run(instance: WebAssembly.Instance): Promise<void>;
}

interface GoRuntimeConstructor {
  new (): GoRuntime;
}

const enabled = process.env.TRUEREPUBLIC_ZKP_WASM_INTEGRATION === '1';
const integration = enabled ? describe : describe.skip;

integration('real maintained-client Groth16 WASM compatibility', () => {
  it('generates the pinned synthetic proof through Go WASM', async () => {
    const wasmPath = requiredEnvironment('TRUEREPUBLIC_ZKP_WASM_PATH');
    const wasmExecPath = requiredEnvironment('TRUEREPUBLIC_WASM_EXEC_PATH');
    const resultPath = requiredEnvironment('TRUEREPUBLIC_ZKP_RESULT_PATH');
    runInThisContext(readFileSync(wasmExecPath, 'utf8'), {
      filename: wasmExecPath,
    });
    const Go = (globalThis as typeof globalThis & { Go?: GoRuntimeConstructor })
      .Go;
    if (!Go) throw new Error('Go WASM runtime did not register global Go');
    const go = new Go();
    const instantiated = await WebAssembly.instantiate(
      readFileSync(wasmPath),
      go.importObject
    );
    void go.run(instantiated.instance);

    const rawProve = (
      globalThis as typeof globalThis & {
        trueRepublicTestOnlyGroth16Prove?: (
          request: string,
          cs: Uint8Array,
          pk: Uint8Array,
          vk: Uint8Array
        ) => TestOnlyGoWasmResponse;
      }
    ).trueRepublicTestOnlyGroth16Prove;
    if (!rawProve) throw new Error('test-only Go prover did not register');

    const fixtureDirectory = resolve(
      process.cwd(),
      '../x/truedemocracy/testdata/zkp'
    );
    const vector = JSON.parse(
      readFileSync(resolve(fixtureDirectory, 'golden_vector.json'), 'utf8')
    ) as GoldenVector;
    const specV2 = JSON.parse(
      readFileSync(
        resolve(process.cwd(), '../configs/security/zkp-circuit.json'),
        'utf8'
      )
    ) as CircuitSpecV2;
    const v2Vector = specV2.vote_context_v2.vector;
    const artifacts: TestOnlyZKPArtifacts = {
      constraintSystem: readFileSync(
        resolve(fixtureDirectory, 'membership_v2.cs')
      ),
      provingKey: readFileSync(resolve(fixtureDirectory, 'membership_v2.pk')),
      verifyingKey: readFileSync(resolve(fixtureDirectory, 'membership_v2.vk')),
    };
    const runtime = new TestOnlyGoWasmRuntime(rawProve);
    const inputs: ProofInputs = {
      chainId: vector.chain_id,
      identitySecret: vector.synthetic_witness_hex,
      merkleRoot: vector.merkle_root_hex,
      merkleProof: {
        root: vector.merkle_root_hex,
        pathIndices: vector.path_indices,
        pathElements: vector.siblings_hex,
        leaf: vector.commitment_hex,
      },
      externalNullifier: vector.external_nullifier_hex,
      rating: vector.rating,
      domainName: vector.domain_name,
      issueName: vector.issue_name,
      suggestionName: vector.suggestion_name,
      rewardRecipient: v2Vector.reward_recipient,
    };
    const result = await new TestOnlyGroth16WasmProver(
      runtime,
      artifacts
    ).generate(inputs);

    expect(result.proof).toMatch(/^[0-9a-f]+$/u);
    expect(result.nullifierHash).toBe(vector.nullifier_hash_hex);
    expect(result.merkleRoot).toBe(vector.merkle_root_hex);
    // GH-209: the proof's public signal is the pinned recipient-bound v2
    // signal; the external nullifier and nullifier stay identical to the
    // recipient-independent v1 fixture values.
    expect(result.publicSignals).toEqual([
      vector.merkle_root_hex,
      vector.nullifier_hash_hex,
      vector.external_nullifier_hex,
      v2Vector.signal_hash_hex,
    ]);
    writeFileSync(resultPath, `${JSON.stringify(result)}\n`, {
      encoding: 'utf8',
      mode: 0o600,
    });
  }, 120_000);
});

function requiredEnvironment(name: string): string {
  const value = process.env[name];
  if (!value) throw new Error(`${name} is required`);
  return value;
}
