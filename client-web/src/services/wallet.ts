import { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import { stringToPath } from '@cosmjs/crypto';
import type { Wallet, CreateWalletParams, ImportWalletParams } from '@/types/wallet';

const DERIVATION_PATH = "m/44'/118'/0'/0/0"; // Cosmos standard
const STORAGE_KEY = 'truerepublic_wallets';

export class WalletService {
  /**
   * Create a new wallet with random mnemonic
   */
  static async createWallet(params: CreateWalletParams): Promise<Wallet> {
    const wallet = await DirectSecp256k1HdWallet.generate(24, {
      prefix: 'true',
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
    const words = params.mnemonic.trim().split(/\s+/);
    if (words.length !== 12 && words.length !== 24) {
      throw new Error('Mnemonic must be 12 or 24 words');
    }

    const wallet = await DirectSecp256k1HdWallet.fromMnemonic(
      params.mnemonic,
      {
        prefix: 'true',
        hdPaths: [stringToPath(DERIVATION_PATH)],
      }
    );

    const [account] = await wallet.getAccounts();

    const importedWallet: Wallet = {
      address: account.address,
      mnemonic: params.mnemonic,
      name: params.name,
      createdAt: Date.now(),
    };

    await this.saveWallet(importedWallet, params.password);

    return importedWallet;
  }

  /**
   * Get wallet instance for signing
   */
  static async getWalletForSigning(
    address: string,
    password: string
  ): Promise<DirectSecp256k1HdWallet> {
    const wallet = await this.getWallet(address, password);

    if (!wallet.mnemonic) {
      throw new Error('Wallet mnemonic not found');
    }

    return DirectSecp256k1HdWallet.fromMnemonic(wallet.mnemonic, {
      prefix: 'true',
      hdPaths: [stringToPath(DERIVATION_PATH)],
    });
  }

  /**
   * Save wallet encrypted
   */
  private static async saveWallet(
    wallet: Wallet,
    password: string
  ): Promise<void> {
    const wallets = this.loadWallets();

    const encrypted = await this.encrypt(wallet.mnemonic || '', password);

    const storedWallet = {
      ...wallet,
      mnemonic: encrypted,
    };

    wallets.push(storedWallet);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(wallets));
  }

  /**
   * Load all wallets (mnemonics remain encrypted)
   */
  static loadWallets(): Wallet[] {
    const stored = localStorage.getItem(STORAGE_KEY);
    return stored ? JSON.parse(stored) : [];
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

    const mnemonic = await this.decrypt(stored.mnemonic || '', password);

    return {
      ...stored,
      mnemonic,
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

    const key = await this.deriveKey(password, salt);

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

    return btoa(String.fromCharCode(...combined));
  }

  /**
   * Decrypt text using Web Crypto API (AES-GCM with PBKDF2-derived key)
   */
  private static async decrypt(encrypted: string, password: string): Promise<string> {
    const combined = new Uint8Array(
      atob(encrypted).split('').map((c) => c.charCodeAt(0))
    );

    const salt = combined.slice(0, 16);
    const iv = combined.slice(16, 28);
    const ciphertext = combined.slice(28);

    const key = await this.deriveKey(password, salt);

    const decrypted = await crypto.subtle.decrypt(
      { name: 'AES-GCM', iv: iv as ArrayBufferView<ArrayBuffer> },
      key,
      ciphertext as ArrayBufferView<ArrayBuffer>
    );

    return new TextDecoder().decode(decrypted);
  }

  /**
   * Derive AES key from password using PBKDF2
   */
  private static async deriveKey(
    password: string,
    salt: Uint8Array
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
        iterations: 100_000,
        hash: 'SHA-256',
      },
      keyMaterial,
      { name: 'AES-GCM', length: 256 },
      false,
      ['encrypt', 'decrypt']
    );
  }
}
