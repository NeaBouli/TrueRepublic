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
    const prover: Groth16Prover = {
      generate: async () => fixture,
    };
    const service = new ZKPService(DEFAULT_CHAIN, undefined, prover);

    expect(service.isSubmittable).toBe(false);
    await expect(service.generateProof({} as ProofInputs)).resolves.toEqual(
      fixture
    );
    expect(service.isSubmittable).toBe(false);
  });
});
