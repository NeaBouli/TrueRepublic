import { useState, type ReactNode } from 'react';
import {
  ArrowDownTrayIcon,
  ExclamationTriangleIcon,
  LockClosedIcon,
  ShieldCheckIcon,
} from '@heroicons/react/24/outline';
import { Button } from '@/components/common/Button';
import { Card } from '@/components/common/Card';
import {
  IDENTITY_CREATION_DISABLED,
  LEGACY_IDENTITY_EXPORT_WARNING,
  PREVIEW_IDENTITY_EXPORT_WARNING,
  useIdentityStore,
  type IdentityProblem,
} from '@/stores/identityStore';
import { currentIdentitySession } from '@/stores/walletStore';
import { downloadIdentityFile } from '@/utils/identityDownload';

const PROBLEM_MESSAGES: Record<IdentityProblem, string> = {
  oversized: 'The stored preview identity is larger than any valid record.',
  'not-json': 'The stored preview identity is not readable data.',
  schema: 'The stored preview identity does not have the expected fields.',
  'non-canonical': 'The stored preview identity was not written by this client in the expected form.',
  'preview-hash-mismatch': "The stored preview identity's commitment or nullifier does not match its secret.",
  'vault-conflict': 'This wallet already holds a different encrypted identity; nothing was overwritten.',
  'legacy-changed': 'The unencrypted preview identity changed after it was migrated; both copies were kept.',
  'verify-mismatch': 'The encrypted copy could not be verified against the original; the original was kept.',
  'vault-corrupt': 'The encrypted identity storage on this device is damaged.',
  'corrupt-marker': 'The identity migration record on this device is damaged.',
  'wrong-password': "The identity could not be opened with this wallet's password.",
  storage: 'Secure identity storage is unavailable in this browser.',
  'migration-failed': 'The preview identity could not be migrated.',
};

const DISABLED_CLASSES = 'w-full min-h-[44px] disabled:bg-gray-200 disabled:text-gray-600';

function Notice({ tone, children, testId }: { tone: 'info' | 'warning'; children: ReactNode; testId: string }) {
  const classes =
    tone === 'warning'
      ? 'bg-yellow-50 border border-yellow-200 text-yellow-900'
      : 'bg-blue-50 border border-blue-200 text-blue-900';
  return (
    <div className={`${classes} rounded-lg p-4 mb-6 last:mb-0 text-sm break-words`} data-testid={testId}>
      {children}
    </div>
  );
}

