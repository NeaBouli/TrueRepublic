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
const RECORD_V2_KEYS = ['alg', 'createdAt', 'envelope', 'iter', 'kdf', 'kind', 'v'];
const CANONICAL_IDENTITY_KEYS = ['commitment', 'createdAt', 'kind', 'secret'];
const RECORD_V1 = 1;
const RECORD_V2 = 2;

/** Authenticated custody record kinds (GH309C1). record-v1 is always preview-v0. */
export const PREVIEW_IDENTITY_KIND = 'preview-v0';
export const CANONICAL_IDENTITY_KIND = 'canonical-bn254-mimc-v1';
export type IdentityKind = typeof PREVIEW_IDENTITY_KIND | typeof CANONICAL_IDENTITY_KIND;

/** Historical preview identity as opened from a record-v1 (FNV placeholder values). */
export interface PreviewCustodyIdentity extends Identity {
  kind: typeof PREVIEW_IDENTITY_KIND;
}

/** Canonical identity: commitment = BN254 MiMC(secret); no stored per-identity nullifier. */
export interface CanonicalIdentity {
  kind: typeof CANONICAL_IDENTITY_KIND;
  secret: string;
  commitment: string;
  createdAt: number;
}

export type CustodyIdentity = PreviewCustodyIdentity | CanonicalIdentity;
const DOCUMENT_KEYS = ['records', 'v'];
const IDENTITY_KEYS = ['commitment', 'createdAt', 'nullifier', 'secret'];

export type IdentityVaultErrorCode =
  | 'invalid-address'
  | 'invalid-identity'
  | 'invalid-password'
  | 'corrupt'
  | 'conflict'
  | 'kind-mismatch'
  | 'locked'
  | 'storage';

const ERROR_MESSAGES: Record<IdentityVaultErrorCode, string> = {
  'invalid-address': 'Identity vault requires a canonical wallet address',
  'invalid-identity': 'Identity data is invalid',
  'invalid-password': 'Identity vault requires the wallet password',
  corrupt: 'Identity vault data is corrupted or unsupported',
  conflict: 'An identity is already stored for this wallet or the vault changed',
  'kind-mismatch': 'The stored identity for this wallet is not a preview identity',
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

interface VaultRecordV1 {
  v: typeof RECORD_V1;
  alg: typeof RECORD_ALG;
  kdf: typeof RECORD_KDF;
  iter: typeof CURRENT_PBKDF2_ITERATIONS;
  envelope: string;
  createdAt: number;
}

interface VaultRecordV2 extends Omit<VaultRecordV1, 'v'> {
  v: typeof RECORD_V2;
  kind: typeof CANONICAL_IDENTITY_KIND;
}

type VaultRecord = VaultRecordV1 | VaultRecordV2;

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

/** record-v1 (preview-v0) AAD; unchanged since GH309B2B. */
export function identityAad(address: string): string {
  return `truerepublic/identity/v1|${address}`;
}

/** record-v2 AAD binds the record version, the exact kind and the canonical address. */
export function canonicalIdentityAad(address: string): string {
  return `truerepublic/identity/v2|${CANONICAL_IDENTITY_KIND}|${address}`;
}

/** Exact canonical plaintext; contains no preview nullifier. */
export function serializeCanonicalIdentity(identity: CanonicalIdentity): string {
  return JSON.stringify({
    kind: CANONICAL_IDENTITY_KIND,
    secret: identity.secret,
    commitment: identity.commitment,
    createdAt: identity.createdAt,
  });
}

function isCanonicalIdentityShape(value: unknown): value is CanonicalIdentity {
  return !(
    !isPlainObject(value) ||
    !hasExactKeys(value, CANONICAL_IDENTITY_KEYS) ||
    value.kind !== CANONICAL_IDENTITY_KIND ||
    typeof value.secret !== 'string' ||
    !HEX_32_BYTES.test(value.secret) ||
    typeof value.commitment !== 'string' ||
    !HEX_32_BYTES.test(value.commitment) ||
    !Number.isSafeInteger(value.createdAt) ||
    (value.createdAt as number) < 0
  );
}

/**
 * Canonical identity: exact shape, field-element secret and commitment =
 * mimcBn254([secret]). The MiMC code is loaded only here, after the cheap
 * shape checks, so normal startup and record-v1 never pull it into the entry
 * chunk. A module-load failure rejects (operational), it is not a verdict.
 */
export async function isValidCanonicalIdentity(value: unknown): Promise<boolean> {
  if (!isCanonicalIdentityShape(value)) return false;
  const { bytesToHex, hexToBytes, mimcBn254 } = await import('./zkpEncoding');
  try {
    return bytesToHex(mimcBn254([hexToBytes(value.secret)])) === value.commitment;
  } catch {
    // Secrets outside the BN254 scalar field are not canonical.
    return false;
  }
}

function isValidRecordCommon(value: Record<string, unknown>): boolean {
  return (
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

/** Strict union: v1 with the v1 key set, or v2 with exactly the canonical kind. */
function isValidRecord(value: unknown): value is VaultRecord {
  if (!isPlainObject(value) || !isValidRecordCommon(value)) return false;
  if (value.v === RECORD_V1) return hasExactKeys(value, RECORD_KEYS);
  if (value.v === RECORD_V2) return hasExactKeys(value, RECORD_V2_KEYS) && value.kind === CANONICAL_IDENTITY_KIND;
  return false;
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
          v: RECORD_V1,
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
   * Decrypt the identity stored for the address with the wallet password and
   * return it with its authenticated kind. record-v1 opens as preview-v0
   * (storage is never rewritten); record-v2 opens only as the canonical kind
   * under its own AAD. Wrong password, wrong address, swapped
   * address/version/kind or tampered records fail with one bounded error.
   */
  async openCustodyIdentity(address: string, password: string): Promise<CustodyIdentity | null> {
    assertAddress(address);
    assertPassword(password);
    const { records } = parseDocument(this.readRaw());
    if (!Object.prototype.hasOwnProperty.call(records, address)) return null;
    const record = records[address];
    const canonical = record.v === RECORD_V2;

    let plaintext: string;
    try {
      ({ plaintext } = await openEnvelope(record.envelope, password, {
        aad: canonical ? canonicalIdentityAad(address) : identityAad(address),
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
    if (canonical) {
      let valid: boolean;
      try {
        valid = await isValidCanonicalIdentity(identity);
      } catch {
        // The canonical validator chunk could not be loaded: operational, not corrupt or locked.
        throw new IdentityVaultError('storage');
      }
      if (!valid || serializeCanonicalIdentity(identity as CanonicalIdentity) !== plaintext) {
        throw new IdentityVaultError('corrupt');
      }
      return identity as CanonicalIdentity;
    }
    if (!isValidIdentity(identity) || serializeIdentity(identity) !== plaintext) {
      throw new IdentityVaultError('corrupt');
    }
    return { ...identity, kind: PREVIEW_IDENTITY_KIND };
  }

  /**
   * Preview-only view used by the preview custody flows: returns the record-v1
   * identity without its kind tag, null when absent, and refuses a canonical
   * record with 'kind-mismatch' instead of reinterpreting it.
   */
  async openIdentity(address: string, password: string): Promise<Identity | null> {
    const opened = await this.openCustodyIdentity(address, password);
    if (!opened) return null;
    if (opened.kind !== PREVIEW_IDENTITY_KIND) throw new IdentityVaultError('kind-mismatch');
    const { secret, commitment, nullifier, createdAt } = opened;
    return { secret, commitment, nullifier, createdAt };
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
