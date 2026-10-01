import { toBech32 } from '@cosmjs/encoding';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { openEnvelope, sealEnvelope } from './custodyEnvelope';
import { LEGACY_IDENTITY_STORAGE_KEY, LegacyIdentityMigration } from './identityMigration';
import {
  CANONICAL_IDENTITY_KIND,
  IDENTITY_VAULT_STORAGE_KEY,
  IdentityVault,
  IdentityVaultError,
  PREVIEW_IDENTITY_KIND,
  canonicalIdentityAad,
  identityAad,
  isValidCanonicalIdentity,
  serializeCanonicalIdentity,
  serializeIdentity,
  type CanonicalIdentity,
  type IdentityVaultErrorCode,
} from './identityVault';
import { previewMockIdentityHash } from './previewIdentityHash';
import { BN254_SCALAR_MODULUS, bytesToHex, hexToBytes, mimcBn254 } from './zkpEncoding';

// Synthetic material only. There is deliberately no production API that
// writes a canonical record yet (GH309C1); tests build record-v2 entries
// directly with the custody envelope and the exported v2 AAD.
const PASSWORD = 'synthetic-schema-password';
const ADDRESS_A = toBech32('truerepublic', new Uint8Array(20).fill(5));
const ADDRESS_B = toBech32('truerepublic', new Uint8Array(20).fill(6));
const CANONICAL_SECRET = '0f'.repeat(32);
const CANONICAL: CanonicalIdentity = {
  kind: CANONICAL_IDENTITY_KIND,
  secret: CANONICAL_SECRET,
  commitment: bytesToHex(mimcBn254([hexToBytes(CANONICAL_SECRET)])),
  createdAt: 1_700_000_000_000,
};

// Frozen record-v1 written by the accepted GH309B2B custody code (salt 0x40..0x4f,
// IV 0x70..0x7b) for ADDRESS 0x09x20 and password 'synthetic-frozen-password'.
const FROZEN_ADDRESS = toBech32('truerepublic', new Uint8Array(20).fill(9));
const FROZEN_PASSWORD = 'synthetic-frozen-password';
const FROZEN_IDENTITY = { secret: '91'.repeat(32), commitment: '92'.repeat(32), nullifier: '93'.repeat(32), createdAt: 1_700_000_000_000 };
const FROZEN_V1_ENVELOPE =
  'v2:QEFCQ0RFRkdISUpLTE1OT3BxcnN0dXZ3eHl6e5StHe6bbRzO2NTj85nqo98Iwu5eolfxD4aniSIzHw97CItaspjfmFFwVxZ6oRzdQ/mCSkXc0QyAFtK/YF5P4iM8VuvLubxcQUzufFMkP4d/4DJbMFnuK7d3QVirfKhqDua6JRJzFGIxwoeIe8D7Rll96ATsce8B1yfxZIG8mN/LFZtICbsk0PEARMd97TNQR3KWsb07zpNOodowNMCKEigbfL7sREpgnexRZxSOGHEtBhuM8WGfWYBBfa2aeZwvBH3zZLY7Ce6GCM3n21akGE4f2g7iDdeXisTjl/5fGAzXRt6dnVmgmlh/a67l2du0EP4C98s4t7x11w4iIkFDKqlILZtBh43hKt8J3LGILzPVjMzPqExW';

const RECORD_BASE = { alg: 'AES-GCM-256', kdf: 'PBKDF2-SHA256', iter: 600_000, createdAt: 1 };

class FifoLocks {
  private held = false;
  private queue: Array<() => void> = [];

  request<T>(_name: string, _options: { mode: 'exclusive' }, callback: () => Promise<T>): Promise<T> {
    return new Promise<T>((resolve, reject) => {
      const grant = () => {
        this.held = true;
        Promise.resolve()
          .then(callback)
          .then(resolve, reject)
          .finally(() => {
            this.held = false;
            this.queue.shift()?.();
          });
      };
      if (this.held) this.queue.push(grant);
      else grant();
    });
  }
}

function storeRecords(records: Record<string, unknown>): string {
  const raw = JSON.stringify({ v: 1, records });
  localStorage.setItem(IDENTITY_VAULT_STORAGE_KEY, raw);
  return raw;
}

async function canonicalRecord(address: string, identity: CanonicalIdentity = CANONICAL) {
  const envelope = await sealEnvelope(serializeCanonicalIdentity(identity), PASSWORD, {
    aad: canonicalIdentityAad(address),
    maxEnvelopeChars: 1024,
  });
  return { v: 2, kind: CANONICAL_IDENTITY_KIND, ...RECORD_BASE, envelope };
}

