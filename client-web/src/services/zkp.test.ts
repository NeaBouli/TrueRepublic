import { describe, expect, it } from 'vitest';
import { ZKPService } from './zkp';
import { DEFAULT_CHAIN } from '@/config/chains';
import type { GeneratedProof, Groth16Prover, ProofInputs } from '@/types/zkp';
import { bytesToHex, computeVoteNullifierScope, hexToBytes, mimcBn254 } from './zkpEncoding';
import { previewMockIdentityHash } from './previewIdentityHash';

// Synthetic canonical-bn254-mimc-v1 fixture: secret 0x0f x32 is a BN254 field
// element and its frozen commitment is MiMC(secret). Never a real identity.
const CANONICAL_SECRET = '0f'.repeat(32);
const CANONICAL_COMMITMENT = '004c1d64b08fcced074a7b3e4ec6b7447982b2a11848596f818f19a9dbc05a27';
const FIXTURE: GeneratedProof = {
  proof: '00',
  nullifierHash: '01',
  merkleRoot: '02',
  publicSignals: ['02', '01'],
};

function canonicalInputs(service: ZKPService, overrides: Partial<ProofInputs> = {}): ProofInputs {
  return {
    chainId: DEFAULT_CHAIN.chainId,
    identitySecret: CANONICAL_SECRET,
    merkleRoot: '04'.repeat(32),
    merkleProof: {
      root: '04'.repeat(32),
      pathIndices: Array.from({ length: 20 }, () => 0),
      pathElements: Array.from({ length: 20 }, () => '00'.repeat(32)),
      leaf: CANONICAL_COMMITMENT,
    },
    externalNullifier: service.computeExternalNullifier('FixtureDomain', 'FixtureIssue', 'FixtureSuggestion'),
    rating: 3,
    domainName: 'FixtureDomain',
    issueName: 'FixtureIssue',
    suggestionName: 'FixtureSuggestion',
    rewardRecipient: 'truerepublic10f4hqttjv4mkzuny94ex2cmfwp5k2mn5kqf890',
    ...overrides,
  };
}

function recordingProver(options: { initialize?: () => Promise<void> } = {}) {
  const forwarded: ProofInputs[] = [];
  const prover: Groth16Prover = {
    ...(options.initialize ? { initialize: options.initialize } : {}),
    generate: async (received) => {
      forwarded.push(received);
      return FIXTURE;
    },
  };
  return { prover, forwarded };
}

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

  it('routes a canonical fixture only through the initialized injected prover; submission stays disabled', async () => {
    const { prover, forwarded } = recordingProver();
    const service = new ZKPService(DEFAULT_CHAIN, undefined, prover);
    const inputs = canonicalInputs(service);

    expect(service.isSubmittable).toBe(false);
    await expect(service.initialize()).resolves.toBeUndefined();
    expect(service.isReady).toBe(true);
    await expect(service.generateProof(inputs)).resolves.toEqual(FIXTURE);
    expect(forwarded).toEqual([inputs]);
    expect(service.isSubmittable).toBe(false);
  });
});

