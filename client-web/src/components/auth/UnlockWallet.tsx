import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useWalletStore } from '@/stores/walletStore';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import { Card } from '@/components/common/Card';
import { LockClosedIcon } from '@heroicons/react/24/outline';

export function UnlockWallet() {
  const navigate = useNavigate();
  const { wallets, unlock, isLoading, loadWallets } = useWalletStore();

  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    loadWallets();
  }, [loadWallets]);

  const handleUnlock = async () => {
    if (!password) {
      setError('Password is required');
      return;
    }

    try {
      setError('');
      await unlock(password);
      navigate('/wallet');
    } catch {
      setError('Incorrect password');
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleUnlock();
    }
  };

  if (wallets.length === 0) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <Card className="max-w-md w-full text-center">
          <div className="mb-6">
            <LockClosedIcon className="h-16 w-16 text-gray-400 mx-auto mb-4" />
            <h2 className="text-2xl font-bold mb-2">No Wallet Found</h2>
            <p className="text-gray-600">
              Create or import a wallet to get started
            </p>
          </div>

          <div className="space-y-3">
            <Button onClick={() => navigate('/create')} className="w-full">
              Create New Wallet
            </Button>
            <Button
              variant="secondary"
              onClick={() => navigate('/import')}
              className="w-full"
            >
              Import Wallet
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <Card className="max-w-md w-full">
        <div className="text-center mb-6">
          <LockClosedIcon className="h-16 w-16 text-primary-600 mx-auto mb-4" />
          <h2 className="text-2xl font-bold mb-2">Unlock Wallet</h2>
          <p className="text-gray-600">Enter your password to continue</p>
        </div>

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-red-800">{error}</p>
          </div>
        )}

        <div className="mb-6">
          <Input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Enter password"
            autoFocus
          />
        </div>

        <Button
          onClick={handleUnlock}
          isLoading={isLoading}
          className="w-full"
        >
          Unlock
        </Button>

        <div className="mt-6 text-center">
          <button
            onClick={() => navigate('/create')}
            className="text-sm text-primary-600 hover:text-primary-700"
          >
            Create a new wallet instead
          </button>
        </div>
      </Card>
    </div>
  );
}
