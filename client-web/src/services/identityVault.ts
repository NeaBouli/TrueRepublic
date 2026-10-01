import { fromBech32, toBech32 } from '@cosmjs/encoding';
import { DEFAULT_CHAIN } from '@/config/chains';
import type { Identity } from '@/types/zkp';
import {
  CURRENT_PBKDF2_ITERATIONS,
  CUSTODY_ENVELOPE_VERSION,
  openEnvelope,
  sealEnvelope,
} from './custodyEnvelope';

/**
 * Encrypted identity vault core (issue #309, GH309B2B).
 *
 * Storage: one localStorage value under IDENTITY_VAULT_STORAGE_KEY holding a
 * versioned, bounded map from canonical wallet address to an AES-GCM custody
 * envelope. The AAD is derived only from the caller's expected address and is
 * never stored, so a record copied under another address cannot be opened.
 * Plaintext, passwords and decrypted identities are never persisted.
 *
 * This module is the core only: no migration, store hydration, wallet
 * lifecycle, UI or identity generation lives here. Every failure is one
 * bounded IdentityVaultError, so callers can keep a wallet unlocked with the
 * identity absent when the vault is unusable.
 */

export const IDENTITY_VAULT_STORAGE_KEY = 'truerepublic_identity_vault_v1';
export const IDENTITY_VAULT_VERSION = 1;
export const MAX_IDENTITY_VAULT_RECORDS = 16;
export const MAX_IDENTITY_VAULT_CHARS = 64 * 1024;
/** Every vault mutation in every tab runs under this exclusive Web Lock. */
export const IDENTITY_VAULT_LOCK_NAME = 'truerepublic-identity-vault';
const MAX_RECORD_ENVELOPE_CHARS = 1_024;
const MAX_ADDRESS_CHARS = 128;
const PASSWORD_MIN_LENGTH = 8; // mirrors the wallet encryption password policy
const ADDRESS_DATA_BYTES = 20;
const RECORD_ALG = 'AES-GCM-256';
const RECORD_KDF = 'PBKDF2-SHA256';
const HEX_32_BYTES = /^[0-9a-f]{64}$/;
const RECORD_KEYS = ['alg', 'createdAt', 'envelope', 'iter', 'kdf', 'v'];
const DOCUMENT_KEYS = ['records', 'v'];
const IDENTITY_KEYS = ['commitment', 'createdAt', 'nullifier', 'secret'];

export type IdentityVaultErrorCode =
  | 'invalid-address'
  | 'invalid-identity'
  | 'invalid-password'
  | 'corrupt'
  | 'conflict'
  | 'locked'
  | 'storage';

const ERROR_MESSAGES: Record<IdentityVaultErrorCode, string> = {
  'invalid-address': 'Identity vault requires a canonical wallet address',
  'invalid-identity': 'Identity data is invalid',
  'invalid-password': 'Identity vault requires the wallet password',
  corrupt: 'Identity vault data is corrupted or unsupported',
  conflict: 'An identity is already stored for this wallet or the vault changed',
  locked: 'Incorrect password or corrupted identity data',
  storage: 'Identity vault storage is unavailable',
};

export class IdentityVaultError extends Error {
  readonly code: IdentityVaultErrorCode;

  constructor(code: IdentityVaultErrorCode) {
    super(ERROR_MESSAGES[code]);
    this.name = 'IdentityVaultError';
    this.code = code;
  }
}

export type IdentityVaultStorage = Pick<Storage, 'getItem' | 'setItem' | 'removeItem'>;

/**
 * Operations available only inside IdentityVault.runExclusive, i.e. while the
 * vault Web Lock is held. The handle is invalid once the callback settles.
 */
export interface IdentityVaultTransaction {
  hasIdentity(address: string): boolean;
  openIdentity(address: string, password: string): Promise<Identity | null>;
  /** Create-if-absent; `beforeWrite` runs after sealing, immediately before the write, and may throw to abort. */
  createIdentity(
    address: string,
    password: string,
    identity: Identity,
    beforeWrite?: () => void
  ): Promise<void>;
}

