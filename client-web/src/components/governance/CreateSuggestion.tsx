import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useWalletStore } from '@/stores/walletStore';
import { GovernanceTxService } from '@/services/governanceTx';
import { WalletService } from '@/services/wallet';
import { Card } from '@/components/common/Card';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import { TextArea } from '@/components/common/TextArea';
import { DEFAULT_CHAIN } from '@/config/chains';
import { formatPnyx } from '@/utils/format';
import {
  ArrowLeftIcon,
  CheckCircleIcon,
  CurrencyDollarIcon,
} from '@heroicons/react/24/outline';

export function CreateSuggestion() {
  const navigate = useNavigate();
  const { domainId, issueId } = useParams<{
    domainId: string;
    issueId: string;
  }>();
  const { currentWallet, password, balances } = useWalletStore();

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [payToPut, setPayToPut] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [txHash, setTxHash] = useState('');
  const [error, setError] = useState('');

  const pnyxBalance = balances.find((b) => b.denom === 'upnyx');

  useEffect(() => {
    if (domainId) {
      const service = new GovernanceTxService(DEFAULT_CHAIN);
      service.calculatePayToPut(domainId).then((calc) => {
        setPayToPut(calc.finalCost);
      });
    }
  }, [domainId]);

  const handleSubmit = async () => {
    if (!currentWallet || !password || !domainId || !issueId) return;

    setIsSubmitting(true);
    setError('');

    try {
      const service = new GovernanceTxService(DEFAULT_CHAIN);
      const wallet = await WalletService.getWalletForSigning(
        currentWallet.address,
        password
      );

      const fee = payToPut
        ? [{ denom: DEFAULT_CHAIN.coinMinimalDenom, amount: payToPut }]
        : [];

      const result = await service.createSuggestion(
        wallet,
        domainId,
        issueId,
        title,
        fee,
        ''
      );

      if (!result.success) {
        throw new Error(result.error || 'Suggestion creation failed');
      }

      setTxHash(result.hash);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Suggestion creation failed');
      setIsSubmitting(false);
    }
  };

  if (txHash) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <Card className="max-w-md w-full text-center">
          <CheckCircleIcon className="h-16 w-16 text-green-600 mx-auto mb-4" />
          <h2 className="text-2xl font-bold mb-2">Suggestion Created!</h2>
          <p className="text-gray-600 mb-6">
            Your suggestion has been published
          </p>

          <div className="bg-gray-50 rounded-lg p-4 mb-6">
            <div className="text-xs text-gray-600 mb-1">Transaction Hash</div>
            <code className="text-xs font-mono break-all">{txHash}</code>
          </div>

          <div className="space-y-3">
            <Button
              onClick={() =>
                navigate(
                  `/governance/domain/${domainId}/issue/${issueId}`
                )
              }
              className="w-full"
            >
              View Suggestions
            </Button>
            <Button
              variant="secondary"
              onClick={() => {
                setTxHash('');
                setTitle('');
                setDescription('');
                setIsSubmitting(false);
              }}
              className="w-full"
            >
              Create Another
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-2xl mx-auto px-4 py-4">
          <button
            onClick={() => navigate(-1)}
            className="flex items-center gap-2 text-gray-600 hover:text-gray-900"
          >
            <ArrowLeftIcon className="h-5 w-5" />
            Back
          </button>
        </div>
      </header>

      <main className="max-w-2xl mx-auto px-4 py-8">
        <Card>
          <h2 className="text-2xl font-bold mb-6">Create Suggestion</h2>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          {payToPut && (
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
              <div className="flex items-center gap-2 mb-2">
                <CurrencyDollarIcon className="h-5 w-5 text-blue-600" />
                <span className="font-semibold text-blue-900">
                  PayToPut Cost
                </span>
              </div>
              <div className="text-2xl font-bold text-blue-900 mb-1">
                {formatPnyx(payToPut)} PNYX
              </div>
              <p className="text-sm text-blue-800">
                This fee goes to the domain treasury and prevents spam (eq.3)
              </p>
            </div>
          )}

          <div className="space-y-4 mb-6">
            <Input
              label="Suggestion Name"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Short, descriptive name"
              maxLength={100}
            />

            <TextArea
              label="Description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Explain your suggestion in detail"
              rows={6}
            />

            {pnyxBalance && payToPut && (
              <div className="text-sm text-gray-600">
                Balance: {formatPnyx(pnyxBalance.amount)} PNYX
                {BigInt(pnyxBalance.amount) < BigInt(payToPut) && (
                  <span className="text-red-600 ml-2">
                    (Insufficient balance)
                  </span>
                )}
              </div>
            )}
          </div>

          <Button
            onClick={handleSubmit}
            isLoading={isSubmitting}
            disabled={
              !title ||
              !description ||
              !payToPut ||
              (!!pnyxBalance &&
                BigInt(pnyxBalance.amount) < BigInt(payToPut))
            }
            className="w-full"
          >
            Create Suggestion ({payToPut ? formatPnyx(payToPut) : '...'} PNYX)
          </Button>
        </Card>
      </main>
    </div>
  );
}
