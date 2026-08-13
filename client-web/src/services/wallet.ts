import type { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import { fromBech32, toBech32 } from '@cosmjs/encoding';
import { DEFAULT_CHAIN } from '@/config/chains';
import type { Wallet, CreateWalletParams, ImportWalletParams } from '@/types/wallet';

const DERIVATION_PATH = "m/44'/118'/0'/0/0"; // Cosmos standard
const STORAGE_KEY = 'truerepublic_wallets';
const LEGACY_BECH32_PREFIX = 'true';

// Service-layer bounds mirror the form validators in utils/validation.ts so
// direct service callers fail closed even when the UI checks were bypassed.
const PASSWORD_MIN_LENGTH = 8;
const WALLET_NAME_MAX_LENGTH = 50;
const ENCRYPTION_VERSION = 'v2';
const CURRENT_PBKDF2_ITERATIONS = 600_000;
const LEGACY_PBKDF2_ITERATIONS = 100_000;
// salt (16) + iv (12) + minimum AES-GCM tag (16)
const MIN_ENCRYPTED_PAYLOAD_BYTES = 44;

async function loadSigningDependencies() {
  const [
    { DirectSecp256k1HdWallet },
    { EnglishMnemonic, stringToPath },
  ] = await Promise.all([
    import('@cosmjs/proto-signing'),
    import('@cosmjs/crypto'),
  ]);
  return { DirectSecp256k1HdWallet, EnglishMnemonic, stringToPath };
}

/**
 * Normalize a caller-supplied mnemonic to its canonical BIP-39 form: NFKC,
 * lowercase, single-space separated. Throws on non-string or empty input so
 * malformed import material fails closed before any key derivation.
 */
function normalizeMnemonic(mnemonic: string): string {
  if (typeof mnemonic !== 'string') {
    throw new Error('Recovery phrase is required');
  }
  const words = mnemonic
    .normalize('NFKC')
    .trim()
    .toLowerCase()
    .split(/\s+/)
    .filter((word) => word.length > 0);
  if (words.length === 0) {
    throw new Error('Recovery phrase is required');
  }
  return words.join(' ');
}

function assertWalletName(name: string): void {
  if (typeof name !== 'string' || name.trim().length === 0) {
    throw new Error('Wallet name is required');
  }
  if (name.length > WALLET_NAME_MAX_LENGTH) {
    throw new Error(`Wallet name must be at most ${WALLET_NAME_MAX_LENGTH} characters`);
  }
}

function assertEncryptionPassword(password: string): void {
  if (typeof password !== 'string' || password.length < PASSWORD_MIN_LENGTH) {
    throw new Error(
      `Encryption password must be at least ${PASSWORD_MIN_LENGTH} characters`
    );
  }
}

function isStoredWalletRecord(value: unknown): value is Wallet {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) {
    return false;
  }
  const record = value as Record<string, unknown>;
  return (
    typeof record.address === 'string' &&
    typeof record.name === 'string' &&
    typeof record.createdAt === 'number' &&
    typeof record.mnemonic === 'string'
  );
}

function migrateStoredAddress(wallet: Wallet): Wallet {
  try {
    const decoded = fromBech32(wallet.address);
    if (decoded.prefix !== LEGACY_BECH32_PREFIX) return wallet;
    return {
      ...wallet,
      address: toBech32(DEFAULT_CHAIN.bech32Prefix, decoded.data),
    };
  } catch {
    // Preserve malformed historical records so users can still explicitly
    // delete them; address migration must never destroy encrypted custody data.
    return wallet;
  }
}

export class WalletService {
  private static sessionGeneration = 0;
  private static activeSigningAddress: string | null = null;

  static activateSigningSession(address: string): void {
    this.sessionGeneration += 1;
    this.activeSigningAddress = address;
  }

  static invalidateSigningSession(): void {
    this.sessionGeneration += 1;
    this.activeSigningAddress = null;
  }

  private static assertSigningSession(address: string, generation: number): void {
    if (
      this.activeSigningAddress !== address ||
      this.sessionGeneration !== generation
    ) {
      throw new Error('Wallet signing session is no longer active');
    }
  }

  /**
   * Create a new wallet with random mnemonic
   */
  static async createWallet(params: CreateWalletParams): Promise<Wallet> {
    assertWalletName(params.name);
    assertEncryptionPassword(params.password);

    const { DirectSecp256k1HdWallet, stringToPath } = await loadSigningDependencies();
    const wallet = await DirectSecp256k1HdWallet.generate(24, {
      prefix: DEFAULT_CHAIN.bech32Prefix,
      hdPaths: [stringToPath(DERIVATION_PATH)],
    });

    const [account] = await wallet.getAccounts();
    const mnemonic = wallet.mnemonic;

    const newWallet: Wallet = {
      address: account.address,
      mnemonic,
      name: params.name,
      createdAt: Date.now(),
    };

    await this.saveWallet(newWallet, params.password);

    return newWallet;
  }

