import { toBech32 } from '@cosmjs/encoding';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { Identity } from '@/types/zkp';
import {
  IDENTITY_VAULT_LOCK_NAME,
  IDENTITY_VAULT_STORAGE_KEY,
  IdentityVault,
  IdentityVaultError,
  MAX_IDENTITY_VAULT_RECORDS,
  identityAad,
  serializeIdentity,
  type IdentityVaultErrorCode,
  type IdentityVaultLocks,
  type IdentityVaultStorage,
} from './identityVault';
import { WalletService } from './wallet';

// Synthetic material only; never a real identity, wallet or password.
const PASSWORD = 'synthetic-vault-password';
const ADDRESS_A = toBech32('truerepublic', new Uint8Array(20).fill(1));
const ADDRESS_B = toBech32('truerepublic', new Uint8Array(20).fill(2));
const IDENTITY: Identity = {
  secret: 'a1'.repeat(32),
  commitment: 'b2'.repeat(32),
  nullifier: 'c3'.repeat(32),
  createdAt: 1_700_000_000_000,
};

function rawVault(): string | null {
  return localStorage.getItem(IDENTITY_VAULT_STORAGE_KEY);
}

function vaultDocument(): { v: number; records: Record<string, Record<string, unknown>> } {
  return JSON.parse(rawVault() ?? '{}');
}

async function expectVaultError(
  action: (() => unknown) | Promise<unknown>,
  code: IdentityVaultErrorCode
): Promise<void> {
  let error: unknown = null;
  try {
    await (typeof action === 'function' ? action() : action);
  } catch (caught) {
    error = caught;
  }
  expect(error).toBeInstanceOf(IdentityVaultError);
  expect((error as IdentityVaultError).code).toBe(code);
}

function bytesToBase64(text: string): string {
  return btoa(text);
}

