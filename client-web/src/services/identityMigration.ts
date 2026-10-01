import type { Identity } from '@/types/zkp';
import {
  IdentityVault,
  IdentityVaultError,
  isCanonicalAddress,
  isValidIdentity,
  serializeIdentity,
  type IdentityVaultStorage,
  type IdentityVaultTransaction,
} from './identityVault';

/**
 * Crash-safe legacy identity migration core (issue #309, GH309B2C).
 *
 * Moves the plaintext zustand record `identity-store` into the encrypted
 * identity vault without ever deleting or reinterpreting it:
 *
 *   (none) -> DETECTED -> PENDING(address) -> WRITTEN -> VERIFIED
 *   invalid legacy input -> QUARANTINED (terminal, never deleted)
 *   conflicting or inconsistent evidence -> ERROR (marker keeps the evidence)
 *
 * The marker under LEGACY_MIGRATION_MARKER_KEY records the state, the SHA-256
 * of the exact raw legacy bytes and the target address. Every step is
 * idempotent: re-running after an interruption at any transition resumes
 * from the marker and the stored bytes. The whole run holds the vault Web
 * Lock; without Web Locks nothing is written. Before every write that follows
 * an await, the caller's session check must still pass.
 *
 * Confirmed policy (Gio, 2026-10-01): plaintext is never removed here, not
 * even after VERIFIED; QUARANTINED records are never deleted; no identity is
 * generated or activated.
 */

export const LEGACY_IDENTITY_STORAGE_KEY = 'identity-store';
export const LEGACY_MIGRATION_MARKER_KEY = 'truerepublic_identity_migration_v1';
export const MAX_LEGACY_IDENTITY_BYTES = 4 * 1024;
const MARKER_VERSION = 1;
const MAX_MARKER_CHARS = 1_024;
const SHA256_HEX = /^[0-9a-f]{64}$/;

export type MigrationState = 'DETECTED' | 'PENDING' | 'WRITTEN' | 'VERIFIED' | 'QUARANTINED' | 'ERROR';

export type QuarantineReason = 'oversized' | 'not-json' | 'schema' | 'non-canonical';
export type MigrationErrorReason = 'vault-conflict' | 'legacy-changed' | 'verify-mismatch' | 'vault-corrupt';
export type MigrationFailureCode =
  | 'session-changed'
  | 'locked'
  | 'conflict'
  | 'storage'
  | 'corrupt-marker'
  | 'invalid-request';

export interface MigrationMarker {
  v: typeof MARKER_VERSION;
  state: MigrationState;
  legacyDigest: string;
  address: string | null;
  reason: QuarantineReason | MigrationErrorReason | null;
  updatedAt: number;
}

export type MigrationStatus =
  | { kind: 'none' }
  | { kind: 'legacy'; marker: MigrationMarker | null };

const FAILURE_MESSAGES: Record<MigrationFailureCode, string> = {
  'session-changed': 'Wallet session changed during identity migration',
  locked: 'Incorrect password or corrupted identity data',
  conflict: 'Identity migration conflicts with an existing vault entry or target',
  storage: 'Identity migration storage is unavailable',
  'corrupt-marker': 'Identity migration record is corrupted',
  'invalid-request': 'Identity migration request is invalid',
};

export class IdentityMigrationError extends Error {
  readonly code: MigrationFailureCode;

  constructor(code: MigrationFailureCode) {
    super(FAILURE_MESSAGES[code]);
    this.name = 'IdentityMigrationError';
    this.code = code;
  }
}

/** Caller-supplied session binding, checked before every post-await write. */
export interface MigrationSession {
  address: string;
  password: string;
  /** True while the same wallet address and session generation are still active. */
  isCurrent(): boolean;
}

type LegacyClassification =
  | { kind: 'absent' }
  | { kind: 'empty'; raw: string }
  | { kind: 'quarantined'; raw: string; reason: QuarantineReason }
  | { kind: 'valid'; raw: string; identity: Identity; identityBytes: string };