  /**
   * Import wallet from mnemonic
   */
  static async importWallet(params: ImportWalletParams): Promise<Wallet> {
    assertWalletName(params.name);
    assertEncryptionPassword(params.password);

    const mnemonic = normalizeMnemonic(params.mnemonic);
    const words = mnemonic.split(' ');
    if (words.length !== 12 && words.length !== 24) {
      throw new Error('Mnemonic must be 12 or 24 words');
    }

    const { DirectSecp256k1HdWallet, EnglishMnemonic, stringToPath } =
      await loadSigningDependencies();

    // Fail closed with bounded messages: wordlist membership first, then the
    // BIP-39 checksum. The raw library errors are deliberately not surfaced —
    // they can embed the full 2048-word list into UI/log output.
    const wordlist = new Set<string>(EnglishMnemonic.wordlist);
    if (words.some((word) => !wordlist.has(word))) {
      throw new Error('Recovery phrase contains a word outside the BIP-39 English wordlist');
    }
    try {
      // The constructor runs mnemonicToEntropy, which enforces the checksum.
      void new EnglishMnemonic(mnemonic);
    } catch {
      throw new Error('Recovery phrase checksum is invalid');
    }

    const wallet = await DirectSecp256k1HdWallet.fromMnemonic(
      mnemonic,
      {
        prefix: DEFAULT_CHAIN.bech32Prefix,
        hdPaths: [stringToPath(DERIVATION_PATH)],
      }
    );

    const [account] = await wallet.getAccounts();

    const importedWallet: Wallet = {
      address: account.address,
      mnemonic,
      name: params.name,
      createdAt: Date.now(),
    };

    await this.saveWallet(importedWallet, params.password);

    return importedWallet;
  }

  /**
   * Get wallet instance for signing. The derived account must match the
   * requested address exactly; a custody record whose decrypted mnemonic
   * derives a different account (tampered or corrupted storage) is rejected
   * instead of silently signing for the wrong wallet.
   */
  static async getWalletForSigning(
    address: string,
    password: string
  ): Promise<DirectSecp256k1HdWallet> {
    const generation = this.sessionGeneration;
    this.assertSigningSession(address, generation);
    const wallet = await this.getWallet(address, password);
    this.assertSigningSession(address, generation);

    if (!wallet.mnemonic) {
      throw new Error('Wallet mnemonic not found');
    }

    const { DirectSecp256k1HdWallet, stringToPath } = await loadSigningDependencies();
    const signer = await DirectSecp256k1HdWallet.fromMnemonic(wallet.mnemonic, {
      prefix: DEFAULT_CHAIN.bech32Prefix,
      hdPaths: [stringToPath(DERIVATION_PATH)],
    });

    const [account] = await signer.getAccounts();
    if (!account || account.address !== address) {
      throw new Error(
        'Stored wallet data does not match the requested address'
      );
    }
    this.assertSigningSession(address, generation);

    // CosmJS checks for signDirect structurally. The proxy keeps that exact
    // signer surface while rejecting both account reads and signature return
    // after lock/switch/delete invalidates the captured session generation.
    return new Proxy(signer, {
      get: (target, property) => {
        if (property === 'getAccounts') {
          return async () => {
            this.assertSigningSession(address, generation);
            const accounts = await target.getAccounts();
            this.assertSigningSession(address, generation);
            return accounts;
          };
        }
        if (property === 'signDirect') {
          return async (...args: Parameters<typeof target.signDirect>) => {
            this.assertSigningSession(address, generation);
            const result = await target.signDirect(...args);
            this.assertSigningSession(address, generation);
            return result;
          };
        }
        if (property === 'mnemonic') {
          this.assertSigningSession(address, generation);
        }
        const value = Reflect.get(target, property, target) as unknown;
        return typeof value === 'function' ? value.bind(target) : value;
      },
    });
  }