// The subset of the Web Locks API the vault needs. Deliberately not exported
// and not injectable: production mutations only ever use navigator.locks.
interface WebLocks {
  request<T>(name: string, options: { mode: 'exclusive' }, callback: () => Promise<T>): Promise<T>;
}

interface VaultRecord {
  v: typeof IDENTITY_VAULT_VERSION;
  alg: typeof RECORD_ALG;
  kdf: typeof RECORD_KDF;
  iter: typeof CURRENT_PBKDF2_ITERATIONS;
  envelope: string;
  createdAt: number;
}

interface VaultDocument {
  v: typeof IDENTITY_VAULT_VERSION;
  records: Record<string, VaultRecord>;
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

function hasExactKeys(value: Record<string, unknown>, keys: string[]): boolean {
  const actual = Object.keys(value).sort();
  return actual.length === keys.length && actual.every((key, index) => key === keys[index]);
}

export function isCanonicalAddress(address: unknown): address is string {
  if (typeof address !== 'string' || address.length === 0 || address.length > MAX_ADDRESS_CHARS) {
    return false;
  }
  try {
    const decoded = fromBech32(address);
    return (
      decoded.prefix === DEFAULT_CHAIN.bech32Prefix &&
      decoded.data.length === ADDRESS_DATA_BYTES &&
      toBech32(decoded.prefix, decoded.data) === address
    );
  } catch {
    return false;
  }
}

function assertAddress(address: string): void {
  if (!isCanonicalAddress(address)) throw new IdentityVaultError('invalid-address');
}

function assertPassword(password: string): void {
  if (typeof password !== 'string' || password.length < PASSWORD_MIN_LENGTH) {
    throw new IdentityVaultError('invalid-password');
  }
}

export function isValidIdentity(value: unknown): value is Identity {
  return (
    isPlainObject(value) &&
    hasExactKeys(value, IDENTITY_KEYS) &&
    typeof value.secret === 'string' &&
    HEX_32_BYTES.test(value.secret) &&
    typeof value.commitment === 'string' &&
    HEX_32_BYTES.test(value.commitment) &&
    typeof value.nullifier === 'string' &&
    HEX_32_BYTES.test(value.nullifier) &&
    Number.isSafeInteger(value.createdAt) &&
    (value.createdAt as number) >= 0
  );
}

/** Canonical, field-ordered serialization; decrypting yields exactly these bytes. */
export function serializeIdentity(identity: Identity): string {
  return JSON.stringify({
    secret: identity.secret,
    commitment: identity.commitment,
    nullifier: identity.nullifier,
    createdAt: identity.createdAt,
  });
}

export function identityAad(address: string): string {
  return `truerepublic/identity/v1|${address}`;
}

function isValidRecord(value: unknown): value is VaultRecord {
  return (
    isPlainObject(value) &&
    hasExactKeys(value, RECORD_KEYS) &&
    value.v === IDENTITY_VAULT_VERSION &&
    value.alg === RECORD_ALG &&
    value.kdf === RECORD_KDF &&
    value.iter === CURRENT_PBKDF2_ITERATIONS &&
    typeof value.envelope === 'string' &&
    value.envelope.length <= MAX_RECORD_ENVELOPE_CHARS &&
    value.envelope.startsWith(`${CUSTODY_ENVELOPE_VERSION}:`) &&
    Number.isSafeInteger(value.createdAt) &&
    (value.createdAt as number) >= 0
  );
}

function parseDocument(raw: string | null): VaultDocument {
  if (raw === null) return { v: IDENTITY_VAULT_VERSION, records: {} };
  if (typeof raw !== 'string' || raw.length > MAX_IDENTITY_VAULT_CHARS) {
    throw new IdentityVaultError('corrupt');
  }
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    throw new IdentityVaultError('corrupt');
  }
  if (
    !isPlainObject(parsed) ||
    !hasExactKeys(parsed, DOCUMENT_KEYS) ||
    parsed.v !== IDENTITY_VAULT_VERSION ||
    !isPlainObject(parsed.records)
  ) {
    throw new IdentityVaultError('corrupt');
  }
  const entries = Object.entries(parsed.records);
  if (entries.length > MAX_IDENTITY_VAULT_RECORDS) throw new IdentityVaultError('corrupt');
  for (const [address, record] of entries) {
    if (!isCanonicalAddress(address) || !isValidRecord(record)) {
      throw new IdentityVaultError('corrupt');
    }
  }
  return parsed as unknown as VaultDocument;
}