async function expectVaultError(action: (() => unknown) | Promise<unknown>, code: IdentityVaultErrorCode) {
  let error: unknown = null;
  try {
    await (typeof action === 'function' ? action() : action);
  } catch (caught) {
    error = caught;
  }
  expect(error).toBeInstanceOf(IdentityVaultError);
  expect((error as IdentityVaultError).code).toBe(code);
}

describe('versioned identity custody schema (GH309C1)', { timeout: 60_000 }, () => {
  let vault: IdentityVault;

  beforeEach(() => {
    localStorage.clear();
    vi.stubGlobal('navigator', { locks: new FifoLocks() });
    vault = new IdentityVault();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('keeps record-v1 bytes and AAD: the frozen record opens unchanged as preview-v0', async () => {
    const raw = storeRecords({ [FROZEN_ADDRESS]: { v: 1, ...RECORD_BASE, envelope: FROZEN_V1_ENVELOPE } });
    await expect(
      openEnvelope(FROZEN_V1_ENVELOPE, FROZEN_PASSWORD, { aad: identityAad(FROZEN_ADDRESS), maxEnvelopeChars: 1024 })
    ).resolves.toMatchObject({ plaintext: serializeIdentity(FROZEN_IDENTITY) });
    await expect(vault.openCustodyIdentity(FROZEN_ADDRESS, FROZEN_PASSWORD)).resolves.toEqual({
      ...FROZEN_IDENTITY,
      kind: PREVIEW_IDENTITY_KIND,
    });
    await expect(vault.openIdentity(FROZEN_ADDRESS, FROZEN_PASSWORD)).resolves.toEqual(FROZEN_IDENTITY);
    // Normalization is in memory only: storage is never rewritten.
    expect(localStorage.getItem(IDENTITY_VAULT_STORAGE_KEY)).toBe(raw);
  });

  it('createIdentity still writes record-v1 without a kind field', async () => {
    const preview = { secret: 'a2'.repeat(32), commitment: previewMockIdentityHash('a2'.repeat(32)), nullifier: '00'.repeat(32), createdAt: 1 };
    await vault.createIdentity(ADDRESS_A, PASSWORD, preview);
    const record = JSON.parse(localStorage.getItem(IDENTITY_VAULT_STORAGE_KEY)!).records[ADDRESS_A];
    expect(Object.keys(record).sort()).toEqual(['alg', 'createdAt', 'envelope', 'iter', 'kdf', 'v']);
    expect(record.v).toBe(1);
    await expect(vault.openCustodyIdentity(ADDRESS_A, PASSWORD)).resolves.toMatchObject({ kind: PREVIEW_IDENTITY_KIND });
  });

  it('opens record-v2 only as the canonical kind under its own AAD', async () => {
    storeRecords({ [ADDRESS_A]: await canonicalRecord(ADDRESS_A) });
    await expect(vault.openCustodyIdentity(ADDRESS_A, PASSWORD)).resolves.toEqual(CANONICAL);
    await expectVaultError(vault.openIdentity(ADDRESS_A, PASSWORD), 'kind-mismatch');
    expect(canonicalIdentityAad(ADDRESS_A)).toBe(`truerepublic/identity/v2|canonical-bn254-mimc-v1|${ADDRESS_A}`);
  });

  it('rejects unknown or malformed version/kind combinations before any key derivation', async () => {
    const v2 = await canonicalRecord(ADDRESS_A);
    const v1 = { v: 1, ...RECORD_BASE, envelope: FROZEN_V1_ENVELOPE };
    const importKey = vi.spyOn(crypto.subtle, 'importKey');
    for (const record of [
      { ...v2, v: 3 },
      { ...v2, v: 0 },
      { ...v2, v: '2' },
      { ...v2, kind: PREVIEW_IDENTITY_KIND },
      { ...v2, kind: 'canonical-bn254-mimc-v2' },
      { ...v2, kind: undefined },
      { ...v1, kind: CANONICAL_IDENTITY_KIND },
      { ...v1, kind: PREVIEW_IDENTITY_KIND },
      { ...v2, iter: 100_000 },
      { ...v2, extra: true },
    ]) {
      const raw = storeRecords({ [ADDRESS_A]: JSON.parse(JSON.stringify(record)) });
      await expectVaultError(() => vault.hasIdentity(ADDRESS_A), 'corrupt');
      await expectVaultError(vault.openCustodyIdentity(ADDRESS_A, PASSWORD), 'corrupt');
      expect(localStorage.getItem(IDENTITY_VAULT_STORAGE_KEY)).toBe(raw);
    }
    expect(importKey).not.toHaveBeenCalled();
  });

  it('fails closed on swapped address, version or kind', async () => {
    const v2 = await canonicalRecord(ADDRESS_A);
    // Canonical record moved to another address.
    storeRecords({ [ADDRESS_B]: v2 });
    await expectVaultError(vault.openCustodyIdentity(ADDRESS_B, PASSWORD), 'locked');
    // Canonical envelope relabeled as record-v1 (preview).
    storeRecords({ [ADDRESS_A]: { v: 1, ...RECORD_BASE, envelope: v2.envelope } });
    await expectVaultError(vault.openCustodyIdentity(ADDRESS_A, PASSWORD), 'locked');
    // Preview envelope relabeled as record-v2 (canonical).
    storeRecords({ [FROZEN_ADDRESS]: { v: 2, kind: CANONICAL_IDENTITY_KIND, ...RECORD_BASE, envelope: FROZEN_V1_ENVELOPE } });
    await expectVaultError(vault.openCustodyIdentity(FROZEN_ADDRESS, FROZEN_PASSWORD), 'locked');
  });

  it('accepts only canonical plaintext with commitment = MiMC(secret) and no preview nullifier', async () => {
    expect(JSON.parse(serializeCanonicalIdentity(CANONICAL))).toEqual({
      kind: CANONICAL_IDENTITY_KIND,
      secret: CANONICAL.secret,
      commitment: CANONICAL.commitment,
      createdAt: CANONICAL.createdAt,
    });
    expect(serializeCanonicalIdentity(CANONICAL)).not.toContain('nullifier');
    await expect(isValidCanonicalIdentity(CANONICAL)).resolves.toBe(true);
    await expect(isValidCanonicalIdentity({ ...CANONICAL, nullifier: '00'.repeat(32) })).resolves.toBe(false);
    await expect(isValidCanonicalIdentity({ ...CANONICAL, commitment: previewMockIdentityHash(CANONICAL.secret) })).resolves.toBe(false);
    const outOfField = BN254_SCALAR_MODULUS.toString(16).padStart(64, '0');
    await expect(isValidCanonicalIdentity({ ...CANONICAL, secret: outOfField, commitment: CANONICAL.commitment })).resolves.toBe(false);

    // A sealed canonical record whose commitment does not match is corrupt after decryption.
    storeRecords({ [ADDRESS_A]: await canonicalRecord(ADDRESS_A, { ...CANONICAL, commitment: '11'.repeat(32) }) });
    await expectVaultError(vault.openCustodyIdentity(ADDRESS_A, PASSWORD), 'corrupt');
  });

  it('never overwrites a canonical record: preview create and legacy migration both refuse', async () => {
    const raw = storeRecords({ [ADDRESS_A]: await canonicalRecord(ADDRESS_A) });
    const preview = { secret: 'b4'.repeat(32), commitment: previewMockIdentityHash('b4'.repeat(32)), nullifier: previewMockIdentityHash(`${'b4'.repeat(32)}00`), createdAt: 1 };
    await expectVaultError(vault.createIdentity(ADDRESS_A, PASSWORD, preview), 'conflict');
    expect(localStorage.getItem(IDENTITY_VAULT_STORAGE_KEY)).toBe(raw);

    localStorage.setItem(LEGACY_IDENTITY_STORAGE_KEY, `{"state":{"identity":${serializeIdentity(preview)}},"version":0}`);
    const migration = new LegacyIdentityMigration();
    await expect(
      migration.migrate({ address: ADDRESS_A, password: PASSWORD, isCurrent: () => true })
    ).rejects.toMatchObject({ code: 'conflict' });
    expect(localStorage.getItem(IDENTITY_VAULT_STORAGE_KEY)).toBe(raw);
    await expect(vault.openCustodyIdentity(ADDRESS_A, PASSWORD)).resolves.toEqual(CANONICAL);
  });

  it('keeps removal explicit and lock-guarded for both record kinds', async () => {
    storeRecords({ [ADDRESS_A]: await canonicalRecord(ADDRESS_A), [FROZEN_ADDRESS]: { v: 1, ...RECORD_BASE, envelope: FROZEN_V1_ENVELOPE } });
    vi.stubGlobal('navigator', {});
    await expectVaultError(vault.removeIdentity(ADDRESS_A), 'storage');
    vi.stubGlobal('navigator', { locks: new FifoLocks() });
    await expect(vault.removeIdentity(ADDRESS_A)).resolves.toBe(true);
    expect(vault.hasIdentity(FROZEN_ADDRESS)).toBe(true);
    expect(vault.hasIdentity(ADDRESS_A)).toBe(false);
  });
});