export function IdentityCustodyNotice() {
  const { identity, status, problem, legacyPresent, migrateLegacy, exportIdentityFile, exportLegacyFile } =
    useIdentityStore();
  const [busy, setBusy] = useState(false);

  const exportLegacy = () => {
    const file = exportLegacyFile();
    if (file) downloadIdentityFile(file);
  };

  const legacyExportButton = legacyPresent ? (
    <Button variant="secondary" onClick={exportLegacy} className="w-full min-h-[44px] flex items-center justify-center gap-2">
      <ArrowDownTrayIcon className="h-5 w-5" aria-hidden="true" />
      Download Unencrypted Data File
    </Button>
  ) : null;

  const legacyWarning = legacyPresent ? (
    <p className="text-xs text-gray-700 mt-3" id="legacy-export-warning">
      {LEGACY_IDENTITY_EXPORT_WARNING}
    </p>
  ) : null;

  const creationDisabled = (
    <>
      <Button type="button" disabled aria-describedby="identity-creation-disabled-reason" className={DISABLED_CLASSES}>
        Identity Creation Disabled in Preview
      </Button>
      <p id="identity-creation-disabled-reason" className="sr-only">
        {IDENTITY_CREATION_DISABLED}
      </p>
    </>
  );

  if (status === 'locked') {
    return (
      <Card>
        <div className="flex items-center gap-3 mb-4">
          <LockClosedIcon className="h-8 w-8 shrink-0 text-gray-500" aria-hidden="true" />
          <h2 className="text-xl font-bold">Preview Identity Locked</h2>
        </div>
        <p role="status" className="text-gray-700">
          Unlock your wallet to use its preview identity. Identity data is never kept while the wallet is locked.
        </p>
      </Card>
    );
  }

  if (status === 'loading') {
    return (
      <Card>
        <h2 className="text-xl font-bold mb-4">Checking Preview Identity</h2>
        <p role="status" className="text-gray-700">
          Checking this wallet&apos;s preview identity…
        </p>
      </Card>
    );
  }

  if (status === 'absent') {
    return (
      <Card>
        <h2 className="text-xl font-bold mb-4">No Preview Identity</h2>
        <Notice tone="info" testId="identity-absent-notice">
          <p role="status">This wallet has no preview identity. {IDENTITY_CREATION_DISABLED}</p>
        </Notice>
        {creationDisabled}
      </Card>
    );
  }

  if (status === 'legacy-pending') {
    const encrypt = async () => {
      const session = currentIdentitySession();
      if (!session) return;
      setBusy(true);
      try {
        await migrateLegacy(session);
      } finally {
        setBusy(false);
      }
    };
    return (
      <Card>
        <div className="flex items-center gap-3 mb-4">
          <ExclamationTriangleIcon className="h-8 w-8 shrink-0 text-yellow-600" aria-hidden="true" />
          <h2 className="text-xl font-bold">Unencrypted Preview Identity Found</h2>
        </div>
        <Notice tone="warning" testId="identity-legacy-notice">
          <p role="status">
            An unencrypted preview identity is stored on this device. You can encrypt it into this wallet with the
            wallet password. The unencrypted copy is kept until you remove it yourself in a later version.
          </p>
        </Notice>
        <div className="space-y-3">
          <Button onClick={encrypt} disabled={busy} className="w-full min-h-[44px]">
            Encrypt Into This Wallet
          </Button>
          {legacyExportButton}
        </div>
        {legacyWarning}
      </Card>
    );
  }

  if (status === 'quarantined' || status === 'error') {
    return (
      <Card>
        <div className="flex items-center gap-3 mb-4">
          <ExclamationTriangleIcon className="h-8 w-8 shrink-0 text-red-600" aria-hidden="true" />
          <h2 className="text-xl font-bold">
            {status === 'quarantined' ? 'Preview Identity Not Used' : 'Preview Identity Unavailable'}
          </h2>
        </div>
        <Notice tone="warning" testId="identity-problem-notice">
          <p role="status">
            {problem ? PROBLEM_MESSAGES[problem] : PROBLEM_MESSAGES['migration-failed']} It is kept unchanged and is
            not used. Your wallet stays unlocked.
          </p>
        </Notice>
        {legacyExportButton && <div className="space-y-3">{legacyExportButton}</div>}
        {legacyWarning}
      </Card>
    );
  }

  // ready
  const exportIdentity = () => {
    const file = exportIdentityFile();
    if (file) downloadIdentityFile(file);
  };
  return (
    <Card>
      <div className="flex items-center gap-3 mb-6">
        <ShieldCheckIcon className="h-8 w-8 shrink-0 text-green-600" aria-hidden="true" />
        <div>
          <h2 className="text-xl font-bold">Preview Identity</h2>
          <p className="text-sm text-green-700 font-medium" data-testid="identity-ready-marker">
            Encrypted in this wallet
          </p>
        </div>
      </div>
      {identity && (
        <div className="mb-6">
          <span className="block text-sm font-medium text-gray-600 mb-1">Commitment (preview, not canonical)</span>
          <div className="font-mono text-xs bg-gray-50 rounded-lg p-3 break-all border border-gray-200">
            {identity.commitment}
          </div>
        </div>
      )}
      <Notice tone="warning" testId="identity-export-warning">
        <p>{PREVIEW_IDENTITY_EXPORT_WARNING}</p>
      </Notice>
      <div className="space-y-3">
        <Button
          variant="secondary"
          onClick={exportIdentity}
          className="w-full min-h-[44px] flex items-center justify-center gap-2"
        >
          <ArrowDownTrayIcon className="h-5 w-5" aria-hidden="true" />
          Download Identity Backup File
        </Button>
        <Button type="button" disabled aria-describedby="identity-delete-disabled-reason" className={DISABLED_CLASSES}>
          Identity Deletion Disabled in Preview
        </Button>
        <p id="identity-delete-disabled-reason" className="sr-only">
          Deleting an identity is disabled until an export-confirmed deletion flow exists (issue #309).
        </p>
      </div>
    </Card>
  );
}
