import { toBech32 } from '@cosmjs/encoding';
import { beforeEach, describe, expect, it } from 'vitest';
import { WalletService } from './wallet';

const STORAGE_KEY = 'truerepublic_wallets';

// BIP-39 test vectors (zero entropy): valid wordlist and checksum.
const MNEMONIC_12 =
  'abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about';
const MNEMONIC_24 = `${'abandon '.repeat(23)}art`;
const PASSWORD = 'correct horse battery staple';

async function legacyEncrypt(plaintext: string, password: string): Promise<string> {
  const salt = crypto.getRandomValues(new Uint8Array(16));
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const material = await crypto.subtle.importKey(
    'raw',
    new TextEncoder().encode(password),
    'PBKDF2',
    false,
    ['deriveKey']
  );
  const key = await crypto.subtle.deriveKey(
    { name: 'PBKDF2', salt, iterations: 100_000, hash: 'SHA-256' },
    material,
    { name: 'AES-GCM', length: 256 },
    false,
    ['encrypt']
  );
  const ciphertext = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv },
    key,
    new TextEncoder().encode(plaintext)
  );
  const combined = new Uint8Array(28 + ciphertext.byteLength);
  combined.set(salt, 0);
  combined.set(iv, 16);
  combined.set(new Uint8Array(ciphertext), 28);
  return btoa(String.fromCharCode(...combined));
}

describe('wallet address migration', () => {
  beforeEach(() => localStorage.clear());

  it('rewrites the retired true prefix without changing encrypted data', () => {
    const data = Uint8Array.from({ length: 20 }, (_, index) => index + 1);
    const stored = {
      address: toBech32('true', data),
      mnemonic: 'encrypted-wallet-payload',
      name: 'Legacy wallet',
      createdAt: 1,
    };
    localStorage.setItem(STORAGE_KEY, JSON.stringify([stored]));

    const [wallet] = WalletService.loadWallets();

    expect(wallet).toEqual({
      ...stored,
      address: toBech32('truerepublic', data),
    });
    expect(JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '[]')).toEqual([
      wallet,
    ]);
  });

  it('preserves canonical and malformed historical records', () => {
    const canonical = {
      address: toBech32('truerepublic', new Uint8Array(20).fill(7)),
      mnemonic: 'encrypted-one',
      name: 'Canonical',
      createdAt: 2,
    };
    const malformed = {
      address: 'not-an-address',
      mnemonic: 'encrypted-two',
      name: 'Malformed',
      createdAt: 3,
    };
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify([canonical, malformed])
    );

    expect(WalletService.loadWallets()).toEqual([canonical, malformed]);
  });
});

