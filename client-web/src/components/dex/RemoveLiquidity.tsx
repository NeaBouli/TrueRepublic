import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useWalletStore } from '@/stores/walletStore';
import { useDEXStore } from '@/stores/dexStore';
import { DEXService } from '@/services/dex';
import { DEXTxService } from '@/services/dexTx';
import { WalletService } from '@/services/wallet';
import { Card } from '@/components/common/Card';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import { DEFAULT_CHAIN } from '@/config/chains';
import { formatPnyx } from '@/utils/format';
import type { LPPosition } from '@/types/dex';
import {
  ArrowLeftIcon,
  CheckCircleIcon,
  InformationCircleIcon,
} from '@heroicons/react/24/outline';

export function RemoveLiquidity() {
  const navigate = useNavigate();
  const { assetDenom } = useParams<{ assetDenom: string }>();
  const { currentWallet, password } = useWalletStore();
  const { pools, loadPools } = useDEXStore();

  const pool = pools.find((p) => p.asset_denom === assetDenom);

  const [shares, setShares] = useState('');
  const [positionPreview, setPositionPreview] = useState<LPPosition | null>(
    null
  );
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [txHash, setTxHash] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    if (pools.length === 0) {
      loadPools();
    }
  }, [pools.length, loadPools]);

  // Preview position value when shares change
  useEffect(() => {
    if (!assetDenom || !shares) {
      setPositionPreview(null);
      return;
    }

    const sharesNum = parseInt(shares, 10);
    if (isNaN(sharesNum) || sharesNum <= 0) {
      setPositionPreview(null);
      return;
    }

    const dexService = new DEXService(DEFAULT_CHAIN);
    dexService.getLPPosition(assetDenom, shares).then(setPositionPreview);
  }, [assetDenom, shares]);

  const handleRemoveLiquidity = async () => {
    if (!currentWallet || !password || !assetDenom || !shares) return;

    setIsSubmitting(true);
    setError('');

    try {
      const service = new DEXTxService(DEFAULT_CHAIN);
      const wallet = await WalletService.getWalletForSigning(
        currentWallet.address,
        password
      );

      const result = await service.removeLiquidity(wallet, {
        asset_denom: assetDenom,
        shares,
      });

      if (!result.success) {
        throw new Error(result.error || 'Remove liquidity failed');
      }

      setTxHash(result.hash);
    } catch (err: unknown) {
      setError(
        err instanceof Error ? err.message : 'Remove liquidity failed'
      );
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
          <h2 className="text-2xl font-bold mb-2">Liquidity Removed!</h2>
          <p className="text-gray-600 mb-6">
            Your funds have been returned
          </p>

          <div className="bg-gray-50 rounded-lg p-4 mb-6">
            <div className="text-xs text-gray-600 mb-1">Transaction Hash</div>
            <code className="text-xs font-mono break-all">{txHash}</code>
          </div>

          <Button onClick={() => navigate('/dex')} className="w-full">
            Back to Pools
          </Button>
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
          <h2 className="text-2xl font-bold mb-2">Remove Liquidity</h2>
          <p className="text-gray-600 mb-6">PNYX / {assetSymbol}</p>

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          )}

          {/* Pool Info */}
          <div className="bg-gray-50 rounded-lg p-4 mb-6">
            <div className="text-sm text-gray-600 mb-2">Pool Reserves</div>
            <div className="space-y-1">
              <div className="flex justify-between text-sm">
                <span>PNYX</span>
                <span className="font-medium">
                  {formatPnyx(pool.pnyx_reserve)}
                </span>
              </div>
              <div className="flex justify-between text-sm">
                <span>{assetSymbol}</span>
                <span className="font-medium">
                  {formatPnyx(pool.asset_reserve)}
                </span>
              </div>
              <div className="flex justify-between text-sm">
                <span>Total Shares</span>
                <span className="font-medium">{pool.total_shares}</span>
              </div>
            </div>
          </div>

          {/* Shares Input */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              LP Shares to Remove
            </label>
            <Input
              type="number"
              value={shares}
              onChange={(e) => setShares(e.target.value)}
              placeholder="Enter number of shares"
            />
          </div>

          {/* Preview Output */}
          {positionPreview && (
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
              <div className="flex items-start gap-2">
                <InformationCircleIcon className="h-5 w-5 text-blue-600 flex-shrink-0 mt-0.5" />
                <div className="text-sm text-blue-900">
                  <p className="font-semibold mb-2">You will receive</p>
                  <div className="space-y-1">
                    <p className="text-blue-800">
                      {formatPnyx(positionPreview.pnyx_value)} PNYX
                    </p>
                    <p className="text-blue-800">
                      {formatPnyx(positionPreview.asset_value)} {assetSymbol}
                    </p>
                  </div>
                  <p className="mt-2 text-xs text-blue-700">
                    Pool share: {(positionPreview.share_of_pool_bps / 100).toFixed(2)}%
                  </p>
                </div>
              </div>
            </div>
          )}

          <Button
            onClick={handleRemoveLiquidity}
            isLoading={isSubmitting}
            disabled={!shares || parseInt(shares, 10) <= 0}
            className="w-full"
          >
            Remove Liquidity
          </Button>
        </Card>
      </main>
    </div>
  );
}