// GH300B3: explicit prover dependency, canonical nullifiers and identity binding.
describe('ZKPService injected prover wiring (GH300B3)', () => {
  it('refuses generation before the injected prover is initialized', async () => {
    const { prover, forwarded } = recordingProver();
    const service = new ZKPService(DEFAULT_CHAIN, undefined, prover);
    await expect(service.generateProof(canonicalInputs(service))).rejects.toThrow('has not been initialized');
    expect(forwarded).toEqual([]);
  });

  it('fails closed when the injected prover fails to initialize', async () => {
    const { prover, forwarded } = recordingProver({
      initialize: () => Promise.reject(new Error('artifact load failed')),
    });
    const service = new ZKPService(DEFAULT_CHAIN, undefined, prover);
    await expect(service.initialize()).rejects.toThrow('failed to initialize');
    expect(service.isReady).toBe(false);
    await expect(service.generateProof(canonicalInputs(service))).rejects.toThrow('has not been initialized');
    expect(forwarded).toEqual([]);
  });

  it('rejects preview (FNV) identities and non-canonical secrets before the prover', async () => {
    const { prover, forwarded } = recordingProver();
    const service = new ZKPService(DEFAULT_CHAIN, undefined, prover);
    await service.initialize();
    const preview = service.generateIdentity();
    expect(preview.commitment).toBe(previewMockIdentityHash(preview.secret));

    await expect(
      service.generateProof(
        canonicalInputs(service, {
          identitySecret: preview.secret,
          merkleProof: { ...canonicalInputs(service).merkleProof, leaf: preview.commitment },
        })
      )
    ).rejects.toThrow(/preview or non-canonical identity/);

    // Deterministic: a field-element secret with its preview (FNV) commitment as leaf.
    await expect(
      service.generateProof(
        canonicalInputs(service, {
          merkleProof: { ...canonicalInputs(service).merkleProof, leaf: previewMockIdentityHash(CANONICAL_SECRET) },
        })
      )
    ).rejects.toThrow('commitment is not MiMC(secret)');

    const outOfField = 'ff'.repeat(32);
    await expect(
      service.generateProof(canonicalInputs(service, { identitySecret: outOfField }))
    ).rejects.toThrow('preview or non-canonical identity secrets cannot generate proofs');

    await expect(
      service.generateProof(canonicalInputs(service, { identitySecret: '03' }))
    ).rejects.toThrow('32-byte canonical identity secret');
    expect(forwarded).toEqual([]);
  });

  it('rejects a foreign chain or a mismatched external nullifier scope', async () => {
    const { prover, forwarded } = recordingProver();
    const service = new ZKPService(DEFAULT_CHAIN, undefined, prover);
    await service.initialize();
    await expect(service.generateProof(canonicalInputs(service, { chainId: 'other-chain-1' }))).rejects.toThrow(
      'different chain'
    );
    await expect(
      service.generateProof(canonicalInputs(service, { externalNullifier: '07'.repeat(32) }))
    ).rejects.toThrow('canonical vote scope');
    expect(forwarded).toEqual([]);
  });

  it('uses the canonical chain-bound nullifier scope and MiMC nullifier', () => {
    const service = new ZKPService(DEFAULT_CHAIN);
    const external = service.computeExternalNullifier('FixtureDomain', 'FixtureIssue', 'FixtureSuggestion');
    expect(external).toBe(
      bytesToHex(computeVoteNullifierScope(DEFAULT_CHAIN.chainId, 'FixtureDomain', 'FixtureIssue', 'FixtureSuggestion'))
    );
    const secret = '231208b9f8df97ba1aef7aaba5364d4cebcef3ae9f6dd861f42c0dd69f08991b';
    expect(service.computeNullifierHash(secret, external)).toBe(
      bytesToHex(mimcBn254([hexToBytes(secret), hexToBytes(external)]))
    );
  });

  it('rejects malformed or non-field nullifier inputs', () => {
    const service = new ZKPService(DEFAULT_CHAIN);
    expect(() => service.computeNullifierHash('01', '02')).toThrow('exactly 32 bytes');
    expect(() => service.computeNullifierHash('AB'.repeat(32), '02'.repeat(32))).toThrow('exactly 32 bytes');
    expect(() => service.computeNullifierHash('ff'.repeat(32), '02'.repeat(32))).toThrow(
      'canonical BN254 field elements'
    );
  });

  it('keeps preview identity generation on the historical preview hash', () => {
    const identity = new ZKPService(DEFAULT_CHAIN).generateIdentity();
    expect(identity.commitment).toBe(previewMockIdentityHash(identity.secret));
    expect(identity.nullifier).toBe(previewMockIdentityHash(`${identity.secret}00`));
  });
});
