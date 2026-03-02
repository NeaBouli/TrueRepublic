import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useWalletStore } from '@/stores/walletStore';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import { TextArea } from '@/components/common/TextArea';
import { Card } from '@/components/common/Card';
import {
  validateWalletName,
  validatePassword,
  validateMnemonic,
} from '@/utils/validation';

export function ImportWallet() {
  const navigate = useNavigate();
  const { importWallet, isLoading } = useWalletStore();

  const [name, setName] = useState('');
  const [mnemonic, setMnemonic] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    const nameValidation = validateWalletName(name);
    if (!nameValidation.valid) {
      newErrors.name = nameValidation.error!;
    }

    const mnemonicValidation = validateMnemonic(mnemonic);
    if (!mnemonicValidation.valid) {
      newErrors.mnemonic = mnemonicValidation.error!;
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

  const handleImport = async () => {
    if (!validateForm()) return;

    try {
      await importWallet(name, mnemonic.trim(), password);
      navigate('/wallet');
    } catch (error: unknown) {
      const message =
        error instanceof Error ? error.message : 'Failed to import wallet';
      setErrors({ general: message });
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <Card className="max-w-md w-full">
        <h2 className="text-2xl font-bold mb-6">Import Wallet</h2>

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

          <TextArea
            label="Recovery Phrase"
            value={mnemonic}
            onChange={(e) => setMnemonic(e.target.value)}
            placeholder="Enter your 12 or 24 word recovery phrase"
            rows={4}
            error={errors.mnemonic}
            helperText="Separate words with spaces"
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
            onClick={handleImport}
            isLoading={isLoading}
            className="w-full"
          >
            Import Wallet
          </Button>

          <Button
            variant="secondary"
            onClick={() => navigate('/create')}
            className="w-full"
          >
            Create New Wallet
          </Button>
        </div>
      </Card>
    </div>
  );
}
