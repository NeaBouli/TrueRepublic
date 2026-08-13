import { describe, expect, it } from 'vitest';
import { ZKPService } from './zkp';
import { DEFAULT_CHAIN } from '@/config/chains';
import type { GeneratedProof, Groth16Prover, ProofInputs } from '@/types/zkp';

describe('ZKPService fail-closed boundary', () => {
  it('never reports the mock prover as submittable', async () => {
    const service = new ZKPService(DEFAULT_CHAIN);

    expect(service.isReady).toBe(false);
    expect(service.isSubmittable).toBe(false);
    await expect(service.initialize()).rejects.toThrow('preview-only');
    expect(service.isReady).toBe(false);
  });

  it('rejects direct mock-proof generation', async () => {
    const service = new ZKPService(DEFAULT_CHAIN);

    await expect(service.generateProof({} as never)).rejects.toThrow(
      'not chain-compatible'
    );
  });

  it('keeps submission disabled when a test-only prover is injected', async () => {
    const fixture: GeneratedProof = {
      proof: '00',
      nullifierHash: '01',
      merkleRoot: '02',
      publicSignals: ['02', '01'],
    };
    const inputs: ProofInputs = {
      identitySecret: '03',
      merkleRoot: '04',
      merkleProof: {
        root: '04',
        pathIndices: Array.from({ length: 20 }, () => 0),
        pathElements: Array.from({ length: 20 }, () => '00'),
        leaf: '05',
      },
      externalNullifier: '06',
      rating: 3,
      domainName: 'FixtureDomain',
      issueName: 'FixtureIssue',
      suggestionName: 'FixtureSuggestion',
    };
    let forwarded: ProofInputs | undefined;
    const prover: Groth16Prover = {
      generate: async (received) => {
        forwarded = received;
        return fixture;
      },
    };
    const service = new ZKPService(DEFAULT_CHAIN, undefined, prover);

    expect(service.isSubmittable).toBe(false);
    await expect(service.generateProof(inputs)).resolves.toEqual(fixture);
    expect(forwarded).toEqual(inputs);
    expect(service.isSubmittable).toBe(false);
  });
});
