import { useState } from 'react';
import { useIdentityStore } from '@/stores/identityStore';
import { Card } from '@/components/common/Card';
import { Button } from '@/components/common/Button';
import { TextArea } from '@/components/common/TextArea';
import {
  ShieldCheckIcon,
  ArrowDownTrayIcon,
  ArrowUpTrayIcon,
  ExclamationTriangleIcon,
} from '@heroicons/react/24/outline';

interface IdentitySetupProps {
  onComplete: () => void;
}

export function IdentitySetup({ onComplete }: IdentitySetupProps) {
  const { createIdentity, exportIdentity, importIdentity } =
    useIdentityStore();
  const [showExport, setShowExport] = useState(false);
  const [importData, setImportData] = useState('');
  const [error, setError] = useState('');
  const [mode, setMode] = useState<'create' | 'import'>('create');

  const handleCreate = () => {
    createIdentity();
    setShowExport(true);
  };

  const handleExport = () => {
    const exported = exportIdentity();
    if (exported) {
      navigator.clipboard.writeText(exported);
    }
  };

  const handleImport = () => {
    try {
      setError('');
      importIdentity(importData);
      onComplete();
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : 'Invalid identity data';
      setError(message);
    }
  };

  if (showExport) {
    return (
      <Card className="max-w-2xl mx-auto">
        <div className="mb-6">
          <div className="flex items-center gap-3 mb-4">
            <ExclamationTriangleIcon className="h-8 w-8 text-yellow-500" />
            <h2 className="text-2xl font-bold">Backup Your Identity</h2>
          </div>
          <p className="text-gray-600">
            Your anonymous identity has been created! You must save it to
            vote anonymously in the future.
          </p>
        </div>

        <div className="bg-yellow-50 border-2 border-yellow-200 rounded-lg p-6 mb-6">
          <h3 className="font-semibold text-yellow-900 mb-3">
            Important Security Notice
          </h3>
          <ul className="text-sm text-yellow-800 space-y-2">
            <li>
              This identity proves your membership without revealing who you
              are
            </li>
            <li>Store it securely -- you need it to vote anonymously</li>
            <li>Never share it with anyone</li>
            <li>
              If you lose it, you will need to create a new identity
            </li>
          </ul>
        </div>

        <div className="space-y-3">
          <Button
            onClick={handleExport}
            className="w-full flex items-center justify-center gap-2"
          >
            <ArrowDownTrayIcon className="h-5 w-5" />
            Export Identity to Clipboard
          </Button>

          <Button
            variant="secondary"
            onClick={onComplete}
            className="w-full"
          >
            I Have Saved My Identity
          </Button>
        </div>
      </Card>
    );
  }

  if (mode === 'import') {
    return (
      <Card className="max-w-2xl mx-auto">
        <h2 className="text-2xl font-bold mb-6">Import Identity</h2>

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-red-800">{error}</p>
          </div>
        )}

        <TextArea
          label="Identity Data"
          value={importData}
          onChange={(e) => setImportData(e.target.value)}
          placeholder="Paste your exported identity data here"
          rows={8}
          className="font-mono text-xs"
        />

        <div className="mt-6 space-y-3">
          <Button
            onClick={handleImport}
            disabled={!importData}
            className="w-full"
          >
            Import Identity
          </Button>
          <Button
            variant="secondary"
            onClick={() => setMode('create')}
            className="w-full"
          >
            Create New Instead
          </Button>
        </div>
      </Card>
    );
  }

  return (
    <Card className="max-w-2xl mx-auto">
      <div className="text-center mb-6">
        <ShieldCheckIcon className="h-16 w-16 text-primary-600 mx-auto mb-4" />
        <h2 className="text-2xl font-bold mb-2">Anonymous Voting Setup</h2>
        <p className="text-gray-600">
          Create an anonymous identity to vote privately
        </p>
      </div>

      <div className="bg-blue-50 border border-blue-200 rounded-lg p-6 mb-6">
        <h3 className="font-semibold text-blue-900 mb-3">How it works</h3>
        <ul className="text-sm text-blue-800 space-y-2">
          <li>Your votes are cryptographically anonymous</li>
          <li>
            Zero-knowledge proofs prove membership without revealing
            identity
          </li>
          <li>No one can link your votes back to you</li>
          <li>
            You can only vote once per suggestion (double-vote prevention)
          </li>
        </ul>
      </div>

      <div className="space-y-3">
        <Button
          onClick={handleCreate}
          className="w-full flex items-center justify-center gap-2"
        >
          <ShieldCheckIcon className="h-5 w-5" />
          Create Anonymous Identity
        </Button>

        <Button
          variant="secondary"
          onClick={() => setMode('import')}
          className="w-full flex items-center justify-center gap-2"
        >
          <ArrowUpTrayIcon className="h-5 w-5" />
          Import Existing Identity
        </Button>
      </div>
    </Card>
  );
}
