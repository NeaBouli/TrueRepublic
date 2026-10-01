import { toBech32 } from '@cosmjs/encoding';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Identity } from '@/types/zkp';
import {
  IdentityMigrationError,
  LEGACY_IDENTITY_STORAGE_KEY,
  LEGACY_MIGRATION_MARKER_KEY,
  LegacyIdentityMigration,
  classifyLegacyIdentity,
  type MigrationFailureCode,
  type MigrationMarker,
  type MigrationSession,
} from './identityMigration';
import { IDENTITY_VAULT_LOCK_NAME, IDENTITY_VAULT_STORAGE_KEY, IdentityVault, serializeIdentity } from './identityVault';
import { previewMockIdentityHash } from './previewIdentityHash';
import { ZKPService } from './zkp';
import { bytesToHex, mimcBn254 } from './zkpEncoding';
import { DEFAULT_CHAIN } from '@/config/chains';

// Synthetic material only; never a real identity, wallet or password.
const PASSWORD = 'synthetic-migration-password';
const ADDRESS_A = toBech32('truerepublic', new Uint8Array(20).fill(3));
const ADDRESS_B = toBech32('truerepublic', new Uint8Array(20).fill(4));
// Frozen historical preview vector, captured from the pre-extraction
// ZKPService.mockMiMCHash (main 1283a445): FNV-1a placeholder, not MiMC.
const FROZEN_SECRET = 'a1'.repeat(32);
const FROZEN_COMMITMENT = 'e10df405'.repeat(8);
const FROZEN_NULLIFIER = 'ab84528d'.repeat(8);
const FROZEN_EMPTY = '811c9dc5'.repeat(8);
const IDENTITY: Identity = {
  secret: FROZEN_SECRET,
  commitment: FROZEN_COMMITMENT,
  nullifier: FROZEN_NULLIFIER,
  createdAt: 1_700_000_000_000,
};
function previewIdentity(secret: string): Identity {
  return {
    secret,
    commitment: previewMockIdentityHash(secret),
    nullifier: previewMockIdentityHash(`${secret}00`),
    createdAt: 1_700_000_000_000,
  };
}
const OTHER_IDENTITY: Identity = previewIdentity('d4'.repeat(32));
const legacyRawFor = (identity: Identity) => `{"state":{"identity":${serializeIdentity(identity)}},"version":0}`;
const LEGACY_RAW = `{"state":{"identity":${serializeIdentity(IDENTITY)}},"version":0}`;

class FifoLocks {
  held = false;
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

/** In-memory storage that logs writes, can fail a chosen write and tracks the lock. */
class TestStorage {
  readonly data = new Map<string, string>();
  writes: string[] = [];
  writesOutsideLock = 0;
  writesAfterStale = 0;
  stale = false;
  failWriteAt: number | null = null;

  constructor(private readonly locks: FifoLocks) {}

  getItem(key: string): string | null {
    return this.data.get(key) ?? null;
  }

  setItem(key: string, value: string): void {
    this.record(key);
    this.data.set(key, value);
  }

  removeItem(key: string): void {
    this.record(key);
    this.data.delete(key);
  }

  private record(key: string): void {
    if (this.failWriteAt !== null && this.writes.length === this.failWriteAt) {
      this.failWriteAt = null;
      throw new DOMException('quota', 'QuotaExceededError');
    }
    if (!this.locks.held) this.writesOutsideLock += 1;
    if (this.stale) this.writesAfterStale += 1;
    this.writes.push(key);
  }
}

function marker(storage: TestStorage): MigrationMarker | null {
  const raw = storage.data.get(LEGACY_MIGRATION_MARKER_KEY);
  return raw ? (JSON.parse(raw) as MigrationMarker) : null;
}

async function sha256Hex(text: string): Promise<string> {
  const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(text));
  return Array.from(new Uint8Array(digest), (byte) => byte.toString(16).padStart(2, '0')).join('');
}

function session(address = ADDRESS_A, isCurrent: () => boolean = () => true): MigrationSession {
  return { address, password: PASSWORD, isCurrent };
}

async function expectFailure(promise: Promise<unknown>, code: MigrationFailureCode): Promise<void> {
  const error = await promise.then(
    () => null,
    (caught: unknown) => caught
  );
  expect(error).toBeInstanceOf(IdentityMigrationError);
  expect((error as IdentityMigrationError).code).toBe(code);
}

