import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDEXStore } from '@/stores/dexStore';
import { Card } from '@/components/common/Card';
import { formatPnyx } from '@/utils/format';
import {
  ArrowsRightLeftIcon,
  ChartBarIcon,
} from '@heroicons/react/24/outline';

export function PoolList() {
  const navigate = useNavigate();
  const { pools, loadPools, isLoading } = useDEXStore();

  useEffect(() => {
    loadPools();
  }, [loadPools]);

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-6xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold">DEX</h1>
              <p className="text-gray-600">
                Trade tokens and provide liquidity
              </p>
            </div>
            <div className="flex items-center gap-3">
              <button
                onClick={() => navigate('/dex/swap')}
                className="btn btn-primary flex items-center gap-2"
              >
                <ArrowsRightLeftIcon className="h-5 w-5" />
                Swap
              </button>
              <button
                onClick={() => navigate('/wallet')}
                className="text-sm text-primary-600 hover:text-primary-700"
              >
                Back to Wallet
              </button>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-4 py-8">
        <div className="mb-6">
          <h2 className="text-lg font-semibold mb-4">Liquidity Pools</h2>

          {isLoading && (
            <div className="text-center py-12 text-gray-500">
              Loading pools...
            </div>
          )}

          {!isLoading && pools.length === 0 && (
            <Card>
              <div className="text-center py-12 text-gray-500">
                <ChartBarIcon className="h-12 w-12 mx-auto mb-3 text-gray-400" />
                <p className="mb-2">No pools yet</p>
                <p className="text-sm">
                  Pools will appear here once created
                </p>
              </div>
            </Card>
          )}

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {pools.map((pool) => (
              <button
                key={pool.asset_denom}
                onClick={() => navigate(`/dex/pool/${pool.asset_denom}`)}
                className="text-left"
              >
                <Card className="cursor-pointer hover:shadow-lg transition-shadow h-full">
                  <div className="mb-4">
                    <div className="flex items-center gap-2 mb-2">
                      <div className="font-bold text-lg">
                        PNYX / {pool.asset_symbol || pool.asset_denom.toUpperCase()}
                      </div>
                    </div>
                    <div className="text-xs text-gray-500">
                      {pool.asset_denom}
                    </div>
                  </div>

                  <div className="space-y-2">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-gray-600">PNYX</span>
                      <span className="font-medium">
                        {formatPnyx(pool.pnyx_reserve)}
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-gray-600">
                        {pool.asset_symbol || pool.asset_denom.toUpperCase()}
                      </span>
                      <span className="font-medium">
                        {formatPnyx(pool.asset_reserve)}
                      </span>
                    </div>
                  </div>

                  <div className="mt-4 pt-4 border-t border-gray-200">
                    <div className="flex items-center justify-between text-xs text-gray-500">
                      <span>Volume</span>
                      <span>{formatPnyx(pool.total_volume_pnyx)} PNYX</span>
                    </div>
                    <div className="flex items-center justify-between text-xs text-gray-500 mt-1">
                      <span>Swaps</span>
                      <span>{pool.swap_count}</span>
                    </div>
                  </div>
                </Card>
              </button>
            ))}
          </div>
        </div>
      </main>
    </div>
  );
}