describe('identity vault core', { timeout: 60_000 }, () => {
  let vault: IdentityVault;

  beforeEach(() => {
    localStorage.clear();
    // happy-dom has no functional Web Locks; the FIFO model stands in for navigator.locks.
    vault = new IdentityVault(undefined, new FifoLocks());
  });

  it('round-trips an identity to the exact canonical bytes for the bound address', async () => {
    expect(vault.hasIdentity(ADDRESS_A)).toBe(false);
    await expect(vault.openIdentity(ADDRESS_A, PASSWORD)).resolves.toBeNull();

    await vault.createIdentity(ADDRESS_A, PASSWORD, IDENTITY);
    expect(vault.hasIdentity(ADDRESS_A)).toBe(true);
    expect(vault.hasIdentity(ADDRESS_B)).toBe(false);

    const opened = await vault.openIdentity(ADDRESS_A, PASSWORD);
    expect(opened).toEqual(IDENTITY);
    expect(serializeIdentity(opened!)).toBe(serializeIdentity(IDENTITY));
  });

  it('persists no plaintext, password, decrypted identity or AAD', async () => {
    await vault.createIdentity(ADDRESS_A, PASSWORD, IDENTITY);
    const raw = rawVault() ?? '';
    for (const forbidden of [
      IDENTITY.secret,
      IDENTITY.commitment,
      IDENTITY.nullifier,
      PASSWORD,
      identityAad(ADDRESS_A),
      bytesToBase64(serializeIdentity(IDENTITY)),
      bytesToBase64(IDENTITY.secret),
      'secret',
    ]) {
      expect(raw).not.toContain(forbidden);
    }
    const document = vaultDocument();
    expect(Object.keys(document).sort()).toEqual(['records', 'v']);
    expect(Object.keys(document.records)).toEqual([ADDRESS_A]);
    expect(Object.keys(document.records[ADDRESS_A]).sort()).toEqual([
      'alg',
      'createdAt',
      'envelope',
      'iter',
      'kdf',
      'v',
    ]);
    expect(document.records[ADDRESS_A]).toMatchObject({
      v: 1,
      alg: 'AES-GCM-256',
      kdf: 'PBKDF2-SHA256',
      iter: 600_000,
    });
  });

  it('never overwrites an existing entry and mutates nothing on conflict', async () => {
    await vault.createIdentity(ADDRESS_A, PASSWORD, IDENTITY);
    const before = rawVault();
    await expectVaultError(
      vault.createIdentity(ADDRESS_A, PASSWORD, { ...IDENTITY, secret: 'd4'.repeat(32) }),
      'conflict'
    );
    expect(rawVault()).toBe(before);
    await expect(vault.openIdentity(ADDRESS_A, PASSWORD)).resolves.toEqual(IDENTITY);
  });

  it('refuses to write when the vault changed during key derivation', async () => {
    const storage = new HookedStorage();
    const foreign = JSON.stringify({ v: 1, records: {} });
    // After the initial read, another writer outside the lock changes the value
    // while the envelope is being sealed.
    storage.onRead = (read) => {
      if (read === 1) setTimeout(() => storage.data.set(IDENTITY_VAULT_STORAGE_KEY, foreign), 0);
    };
    await expectVaultError(new IdentityVault(storage, new FifoLocks()).createIdentity(ADDRESS_A, PASSWORD, IDENTITY), 'conflict');
    expect(storage.data.get(IDENTITY_VAULT_STORAGE_KEY)).toBe(foreign);
  });

  it('uses a fresh salt and IV per record, so equal inputs never share ciphertext', async () => {
    await vault.createIdentity(ADDRESS_A, PASSWORD, IDENTITY);
    await vault.createIdentity(ADDRESS_B, PASSWORD, IDENTITY);
    const { records } = vaultDocument();
    const a = atob(String(records[ADDRESS_A].envelope).slice(3));
    const b = atob(String(records[ADDRESS_B].envelope).slice(3));
    expect(a.slice(0, 16)).not.toBe(b.slice(0, 16));
    expect(a.slice(16, 28)).not.toBe(b.slice(16, 28));
    expect(a.slice(28)).not.toBe(b.slice(28));
  });

  it('fails closed for a wrong password, a swapped map entry and tampered ciphertext', async () => {
    await vault.createIdentity(ADDRESS_A, PASSWORD, IDENTITY);
    await expectVaultError(vault.openIdentity(ADDRESS_A, 'wrong-vault-password'), 'locked');

    const document = vaultDocument();
    localStorage.setItem(
      IDENTITY_VAULT_STORAGE_KEY,
      JSON.stringify({ v: 1, records: { [ADDRESS_B]: document.records[ADDRESS_A] } })
    );
    await expectVaultError(vault.openIdentity(ADDRESS_B, PASSWORD), 'locked');

    const envelope = String(document.records[ADDRESS_A].envelope);
    const bytes = Uint8Array.from(atob(envelope.slice(3)), (c) => c.charCodeAt(0));
    bytes[bytes.length - 1] ^= 0x01;
    const tampered = `v2:${btoa(String.fromCharCode(...bytes))}`;
    localStorage.setItem(
      IDENTITY_VAULT_STORAGE_KEY,
      JSON.stringify({ v: 1, records: { [ADDRESS_A]: { ...document.records[ADDRESS_A], envelope: tampered } } })
    );
    await expectVaultError(vault.openIdentity(ADDRESS_A, PASSWORD), 'locked');
  });

  it('rejects malformed, unsupported and oversized vault data without mutating it', async () => {
    await vault.createIdentity(ADDRESS_A, PASSWORD, IDENTITY);
    const record = vaultDocument().records[ADDRESS_A];
    const manyRecords = Object.fromEntries(
      Array.from({ length: MAX_IDENTITY_VAULT_RECORDS + 1 }, (_, i) => [
        toBech32('truerepublic', new Uint8Array(20).fill(i + 10)),
        record,
      ])
    );
    const corrupt: string[] = [
      'not json',
      JSON.stringify([]),
      JSON.stringify({ v: 2, records: {} }),
      JSON.stringify({ v: 1, records: {}, extra: true }),
      JSON.stringify({ v: 1, records: [] }),
      JSON.stringify({ v: 1, records: { [ADDRESS_A]: { ...record, alg: 'AES-CBC' } } }),
      JSON.stringify({ v: 1, records: { [ADDRESS_A]: { ...record, kdf: 'scrypt' } } }),
      JSON.stringify({ v: 1, records: { [ADDRESS_A]: { ...record, iter: 100_000 } } }),
      JSON.stringify({ v: 1, records: { [ADDRESS_A]: { ...record, v: 2 } } }),
      JSON.stringify({ v: 1, records: { [ADDRESS_A]: { ...record, aad: 'trusted' } } }),
      JSON.stringify({ v: 1, records: { [ADDRESS_A]: { ...record, envelope: `v2:${'A'.repeat(2_000)}` } } }),
      JSON.stringify({ v: 1, records: { [ADDRESS_A]: { ...record, envelope: 'AAAA' } } }),
      JSON.stringify({ v: 1, records: { [ADDRESS_A.toUpperCase()]: record } }),
      JSON.stringify({ v: 1, records: { [toBech32('cosmos', new Uint8Array(20).fill(1))]: record } }),
      JSON.stringify({ v: 1, records: manyRecords }),
      `{"v":1,"records":{},"pad":"${'x'.repeat(70_000)}"}`,
    ];
    const importKey = vi.spyOn(crypto.subtle, 'importKey');
    for (const raw of corrupt) {
      localStorage.setItem(IDENTITY_VAULT_STORAGE_KEY, raw);
      await expectVaultError(() => vault.hasIdentity(ADDRESS_A), 'corrupt');
      await expectVaultError(vault.openIdentity(ADDRESS_A, PASSWORD), 'corrupt');
      await expectVaultError(vault.createIdentity(ADDRESS_B, PASSWORD, IDENTITY), 'corrupt');
      await expectVaultError(vault.removeIdentity(ADDRESS_A), 'corrupt');
      expect(rawVault()).toBe(raw);
    }
    expect(importKey).not.toHaveBeenCalled();
    importKey.mockRestore();
  });

  it('rejects noncanonical addresses, invalid identities and short passwords before any work', async () => {
    const importKey = vi.spyOn(crypto.subtle, 'importKey');
    for (const address of [
      '',
      ADDRESS_A.toUpperCase(),
      toBech32('cosmos', new Uint8Array(20).fill(1)),
      toBech32('truerepublic', new Uint8Array(32).fill(1)),
      `${ADDRESS_A}x`,
    ]) {
      await expectVaultError(() => vault.hasIdentity(address), 'invalid-address');
      await expectVaultError(vault.createIdentity(address, PASSWORD, IDENTITY), 'invalid-address');
    }
    for (const identity of [
      { ...IDENTITY, secret: 'A1'.repeat(32) },
      { ...IDENTITY, secret: 'a1'.repeat(31) },
      { ...IDENTITY, commitment: 'zz'.repeat(32) },
      { ...IDENTITY, createdAt: -1 },
      { ...IDENTITY, extra: true },
      { secret: IDENTITY.secret, commitment: IDENTITY.commitment, createdAt: 1 },
    ]) {
      await expectVaultError(
        vault.createIdentity(ADDRESS_A, PASSWORD, identity as unknown as Identity),
        'invalid-identity'
      );
    }
    await expectVaultError(vault.createIdentity(ADDRESS_A, 'short', IDENTITY), 'invalid-password');
    await expectVaultError(vault.openIdentity(ADDRESS_A, 'short'), 'invalid-password');
    expect(importKey).not.toHaveBeenCalled();
    expect(rawVault()).toBeNull();
    importKey.mockRestore();
  });

  it('maps quota and security storage failures to a bounded error and writes nothing', async () => {
    const backing = new Map<string, string>();
    const failingSet: IdentityVaultStorage = {
      getItem: (key) => backing.get(key) ?? null,
      setItem: () => {
        throw new DOMException('quota', 'QuotaExceededError');
      },
      removeItem: (key) => void backing.delete(key),
    };
    await expectVaultError(new IdentityVault(failingSet, new FifoLocks()).createIdentity(ADDRESS_A, PASSWORD, IDENTITY), 'storage');
    expect(backing.size).toBe(0);

    const blocked: IdentityVaultStorage = {
      getItem: () => {
        throw new DOMException('denied', 'SecurityError');
      },
      setItem: () => undefined,
      removeItem: () => undefined,
    };
    await expectVaultError(() => new IdentityVault(blocked).hasIdentity(ADDRESS_A), 'storage');
  });

  it('removes only on explicit request and drops the key with the last record', async () => {
    await vault.createIdentity(ADDRESS_A, PASSWORD, IDENTITY);
    await vault.createIdentity(ADDRESS_B, PASSWORD, IDENTITY);
    await expect(vault.removeIdentity(ADDRESS_A)).resolves.toBe(true);
    await expect(vault.removeIdentity(ADDRESS_A)).resolves.toBe(false);
    expect(vault.hasIdentity(ADDRESS_B)).toBe(true);
    await expect(vault.removeIdentity(ADDRESS_B)).resolves.toBe(true);
    expect(rawVault()).toBeNull();
  });

  it('does not affect wallet unlock when the vault is corrupted', async () => {
    const wallet = await WalletService.importWallet({
      name: 'Vault independent',
      mnemonic:
        'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about',
      password: PASSWORD,
    });
    localStorage.setItem(IDENTITY_VAULT_STORAGE_KEY, 'not json');
    await expect(WalletService.getWallet(wallet.address, PASSWORD)).resolves.toMatchObject({
      address: wallet.address,
    });
    await expectVaultError(vault.openIdentity(wallet.address, PASSWORD), 'corrupt');
  });

});

