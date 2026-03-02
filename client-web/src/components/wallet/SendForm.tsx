import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useWalletStore } from '@/stores/walletStore';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import { Card } from '@/components/common/Card';
import { formatPnyx, parsePnyx } from '@/utils/format';
import { ArrowLeftIcon, CheckCircleIcon } from '@heroicons/react/24/outline';

export function SendForm() {
  const navigate = useNavigate();
  const { balances, sendTokens, isLoading } = useWalletStore();

  const [recipient, setRecipient] = useState('');
  const [amount, setAmount] = useState('');
  const [memo, setMemo] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [txHash, setTxHash] = useState('');

  const pnyxBalance = balances.find((b) => b.denom === 'pnyx');
  const availableBalance = pnyxBalance?.amount || '0';

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!recipient) {
      newErrors.recipient = 'Recipient address is required';
    } else if (!recipient.startsWith('true1')) {
      newErrors.recipient = 'Invalid address (must start with true1)';
    }

    if (!amount || parseFloat(amount) <= 0) {
      newErrors.amount = 'Amount must be greater than 0';
    }

    const amountMicro = parsePnyx(amount);
    if (BigInt(amountMicro) > BigInt(availableBalance)) {
      newErrors.amount = 'Insufficient balance';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSend = async () => {
    if (!validateForm()) return;

    try {
      const result = await sendTokens({
        to: recipient,
        amount: parsePnyx(amount),
        denom: 'pnyx',
        memo: memo || undefined,
      });

      setTxHash(result.hash);
    } catch (error: unknown) {
      const message =
        error instanceof Error ? error.message : 'Transaction failed';
      setErrors({ general: message });
    }
  };

  const handleSetMax = () => {
    const maxAmount = BigInt(availableBalance) - BigInt(10000);
    if (maxAmount > 0) {
      setAmount(formatPnyx(maxAmount.toString()));
    }
  };

  // Success screen
  if (txHash) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <Card className="max-w-md w-full text-center">
          <CheckCircleIcon className="h-16 w-16 text-green-600 mx-auto mb-4" />
          <h2 className="text-2xl font-bold mb-2">Transaction Sent!</h2>
          <p className="text-gray-600 mb-6">
            Your transaction has been broadcast to the network
          </p>

          <div className="bg-gray-50 rounded-lg p-4 mb-6">
            <div className="text-xs text-gray-600 mb-1">Transaction Hash</div>
            <code className="text-xs font-mono break-all">{txHash}</code>
          </div>

          <div className="space-y-3">
            <Button onClick={() => navigate('/wallet')} className="w-full">
              Back to Wallet
            </Button>
            <Button
              variant="secondary"
              onClick={() => {
                setTxHash('');
                setRecipient('');
                setAmount('');
                setMemo('');
              }}
              className="w-full"
            >
              Send Another
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  // Send form
  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-2xl mx-auto px-4 py-4">
          <button
            onClick={() => navigate('/wallet')}
            className="flex items-center gap-2 text-gray-600 hover:text-gray-900"
          >
            <ArrowLeftIcon className="h-5 w-5" />
            Back
          </button>
        </div>
      </header>

      <main className="max-w-2xl mx-auto px-4 py-8">
        <Card>
          <h2 className="text-2xl font-bold mb-6">Send PNYX</h2>

          {errors.general && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
              <p className="text-sm text-red-800">{errors.general}</p>
            </div>
          )}

          <div className="space-y-4">
            <Input
              label="Recipient Address"
              value={recipient}
              onChange={(e) => setRecipient(e.target.value)}
              placeholder="true1..."
              error={errors.recipient}
            />

            <div>
              <div className="flex items-center justify-between mb-1">
                <label className="block text-sm font-medium text-gray-700">
                  Amount
                </label>
                <button
                  onClick={handleSetMax}
                  className="text-sm text-primary-600 hover:text-primary-700"
                >
                  Max
                </button>
              </div>
              <Input
                type="number"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                placeholder="0.00"
                error={errors.amount}
                helperText={`Available: ${formatPnyx(availableBalance)} PNYX`}
              />
            </div>

            <Input
              label="Memo (Optional)"
              value={memo}
              onChange={(e) => setMemo(e.target.value)}
              placeholder="Add a note"
            />
          </div>

          <div className="mt-6 bg-gray-50 rounded-lg p-4">
            <div className="flex items-center justify-between text-sm">
              <span className="text-gray-600">Estimated Fee</span>
              <span className="font-medium">~0.005 PNYX</span>
            </div>
            <div className="flex items-center justify-between text-sm mt-2">
              <span className="text-gray-600">You will send</span>
              <span className="font-bold text-lg">{amount || '0'} PNYX</span>
            </div>
          </div>

          <Button
            onClick={handleSend}
            isLoading={isLoading}
            disabled={!recipient || !amount}
            className="w-full mt-6"
          >
            Send Transaction
          </Button>
        </Card>
      </main>
    </div>
  );
}
