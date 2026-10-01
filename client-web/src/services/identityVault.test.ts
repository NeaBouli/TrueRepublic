import { toBech32 } from '@cosmjs/encoding';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { Identity } from '@/types/zkp';
import {
  IDENTITY_VAULT_STORAGE_KEY,
  IdentityVault,
  IdentityVaultError,
  MAX_IDENTITY_VAULT_RECORDS,
  identityAad,
  serializeIdentity,
  type IdentityVaultErrorCode,
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
    vault = new IdentityVault();
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
    const pending = vault.createIdentity(ADDRESS_A, PASSWORD, IDENTITY);
    const foreign = JSON.stringify({ v: 1, records: {} });
    localStorage.setItem(IDENTITY_VAULT_STORAGE_KEY, foreign);
    await expectVaultError(pending, 'conflict');
    expect(rawVault()).toBe(foreign);
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
      await expectVaultError(() => vault.removeIdentity(ADDRESS_A), 'corrupt');
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
    await expectVaultError(new IdentityVault(failingSet).createIdentity(ADDRESS_A, PASSWORD, IDENTITY), 'storage');
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
    expect(vault.removeIdentity(ADDRESS_A)).toBe(true);
    expect(vault.removeIdentity(ADDRESS_A)).toBe(false);
    expect(vault.hasIdentity(ADDRESS_B)).toBe(true);
    expect(vault.removeIdentity(ADDRESS_B)).toBe(true);
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
