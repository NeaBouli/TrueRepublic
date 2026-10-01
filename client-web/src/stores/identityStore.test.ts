import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { Identity } from '@/types/zkp';
import { IDENTITY_VAULT_STORAGE_KEY, IdentityVault, serializeIdentity } from '@/services/identityVault';
import { LEGACY_IDENTITY_STORAGE_KEY, LEGACY_MIGRATION_MARKER_KEY } from '@/services/identityMigration';
import { previewMockIdentityHash } from '@/services/previewIdentityHash';
import { MembershipService } from '@/services/membership';
import { ZKPService } from '@/services/zkp';
import { useIdentityStore } from './identityStore';
import { WALLET_DELETE_BLOCKED_BY_IDENTITY, currentIdentitySession, useWalletStore } from './walletStore';

// Unit tests stay offline: balance refresh after unlock is stubbed.
vi.mock('@/services/blockchain', () => ({
  BlockchainService: vi.fn(() => ({ getBalance: vi.fn().mockResolvedValue([]) })),
}));

// Synthetic material only; never a real wallet, identity or password.
const PASSWORD = 'synthetic-lifecycle-password';
const SECRET = 'e7'.repeat(32);
const IDENTITY: Identity = {
  secret: SECRET,
  commitment: previewMockIdentityHash(SECRET),
  nullifier: previewMockIdentityHash(`${SECRET}00`),
  createdAt: 1_700_000_000_000,
};
const LEGACY_RAW = `{"state":{"identity":${serializeIdentity(IDENTITY)}},"version":0}`;

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

async function settled(): Promise<void> {
  for (let i = 0; i < 200 && useIdentityStore.getState().status === 'loading'; i += 1) {
    await new Promise((resolve) => setTimeout(resolve, 10));
  }
}

async function createUnlockedWallet(name = 'Lifecycle'): Promise<string> {
  const wallet = await useWalletStore.getState().createWallet(name, PASSWORD);
  await settled();
  return wallet.address;
}

function allStorageValues(): string {
  return Array.from({ length: localStorage.length }, (_, i) => localStorage.key(i)!)
    .filter((key) => key !== LEGACY_IDENTITY_STORAGE_KEY)
    .map((key) => `${key}=${localStorage.getItem(key)}`)
    .join('\n');
}

