import { create } from 'zustand';
import type { Identity } from '@/types/zkp';
import {
  IdentityMigrationError,
  LEGACY_IDENTITY_STORAGE_KEY,
  LegacyIdentityMigration,
  type MigrationMarker,
} from '@/services/identityMigration';
import { IdentityVault, IdentityVaultError, serializeIdentity } from '@/services/identityVault';

/**
 * Memory-only preview identity session (issue #309).
 *
 * The decrypted identity exists only in this non-persisted store while its
 * wallet is unlocked. There is deliberately no persist middleware: the
 * legacy plaintext `identity-store` value is never hydrated, rewritten or
 * removed here; only LegacyIdentityMigration reads it. Creating or importing
 * preview identities is disabled until GH309C.
 */

export type IdentityStatus =
  | 'locked'
  | 'loading'
  | 'absent'
  | 'legacy-pending'
  | 'quarantined'
  | 'error'
  | 'ready';

export type IdentityProblem =
  | 'oversized'
  | 'not-json'
  | 'schema'
  | 'non-canonical'
  | 'preview-hash-mismatch'
  | 'vault-conflict'
  | 'legacy-changed'
  | 'verify-mismatch'
  | 'vault-corrupt'
  | 'corrupt-marker'
  | 'wrong-password'
  | 'storage'
  | 'migration-failed';

/** Caller-bound wallet session; `isCurrent` must reflect generation, address and lock state. */
export interface IdentitySession {
  address: string;
  password: string;
  isCurrent(): boolean;
}

export interface IdentityExportFile {
  filename: string;
  mimeType: string;
  content: string;
}

interface IdentityStore {
  identity: Identity | null;
  hasIdentity: boolean;
  status: IdentityStatus;
  problem: IdentityProblem | null;
  /** True when an unencrypted legacy record exists on this device (export offered). */
  legacyPresent: boolean;
  /** Bumped by every invalidate(); async completions from older tokens are dropped. */
  token: number;

  invalidate: () => void;
  load: (session: IdentitySession) => Promise<void>;
  migrateLegacy: (session: IdentitySession) => Promise<void>;
  exportIdentityFile: () => IdentityExportFile | null;
  exportLegacyFile: () => IdentityExportFile | null;
}

const LOCKED = {
  identity: null,
  hasIdentity: false,
  status: 'locked' as const,
  problem: null,
  legacyPresent: false,
};

export const IDENTITY_CREATION_DISABLED =
  'Creating or importing preview identities is disabled in this preview (issue #309).';
export const PREVIEW_IDENTITY_EXPORT_WARNING =
  'Preview identity (issue #309): not canonical and not production-valid. It contains your identity secret — keep the file offline and never share it.';
export const LEGACY_IDENTITY_EXPORT_WARNING =
  'Unencrypted preview identity data exactly as stored on this device. It may contain an identity secret — keep the file offline and never share it.';

function readLegacyRaw(): string | null {
  try {
    return globalThis.localStorage?.getItem(LEGACY_IDENTITY_STORAGE_KEY) ?? null;
  } catch {
    return null;
  }
}

function problemFromMarker(marker: MigrationMarker): IdentityProblem {
  return (marker.reason ?? 'migration-failed') as IdentityProblem;
}

function problemFromError(error: unknown): IdentityProblem {
  if (error instanceof IdentityVaultError) {
    if (error.code === 'locked') return 'wrong-password';
    if (error.code === 'corrupt') return 'vault-corrupt';
    if (error.code === 'kind-mismatch') return 'vault-conflict';
    return 'storage';
  }
  if (error instanceof IdentityMigrationError) {
    if (error.code === 'locked') return 'wrong-password';
    if (error.code === 'corrupt-marker') return 'corrupt-marker';
    if (error.code === 'conflict') return 'vault-conflict';
    if (error.code === 'storage') return 'storage';
  }
  return 'migration-failed';
}

export const useIdentityStore = create<IdentityStore>()((set, get) => {
  /** Apply a result only if no wallet transition happened since `token` was taken. */
  const commit = (token: number, session: IdentitySession, next: Partial<IdentityStore>) => {
    if (get().token !== token || !session.isCurrent()) return false;
    set(next);
    return true;
  };

  const resolve = async (token: number, session: IdentitySession): Promise<void> => {
    const vault = new IdentityVault();
    const migration = new LegacyIdentityMigration();
    const legacyPresent = readLegacyRaw() !== null;
    try {
      if (vault.hasIdentity(session.address)) {
        const identity = await vault.openIdentity(session.address, session.password);
        if (identity) {
          commit(token, session, { identity, hasIdentity: true, status: 'ready', problem: null, legacyPresent });
          return;
        }
      }
      const detected = await migration.detect(() => session.isCurrent() && get().token === token);
      if (detected.kind === 'none' || !detected.marker) {
        commit(token, session, { ...LOCKED, status: 'absent', legacyPresent });
        return;
      }
      const marker = detected.marker;
      if (marker.state === 'QUARANTINED') {
        commit(token, session, { ...LOCKED, status: 'quarantined', problem: problemFromMarker(marker), legacyPresent });
      } else if (marker.state === 'ERROR') {
        commit(token, session, { ...LOCKED, status: 'error', problem: problemFromMarker(marker), legacyPresent });
      } else if ((marker.state === 'WRITTEN' || marker.state === 'VERIFIED') && marker.address !== session.address) {
        // The legacy record already belongs to another wallet's vault entry.
        commit(token, session, { ...LOCKED, status: 'absent', legacyPresent });
      } else {
        commit(token, session, { ...LOCKED, status: 'legacy-pending', legacyPresent });
      }
    } catch (error) {
      commit(token, session, { ...LOCKED, status: 'error', problem: problemFromError(error), legacyPresent });
    }
  };

  return {
    ...LOCKED,
    token: 0,

    invalidate: () => {
      set((state) => ({ ...LOCKED, token: state.token + 1 }));
    },

    load: async (session) => {
      const token = get().token;
      if (!commit(token, session, { ...LOCKED, status: 'loading' })) return;
      await resolve(token, session);
    },

    migrateLegacy: async (session) => {
      const token = get().token;
      if (!commit(token, session, { status: 'loading', problem: null })) return;
      try {
        const marker = await new LegacyIdentityMigration().migrate(session);
        if (marker && marker.state !== 'VERIFIED') {
          commit(token, session, {
            ...LOCKED,
            status: marker.state === 'QUARANTINED' ? 'quarantined' : 'error',
            problem: problemFromMarker(marker),
            legacyPresent: readLegacyRaw() !== null,
          });
          return;
        }
      } catch (error) {
        commit(token, session, {
          ...LOCKED,
          status: 'error',
          problem: problemFromError(error),
          legacyPresent: readLegacyRaw() !== null,
        });
        return;
      }
      await resolve(token, session);
    },

    exportIdentityFile: () => {
      const { identity, status } = get();
      if (status !== 'ready' || !identity) return null;
      return {
        filename: 'truerepublic-preview-identity.json',
        mimeType: 'application/json',
        content: `${JSON.stringify(
          {
            format: 'truerepublic-preview-identity',
            version: 1,
            canonical: false,
            warning: PREVIEW_IDENTITY_EXPORT_WARNING,
            identity: JSON.parse(serializeIdentity(identity)),
          },
          null,
          2
        )}\n`,
      };
    },

    exportLegacyFile: () => {
      if (get().status === 'locked') return null;
      const raw = readLegacyRaw();
      if (raw === null) return null;
      return {
        filename: 'truerepublic-legacy-identity-store.txt',
        mimeType: 'text/plain',
        content: raw,
      };
    },
  };
});