describe('wallet creation and import hardening', () => {
  beforeEach(() => {
    localStorage.clear();
    WalletService.invalidateSigningSession();
  });

  it('persists only encrypted wallet material at rest', async () => {
    const wallet = await WalletService.createWallet({
      name: 'Main',
      password: PASSWORD,
    });

    const raw = localStorage.getItem(STORAGE_KEY) ?? '';
    expect(wallet.mnemonic).toBeDefined();
    expect(raw).not.toContain(wallet.mnemonic!);

    const [stored] = JSON.parse(raw) as { mnemonic: string }[];
    expect(stored.mnemonic).not.toBe(wallet.mnemonic);
    expect(stored.mnemonic.startsWith('v2:')).toBe(true);
    // salt + iv + ciphertext + tag, base64 encoded
    expect(stored.mnemonic.length).toBeGreaterThan(58);
  });

  it('round-trips an imported wallet through encryption with the right password', async () => {
    const imported = await WalletService.importWallet({
      name: 'Imported',
      mnemonic: MNEMONIC_12,
      password: PASSWORD,
    });

    const decrypted = await WalletService.getWallet(imported.address, PASSWORD);
    expect(decrypted.mnemonic).toBe(MNEMONIC_12);
  });

  it('reads and transparently upgrades the legacy 100k PBKDF2 payload', async () => {
    const imported = await WalletService.importWallet({
      name: 'Legacy encrypted',
      mnemonic: MNEMONIC_12,
      password: PASSWORD,
    });
    const [stored] = WalletService.loadWallets();
    const legacy = await legacyEncrypt(MNEMONIC_12, PASSWORD);
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify([{ ...stored, mnemonic: legacy }])
    );

    await expect(
      WalletService.getWallet(imported.address, PASSWORD)
    ).resolves.toMatchObject({ mnemonic: MNEMONIC_12 });
    expect(WalletService.loadWallets()[0].mnemonic?.startsWith('v2:')).toBe(
      true
    );
  });

  it('accepts a valid 24-word mnemonic', async () => {
    const imported = await WalletService.importWallet({
      name: 'Long phrase',
      mnemonic: MNEMONIC_24,
      password: PASSWORD,
    });

    expect(imported.address.startsWith('truerepublic1')).toBe(true);
  });

  it('normalizes case, unicode, and whitespace before deriving', async () => {
    const canonical = await WalletService.importWallet({
      name: 'Canonical',
      mnemonic: MNEMONIC_12,
      password: PASSWORD,
    });
    localStorage.clear();

    const messy = await WalletService.importWallet({
      name: 'Messy',
      mnemonic: `  ${MNEMONIC_12.toUpperCase().split(' ').join('   ')}\n`,
      password: PASSWORD,
    });

    expect(messy.address).toBe(canonical.address);
    expect(messy.mnemonic).toBe(MNEMONIC_12);
  });

  it.each([
    ['11 words', 'abandon '.repeat(11).trim()],
    ['13 words', `${MNEMONIC_12} abandon`],
    ['23 words', 'abandon '.repeat(23).trim()],
  ])('rejects a mnemonic with %s', async (_label, mnemonic) => {
    await expect(
      WalletService.importWallet({ name: 'Bad', mnemonic, password: PASSWORD })
    ).rejects.toThrow('Mnemonic must be 12 or 24 words');
    expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
  });

  it('rejects an empty or non-string mnemonic', async () => {
    await expect(
      WalletService.importWallet({ name: 'Bad', mnemonic: '   ', password: PASSWORD })
    ).rejects.toThrow('Recovery phrase is required');
    await expect(
      WalletService.importWallet({
        name: 'Bad',
        mnemonic: undefined as unknown as string,
        password: PASSWORD,
      })
    ).rejects.toThrow('Recovery phrase is required');
    expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
  });

  it('rejects a word outside the BIP-39 wordlist with a bounded message', async () => {
    const mnemonic = MNEMONIC_12.replace('about', 'zzzz');
    await expect(
      WalletService.importWallet({ name: 'Bad', mnemonic, password: PASSWORD })
    ).rejects.toThrow('wordlist');
    expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
  });

  it('rejects a mnemonic with an invalid checksum', async () => {
    const mnemonic = 'abandon '.repeat(12).trim();
    await expect(
      WalletService.importWallet({ name: 'Bad', mnemonic, password: PASSWORD })
    ).rejects.toThrow('checksum');
    expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
  });

  it('rejects weak encryption passwords at the service layer', async () => {
    await expect(
      WalletService.importWallet({
        name: 'Bad',
        mnemonic: MNEMONIC_12,
        password: 'short',
      })
    ).rejects.toThrow('at least 8 characters');
    await expect(
      WalletService.createWallet({ name: 'Bad', password: '' })
    ).rejects.toThrow('at least 8 characters');
    expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
  });

  it('rejects empty and overlong wallet names', async () => {
    await expect(
      WalletService.importWallet({
        name: '   ',
        mnemonic: MNEMONIC_12,
        password: PASSWORD,
      })
    ).rejects.toThrow('Wallet name is required');
    await expect(
      WalletService.createWallet({ name: 'x'.repeat(51), password: PASSWORD })
    ).rejects.toThrow('at most 50 characters');
    expect(localStorage.getItem(STORAGE_KEY)).toBeNull();
  });

  it('rejects importing a wallet address that already exists', async () => {
    await WalletService.importWallet({
      name: 'First',
      mnemonic: MNEMONIC_12,
      password: PASSWORD,
    });

    await expect(
      WalletService.importWallet({
        name: 'Second',
        mnemonic: MNEMONIC_12,
        password: 'a different password',
      })
    ).rejects.toThrow('already exists');

    expect(WalletService.loadWallets()).toHaveLength(1);
  });

  it('merges concurrent saves without dropping a completed wallet', async () => {
    const results = await Promise.all([
      WalletService.importWallet({
        name: 'Twelve words',
        mnemonic: MNEMONIC_12,
        password: PASSWORD,
      }),
      WalletService.importWallet({
        name: 'Twenty-four words',
        mnemonic: MNEMONIC_24,
        password: PASSWORD,
      }),
    ]);

    expect(new Set(results.map((wallet) => wallet.address)).size).toBe(2);
    expect(WalletService.loadWallets().map((wallet) => wallet.address).sort()).toEqual(
      results.map((wallet) => wallet.address).sort()
    );
  });

  it('fails closed with a bounded error on a wrong password', async () => {
    const imported = await WalletService.importWallet({
      name: 'Locked',
      mnemonic: MNEMONIC_12,
      password: PASSWORD,
    });

    await expect(
      WalletService.getWallet(imported.address, 'wrong password')
    ).rejects.toThrow('Incorrect password or corrupted wallet data');
  });

  it('fails closed on a corrupted stored payload', async () => {
    const imported = await WalletService.importWallet({
      name: 'Corruptible',
      mnemonic: MNEMONIC_12,
      password: PASSWORD,
    });

    const [stored] = WalletService.loadWallets();
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify([{ ...stored, mnemonic: '!!!not-base64!!!' }])
    );
    await expect(
      WalletService.getWallet(imported.address, PASSWORD)
    ).rejects.toThrow('Incorrect password or corrupted wallet data');

    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify([{ ...stored, mnemonic: btoa('too-short') }])
    );
    await expect(
      WalletService.getWallet(imported.address, PASSWORD)
    ).rejects.toThrow('Incorrect password or corrupted wallet data');
  });

  it('derives the signer account from the unlocked wallet', async () => {
    const imported = await WalletService.importWallet({
      name: 'Signer',
      mnemonic: MNEMONIC_12,
      password: PASSWORD,
    });

    WalletService.activateSigningSession(imported.address);
    const signer = await WalletService.getWalletForSigning(
      imported.address,
      PASSWORD
    );
    const [account] = await signer.getAccounts();
    expect(account.address).toBe(imported.address);
  });

  it('rejects signing when the stored record derives a different account', async () => {
    const imported = await WalletService.importWallet({
      name: 'Tampered',
      mnemonic: MNEMONIC_12,
      password: PASSWORD,
    });

    // Tamper: keep the encrypted mnemonic but claim a different address.
    const tamperedAddress = toBech32('truerepublic', new Uint8Array(20).fill(9));
    const [stored] = WalletService.loadWallets();
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify([{ ...stored, address: tamperedAddress }])
    );
    WalletService.activateSigningSession(tamperedAddress);

    await expect(
      WalletService.getWalletForSigning(tamperedAddress, PASSWORD)
    ).rejects.toThrow('does not match the requested address');
    expect(tamperedAddress).not.toBe(imported.address);
  });

  it('fails closed on structurally invalid records without destroying them', async () => {
    await WalletService.importWallet({
      name: 'Kept',
      mnemonic: MNEMONIC_12,
      password: PASSWORD,
    });
    const valid = WalletService.loadWallets()[0];
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify([valid, { address: 42 }, 'garbage', null])
    );

    expect(() => WalletService.loadWallets()).toThrow(
      'invalid wallet record'
    );
    expect(
      (JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '[]') as unknown[])
    ).toHaveLength(4);
  });

  it('invalidates an already derived signer on lock or wallet switch', async () => {
    const imported = await WalletService.importWallet({
      name: 'Session-bound',
      mnemonic: MNEMONIC_12,
      password: PASSWORD,
    });
    WalletService.activateSigningSession(imported.address);
    const signer = await WalletService.getWalletForSigning(
      imported.address,
      PASSWORD
    );

    WalletService.invalidateSigningSession();

    await expect(signer.getAccounts()).rejects.toThrow(
      'signing session is no longer active'
    );
    expect(() => signer.mnemonic).toThrow(
      'signing session is no longer active'
    );
  });

  it('fails closed when storage does not contain a wallet list', async () => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ address: 'x' }));
    expect(() => WalletService.loadWallets()).toThrow(
      'Stored wallet data is not a wallet list'
    );
  });

  it('fails closed with a bounded error when storage is not valid JSON', () => {
    localStorage.setItem(STORAGE_KEY, '{"mnemonic":"sensitive-fragment"');

    expect(() => WalletService.loadWallets()).toThrow(
      'Stored wallet data is not valid JSON'
    );
  });
});
