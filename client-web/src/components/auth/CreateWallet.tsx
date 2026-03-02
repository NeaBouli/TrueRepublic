import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useWalletStore } from '@/stores/walletStore';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import { Card } from '@/components/common/Card';
import { validateWalletName, validatePassword } from '@/utils/validation';
import { ExclamationTriangleIcon } from '@heroicons/react/24/outline';

export function CreateWallet() {
  const navigate = useNavigate();
  const { createWallet, isLoading } = useWalletStore();

  const [name, setName] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [mnemonic, setMnemonic] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    const nameValidation = validateWalletName(name);
    if (!nameValidation.valid) {
      newErrors.name = nameValidation.error!;
    }

    const passwordValidation = validatePassword(password);
    if (!passwordValidation.valid) {
      newErrors.password = passwordValidation.error!;
    }

    if (password !== confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleCreate = async () => {
    if (!validateForm()) return;

    try {
      const wallet = await createWallet(name, password);
      // The wallet returned from createWallet still has the mnemonic
      // before the store strips it
      setMnemonic(wallet.mnemonic || '');
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Failed to create wallet';
      setErrors({ general: message });
    }
  };

  const handleConfirmSaved = () => {
    setSaved(true);
    navigate('/wallet');
  };

  // Mnemonic display screen
  if (mnemonic && !saved) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <Card className="max-w-2xl w-full">
          <div className="mb-6">
            <div className="flex items-center gap-3 mb-4">
              <ExclamationTriangleIcon className="h-8 w-8 text-yellow-500" />
              <h2 className="text-2xl font-bold">Save Your Recovery Phrase</h2>
            </div>
            <p className="text-gray-600">
              Write down these 24 words in order and store them safely.
              You'll need them to recover your wallet.
            </p>
          </div>

          <div className="bg-yellow-50 border-2 border-yellow-200 rounded-lg p-6 mb-6">
            <div className="grid grid-cols-3 gap-4">
              {mnemonic.split(' ').map((word, i) => (
                <div key={i} className="flex items-center gap-2">
                  <span className="text-gray-400 text-sm w-6">{i + 1}.</span>
                  <span className="font-mono font-medium">{word}</span>
                </div>
              ))}
            </div>
          </div>

          <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
            <div className="flex gap-3">
              <ExclamationTriangleIcon className="h-5 w-5 text-red-600 flex-shrink-0 mt-0.5" />
              <div className="text-sm text-red-800">
                <p className="font-semibold mb-1">Warning:</p>
                <ul className="list-disc list-inside space-y-1">
                  <li>Never share your recovery phrase with anyone</li>
                  <li>Store it offline in a secure location</li>
                  <li>If you lose it, you cannot recover your wallet</li>
                  <li>Anyone with this phrase can access your funds</li>
                </ul>
              </div>
            </div>
          </div>

          <Button onClick={handleConfirmSaved} className="w-full">
            I Have Saved My Recovery Phrase
          </Button>
        </Card>
      </div>
    );
  }

  // Create wallet form
  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <Card className="max-w-md w-full">
        <h2 className="text-2xl font-bold mb-6">Create New Wallet</h2>

        {errors.general && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-red-800">{errors.general}</p>
          </div>
        )}

        <div className="space-y-4">
          <Input
            label="Wallet Name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="My Wallet"
            error={errors.name}
          />

          <Input
            label="Password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="At least 8 characters"
            error={errors.password}
            helperText="Used to encrypt your wallet"
          />

          <Input
            label="Confirm Password"
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            placeholder="Re-enter password"
            error={errors.confirmPassword}
          />
        </div>

        <div className="mt-6 space-y-3">
          <Button
            onClick={handleCreate}
            isLoading={isLoading}
            className="w-full"
          >
            Create Wallet
          </Button>

          <Button
            variant="secondary"
            onClick={() => navigate('/import')}
            className="w-full"
          >
            Import Existing Wallet
          </Button>
        </div>
      </Card>
    </div>
  );
}