// Deterministic model of the cross-tab race (GH309B2B1). Two vault instances
// share one storage. A storage hook fires at the creator's final re-read,
// i.e. the former re-read-to-write window, and runs another tab's commit
// there through the same lock manager. FifoLocks behaves like Web Locks
// (exclusive, FIFO; a free lock is granted at once); NoLocks never
// serializes and models the pre-GH309B2B1 behaviour as a negative control.
class FifoLocks implements IdentityVaultLocks {
  held = false;
  names: string[] = [];
  private queue: Array<() => void> = [];

  // Exclusive and FIFO like Web Locks. A free lock is granted synchronously:
  // the strictest adversary, because an unlocked writer then runs inside any
  // window the vault leaves open.
  request<T>(name: string, options: { mode: 'exclusive' }, callback: () => Promise<T>): Promise<T> {
    this.names.push(`${name}:${options.mode}`);
    return new Promise<T>((resolve, reject) => {
      const grant = () => {
        this.held = true;
        let result: Promise<T>;
        try {
          result = Promise.resolve(callback());
        } catch (error) {
          result = Promise.reject(error);
        }
        result.then(resolve, reject).finally(() => {
          this.held = false;
          this.queue.shift()?.();
        });
      };
      if (this.held) this.queue.push(grant);
      else grant();
    });
  }
}

