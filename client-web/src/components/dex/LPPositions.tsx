import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDEXStore } from '@/stores/dexStore';
import { DEXService } from '@/services/dex';
import { Card } from '@/components/common/Card';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import { DEFAULT_CHAIN } from '@/config/chains';
import { formatPnyx } from '@/utils/format';
import type { Pool, LPPosition } from '@/types/dex';
import {
  ArrowLeftIcon,
  InformationCircleIcon,
  PlusIcon,
  MinusIcon,
} from '@heroicons/react/24/outline';

interface PoolPosition {
  pool: Pool;
  shares: string;
  position: LPPosition | null;
}

export function LPPositions() {
  const navigate = useNavigate();
  const { pools, loadPools, isLoading } = useDEXStore();

  const [positions, setPositions] = useState<PoolPosition[]>([]);
  const [sharesInput, setSharesInput] = useState<Record<string, string>>({});

  useEffect(() => {
    loadPools();
  }, [loadPools]);

  useEffect(() => {
    setPositions(pools.map((pool) => ({ pool, shares: '', position: null })));
  }, [pools]);

  const handleCheckPosition = async (assetDenom: string) => {
    const shares = sharesInput[assetDenom];
    if (!shares || parseInt(shares, 10) <= 0) return;

    const dexService = new DEXService(DEFAULT_CHAIN);
    const position = await dexService.getLPPosition(assetDenom, shares);

    setPositions((prev) =>
      prev.map((p) =>
        p.pool.asset_denom === assetDenom
          ? { ...p, shares, position }
          : p
      )
    );
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-6xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <button
                onClick={() => navigate('/dex')}
                className="flex items-center gap-2 text-gray-600 hover:text-gray-900"
              >
                <ArrowLeftIcon className="h-5 w-5" />
              </button>
              <div>
                <h1 className="text-2xl font-bold">LP Positions</h1>
                <p className="text-gray-600">
                  Check your liquidity positions
                </p>
              </div>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-4 py-8">
        {/* Info Banner */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
          <div className="flex items-start gap-2">
            <InformationCircleIcon className="h-5 w-5 text-blue-600 flex-shrink-0 mt-0.5" />
            <p className="text-sm text-blue-900">
              Enter your LP share count for any pool to see the current value of
              your position. LP shares are received when you add liquidity.
            </p>
          </div>
        </div>

        {isLoading && (
          <div className="text-center py-12 text-gray-500">
            Loading pools...
          </div>
        )}

        {!isLoading && positions.length === 0 && (
          <Card>
            <div className="text-center py-12 text-gray-500">
              <p className="mb-4">No pools available</p>
              <Button onClick={() => navigate('/dex')}>Back to DEX</Button>
            </div>
          </Card>
        )}

        <div className="space-y-4">
          {positions.map(({ pool, position }) => {
            const assetSymbol =
              pool.asset_symbol || pool.asset_denom.toUpperCase();
            const inputShares = sharesInput[pool.asset_denom] || '';

            return (
              <Card key={pool.asset_denom}>
                <div className="flex items-center justify-between mb-4">
                  <div>
                    <h3 className="font-bold text-lg">
                      PNYX / {assetSymbol}
                    </h3>
                    <div className="text-xs text-gray-500">
                      {pool.asset_denom}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button
                      onClick={() =>
                        navigate(`/dex/pool/${pool.asset_denom}/add`)
                      }
                      className="!py-1.5 !px-3 !text-sm"
                    >
                      <PlusIcon className="h-4 w-4 mr-1 inline" />
                      Add
                    </Button>
                    <Button
                      onClick={() =>
                        navigate(`/dex/pool/${pool.asset_denom}/remove`)
                      }
                      className="!py-1.5 !px-3 !text-sm bg-gray-600 hover:bg-gray-700"
                    >
                      <MinusIcon className="h-4 w-4 mr-1 inline" />
                      Remove
                    </Button>
                  </div>
                </div>

                {/* Pool reserves */}
                <div className="bg-gray-50 rounded-lg p-3 mb-4">
                  <div className="text-xs text-gray-600 mb-2">
                    Pool Reserves
                  </div>
                  <div className="flex items-center justify-between text-sm">
                    <span>{formatPnyx(pool.pnyx_reserve)} PNYX</span>
                    <span>
                      {formatPnyx(pool.asset_reserve)} {assetSymbol}
                    </span>
                  </div>
                  <div className="text-xs text-gray-500 mt-1">
                    Total Shares: {pool.total_shares}
                  </div>
                </div>

                {/* Check position */}
                <div className="flex items-end gap-3">
                  <div className="flex-1">
                    <label className="block text-xs font-medium text-gray-700 mb-1">
                      Your LP Shares
                    </label>
                    <Input
                      type="number"
                      value={inputShares}
                      onChange={(e) =>
                        setSharesInput((prev) => ({
                          ...prev,
                          [pool.asset_denom]: e.target.value,
                        }))
                      }
                      placeholder="Enter share count"
                    />
                  </div>
                  <Button
                    onClick={() => handleCheckPosition(pool.asset_denom)}
                    disabled={!inputShares || parseInt(inputShares, 10) <= 0}
                    className="!py-2"
                  >
                    Check
                  </Button>
                </div>

                {/* Position result */}
                {position && (
                  <div className="mt-4 bg-green-50 border border-green-200 rounded-lg p-4">
                    <div className="text-sm font-semibold text-green-900 mb-2">
                      Position Value
                    </div>
                    <div className="space-y-1">
                      <div className="flex justify-between text-sm text-green-800">
                        <span>PNYX</span>
                        <span className="font-medium">
                          {formatPnyx(position.pnyx_value)}
                        </span>
                      </div>
                      <div className="flex justify-between text-sm text-green-800">
                        <span>{assetSymbol}</span>
                        <span className="font-medium">
                          {formatPnyx(position.asset_value)}
                        </span>
                      </div>
                      <div className="text-xs text-green-700 mt-2">
                        Pool share:{' '}
                        {(position.share_of_pool_bps / 100).toFixed(2)}%
                      </div>
                    </div>
                  </div>
                )}
              </Card>
            );
          })}
        </div>
      </main>
    </div>
  );
}
