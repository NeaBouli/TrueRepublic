import { useEffect } from 'react';
import { useWalletStore } from '@/stores/walletStore';
import { Card } from '@/components/common/Card';
import { formatPnyx } from '@/utils/format';
import { ArrowPathIcon } from '@heroicons/react/24/outline';

export function BalanceCard() {
  const { balances, refreshBalance, isLoading } = useWalletStore();

  useEffect(() => {
    refreshBalance();

    const interval = setInterval(refreshBalance, 30000);
    return () => clearInterval(interval);
  }, [refreshBalance]);

  const pnyxBalance = balances.find((b) => b.denom === 'pnyx');
  const ibcBalances = balances.filter((b) => b.denom !== 'pnyx');

  return (
    <Card>
      <div className="flex items-center justify-between mb-6">
        <h3 className="text-lg font-semibold">Balances</h3>
        <button
          onClick={refreshBalance}
          disabled={isLoading}
          className="p-2 hover:bg-gray-100 rounded-lg transition-colors disabled:opacity-50"
          title="Refresh balances"
        >
          <ArrowPathIcon
            className={`h-5 w-5 text-gray-600 ${isLoading ? 'animate-spin' : ''}`}
          />
        </button>
      </div>

      {/* PNYX Balance */}
      <div className="bg-gradient-to-br from-primary-500 to-primary-700 rounded-xl p-6 text-white mb-4">
        <div className="text-sm opacity-90 mb-1">PNYX Balance</div>
        <div className="text-4xl font-bold mb-1">
          {formatPnyx(pnyxBalance?.amount || '0')}
        </div>
        <div className="text-sm opacity-75">PNYX</div>
      </div>

      {/* IBC Assets */}
      {ibcBalances.length > 0 && (
        <div className="space-y-2">
          <div className="text-sm font-medium text-gray-600 mb-2">
            Other Assets
          </div>
          {ibcBalances.map((balance) => (
            <div
              key={balance.denom}
              className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
            >
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 bg-gray-300 rounded-full flex items-center justify-center">
                  <span className="text-xs font-bold text-gray-600">
                    {balance.denom.slice(0, 3).toUpperCase()}
                  </span>
                </div>
                <div>
                  <div className="font-medium text-sm">
                    {balance.denom.startsWith('ibc/')
                      ? `IBC/${balance.denom.slice(4, 12)}...`
                      : balance.denom}
                  </div>
                </div>
              </div>
              <div className="text-right">
                <div className="font-semibold">{formatPnyx(balance.amount)}</div>
              </div>
            </div>
          ))}
        </div>
      )}

      {balances.length === 0 && !isLoading && (
        <div className="text-center py-8 text-gray-500">
          <p>No balances found</p>
          <p className="text-sm mt-1">Your wallet is empty</p>
        </div>
      )}
    </Card>
  );
}