class NoLocks implements IdentityVaultLocks {
  request<T>(_name: string, _options: { mode: 'exclusive' }, callback: () => Promise<T>): Promise<T> {
    return callback();
  }
}

class HookedStorage implements IdentityVaultStorage {
  readonly data = new Map<string, string>();
  reads = 0;
  onRead: ((read: number) => void) | null = null;
  writesOutsideLock = 0;

  constructor(private readonly locks?: FifoLocks) {}

  getItem(key: string): string | null {
    const value = this.data.get(key) ?? null;
    this.reads += 1;
    this.onRead?.(this.reads);
    return value;
  }

  setItem(key: string, value: string): void {
    if (this.locks && !this.locks.held) this.writesOutsideLock += 1;
    this.data.set(key, value);
  }

  removeItem(key: string): void {
    if (this.locks && !this.locks.held) this.writesOutsideLock += 1;
    this.data.delete(key);
  }
}

function recordsIn(storage: HookedStorage): string[] {
  const raw = storage.data.get(IDENTITY_VAULT_STORAGE_KEY);
  return raw ? Object.keys(JSON.parse(raw).records).sort() : [];
}

// Another tab committing B through the same locked read-modify-write path.
// Its storage reads must not re-trigger the creator's hook.
function commitInOtherTab(locks: IdentityVaultLocks, storage: HookedStorage, record: Record<string, unknown>) {
  return locks.request(IDENTITY_VAULT_LOCK_NAME, { mode: 'exclusive' }, async () => {
    const raw = storage.data.get(IDENTITY_VAULT_STORAGE_KEY);
    const document = raw ? JSON.parse(raw) : { v: 1, records: {} };
    document.records[ADDRESS_B] = record;
    storage.setItem(IDENTITY_VAULT_STORAGE_KEY, JSON.stringify(document));
  });
}