describe('historical preview hash', () => {
  it('matches the frozen pre-extraction vectors and backs ZKPService.generateIdentity', () => {
    expect(previewMockIdentityHash(FROZEN_SECRET)).toBe(FROZEN_COMMITMENT);
    expect(previewMockIdentityHash(`${FROZEN_SECRET}00`)).toBe(FROZEN_NULLIFIER);
    expect(previewMockIdentityHash('')).toBe(FROZEN_EMPTY);
    const generated = new ZKPService(DEFAULT_CHAIN).generateIdentity();
    expect(generated.commitment).toBe(previewMockIdentityHash(generated.secret));
    expect(generated.nullifier).toBe(previewMockIdentityHash(`${generated.secret}00`));
  });
});

describe('legacy identity classification', () => {
  it('accepts only the exact persisted schema and byte layout', () => {
    expect(classifyLegacyIdentity(null)).toEqual({ kind: 'absent' });
    expect(classifyLegacyIdentity('{"state":{"identity":null},"version":0}').kind).toBe('empty');
    expect(classifyLegacyIdentity(LEGACY_RAW)).toMatchObject({ kind: 'valid', identityBytes: serializeIdentity(IDENTITY) });

    const reordered = `{"state":{"identity":${JSON.stringify({ commitment: IDENTITY.commitment, secret: IDENTITY.secret, nullifier: IDENTITY.nullifier, createdAt: IDENTITY.createdAt })}},"version":0}`;
    const cases: Array<[string, string]> = [
      ['x'.repeat(4_097), 'oversized'],
      [`{"pad":"${'€'.repeat(1_400)}"}`, 'oversized'],
      ['not json', 'not-json'],
      ['[]', 'schema'],
      [`{"state":{"identity":${serializeIdentity(IDENTITY)}},"version":1}`, 'schema'],
      [`{"state":{"identity":${serializeIdentity(IDENTITY)},"hasIdentity":true},"version":0}`, 'schema'],
      [`{"state":{"identity":${serializeIdentity(IDENTITY)}},"version":0,"x":1}`, 'schema'],
      [`{"state":{"identity":${JSON.stringify({ ...IDENTITY, secret: 'zz' })}},"version":0}`, 'schema'],
      [`{"state":{"identity":${JSON.stringify({ ...IDENTITY, extra: 1 })}},"version":0}`, 'schema'],
      [`{"state":{"identity":${JSON.stringify({ secret: IDENTITY.secret, commitment: IDENTITY.commitment })}},"version":0}`, 'schema'],
      [reordered, 'non-canonical'],
      [` ${LEGACY_RAW}`, 'non-canonical'],
      ['{"state":{"identity":null}, "version":0}', 'non-canonical'],
    ];
    for (const [raw, reason] of cases) {
      expect(classifyLegacyIdentity(raw)).toMatchObject({ kind: 'quarantined', reason });
    }
  });

  it('quarantines records whose commitment or nullifier does not follow the preview hash', () => {
    // A real BN254-MiMC output (canonical field element input) must never pass as a preview commitment.
    const bn254Commitment = bytesToHex(mimcBn254([new Uint8Array(32).fill(1)]));
    expect(bn254Commitment).toMatch(/^[0-9a-f]{64}$/);
    for (const identity of [
      { ...IDENTITY, commitment: `${'0'}${FROZEN_COMMITMENT.slice(1)}` },
      { ...IDENTITY, nullifier: `${'0'}${FROZEN_NULLIFIER.slice(1)}` },
      { ...IDENTITY, commitment: FROZEN_NULLIFIER, nullifier: FROZEN_COMMITMENT },
      { ...IDENTITY, commitment: bn254Commitment },
    ]) {
      expect(classifyLegacyIdentity(legacyRawFor(identity))).toMatchObject({
        kind: 'quarantined',
        reason: 'preview-hash-mismatch',
      });
    }
    expect(classifyLegacyIdentity(legacyRawFor(OTHER_IDENTITY)).kind).toBe('valid');
  });
});

