import { toBech32 } from '@cosmjs/encoding';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { sealEnvelope } from './custodyEnvelope';
import {
  CANONICAL_IDENTITY_KIND,
  IDENTITY_VAULT_STORAGE_KEY,
  IdentityVault,
  IdentityVaultError,
  PREVIEW_IDENTITY_KIND,
  canonicalIdentityAad,
  serializeCanonicalIdentity,
  type CanonicalIdentity,
} from './identityVault';

// GH309C1a: this file never imports zkpEncoding statically. Every attempt to
// load it hits a factory that counts the attempt and fails, which models a
// chunk-load failure and proves when the canonical validator module is loaded.
const loader = vi.hoisted(() => ({ attempts: 0 }));
vi.mock('./zkpEncoding', () => {
  loader.attempts += 1;
  throw new Error('simulated canonical validator chunk load failure');
});

// Synthetic material only. Frozen record-v1 (salt 0x40.., IV 0x70..) and the
// frozen BN254 MiMC commitment of secret 0x0f x32, both captured from the
// accepted custody and encoding code (identityVaultSchema.test.ts uses the same values).
const FROZEN_ADDRESS = toBech32('truerepublic', new Uint8Array(20).fill(9));
const FROZEN_PASSWORD = 'synthetic-frozen-password';
const FROZEN_V1_ENVELOPE =
  'v2:QEFCQ0RFRkdISUpLTE1OT3BxcnN0dXZ3eHl6e5StHe6bbRzO2NTj85nqo98Iwu5eolfxD4aniSIzHw97CItaspjfmFFwVxZ6oRzdQ/mCSkXc0QyAFtK/YF5P4iM8VuvLubxcQUzufFMkP4d/4DJbMFnuK7d3QVirfKhqDua6JRJzFGIxwoeIe8D7Rll96ATsce8B1yfxZIG8mN/LFZtICbsk0PEARMd97TNQR3KWsb07zpNOodowNMCKEigbfL7sREpgnexRZxSOGHEtBhuM8WGfWYBBfa2aeZwvBH3zZLY7Ce6GCM3n21akGE4f2g7iDdeXisTjl/5fGAzXRt6dnVmgmlh/a67l2du0EP4C98s4t7x11w4iIkFDKqlILZtBh43hKt8J3LGILzPVjMzPqExW';
const PASSWORD = 'synthetic-schema-password';
const ADDRESS = toBech32('truerepublic', new Uint8Array(20).fill(5));
const CANONICAL: CanonicalIdentity = {
  kind: CANONICAL_IDENTITY_KIND,
  secret: '0f'.repeat(32),
  commitment: '004c1d64b08fcced074a7b3e4ec6b7447982b2a11848596f818f19a9dbc05a27',
  createdAt: 1_700_000_000_000,
};
const RECORD_BASE = { alg: 'AES-GCM-256', kdf: 'PBKDF2-SHA256', iter: 600_000, createdAt: 1 };

describe('lazy canonical validator chunk (GH309C1a)', { timeout: 60_000 }, () => {
  beforeEach(() => {
    localStorage.clear();
    loader.attempts = 0;
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('opens a frozen record-v1 without ever loading the canonical validator module', async () => {
    const raw = JSON.stringify({ v: 1, records: { [FROZEN_ADDRESS]: { v: 1, ...RECORD_BASE, envelope: FROZEN_V1_ENVELOPE } } });
    localStorage.setItem(IDENTITY_VAULT_STORAGE_KEY, raw);
    await expect(new IdentityVault().openCustodyIdentity(FROZEN_ADDRESS, FROZEN_PASSWORD)).resolves.toMatchObject({
      kind: PREVIEW_IDENTITY_KIND,
      secret: '91'.repeat(32),
    });
    expect(loader.attempts).toBe(0);
    expect(localStorage.getItem(IDENTITY_VAULT_STORAGE_KEY)).toBe(raw);
  });

  it('maps a validator load failure on a valid record-v2 to storage and leaves the record unchanged', async () => {
    const envelope = await sealEnvelope(serializeCanonicalIdentity(CANONICAL), PASSWORD, {
      aad: canonicalIdentityAad(ADDRESS),
      maxEnvelopeChars: 1024,
    });
    const raw = JSON.stringify({
      v: 1,
      records: { [ADDRESS]: { v: 2, kind: CANONICAL_IDENTITY_KIND, ...RECORD_BASE, envelope } },
    });
    localStorage.setItem(IDENTITY_VAULT_STORAGE_KEY, raw);

    let error: unknown = null;
    try {
      await new IdentityVault().openCustodyIdentity(ADDRESS, PASSWORD);
    } catch (caught) {
      error = caught;
    }
    expect(error).toBeInstanceOf(IdentityVaultError);
    expect((error as IdentityVaultError).code).toBe('storage');
    expect(loader.attempts).toBe(1);
    expect(localStorage.getItem(IDENTITY_VAULT_STORAGE_KEY)).toBe(raw);
  });

  it('does not attempt the load for a record-v2 that fails decryption or shape first', async () => {
    const envelope = await sealEnvelope(serializeCanonicalIdentity(CANONICAL), PASSWORD, {
      aad: canonicalIdentityAad(ADDRESS),
      maxEnvelopeChars: 1024,
    });
    localStorage.setItem(
      IDENTITY_VAULT_STORAGE_KEY,
      JSON.stringify({ v: 1, records: { [ADDRESS]: { v: 2, kind: CANONICAL_IDENTITY_KIND, ...RECORD_BASE, envelope } } })
    );
    await expect(new IdentityVault().openCustodyIdentity(ADDRESS, 'wrong-schema-password')).rejects.toMatchObject({
      code: 'locked',
    });

    const badShape = await sealEnvelope(JSON.stringify({ ...CANONICAL, nullifier: '00'.repeat(32) }), PASSWORD, {
      aad: canonicalIdentityAad(ADDRESS),
      maxEnvelopeChars: 1024,
    });
    localStorage.setItem(
      IDENTITY_VAULT_STORAGE_KEY,
      JSON.stringify({ v: 1, records: { [ADDRESS]: { v: 2, kind: CANONICAL_IDENTITY_KIND, ...RECORD_BASE, envelope: badShape } } })
    );
    await expect(new IdentityVault().openCustodyIdentity(ADDRESS, PASSWORD)).rejects.toMatchObject({ code: 'corrupt' });
    expect(loader.attempts).toBe(0);
  });
});
