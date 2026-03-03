import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useWalletStore } from '@/stores/walletStore';
import { useDEXStore } from '@/stores/dexStore';
import { DEXTxService } from '@/services/dexTx';
import { WalletService } from '@/services/wallet';
import { Card } from '@/components/common/Card';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import { DEFAULT_CHAIN } from '@/config/chains';
import { formatPnyx } from '@/utils/format';
import {
  ArrowLeftIcon,
  CheckCircleIcon,
  InformationCircleIcon,
} from '@heroicons/react/24/outline';

export function AddLiquidity() {
  const navigate = useNavigate();
  const { assetDenom } = useParams<{ assetDenom: string }>();
  const { currentWallet, password, balances } = useWalletStore();
  const { pools, loadPools } = useDEXStore();

  const pool = pools.find((p) => p.asset_denom === assetDenom);

  const [pnyxAmount, setPnyxAmount] = useState('');
  const [assetAmount, setAssetAmount] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [txHash, setTxHash] = useState('');
  const [error, setError] = useState('');

  const pnyxBalance = balances.find((b) => b.denom === 'upnyx');
  const assetBalance = balances.find((b) => b.denom === assetDenom);

  useEffect(() => {
    if (pools.length === 0) {
      loadPools();
    }
  }, [pools.length, loadPools]);

  // Auto-calculate asset amount based on pool ratio
  useEffect(() => {
    if (!pool || !pnyxAmount) {
      setAssetAmount('');
      return;
    }

    const pnyxAmt = parseInt(pnyxAmount, 10);
    if (isNaN(pnyxAmt) || pnyxAmt <= 0) {
      setAssetAmount('');
      return;
    }

    const pnyxRes = BigInt(pool.pnyx_reserve);
    const assetRes = BigInt(pool.asset_reserve);

    if (pnyxRes === 0n) {
      setAssetAmount('');
      return;
    }

    // Proportional deposit: asset_amt = pnyx_amt * asset_reserve / pnyx_reserve
    const calculatedAsset = (BigInt(pnyxAmt) * assetRes) / pnyxRes;
    setAssetAmount(calculatedAsset.toString());
  }, [pnyxAmount, pool]);

  const estimatedShares = (() => {
    if (!pool || !pnyxAmount) return '0';
    const pnyxAmt = parseInt(pnyxAmount, 10);
    if (isNaN(pnyxAmt) || pnyxAmt <= 0) return '0';

    const totalShares = BigInt(pool.total_shares);
    const pnyxRes = BigInt(pool.pnyx_reserve);
    if (pnyxRes === 0n) return '0';

    return ((BigInt(pnyxAmt) * totalShares) / pnyxRes).toString();
  })();

  const handleSetMax = () => {
    if (pnyxBalance) {
      const max = BigInt(pnyxBalance.amount) - 10000n;
      if (max > 0n) {
        setPnyxAmount(max.toString());
      }
    }
  };

  const handleAddLiquidity = async () => {
    if (!currentWallet || !password || !assetDenom || !pnyxAmount || !assetAmount)
      return;

    setIsSubmitting(true);
    setError('');

    try {
      const service = new DEXTxService(DEFAULT_CHAIN);
      const wallet = await WalletService.getWalletForSigning(
        currentWallet.address,
        password
      );

      const result = await service.addLiquidity(wallet, {
        asset_denom: assetDenom,
        pnyx_amt: pnyxAmount,
        asset_amt: assetAmount,
      });

      if (!result.success) {
        throw new Error(result.error || 'Add liquidity failed');
      }

      setTxHash(result.hash);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Add liquidity failed');
      setIsSubmitting(false);
    }
  };

  const assetSymbol =
    pool?.asset_symbol || assetDenom?.toUpperCase() || 'ASSET';

  if (txHash) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <Card className="max-w-md w-full text-center">
          <CheckCircleIcon className="h-16 w-16 text-green-600 mx-auto mb-4" />
          <h2 className="text-2xl font-bold mb-2">Liquidity Added!</h2>
          <p className="text-gray-600 mb-6">
            You are now earning trading fees
          </p>

          <div className="bg-gray-50 rounded-lg p-4 mb-6">
            <div className="text-xs text-gray-600 mb-1">Transaction Hash</div>
            <code className="text-xs font-mono break-all">{txHash}</code>
          </div>

          <div className="space-y-3">
            <Button onClick={() => navigate('/dex')} className="w-full">
              Back to Pools
            </Button>
          </div>
        </Card>
      </div>
    );
  }

  if (!pool) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <Card className="max-w-md w-full text-center">
          <p className="text-gray-600">Pool not found</p>
          <Button onClick={() => navigate('/dex')} className="w-full mt-4">
            Back to Pools
          </Button>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-2xl mx-auto px-4 py-4">
          <button
            onClick={() => navigate('/dex')}
            className="flex items-center gap-2 text-gray-600 hover:text-gray-900"
          >
            <ArrowLeftIcon className="h-5 w-5" />
            Back to Pools
          </button>
        </div>
      </header>

      <main className="max-w-2xl mx-auto px-4 py-8">
        <Card>
          <h2 className="text-2xl font-bold mb-2">Add Liquidity</h2>
          <p className="text-gray-600 mb-6">PNYX / {assetSymbol}</p>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          <div className="space-y-4 mb-6">
            {/* PNYX Amount */}
            <div>
              <div className="flex items-center justify-between mb-1">
                <label className="block text-sm font-medium text-gray-700">
                  PNYX Amount
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
                value={pnyxAmount}
                onChange={(e) => setPnyxAmount(e.target.value)}
                placeholder="0"
              />
              {pnyxBalance && (
                <div className="text-xs text-gray-500 mt-1">
                  Balance: {formatPnyx(pnyxBalance.amount)} PNYX
                </div>
              )}
            </div>

            {/* Asset Amount (auto-calculated) */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                {assetSymbol} Amount
              </label>
              <Input
                type="number"
                value={assetAmount}
                readOnly
                className="bg-gray-50"
              />
              {assetBalance && (
                <div className="text-xs text-gray-500 mt-1">
                  Balance: {formatPnyx(assetBalance.amount)} {assetSymbol}
                </div>
              )}
            </div>
          </div>

          {/* Pool Info */}
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
            <div className="flex items-start gap-2">
              <InformationCircleIcon className="h-5 w-5 text-blue-600 flex-shrink-0 mt-0.5" />
              <div className="text-sm text-blue-900">
                <p className="font-semibold mb-1">Pool Reserves</p>
                <p className="text-blue-800">
                  {formatPnyx(pool.pnyx_reserve)} PNYX /{' '}
                  {formatPnyx(pool.asset_reserve)} {assetSymbol}
                </p>
                {estimatedShares !== '0' && (
                  <p className="mt-2 text-blue-800">
                    Estimated LP Shares: <strong>{estimatedShares}</strong>
                  </p>
                )}
              </div>
            </div>
          </div>

          <Button
            onClick={handleAddLiquidity}
            isLoading={isSubmitting}
            disabled={!pnyxAmount || !assetAmount}
            className="w-full"
          >
            Add Liquidity
          </Button>
        </Card>
      </main>
    </div>
  );
}
