import { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useWalletStore } from '@/stores/walletStore';
import {
  Bars3Icon,
  XMarkIcon,
  WalletIcon,
  ArrowsRightLeftIcon,
  ChatBubbleLeftRightIcon,
  GlobeAltIcon,
  ShieldCheckIcon,
} from '@heroicons/react/24/outline';

export function MobileNav() {
  const navigate = useNavigate();
  const location = useLocation();
  const [isOpen, setIsOpen] = useState(false);
  const { currentWallet } = useWalletStore();

  if (!currentWallet) return null;

  const navItems = [
    { path: '/wallet', label: 'Wallet', icon: WalletIcon },
    { path: '/dex', label: 'DEX', icon: ArrowsRightLeftIcon },
    { path: '/governance', label: 'Governance', icon: ChatBubbleLeftRightIcon },
    { path: '/network', label: 'Network', icon: GlobeAltIcon },
    { path: '/identity', label: 'Identity', icon: ShieldCheckIcon },
  ];

  return (
    <>
      {/* Mobile FAB */}
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="lg:hidden fixed bottom-4 right-4 z-50 p-4 bg-primary-600 text-white rounded-full shadow-lg hover:bg-primary-700 transition-colors"
        aria-label="Navigation menu"
      >
        {isOpen ? (
          <XMarkIcon className="h-6 w-6" />
        ) : (
          <Bars3Icon className="h-6 w-6" />
        )}
      </button>

      {/* Overlay */}
      {isOpen && (
        <>
          <div
            className="lg:hidden fixed inset-0 bg-black bg-opacity-50 z-40"
            onClick={() => setIsOpen(false)}
          />
          <div className="lg:hidden fixed bottom-0 left-0 right-0 bg-white rounded-t-2xl shadow-2xl z-40 p-6 max-h-[80vh] overflow-y-auto">
            <h3 className="text-lg font-bold mb-4">Navigation</h3>
            <nav className="space-y-2">
              {navItems.map((item) => {
                const Icon = item.icon;
                const isActive = location.pathname.startsWith(item.path);

                return (
                  <button
                    key={item.path}
                    onClick={() => {
                      navigate(item.path);
                      setIsOpen(false);
                    }}
                    className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg transition-colors ${
                      isActive
                        ? 'bg-primary-600 text-white'
                        : 'bg-gray-50 text-gray-700 hover:bg-gray-100'
                    }`}
                  >
                    <Icon className="h-5 w-5" />
                    <span className="font-medium">{item.label}</span>
                  </button>
                );
              })}
            </nav>
          </div>
        </>
      )}
    </>
  );
}