export class IdentityVault {
  constructor(private readonly storageOverride?: IdentityVaultStorage) {}

  /**
   * Serialize a mutation across tabs. Without Web Locks no mutation runs, so
   * concurrent read-modify-write cycles can never lose or resurrect records.
   */
  private async withMutationLock<T>(mutation: () => Promise<T>): Promise<T> {
    const locks = globalThis.navigator?.locks as WebLocks | undefined;
    if (!locks || typeof locks.request !== 'function') throw new IdentityVaultError('storage');
    let result: Promise<T>;
    try {
      result = locks.request(IDENTITY_VAULT_LOCK_NAME, { mode: 'exclusive' }, mutation);
    } catch {
      throw new IdentityVaultError('storage');
    }
    try {
      return await result;
    } catch (error) {
      throw error instanceof IdentityVaultError ? error : new IdentityVaultError('storage');
    }
  }

  private get storage(): IdentityVaultStorage {
    const storage = this.storageOverride ?? globalThis.localStorage;
    if (!storage) throw new IdentityVaultError('storage');
    return storage;
  }

  private readRaw(): string | null {
    try {
      return this.storage.getItem(IDENTITY_VAULT_STORAGE_KEY);
    } catch {
      throw new IdentityVaultError('storage');
    }
  }

  private writeDocument(document: VaultDocument): void {
    const addresses = Object.keys(document.records);
    try {
      if (addresses.length === 0) {
        this.storage.removeItem(IDENTITY_VAULT_STORAGE_KEY);
        return;
      }
      const serialized = JSON.stringify(document);
      if (serialized.length > MAX_IDENTITY_VAULT_CHARS) throw new IdentityVaultError('storage');
      this.storage.setItem(IDENTITY_VAULT_STORAGE_KEY, serialized);
    } catch {
      throw new IdentityVaultError('storage');
    }
  }

  /** Whether an encrypted identity exists for the address; never decrypts. */
  hasIdentity(address: string): boolean {
    assertAddress(address);
    return Object.prototype.hasOwnProperty.call(parseDocument(this.readRaw()).records, address);
  }

  /**
   * Store an identity for the address only if none exists. Existing entries
   * are never overwritten; the whole read-seal-write cycle holds the vault
   * lock, and a change by a writer outside the lock is still a conflict.
   * On any error nothing is written.
   */
  async createIdentity(address: string, password: string, identity: Identity): Promise<void> {
    assertAddress(address);
    assertPassword(password);
    if (!isValidIdentity(identity)) throw new IdentityVaultError('invalid-identity');
    await this.withMutationLock(() => this.createLocked(address, password, identity));
  }

