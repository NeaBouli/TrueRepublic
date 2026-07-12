import { describe, expect, it } from 'vitest';
import { ZKPService } from './zkp';
import { DEFAULT_CHAIN } from '@/config/chains';

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
});