describe('legacy identity migration', { timeout: 120_000 }, () => {
  let locks: FifoLocks;
  let storage: TestStorage;
  let migration: LegacyIdentityMigration;

  beforeEach(() => {
    locks = new FifoLocks();
    vi.stubGlobal('navigator', { locks });
    storage = new TestStorage(locks);
    storage.data.set(LEGACY_IDENTITY_STORAGE_KEY, LEGACY_RAW);
    migration = new LegacyIdentityMigration(storage);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('migrates DETECTED -> PENDING -> WRITTEN -> VERIFIED and keeps the plaintext', async () => {
    expect(migration.inspect()).toEqual({ kind: 'legacy', marker: null });
    expect(storage.writes).toEqual([]);

    const result = await migration.migrate(session());
    expect(result).toMatchObject({
      state: 'VERIFIED',
      address: ADDRESS_A,
      reason: null,
      legacyDigest: await sha256Hex(LEGACY_RAW),
    });
    expect(storage.writes).toEqual([
      LEGACY_MIGRATION_MARKER_KEY, // DETECTED
      LEGACY_MIGRATION_MARKER_KEY, // PENDING
      IDENTITY_VAULT_STORAGE_KEY, // encrypted record
      LEGACY_MIGRATION_MARKER_KEY, // WRITTEN
      LEGACY_MIGRATION_MARKER_KEY, // VERIFIED
    ]);
    expect(storage.data.get(LEGACY_IDENTITY_STORAGE_KEY)).toBe(LEGACY_RAW);
    expect(storage.writesOutsideLock).toBe(0);
    await expect(new IdentityVault(storage).openIdentity(ADDRESS_A, PASSWORD)).resolves.toEqual(IDENTITY);
    expect(storage.data.get(IDENTITY_VAULT_STORAGE_KEY)).not.toContain(IDENTITY.secret);
  });

  it('is idempotent once VERIFIED: no further writes', async () => {
    await migration.migrate(session());
    const writes = storage.writes.length;
    await expect(migration.migrate(session())).resolves.toMatchObject({ state: 'VERIFIED' });
    expect(storage.writes.length).toBe(writes);
  });

  it('resumes to the same VERIFIED end state after a crash at every write', async () => {
    const reference = new TestStorage(locks);
    reference.data.set(LEGACY_IDENTITY_STORAGE_KEY, LEGACY_RAW);
    await new LegacyIdentityMigration(reference).migrate(session());
    const totalWrites = reference.writes.length;

    for (let crashAt = 0; crashAt < totalWrites; crashAt += 1) {
      const crashed = new TestStorage(locks);
      crashed.data.set(LEGACY_IDENTITY_STORAGE_KEY, LEGACY_RAW);
      crashed.failWriteAt = crashAt;
      await expectFailure(new LegacyIdentityMigration(crashed).migrate(session()), 'storage');
      expect(crashed.data.get(LEGACY_IDENTITY_STORAGE_KEY)).toBe(LEGACY_RAW);

      await expect(new LegacyIdentityMigration(crashed).migrate(session())).resolves.toMatchObject({
        state: 'VERIFIED',
        address: ADDRESS_A,
      });
      const vault = JSON.parse(crashed.data.get(IDENTITY_VAULT_STORAGE_KEY)!);
      expect(Object.keys(vault.records)).toEqual([ADDRESS_A]);
      await expect(new IdentityVault(crashed).openIdentity(ADDRESS_A, PASSWORD)).resolves.toEqual(IDENTITY);
      expect(crashed.data.get(LEGACY_IDENTITY_STORAGE_KEY)).toBe(LEGACY_RAW);
    }
  });

  it('never writes after the caller session changes, at any await point', async () => {
    let completedCalls = 0;
    for (let allowed = 0; ; allowed += 1) {
      const run = new TestStorage(locks);
      run.data.set(LEGACY_IDENTITY_STORAGE_KEY, LEGACY_RAW);
      let calls = 0;
      const isCurrent = () => {
        calls += 1;
        if (calls > allowed) {
          run.stale = true;
          return false;
        }
        return true;
      };
      const outcome = await new LegacyIdentityMigration(run).migrate(session(ADDRESS_A, isCurrent)).then(
        () => 'done',
        (error: IdentityMigrationError) => error.code
      );
      expect(run.writesAfterStale).toBe(0);
      expect(run.data.get(LEGACY_IDENTITY_STORAGE_KEY)).toBe(LEGACY_RAW);
      if (outcome === 'done') {
        completedCalls = calls;
        break;
      }
      expect(outcome).toBe('session-changed');
      // A later run with a valid session still completes from the retained marker.
      await expect(new LegacyIdentityMigration(run).migrate(session())).resolves.toMatchObject({ state: 'VERIFIED' });
    }
    // Every write is preceded by a session check: DETECTED, PENDING, vault, WRITTEN, VERIFIED.
    expect(completedCalls).toBe(5);
  });

  it('quarantines invalid legacy input, never deletes it and never touches the vault', async () => {
    const bad = `{"state":{"identity":${JSON.stringify({ ...IDENTITY, extra: true })}},"version":0}`;
    storage.data.set(LEGACY_IDENTITY_STORAGE_KEY, bad);
    await expect(migration.migrate(session())).resolves.toMatchObject({ state: 'QUARANTINED', reason: 'schema' });
    await expect(migration.migrate(session())).resolves.toMatchObject({ state: 'QUARANTINED' });
    expect(storage.writes).toEqual([LEGACY_MIGRATION_MARKER_KEY]);
    expect(storage.data.get(LEGACY_IDENTITY_STORAGE_KEY)).toBe(bad);
    expect(storage.data.has(IDENTITY_VAULT_STORAGE_KEY)).toBe(false);
  });

  it('quarantines a preview-hash mismatch without a vault write or plaintext deletion', async () => {
    const forged = legacyRawFor({ ...IDENTITY, nullifier: OTHER_IDENTITY.nullifier });
    storage.data.set(LEGACY_IDENTITY_STORAGE_KEY, forged);
    await expect(migration.migrate(session())).resolves.toMatchObject({
      state: 'QUARANTINED',
      reason: 'preview-hash-mismatch',
      address: null,
    });
    expect(storage.writes).toEqual([LEGACY_MIGRATION_MARKER_KEY]);
    expect(storage.data.has(IDENTITY_VAULT_STORAGE_KEY)).toBe(false);
    expect(storage.data.get(LEGACY_IDENTITY_STORAGE_KEY)).toBe(forged);
  });

  it('accepts only the legal marker state/address/reason combinations', async () => {
    const digest = await sha256Hex(LEGACY_RAW);
    const states = ['DETECTED', 'PENDING', 'WRITTEN', 'VERIFIED', 'QUARANTINED', 'ERROR'];
    const addresses = [null, ADDRESS_A, ADDRESS_A.toUpperCase()];
    const reasons = [
      null,
      'oversized',
      'not-json',
      'schema',
      'non-canonical',
      'preview-hash-mismatch',
      'vault-conflict',
      'legacy-changed',
      'verify-mismatch',
      'vault-corrupt',
      'unknown',
    ];
    const legal = new Set([
      'DETECTED|null|null',
      ...['PENDING', 'WRITTEN', 'VERIFIED'].map((state) => `${state}|${ADDRESS_A}|null`),
      ...['oversized', 'not-json', 'schema', 'non-canonical', 'preview-hash-mismatch'].map(
        (reason) => `QUARANTINED|null|${reason}`
      ),
      ...['vault-conflict', 'verify-mismatch', 'vault-corrupt', 'legacy-changed'].map(
        (reason) => `ERROR|${ADDRESS_A}|${reason}`
      ),
      'ERROR|null|legacy-changed',
    ]);
    let checked = 0;
    for (const state of states) {
      for (const address of addresses) {
        for (const reason of reasons) {
          const raw = JSON.stringify({ v: 1, state, legacyDigest: digest, address, reason, updatedAt: 1 });
          storage.data.set(LEGACY_MIGRATION_MARKER_KEY, raw);
          const key = `${state}|${address}|${reason}`;
          if (legal.has(key)) {
            expect(migration.inspect(), key).toMatchObject({ kind: 'legacy', marker: { state } });
          } else {
            expect(() => migration.inspect(), key).toThrow(IdentityMigrationError);
          }
          checked += 1;
        }
      }
    }
    expect(checked).toBe(198);
    expect(legal.size).toBe(14);

    // A contradictory but complete marker is left untouched by migrate().
    const contradictory = JSON.stringify({ v: 1, state: 'VERIFIED', legacyDigest: digest, address: null, reason: null, updatedAt: 1 });
    storage.data.set(LEGACY_MIGRATION_MARKER_KEY, contradictory);
    const writes = storage.writes.length;
    await expectFailure(migration.migrate(session()), 'corrupt-marker');
    expect(storage.data.get(LEGACY_MIGRATION_MARKER_KEY)).toBe(contradictory);
    expect(storage.writes.length).toBe(writes);
  });

  it('detect() records DETECTED or QUARANTINED without touching the vault', async () => {
    await expect(migration.detect(() => true)).resolves.toMatchObject({ kind: 'legacy', marker: { state: 'DETECTED' } });
    expect(storage.writes).toEqual([LEGACY_MIGRATION_MARKER_KEY]);
    await migration.detect(() => true);
    expect(storage.writes).toEqual([LEGACY_MIGRATION_MARKER_KEY]);
    // A stale session blocks the first write; with nothing left to write, no check is needed.
    await expect(migration.detect(() => false)).resolves.toMatchObject({ marker: { state: 'DETECTED' } });
    const fresh = new TestStorage(locks);
    fresh.data.set(LEGACY_IDENTITY_STORAGE_KEY, LEGACY_RAW);
    await expectFailure(new LegacyIdentityMigration(fresh).detect(() => false), 'session-changed');
    expect(fresh.writes).toEqual([]);
  });

  it('never overwrites a different vault identity and records the conflict as evidence', async () => {
    await new IdentityVault(storage).createIdentity(ADDRESS_A, PASSWORD, OTHER_IDENTITY);
    const vaultBefore = storage.data.get(IDENTITY_VAULT_STORAGE_KEY);
    await expect(migration.migrate(session())).resolves.toMatchObject({ state: 'ERROR', reason: 'vault-conflict' });
    expect(storage.data.get(IDENTITY_VAULT_STORAGE_KEY)).toBe(vaultBefore);
    expect(storage.data.get(LEGACY_IDENTITY_STORAGE_KEY)).toBe(LEGACY_RAW);
    await expect(new IdentityVault(storage).openIdentity(ADDRESS_A, PASSWORD)).resolves.toEqual(OTHER_IDENTITY);
  });

  it('refuses VERIFIED when the read-back identity differs from the legacy bytes', async () => {
    await new IdentityVault(storage).createIdentity(ADDRESS_A, PASSWORD, OTHER_IDENTITY);
    storage.data.set(
      LEGACY_MIGRATION_MARKER_KEY,
      JSON.stringify({ v: 1, state: 'WRITTEN', legacyDigest: await sha256Hex(LEGACY_RAW), address: ADDRESS_A, reason: null, updatedAt: 1 })
    );
    await expect(migration.migrate(session())).resolves.toMatchObject({ state: 'ERROR', reason: 'verify-mismatch' });
    expect(storage.data.get(LEGACY_IDENTITY_STORAGE_KEY)).toBe(LEGACY_RAW);
  });

  it('accepts an identical vault identity left by an interrupted run', async () => {
    await new IdentityVault(storage).createIdentity(ADDRESS_A, PASSWORD, IDENTITY);
    await expect(migration.migrate(session())).resolves.toMatchObject({ state: 'VERIFIED' });
  });

  it('retargets only while nothing was written; a WRITTEN target cannot be changed', async () => {
    storage.data.set(
      LEGACY_MIGRATION_MARKER_KEY,
      JSON.stringify({ v: 1, state: 'PENDING', legacyDigest: await sha256Hex(LEGACY_RAW), address: ADDRESS_B, reason: null, updatedAt: 1 })
    );
    await expect(migration.migrate(session(ADDRESS_A))).resolves.toMatchObject({ state: 'VERIFIED', address: ADDRESS_A });

    const before = new Map(storage.data);
    storage.data.set(
      LEGACY_MIGRATION_MARKER_KEY,
      JSON.stringify({ ...marker(storage), state: 'WRITTEN' })
    );
    const markerBefore = storage.data.get(LEGACY_MIGRATION_MARKER_KEY);
    await expectFailure(migration.migrate(session(ADDRESS_B)), 'conflict');
    expect(storage.data.get(LEGACY_MIGRATION_MARKER_KEY)).toBe(markerBefore);
    expect(storage.data.get(IDENTITY_VAULT_STORAGE_KEY)).toBe(before.get(IDENTITY_VAULT_STORAGE_KEY));
  });

  it('keeps evidence when the legacy bytes change after VERIFIED', async () => {
    await migration.migrate(session());
    const verified = marker(storage)!;
    const changed = legacyRawFor(OTHER_IDENTITY);
    storage.data.set(LEGACY_IDENTITY_STORAGE_KEY, changed);
    await expect(migration.migrate(session())).resolves.toMatchObject({
      state: 'ERROR',
      reason: 'legacy-changed',
      legacyDigest: verified.legacyDigest,
      address: ADDRESS_A,
    });
    await expect(new IdentityVault(storage).openIdentity(ADDRESS_A, PASSWORD)).resolves.toEqual(IDENTITY);
    expect(storage.data.get(LEGACY_IDENTITY_STORAGE_KEY)).toBe(changed);
  });

  it('re-detects new bytes when nothing was written for the old ones', async () => {
    await migration.detect(() => true);
    const changed = legacyRawFor(OTHER_IDENTITY);
    storage.data.set(LEGACY_IDENTITY_STORAGE_KEY, changed);
    await expect(migration.migrate(session())).resolves.toMatchObject({
      state: 'VERIFIED',
      legacyDigest: await sha256Hex(changed),
    });
    await expect(new IdentityVault(storage).openIdentity(ADDRESS_A, PASSWORD)).resolves.toEqual(OTHER_IDENTITY);
  });

  it('maps a wrong password on an existing record to locked without changing the marker', async () => {
    await new IdentityVault(storage).createIdentity(ADDRESS_A, PASSWORD, IDENTITY);
    await migration.detect(() => true);
    const markerBefore = storage.data.get(LEGACY_MIGRATION_MARKER_KEY);
    await expectFailure(migration.migrate({ ...session(), password: 'wrong-migration-password' }), 'locked');
    expect(marker(storage)).toMatchObject({ state: 'PENDING' });
    expect(markerBefore).toContain('DETECTED');
  });

  it('keeps legacy and marker evidence on storage failure and corrupt markers', async () => {
    storage.failWriteAt = 2; // the encrypted vault write
    await expectFailure(migration.migrate(session()), 'storage');
    expect(marker(storage)).toMatchObject({ state: 'PENDING', address: ADDRESS_A });
    expect(storage.data.has(IDENTITY_VAULT_STORAGE_KEY)).toBe(false);
    expect(storage.data.get(LEGACY_IDENTITY_STORAGE_KEY)).toBe(LEGACY_RAW);

    storage.data.set(LEGACY_MIGRATION_MARKER_KEY, '{"v":1,"state":"VERIFIED"}');
    const writes = storage.writes.length;
    await expectFailure(migration.migrate(session()), 'corrupt-marker');
    expect(storage.data.get(LEGACY_MIGRATION_MARKER_KEY)).toBe('{"v":1,"state":"VERIFIED"}');
    expect(storage.writes.length).toBe(writes);
  });

  it('records a corrupted vault as evidence without touching it', async () => {
    storage.data.set(IDENTITY_VAULT_STORAGE_KEY, 'not json');
    await expect(migration.migrate(session())).resolves.toMatchObject({ state: 'ERROR', reason: 'vault-corrupt' });
    expect(storage.data.get(IDENTITY_VAULT_STORAGE_KEY)).toBe('not json');
    expect(storage.data.get(LEGACY_IDENTITY_STORAGE_KEY)).toBe(LEGACY_RAW);
  });

  it('performs no write at all without Web Locks', async () => {
    vi.stubGlobal('navigator', {});
    await expectFailure(migration.migrate(session()), 'storage');
    await expectFailure(migration.detect(() => true), 'storage');
    expect(storage.writes).toEqual([]);
  });

  it('runs the whole transaction under the vault lock', async () => {
    const names: string[] = [];
    const recording = {
      request<T>(name: string, options: { mode: 'exclusive' }, callback: () => Promise<T>) {
        names.push(`${name}:${options.mode}`);
        return locks.request(name, options, callback);
      },
    };
    vi.stubGlobal('navigator', { locks: recording });
    await migration.migrate(session());
    expect(names).toEqual([`${IDENTITY_VAULT_LOCK_NAME}:exclusive`]);
    expect(storage.writesOutsideLock).toBe(0);
  });

  it('rejects invalid requests before any read or write', async () => {
    for (const bad of [
      { ...session(), address: ADDRESS_A.toUpperCase() },
      { ...session(), password: 'short' },
      { address: ADDRESS_A, password: PASSWORD } as unknown as MigrationSession,
    ]) {
      await expectFailure(migration.migrate(bad), 'invalid-request');
    }
    expect(storage.writes).toEqual([]);
  });

  it('does nothing for absent or empty legacy records', async () => {
    storage.data.delete(LEGACY_IDENTITY_STORAGE_KEY);
    await expect(migration.migrate(session())).resolves.toBeNull();
    storage.data.set(LEGACY_IDENTITY_STORAGE_KEY, '{"state":{"identity":null},"version":0}');
    await expect(migration.migrate(session())).resolves.toBeNull();
    expect(migration.inspect()).toEqual({ kind: 'none' });
    expect(storage.writes).toEqual([]);
  });
});
