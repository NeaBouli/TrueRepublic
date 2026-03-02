import { useState } from 'react';
import { useWalletStore } from '@/stores/walletStore';
import { Card } from '@/components/common/Card';
import { formatAddress, copyToClipboard } from '@/utils/format';
import {
  ClipboardDocumentIcon,
  CheckIcon,
  ArrowRightOnRectangleIcon,
} from '@heroicons/react/24/outline';
import { useNavigate } from 'react-router-dom';

export function AccountInfo() {
  const navigate = useNavigate();
  const { currentWallet, lock } = useWalletStore();
  const [copied, setCopied] = useState(false);

  if (!currentWallet) return null;

  const handleCopy = async () => {
    const success = await copyToClipboard(currentWallet.address);
    if (success) {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const handleLock = () => {
    lock();
    navigate('/unlock');
  };

  return (
    <Card>
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold">{currentWallet.name}</h3>
        <button
          onClick={handleLock}
          className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
          title="Lock wallet"
        >
          <ArrowRightOnRectangleIcon className="h-5 w-5 text-gray-600" />
        </button>
      </div>

      <div className="bg-gray-50 rounded-lg p-4">
        <div className="text-xs text-gray-600 mb-1">Address</div>
        <div className="flex items-center justify-between gap-2">
          <code className="text-sm font-mono text-gray-900">
            {formatAddress(currentWallet.address, 12)}
          </code>
          <button
            onClick={handleCopy}
            className="p-2 hover:bg-gray-200 rounded transition-colors"
          >
            {copied ? (
              <CheckIcon className="h-4 w-4 text-green-600" />
            ) : (
              <ClipboardDocumentIcon className="h-4 w-4 text-gray-600" />
            )}
          </button>
        </div>
      </div>

      <div className="mt-4 text-xs text-gray-500">
        Created {new Date(currentWallet.createdAt).toLocaleDateString()}
      </div>
    </Card>
  );
}