describe('identity vault cross-tab serialization (GH309B2B1)', { timeout: 60_000 }, () => {
  async function sealedRecordFor(address: string): Promise<Record<string, unknown>> {
    const scratch = new HookedStorage();
    await new IdentityVault(scratch, new FifoLocks()).createIdentity(address, PASSWORD, IDENTITY);
    return JSON.parse(scratch.data.get(IDENTITY_VAULT_STORAGE_KEY)!).records[address];
  }

  it('uses navigator.locks by default and fails closed when it is missing', async () => {
    localStorage.clear();
    const locks = new FifoLocks();
    vi.stubGlobal('navigator', { locks });
    try {
      await new IdentityVault().createIdentity(ADDRESS_A, PASSWORD, IDENTITY);
      expect(locks.names).toEqual([`${IDENTITY_VAULT_LOCK_NAME}:exclusive`]);
      vi.stubGlobal('navigator', {});
      await expectVaultError(new IdentityVault().removeIdentity(ADDRESS_A), 'storage');
      expect(new IdentityVault().hasIdentity(ADDRESS_A)).toBe(true);
    } finally {
      vi.unstubAllGlobals();
      localStorage.clear();
    }
  });

  it('fails closed without Web Locks and writes nothing', async () => {
    const storage = new HookedStorage();
    const vault = new IdentityVault(storage, null);
    await expectVaultError(vault.createIdentity(ADDRESS_A, PASSWORD, IDENTITY), 'storage');
    await expectVaultError(vault.removeIdentity(ADDRESS_A), 'storage');
    expect(storage.data.size).toBe(0);
    expect(vault.hasIdentity(ADDRESS_A)).toBe(false);
  });

  it('runs every mutation under the named exclusive lock and never writes outside it', async () => {
    const locks = new FifoLocks();
    const storage = new HookedStorage(locks);
    const vault = new IdentityVault(storage, locks);
    await vault.createIdentity(ADDRESS_A, PASSWORD, IDENTITY);
    await vault.removeIdentity(ADDRESS_A);
    expect(locks.names).toEqual([
      `${IDENTITY_VAULT_LOCK_NAME}:exclusive`,
      `${IDENTITY_VAULT_LOCK_NAME}:exclusive`,
    ]);
    expect(storage.writesOutsideLock).toBe(0);
  });

  it('create/create: a commit in the re-read-to-write window cannot be lost', async () => {
    const recordB = await sealedRecordFor(ADDRESS_B);
    const run = async (locks: IdentityVaultLocks & Partial<FifoLocks>) => {
      const storage = new HookedStorage();
      let other: Promise<void> | null = null;
      // Read 1 = creator's initial read, read 2 = its final re-read.
      storage.onRead = (read) => {
        if (read === 2 && !other) other = commitInOtherTab(locks, storage, recordB);
      };
      const created = new IdentityVault(storage, locks).createIdentity(ADDRESS_A, PASSWORD, IDENTITY);
      const outcome = await created.then(
        () => 'ok',
        (error: IdentityVaultError) => error.code
      );
      await other;
      return { outcome, records: recordsIn(storage) };
    };

    await expect(run(new FifoLocks())).resolves.toEqual({
      outcome: 'ok',
      records: [ADDRESS_A, ADDRESS_B].sort(),
    });
    // Negative control: without serialization the same interleaving loses B.
    await expect(run(new NoLocks())).resolves.toEqual({ outcome: 'ok', records: [ADDRESS_A] });
  });

  it('create/remove: a removal in the re-read-to-write window cannot be resurrected', async () => {
    const recordB = await sealedRecordFor(ADDRESS_B);
    const run = async (locks: IdentityVaultLocks) => {
      const storage = new HookedStorage();
      storage.data.set(IDENTITY_VAULT_STORAGE_KEY, JSON.stringify({ v: 1, records: { [ADDRESS_B]: recordB } }));
      const otherTab = new IdentityVault(storage, locks);
      let removal: Promise<boolean> | null = null;
      storage.onRead = (read) => {
        if (read === 2 && !removal) {
          storage.onRead = null;
          removal = otherTab.removeIdentity(ADDRESS_B);
        }
      };
      await new IdentityVault(storage, locks).createIdentity(ADDRESS_A, PASSWORD, IDENTITY);
      const removed = await removal;
      return { removed, records: recordsIn(storage) };
    };

    await expect(run(new FifoLocks())).resolves.toEqual({ removed: true, records: [ADDRESS_A] });
    // Negative control: without serialization the removed B comes back.
    await expect(run(new NoLocks())).resolves.toEqual({
      removed: true,
      records: [ADDRESS_A, ADDRESS_B].sort(),
    });
  });

  it('concurrent creates from two tabs keep both records', async () => {
    const locks = new FifoLocks();
    const storage = new HookedStorage(locks);
    await Promise.all([
      new IdentityVault(storage, locks).createIdentity(ADDRESS_A, PASSWORD, IDENTITY),
      new IdentityVault(storage, locks).createIdentity(ADDRESS_B, PASSWORD, IDENTITY),
    ]);
    expect(recordsIn(storage)).toEqual([ADDRESS_A, ADDRESS_B].sort());
    expect(storage.writesOutsideLock).toBe(0);
    await expect(new IdentityVault(storage, locks).openIdentity(ADDRESS_A, PASSWORD)).resolves.toEqual(IDENTITY);
    await expect(new IdentityVault(storage, locks).openIdentity(ADDRESS_B, PASSWORD)).resolves.toEqual(IDENTITY);
  });
});
