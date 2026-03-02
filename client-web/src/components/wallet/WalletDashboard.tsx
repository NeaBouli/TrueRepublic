import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useWalletStore } from '@/stores/walletStore';
import { AccountInfo } from './AccountInfo';
import { BalanceCard } from './BalanceCard';
import { Button } from '@/components/common/Button';
import {
  PaperAirplaneIcon,
  ArrowDownTrayIcon,
  Cog6ToothIcon,
} from '@heroicons/react/24/outline';

export function WalletDashboard() {
  const navigate = useNavigate();
  const { currentWallet, isLocked } = useWalletStore();

  useEffect(() => {
    if (isLocked || !currentWallet) {
      navigate('/unlock');
    }
  }, [isLocked, currentWallet, navigate]);

  if (!currentWallet) return null;

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-4xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <img
                src="https://raw.githubusercontent.com/NeaBouli/TrueRepublic/main/assets/logo.png"
                alt="TrueRepublic"
                className="h-10"
              />
              <h1 className="text-xl font-bold text-gray-900">TrueRepublic</h1>
            </div>
            <button
              onClick={() => navigate('/settings')}
              className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
            >
              <Cog6ToothIcon className="h-6 w-6 text-gray-600" />
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-4xl mx-auto px-4 py-8">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {/* Left Column - Account Info */}
          <div className="md:col-span-1">
            <AccountInfo />

            {/* Quick Actions */}
            <div className="mt-6 space-y-3">
              <Button
                onClick={() => navigate('/send')}
                className="w-full flex items-center justify-center gap-2"
              >
                <PaperAirplaneIcon className="h-5 w-5" />
                Send
              </Button>
              <Button
                variant="secondary"
                className="w-full flex items-center justify-center gap-2"
              >
                <ArrowDownTrayIcon className="h-5 w-5" />
                Receive
              </Button>
            </div>
          </div>

          {/* Right Column - Balances */}
          <div className="md:col-span-2">
            <BalanceCard />

            {/* Coming Soon */}
            <div className="mt-6 grid grid-cols-2 gap-4">
              <div className="bg-white rounded-xl p-6 text-center border-2 border-dashed border-gray-300">
                <div className="text-gray-400 text-2xl mb-2">&#x1F5F3;</div>
                <div className="text-sm font-medium text-gray-600">Governance</div>
                <div className="text-xs text-gray-500 mt-1">Coming Soon</div>
              </div>
              <div className="bg-white rounded-xl p-6 text-center border-2 border-dashed border-gray-300">
                <div className="text-gray-400 text-2xl mb-2">&#x1F4B1;</div>
                <div className="text-sm font-medium text-gray-600">DEX</div>
                <div className="text-xs text-gray-500 mt-1">Coming Soon</div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