describe('identity custody lifecycle', { timeout: 120_000 }, () => {
  const clipboard = { writeText: vi.fn() };

  beforeEach(() => {
    localStorage.clear();
    vi.stubGlobal('navigator', { locks: new FifoLocks(), clipboard });
    useWalletStore.getState().lock();
    useWalletStore.setState({ wallets: [] });
    clipboard.writeText.mockClear();
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it('starts locked and never hydrates the legacy plaintext record', () => {
    localStorage.setItem(LEGACY_IDENTITY_STORAGE_KEY, LEGACY_RAW);
    const state = useIdentityStore.getState();
    expect(state.status).toBe('locked');
    expect(state.identity).toBeNull();
    expect(state.hasIdentity).toBe(false);
    // No persist middleware: the legacy value is neither read into state nor rewritten.
    useIdentityStore.getState().invalidate();
    expect(localStorage.getItem(LEGACY_IDENTITY_STORAGE_KEY)).toBe(LEGACY_RAW);
  });

  it('shows absent after unlock without vault or legacy data', async () => {
    await createUnlockedWallet();
    expect(useIdentityStore.getState()).toMatchObject({ status: 'absent', identity: null, legacyPresent: false });
  });

  it('detects legacy data, migrates only on explicit request and keeps the plaintext', async () => {
    localStorage.setItem(LEGACY_IDENTITY_STORAGE_KEY, LEGACY_RAW);
    const address = await createUnlockedWallet();
    expect(useIdentityStore.getState()).toMatchObject({ status: 'legacy-pending', legacyPresent: true });
    expect(localStorage.getItem(IDENTITY_VAULT_STORAGE_KEY)).toBeNull();

    await useIdentityStore.getState().migrateLegacy(currentIdentitySession()!);
    expect(useIdentityStore.getState()).toMatchObject({ status: 'ready', hasIdentity: true, identity: IDENTITY });
    expect(localStorage.getItem(LEGACY_IDENTITY_STORAGE_KEY)).toBe(LEGACY_RAW);
    expect(JSON.parse(localStorage.getItem(LEGACY_MIGRATION_MARKER_KEY)!)).toMatchObject({ state: 'VERIFIED', address });
    expect(allStorageValues()).not.toContain(SECRET);
  });

  it('opens an existing vault identity after unlock, memory only', async () => {
    const address = await createUnlockedWallet();
    await new IdentityVault().createIdentity(address, PASSWORD, IDENTITY);
    useWalletStore.getState().lock();
    await useWalletStore.getState().unlock(PASSWORD);
    await settled();
    expect(useIdentityStore.getState()).toMatchObject({ status: 'ready', identity: IDENTITY });
    expect(allStorageValues()).not.toContain(SECRET);
  });

  it('maps quarantined, corrupt vault and missing Web Locks to bounded states', async () => {
    localStorage.setItem(LEGACY_IDENTITY_STORAGE_KEY, LEGACY_RAW.replace(IDENTITY.nullifier, 'f'.repeat(64)));
    await createUnlockedWallet('Quarantine');
    expect(useIdentityStore.getState()).toMatchObject({ status: 'quarantined', problem: 'preview-hash-mismatch' });

    useWalletStore.getState().lock();
    localStorage.clear();
    localStorage.setItem(IDENTITY_VAULT_STORAGE_KEY, 'not json');
    await createUnlockedWallet('Corrupt');
    expect(useIdentityStore.getState()).toMatchObject({ status: 'error', problem: 'vault-corrupt' });
    expect(useWalletStore.getState().isLocked).toBe(false);

    useWalletStore.getState().lock();
    localStorage.clear();
    localStorage.setItem(LEGACY_IDENTITY_STORAGE_KEY, LEGACY_RAW);
    vi.stubGlobal('navigator', { clipboard });
    await createUnlockedWallet('NoLocks');
    expect(useIdentityStore.getState()).toMatchObject({ status: 'error', problem: 'storage' });
    expect(localStorage.getItem(LEGACY_MIGRATION_MARKER_KEY)).toBeNull();
  });

  it('reports a wrong password for a record sealed with another password, wallet stays unlocked', async () => {
    const address = await createUnlockedWallet();
    await new IdentityVault().createIdentity(address, 'different-synthetic-password', IDENTITY);
    useWalletStore.getState().lock();
    await useWalletStore.getState().unlock(PASSWORD);
    await settled();
    expect(useIdentityStore.getState()).toMatchObject({ status: 'error', problem: 'wrong-password', identity: null });
    expect(useWalletStore.getState().isLocked).toBe(false);
  });

  it('invalidates synchronously on lock, switch, create, import and current-wallet delete', async () => {
    const address = await createUnlockedWallet();
    await new IdentityVault().createIdentity(address, PASSWORD, IDENTITY);
    const ready = async () => {
      useWalletStore.getState().lock();
      await useWalletStore.getState().unlock(PASSWORD);
      await settled();
      expect(useIdentityStore.getState().status).toBe('ready');
    };

    await ready();
    useWalletStore.getState().lock();
    expect(useIdentityStore.getState()).toMatchObject({ status: 'locked', identity: null });

    await ready();
    const switching = useWalletStore.getState().switchWallet(address, PASSWORD);
    expect(useIdentityStore.getState().identity).toBeNull();
    await switching;
    await settled();

    await ready();
    const creating = useWalletStore.getState().createWallet('Second', PASSWORD);
    expect(useIdentityStore.getState().identity).toBeNull();
    await creating;
    await settled();

    await ready();
    const importing = useWalletStore
      .getState()
      .importWallet(
        'Imported',
        'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about',
        PASSWORD
      );
    expect(useIdentityStore.getState().identity).toBeNull();
    const imported = await importing;
    await settled();

    // Deleting the current wallet (no vault record for it) also invalidates synchronously.
    useIdentityStore.setState({ identity: IDENTITY, hasIdentity: true, status: 'ready' });
    useWalletStore.getState().deleteWallet(imported.address);
    expect(useIdentityStore.getState()).toMatchObject({ status: 'locked', identity: null });
  });

  it('drops stale completions after a wallet transition', async () => {
    const address = await createUnlockedWallet();
    await new IdentityVault().createIdentity(address, PASSWORD, IDENTITY);
    useWalletStore.getState().lock();
    const unlocking = useWalletStore.getState().unlock(PASSWORD);
    await unlocking;
    // The vault open is now in flight; lock before it completes.
    expect(useIdentityStore.getState().status).toBe('loading');
    useWalletStore.getState().lock();
    await new Promise((resolve) => setTimeout(resolve, 1_500));
    expect(useIdentityStore.getState()).toMatchObject({ status: 'locked', identity: null });

    // A stale migrate request with a superseded session does nothing.
    const staleSession = { address, password: PASSWORD, isCurrent: () => false };
    await useIdentityStore.getState().migrateLegacy(staleSession);
    await useIdentityStore.getState().load(staleSession);
    expect(useIdentityStore.getState()).toMatchObject({ status: 'locked', identity: null });
  });

  it('refuses wallet deletion while a vault record exists or the vault is unreadable', async () => {
    const address = await createUnlockedWallet();
    await new IdentityVault().createIdentity(address, PASSWORD, IDENTITY);
    expect(() => useWalletStore.getState().deleteWallet(address)).toThrow(WALLET_DELETE_BLOCKED_BY_IDENTITY);
    expect(useWalletStore.getState().wallets.map((wallet) => wallet.address)).toContain(address);
    await expect(new IdentityVault().openIdentity(address, PASSWORD)).resolves.toEqual(IDENTITY);

    localStorage.setItem(IDENTITY_VAULT_STORAGE_KEY, 'not json');
    expect(() => useWalletStore.getState().deleteWallet(address)).toThrow(WALLET_DELETE_BLOCKED_BY_IDENTITY);
    expect(localStorage.getItem(IDENTITY_VAULT_STORAGE_KEY)).toBe('not json');

    localStorage.removeItem(IDENTITY_VAULT_STORAGE_KEY);
    useWalletStore.getState().deleteWallet(address);
    expect(useWalletStore.getState().wallets.map((wallet) => wallet.address)).not.toContain(address);
  });

  it('exports only on request, as files with warnings, and never generates, registers or copies', async () => {
    const generate = vi.spyOn(ZKPService.prototype, 'generateIdentity');
    const register = vi.spyOn(MembershipService.prototype, 'registerIdentity');
    expect(useIdentityStore.getState().exportIdentityFile()).toBeNull();
    expect(useIdentityStore.getState().exportLegacyFile()).toBeNull();

    localStorage.setItem(LEGACY_IDENTITY_STORAGE_KEY, LEGACY_RAW);
    await createUnlockedWallet();
    expect(useIdentityStore.getState().exportLegacyFile()).toMatchObject({ content: LEGACY_RAW, mimeType: 'text/plain' });
    await useIdentityStore.getState().migrateLegacy(currentIdentitySession()!);
    const file = useIdentityStore.getState().exportIdentityFile()!;
    const parsed = JSON.parse(file.content);
    expect(parsed).toMatchObject({ format: 'truerepublic-preview-identity', version: 1, canonical: false, identity: IDENTITY });
    expect(parsed.warning).toMatch(/not canonical and not production-valid/);

    useWalletStore.getState().lock();
    expect(useIdentityStore.getState().exportIdentityFile()).toBeNull();
    expect(generate).not.toHaveBeenCalled();
    expect(register).not.toHaveBeenCalled();
    expect(clipboard.writeText).not.toHaveBeenCalled();
  });
});