const STATES: readonly MigrationState[] = ['DETECTED', 'PENDING', 'WRITTEN', 'VERIFIED', 'QUARANTINED', 'ERROR'];
const REASONS: readonly string[] = [
  'oversized',
  'not-json',
  'schema',
  'non-canonical',
  'vault-conflict',
  'legacy-changed',
  'verify-mismatch',
  'vault-corrupt',
];
const MARKER_KEYS = ['address', 'legacyDigest', 'reason', 'state', 'updatedAt', 'v'];

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

/**
 * Classify the raw legacy value without interpreting anything beyond the exact
 * persisted schema `{"state":{"identity":<canonical identity>},"version":0}`.
 * Size is bounded in UTF-8 bytes before parsing.
 */
export function classifyLegacyIdentity(raw: string | null): LegacyClassification {
  if (raw === null) return { kind: 'absent' };
  if (raw.length > MAX_LEGACY_IDENTITY_BYTES || new TextEncoder().encode(raw).length > MAX_LEGACY_IDENTITY_BYTES) {
    return { kind: 'quarantined', raw, reason: 'oversized' };
  }
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    return { kind: 'quarantined', raw, reason: 'not-json' };
  }
  if (
    !isPlainObject(parsed) ||
    Object.keys(parsed).sort().join(',') !== 'state,version' ||
    parsed.version !== 0 ||
    !isPlainObject(parsed.state) ||
    Object.keys(parsed.state).join(',') !== 'identity'
  ) {
    return { kind: 'quarantined', raw, reason: 'schema' };
  }
  const identity = parsed.state.identity;
  if (identity === null) {
    return raw === '{"state":{"identity":null},"version":0}'
      ? { kind: 'empty', raw }
      : { kind: 'quarantined', raw, reason: 'non-canonical' };
  }
  if (!isValidIdentity(identity)) return { kind: 'quarantined', raw, reason: 'schema' };
  const identityBytes = serializeIdentity(identity);
  // Byte-exact layout: anything zustand would not have written verbatim is not reinterpreted.
  if (raw !== `{"state":{"identity":${identityBytes}},"version":0}`) {
    return { kind: 'quarantined', raw, reason: 'non-canonical' };
  }
  return { kind: 'valid', raw, identity, identityBytes };
}

async function sha256Hex(text: string): Promise<string> {
  const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(text) as ArrayBufferView<ArrayBuffer>);
  return Array.from(new Uint8Array(digest), (byte) => byte.toString(16).padStart(2, '0')).join('');
}

function parseMarker(raw: string | null): MigrationMarker | null {
  if (raw === null) return null;
  if (raw.length > MAX_MARKER_CHARS) throw new IdentityMigrationError('corrupt-marker');
  let parsed: unknown;
  try {
    parsed = JSON.parse(raw);
  } catch {
    throw new IdentityMigrationError('corrupt-marker');
  }
  if (
    !isPlainObject(parsed) ||
    Object.keys(parsed).sort().join(',') !== MARKER_KEYS.join(',') ||
    parsed.v !== MARKER_VERSION ||
    !STATES.includes(parsed.state as MigrationState) ||
    typeof parsed.legacyDigest !== 'string' ||
    !SHA256_HEX.test(parsed.legacyDigest) ||
    !(parsed.address === null || (typeof parsed.address === 'string' && parsed.address.length <= 128)) ||
    !(parsed.reason === null || REASONS.includes(parsed.reason as string)) ||
    !Number.isSafeInteger(parsed.updatedAt)
  ) {
    throw new IdentityMigrationError('corrupt-marker');
  }
  return parsed as unknown as MigrationMarker;
}

export class LegacyIdentityMigration {
  private readonly vault: IdentityVault;

  constructor(private readonly storageOverride?: IdentityVaultStorage) {
    this.vault = new IdentityVault(storageOverride);
  }

  private get storage(): IdentityVaultStorage {
    const storage = this.storageOverride ?? globalThis.localStorage;
    if (!storage) throw new IdentityMigrationError('storage');
    return storage;
  }

