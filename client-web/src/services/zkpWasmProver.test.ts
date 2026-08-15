import { describe, expect, it } from 'vitest';
import type { ProofInputs } from '@/types/zkp';
import {
  TestOnlyGroth16WasmProver,
  TestOnlyGoWasmRuntime,
  type TestOnlyZKPArtifacts,
  type TestOnlyZKPRuntime,
} from './zkpWasmProver';

const artifacts: TestOnlyZKPArtifacts = {
  constraintSystem: new Uint8Array(),
  provingKey: new Uint8Array(),
  verifyingKey: new Uint8Array(),
};
const inputs: ProofInputs = {
  chainId: 'truerepublic-zkp-fixture-1',
  identitySecret:
    '231208b9f8df97ba1aef7aaba5364d4cebcef3ae9f6dd861f42c0dd69f08991b',
  merkleRoot:
    '05ba2fb0b13a7d9162b742aec08904c50caf18b5b86b7d94ef3c75fece6feca8',
  merkleProof: {
    root: '05ba2fb0b13a7d9162b742aec08904c50caf18b5b86b7d94ef3c75fece6feca8',
    pathIndices: Array.from({ length: 20 }, () => 0),
    pathElements: Array.from({ length: 20 }, () => '00'.repeat(32)),
    leaf: '2f861838e73c02224caa920f32e6340501b2a924f9ab9d1d6f28c9dd90017a82',
  },
  externalNullifier:
    '0e5c757eadbc0d0b85d1835012ad9fb25533781644e4c428527b7de09326e283',
  rating: 3,
  domainName: 'FixtureDomain',
  issueName: 'FixtureIssue',
  suggestionName: 'FixtureSuggestion',
  rewardRecipient: 'truerepublic10f4hqttjv4mkzuny94ex2cmfwp5k2mn5kqf890',
};

describe('test-only Groth16 WASM client boundary', () => {
  it('rejects a vote context that does not bind the supplied nullifier', async () => {
    const runtime: TestOnlyZKPRuntime = {
      prove: async () => {
        throw new Error('must not run');
      },
    };
    const prover = new TestOnlyGroth16WasmProver(runtime, artifacts);

    await expect(
      prover.generate({ ...inputs, suggestionName: 'different' })
    ).rejects.toThrow('canonical vote scope');
  });

  it('rejects malformed runtime output', async () => {
    const runtime: TestOnlyZKPRuntime = {
      prove: async () => '{"schema":"unexpected"}',
    };
    const prover = new TestOnlyGroth16WasmProver(runtime, artifacts);

    await expect(prover.generate(inputs)).rejects.toThrow(
      'missing or unknown fields'
    );
  });

  it('rejects mismatched Merkle roots before invoking WASM', async () => {
    const runtime: TestOnlyZKPRuntime = {
      prove: async () => {
        throw new Error('must not run');
      },
    };
    const prover = new TestOnlyGroth16WasmProver(runtime, artifacts);

    await expect(
      prover.generate({
        ...inputs,
        merkleProof: { ...inputs.merkleProof, root: '01'.repeat(32) },
      })
    ).rejects.toThrow('root binding');
  });

  it('rejects a Merkle leaf not bound to the private identity', async () => {
    const runtime: TestOnlyZKPRuntime = {
      prove: async () => {
        throw new Error('must not run');
      },
    };
    const prover = new TestOnlyGroth16WasmProver(runtime, artifacts);
    await expect(
      prover.generate({
        ...inputs,
        merkleProof: { ...inputs.merkleProof, leaf: '00'.repeat(32) },
      })
    ).rejects.toThrow('identity commitment');
  });

  it('rejects a missing or non-canonical reward recipient before proving', async () => {
    const runtime: TestOnlyZKPRuntime = {
      prove: async () => {
        throw new Error('must not run');
      },
    };
    const prover = new TestOnlyGroth16WasmProver(runtime, artifacts);

    await expect(
      prover.generate({ ...inputs, rewardRecipient: '' })
    ).rejects.toThrow('reward recipient');
    await expect(
      prover.generate({
        ...inputs,
        rewardRecipient: inputs.rewardRecipient.toUpperCase(),
      })
    ).rejects.toThrow('reward recipient');
    await expect(
      prover.generate({ ...inputs, rewardRecipient: 'cosmos1qqqqqqqq' })
    ).rejects.toThrow('reward recipient');
    await expect(
      prover.generate({
        ...inputs,
        rewardRecipient: `${inputs.rewardRecipient.slice(0, -1)}q`,
      })
    ).rejects.toThrow('reward recipient');
  });

  it('fails closed on a trapped or malformed Go WASM runtime', async () => {
    const trapped = new TestOnlyGoWasmRuntime(() => {
      throw new Error('trap');
    });
    await expect(trapped.prove('{}', artifacts)).rejects.toThrow(
      'runtime failed: trap'
    );

    const malformed = new TestOnlyGoWasmRuntime(
      () => ({ ok: true, result: '{}', error: '', extra: true }) as never
    );
    await expect(malformed.prove('{}', artifacts)).rejects.toThrow(
      'malformed envelope'
    );
  });

  it('has no mutable initialization state across concurrent runtime calls', async () => {
    let calls = 0;
    const runtime = new TestOnlyGoWasmRuntime(() => {
      calls += 1;
      return { ok: true, result: '{}', error: '' };
    });
    await expect(
      Promise.all([runtime.prove('{}', artifacts), runtime.prove('{}', artifacts)])
    ).resolves.toEqual(['{}', '{}']);
    expect(calls).toBe(2);
  });
});