  /**
   * Run a multi-step transaction under the single vault lock (Web Locks are not
   * re-entrant). Errors thrown by `work` itself propagate unchanged; the
   * transaction handle rejects every call after the callback has settled.
   */
  async runExclusive<T>(work: (tx: IdentityVaultTransaction) => Promise<T>): Promise<T> {
    let failure: { error: unknown } | null = null;
    const result = await this.withMutationLock(async () => {
      let active = true;
      const guard = () => {
        if (!active) throw new IdentityVaultError('storage');
      };
      const tx: IdentityVaultTransaction = {
        hasIdentity: (address) => {
          guard();
          return this.hasIdentity(address);
        },
        openIdentity: async (address, password) => {
          guard();
          return this.openIdentity(address, password);
        },
        createIdentity: async (address, password, identity, beforeWrite) => {
          guard();
          assertAddress(address);
          assertPassword(password);
          if (!isValidIdentity(identity)) throw new IdentityVaultError('invalid-identity');
          await this.createLocked(address, password, identity, () => {
            guard();
            beforeWrite?.();
          });
        },
      };
      try {
        return await work(tx);
      } catch (error) {
        failure = { error };
        return undefined;
      } finally {
        active = false;
      }
    });
    if (failure) throw (failure as { error: unknown }).error;
    return result as T;
  }

  private async createLocked(
    address: string,
    password: string,
    identity: Identity,
    beforeWrite?: () => void
  ): Promise<void> {
    const before = this.readRaw();
    const document = parseDocument(before);
    if (Object.prototype.hasOwnProperty.call(document.records, address)) {
      throw new IdentityVaultError('conflict');
    }
    if (Object.keys(document.records).length >= MAX_IDENTITY_VAULT_RECORDS) {
      throw new IdentityVaultError('storage');
    }

    let envelope: string;
    try {
      envelope = await sealEnvelope(serializeIdentity(identity), password, {
        aad: identityAad(address),
        maxEnvelopeChars: MAX_RECORD_ENVELOPE_CHARS,
      });
    } catch {
      throw new IdentityVaultError('storage');
    }

    // Key derivation is asynchronous: refuse to write over anything that changed meanwhile.
    if (this.readRaw() !== before) throw new IdentityVaultError('conflict');
    beforeWrite?.();

    this.writeDocument({
      v: IDENTITY_VAULT_VERSION,
      records: {
        ...document.records,
        [address]: {
          v: IDENTITY_VAULT_VERSION,
          alg: RECORD_ALG,
          kdf: RECORD_KDF,
          iter: CURRENT_PBKDF2_ITERATIONS,
          envelope,
          createdAt: Date.now(),
        },
      },
    });
  }

  /**
   * Decrypt the identity stored for the address with the wallet password.
   * Returns null when no identity exists. Wrong password, wrong address,
   * swapped or tampered records fail with one bounded error.
   */
  async openIdentity(address: string, password: string): Promise<Identity | null> {
    assertAddress(address);
    assertPassword(password);
    const { records } = parseDocument(this.readRaw());
    if (!Object.prototype.hasOwnProperty.call(records, address)) return null;
    const record = records[address];

    let plaintext: string;
    try {
      ({ plaintext } = await openEnvelope(record.envelope, password, {
        aad: identityAad(address),
        maxEnvelopeChars: MAX_RECORD_ENVELOPE_CHARS,
      }));
    } catch {
      throw new IdentityVaultError('locked');
    }

    let identity: unknown;
    try {
      identity = JSON.parse(plaintext);
    } catch {
      throw new IdentityVaultError('corrupt');
    }
    if (!isValidIdentity(identity) || serializeIdentity(identity) !== plaintext) {
      throw new IdentityVaultError('corrupt');
    }
    return identity;
  }

  /** Remove the entry for the address on explicit caller request, under the vault lock. */
  async removeIdentity(address: string): Promise<boolean> {
    assertAddress(address);
    return this.withMutationLock(async () => this.removeLocked(address));
  }

  private removeLocked(address: string): boolean {
    const document = parseDocument(this.readRaw());
    if (!Object.prototype.hasOwnProperty.call(document.records, address)) return false;
    const records = { ...document.records };
    delete records[address];
    this.writeDocument({ v: IDENTITY_VAULT_VERSION, records });
    return true;
  }
}