  private read(key: string): string | null {
    try {
      return this.storage.getItem(key);
    } catch {
      throw new IdentityMigrationError('storage');
    }
  }

  private writeMarker(marker: Omit<MigrationMarker, 'v' | 'updatedAt'>): MigrationMarker {
    const record: MigrationMarker = { v: MARKER_VERSION, ...marker, updatedAt: Date.now() };
    try {
      this.storage.setItem(LEGACY_MIGRATION_MARKER_KEY, JSON.stringify(record));
    } catch {
      throw new IdentityMigrationError('storage');
    }
    return record;
  }

  /** Read-only view of the legacy record and marker; never writes, never decrypts. */
  inspect(): MigrationStatus {
    const legacy = classifyLegacyIdentity(this.read(LEGACY_IDENTITY_STORAGE_KEY));
    if (legacy.kind === 'absent' || legacy.kind === 'empty') return { kind: 'none' };
    return { kind: 'legacy', marker: parseMarker(this.read(LEGACY_MIGRATION_MARKER_KEY)) };
  }

  /**
   * Classify only: record DETECTED or QUARANTINED for the current legacy bytes.
   * Idempotent; never touches the vault.
   */
  async detect(isCurrent: () => boolean): Promise<MigrationStatus> {
    const checkSession = () => {
      if (!isCurrent()) throw new IdentityMigrationError('session-changed');
    };
    return this.inLock(async () => {
      const legacy = classifyLegacyIdentity(this.read(LEGACY_IDENTITY_STORAGE_KEY));
      if (legacy.kind === 'absent' || legacy.kind === 'empty') return { kind: 'none' };
      const marker = parseMarker(this.read(LEGACY_MIGRATION_MARKER_KEY));
      const digest = await sha256Hex(legacy.raw);
      return { kind: 'legacy', marker: this.classified(legacy, marker, digest, checkSession) };
    });
  }

  /**
   * Run (or resume) migration of the legacy identity into the vault entry of
   * `session.address`. Returns the resulting marker; the plaintext legacy
   * record is always retained.
   */
  async migrate(session: MigrationSession): Promise<MigrationMarker | null> {
    if (
      !isPlainObject(session) ||
      typeof session.isCurrent !== 'function' ||
      !isCanonicalAddress(session.address) ||
      typeof session.password !== 'string' ||
      session.password.length < 8
    ) {
      throw new IdentityMigrationError('invalid-request');
    }
    const checkSession = () => {
      if (!session.isCurrent()) throw new IdentityMigrationError('session-changed');
    };
    return this.inLock(async (tx) => {
      const legacy = classifyLegacyIdentity(this.read(LEGACY_IDENTITY_STORAGE_KEY));
      if (legacy.kind === 'absent' || legacy.kind === 'empty') return null;
      let marker = parseMarker(this.read(LEGACY_MIGRATION_MARKER_KEY));
      const digest = await sha256Hex(legacy.raw);

      marker = this.classified(legacy, marker, digest, checkSession);
      if (legacy.kind !== 'valid' || marker.state === 'QUARANTINED' || marker.state === 'ERROR') {
        return marker;
      }
      if (marker.state === 'VERIFIED') return marker;

      if (marker.state === 'DETECTED' || (marker.state === 'PENDING' && marker.address !== session.address)) {
        // Nothing has been written to the vault for this digest in these states,
        // so (re)targeting the caller's address is lossless.
        checkSession();
        marker = this.writeMarker({ state: 'PENDING', legacyDigest: digest, address: session.address, reason: null });
      } else if (marker.address !== session.address) {
        throw new IdentityMigrationError('conflict');
      }

      if (marker.state === 'PENDING') {
        marker = await this.writeVault(tx, legacy, digest, session, checkSession);
        if (marker.state === 'ERROR') return marker;
      }

      return this.verify(tx, legacy, digest, session, checkSession);
    });
  }

