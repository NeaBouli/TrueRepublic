import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useIdentityStore } from '@/stores/identityStore';
import { Card } from '@/components/common/Card';
import { Button } from '@/components/common/Button';
import { IdentitySetup } from './IdentitySetup';
import {
  ArrowLeftIcon,
  ShieldCheckIcon,
  ClipboardDocumentIcon,
  ExclamationTriangleIcon,
} from '@heroicons/react/24/outline';

export function IdentityManager() {
  const navigate = useNavigate();
  const { identity, hasIdentity, exportIdentity, clearIdentity } =
    useIdentityStore();
  const [showConfirmDelete, setShowConfirmDelete] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleExport = () => {
    const exported = exportIdentity();
    if (exported) {
      navigator.clipboard.writeText(exported);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleDelete = () => {
    clearIdentity();
    setShowConfirmDelete(false);
  };

  if (!hasIdentity) {
    return (
      <div className="min-h-screen bg-gray-50">
        <header className="bg-white border-b border-gray-200">
          <div className="max-w-4xl mx-auto px-4 py-4">
            <button
              onClick={() => navigate(-1)}
              className="flex items-center gap-2 text-gray-600 hover:text-gray-900"
            >
              <ArrowLeftIcon className="h-5 w-5" />
              Back
            </button>
          </div>
        </header>
        <main className="max-w-4xl mx-auto px-4 py-8">
          <IdentitySetup onComplete={() => {}} />
        </main>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-4xl mx-auto px-4 py-4">
          <button
            onClick={() => navigate(-1)}
            className="flex items-center gap-2 text-gray-600 hover:text-gray-900"
          >
            <ArrowLeftIcon className="h-5 w-5" />
            Back
          </button>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8">
        <Card>
          <div className="flex items-center gap-3 mb-6">
            <ShieldCheckIcon className="h-8 w-8 text-green-600" />
            <div>
              <h2 className="text-xl font-bold">Anonymous Identity</h2>
              <p className="text-sm text-green-600 font-medium">Active</p>
            </div>
          </div>

          <div className="space-y-4 mb-6">
            <div>
              <label className="block text-sm font-medium text-gray-500 mb-1">
                Commitment
              </label>
              <div className="font-mono text-xs bg-gray-50 rounded-lg p-3 break-all border border-gray-200">
                {identity!.commitment}
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-500 mb-1">
                Created
              </label>
              <div className="text-sm text-gray-900">
                {new Date(identity!.createdAt).toLocaleDateString(undefined, {
                  year: 'numeric',
                  month: 'long',
                  day: 'numeric',
                  hour: '2-digit',
                  minute: '2-digit',
                })}
              </div>
            </div>
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
            <div className="text-sm text-blue-900">
              <strong>Zero-Knowledge Identity</strong>
              <p className="mt-1 text-blue-800">
                This identity allows you to vote anonymously using
                zero-knowledge proofs. Your votes cannot be linked back to you.
              </p>
            </div>
          </div>

          <div className="space-y-3">
            <Button
              onClick={handleExport}
              variant="secondary"
              className="w-full flex items-center justify-center gap-2"
            >
              <ClipboardDocumentIcon className="h-5 w-5" />
              {copied ? 'Copied to Clipboard!' : 'Export Identity Backup'}
            </Button>

            {!showConfirmDelete ? (
              <button
                onClick={() => setShowConfirmDelete(true)}
                className="w-full px-4 py-2 text-red-600 hover:bg-red-50 rounded-lg text-sm font-medium transition-colors"
              >
                Delete Identity
              </button>
            ) : (
              <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                <div className="flex items-start gap-2 mb-3">
                  <ExclamationTriangleIcon className="h-5 w-5 text-red-600 flex-shrink-0 mt-0.5" />
                  <div>
                    <div className="font-medium text-red-900">
                      Delete Identity?
                    </div>
                    <div className="text-sm text-red-800 mt-1">
                      This cannot be undone. You will need to create a new
                      identity to vote anonymously.
                    </div>
                  </div>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={handleDelete}
                    className="flex-1 px-4 py-2 bg-red-600 text-white text-sm font-medium rounded-lg hover:bg-red-700 transition-colors"
                  >
                    Confirm Delete
                  </button>
                  <button
                    onClick={() => setShowConfirmDelete(false)}
                    className="flex-1 px-4 py-2 bg-white text-gray-700 text-sm font-medium rounded-lg border border-gray-300 hover:bg-gray-50 transition-colors"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            )}
          </div>
        </Card>
      </main>
    </div>
  );
}