  /**
   * Save wallet encrypted
   */
  private static async saveWallet(
    wallet: Wallet,
    password: string
  ): Promise<void> {
    if (!wallet.mnemonic) {
      throw new Error('Refusing to persist a wallet without secret material');
    }
    const encrypted = await this.encrypt(wallet.mnemonic, password);

    // Web Crypto yields control. Read only after it completes, then keep this
    // check/merge/write sequence synchronous so concurrent saves cannot write
    // stale snapshots over one another.
    const wallets = this.loadWallets();
    if (wallets.some((stored) => stored.address === wallet.address)) {
      throw new Error('A wallet with this address already exists');
    }

    const storedWallet = {
      ...wallet,
      mnemonic: encrypted,
    };

    wallets.push(storedWallet);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(wallets));
  }

  /**
   * Load all wallets (mnemonics remain encrypted). Malformed storage rejects
   * as a whole and remains untouched, so corrupted entries are never silently
   * destroyed or acted upon.
   */
  static loadWallets(): Wallet[] {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (!stored) return [];

    let parsed: unknown;
    try {
      parsed = JSON.parse(stored);
    } catch {
      throw new Error('Stored wallet data is not valid JSON');
    }
    if (!Array.isArray(parsed)) {
      throw new Error('Stored wallet data is not a wallet list');
    }

    if (parsed.some((value) => !isStoredWalletRecord(value))) {
      throw new Error('Stored wallet data contains an invalid wallet record');
    }

    let migratedAny = false;
    const wallets = parsed.map((value) => {
      // The structural check above narrows every entry.
      const wallet = value as Wallet;
      const migrated = migrateStoredAddress(wallet);
      if (migrated !== wallet) migratedAny = true;
      return migrated;
    });
    if (migratedAny) {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(wallets));
    }
    return wallets;
  }

  /**
   * Get wallet by address (decrypts mnemonic)
   */
  static async getWallet(
    address: string,
    password: string
  ): Promise<Wallet> {
    const wallets = this.loadWallets();
    const stored = wallets.find((w) => w.address === address);

    if (!stored) {
      throw new Error('Wallet not found');
    }

    const decrypted = await this.decrypt(stored.mnemonic || '', password);
    if (decrypted.needsUpgrade) {
      const encrypted = await this.encrypt(decrypted.plaintext, password);
      const upgraded = wallets.map((wallet) =>
        wallet.address === address ? { ...wallet, mnemonic: encrypted } : wallet
      );
      localStorage.setItem(STORAGE_KEY, JSON.stringify(upgraded));
    }

    return {
      ...stored,
      mnemonic: decrypted.plaintext,
    };
  }

  /**
   * Delete wallet
   */
  static deleteWallet(address: string): void {
    const wallets = this.loadWallets();
    const filtered = wallets.filter((w) => w.address !== address);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(filtered));
  }

  /**
   * Encrypt text using Web Crypto API (AES-GCM with PBKDF2-derived key)
   */
  private static async encrypt(text: string, password: string): Promise<string> {
    const encoder = new TextEncoder();
    const salt = crypto.getRandomValues(new Uint8Array(16));
    const iv = crypto.getRandomValues(new Uint8Array(12));

    const key = await this.deriveKey(password, salt, CURRENT_PBKDF2_ITERATIONS);

    const plaintext = encoder.encode(text);
    const encrypted = await crypto.subtle.encrypt(
      { name: 'AES-GCM', iv: iv as ArrayBufferView<ArrayBuffer> },
      key,
      plaintext as ArrayBufferView<ArrayBuffer>
    );

    // Combine salt + iv + ciphertext
    const combined = new Uint8Array(salt.length + iv.length + encrypted.byteLength);
    combined.set(salt, 0);
    combined.set(iv, salt.length);
    combined.set(new Uint8Array(encrypted), salt.length + iv.length);

    return `${ENCRYPTION_VERSION}:${btoa(String.fromCharCode(...combined))}`;
  }

  /**
   * Decrypt text using Web Crypto API (AES-GCM with PBKDF2-derived key).
   * Malformed payloads and authentication failures (wrong password or
   * tampered ciphertext) all surface as one bounded error; the raw Web
   * Crypto failure is never propagated into UI or logs.
   */
  private static async decrypt(
    encrypted: string,
    password: string
  ): Promise<{ plaintext: string; needsUpgrade: boolean }> {
    const failure = 'Incorrect password or corrupted wallet data';
    const currentPrefix = `${ENCRYPTION_VERSION}:`;
    const isCurrent = encrypted.startsWith(currentPrefix);
    const payload = isCurrent ? encrypted.slice(currentPrefix.length) : encrypted;

    let combined: Uint8Array;
    try {
      combined = new Uint8Array(
        atob(payload).split('').map((c) => c.charCodeAt(0))
      );
    } catch {
      throw new Error(failure);
    }
    if (combined.length < MIN_ENCRYPTED_PAYLOAD_BYTES) {
      throw new Error(failure);
    }

    const salt = combined.slice(0, 16);
    const iv = combined.slice(16, 28);
    const ciphertext = combined.slice(28);

    try {
      const key = await this.deriveKey(
        password,
        salt,
        isCurrent ? CURRENT_PBKDF2_ITERATIONS : LEGACY_PBKDF2_ITERATIONS
      );

      const decrypted = await crypto.subtle.decrypt(
        { name: 'AES-GCM', iv: iv as ArrayBufferView<ArrayBuffer> },
        key,
        ciphertext as ArrayBufferView<ArrayBuffer>
      );

      return {
        plaintext: new TextDecoder().decode(decrypted),
        needsUpgrade: !isCurrent,
      };
    } catch {
      throw new Error(failure);
    }
  }

  /**
   * Derive AES key from password using PBKDF2
   */
  private static async deriveKey(
    password: string,
    salt: Uint8Array,
    iterations: number
  ): Promise<CryptoKey> {
    const encoder = new TextEncoder();
    const passwordBytes = encoder.encode(password);
    const keyMaterial = await crypto.subtle.importKey(
      'raw',
      passwordBytes as ArrayBufferView<ArrayBuffer>,
      'PBKDF2',
      false,
      ['deriveKey']
    );

    return crypto.subtle.deriveKey(
      {
        name: 'PBKDF2',
        salt: salt as ArrayBufferView<ArrayBuffer>,
        iterations,
        hash: 'SHA-256',
      },
      keyMaterial,
      { name: 'AES-GCM', length: 256 },
      false,
      ['encrypt', 'decrypt']
    );
  }
}