  private async inLock<T>(work: (tx: IdentityVaultTransaction) => Promise<T>): Promise<T> {
    try {
      return await this.vault.runExclusive(work);
    } catch (error) {
      if (error instanceof IdentityMigrationError) throw error;
      if (error instanceof IdentityVaultError) {
        if (error.code === 'locked') throw new IdentityMigrationError('locked');
        if (error.code === 'conflict') throw new IdentityMigrationError('conflict');
      }
      throw new IdentityMigrationError('storage');
    }
  }

  /** Bring the marker in line with the current legacy bytes; never loses evidence. */
  private classified(
    legacy: Exclude<LegacyClassification, { kind: 'absent' } | { kind: 'empty' }>,
    marker: MigrationMarker | null,
    digest: string,
    checkSession: () => void
  ): MigrationMarker {
    const ensure = (next: Omit<MigrationMarker, 'v' | 'updatedAt'>) => {
      checkSession();
      return this.writeMarker(next);
    };
    if (marker && marker.legacyDigest !== digest) {
      if (marker.state === 'DETECTED' || marker.state === 'PENDING') {
        // Nothing was written for the old bytes; re-classify the new ones.
        marker = null;
      } else if (marker.state !== 'ERROR') {
        // WRITTEN/VERIFIED/QUARANTINED evidence belongs to other bytes: keep it visible.
        return ensure({ state: 'ERROR', legacyDigest: marker.legacyDigest, address: marker.address, reason: 'legacy-changed' });
      } else {
        return marker;
      }
    }
    if (legacy.kind === 'quarantined') {
      if (marker?.state === 'QUARANTINED') return marker;
      return ensure({ state: 'QUARANTINED', legacyDigest: digest, address: null, reason: legacy.reason });
    }
    if (!marker) return ensure({ state: 'DETECTED', legacyDigest: digest, address: null, reason: null });
    return marker;
  }

  private async writeVault(
    tx: IdentityVaultTransaction,
    legacy: Extract<LegacyClassification, { kind: 'valid' }>,
    digest: string,
    session: MigrationSession,
    checkSession: () => void
  ): Promise<MigrationMarker> {
    let existing: Identity | null;
    try {
      existing = tx.hasIdentity(session.address) ? await tx.openIdentity(session.address, session.password) : null;
    } catch (error) {
      if (error instanceof IdentityVaultError && error.code === 'corrupt') {
        checkSession();
        return this.writeMarker({ state: 'ERROR', legacyDigest: digest, address: session.address, reason: 'vault-corrupt' });
      }
      throw error;
    }
    if (existing) {
      // Resume after a crash between the vault write and the WRITTEN marker,
      // or a foreign identity that must never be overwritten.
      if (serializeIdentity(existing) !== legacy.identityBytes) {
        checkSession();
        return this.writeMarker({ state: 'ERROR', legacyDigest: digest, address: session.address, reason: 'vault-conflict' });
      }
    } else {
      await tx.createIdentity(session.address, session.password, legacy.identity, checkSession);
    }
    checkSession();
    return this.writeMarker({ state: 'WRITTEN', legacyDigest: digest, address: session.address, reason: null });
  }

  private async verify(
    tx: IdentityVaultTransaction,
    legacy: Extract<LegacyClassification, { kind: 'valid' }>,
    digest: string,
    session: MigrationSession,
    checkSession: () => void
  ): Promise<MigrationMarker> {
    const opened = await tx.openIdentity(session.address, session.password);
    const openedBytes = opened ? serializeIdentity(opened) : null;
    const rawAfter = this.read(LEGACY_IDENTITY_STORAGE_KEY);
    const ok =
      openedBytes === legacy.identityBytes &&
      rawAfter === legacy.raw &&
      openedBytes !== null &&
      (await sha256Hex(openedBytes)) === (await sha256Hex(legacy.identityBytes)) &&
      (await sha256Hex(rawAfter)) === digest;
    checkSession();
    return this.writeMarker({
      state: ok ? 'VERIFIED' : 'ERROR',
      legacyDigest: digest,
      address: session.address,
      reason: ok ? null : 'verify-mismatch',
    });
  }
}
