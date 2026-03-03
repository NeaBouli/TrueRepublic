import { useState } from 'react';
import { useWalletStore } from '@/stores/walletStore';
import { GovernanceTxService } from '@/services/governanceTx';
import { WalletService } from '@/services/wallet';
import { DEFAULT_CHAIN } from '@/config/chains';
import { CheckCircleIcon } from '@heroicons/react/24/outline';

interface StonePlacementProps {
  domainName: string;
  issueName: string;
  suggestionName: string;
  currentStones: {
    green: number;
    yellow: number;
    red: number;
  };
  zone: string;
  onPlaced?: () => void;
}

/**
 * Stone placement UI.
 * Go MsgPlaceStoneOnSuggestion has no color field — stones are a single
 * endorsement counter. The zone (green/yellow/red) is determined by the
 * lifecycle system based on stone count vs member threshold.
 */
export function StonePlacement({
  domainName,
  issueName,
  suggestionName,
  currentStones,
  zone,
  onPlaced,
}: StonePlacementProps) {
  const { currentWallet, password } = useWalletStore();
  const [isPlacing, setIsPlacing] = useState(false);
  const [txHash, setTxHash] = useState('');
  const [error, setError] = useState('');

  const handlePlaceStone = async () => {
    if (!currentWallet || !password) return;

    setIsPlacing(true);
    setError('');

    try {
      const service = new GovernanceTxService(DEFAULT_CHAIN);
      const wallet = await WalletService.getWalletForSigning(
        currentWallet.address,
        password
      );

      const result = await service.placeStoneOnSuggestion(
        wallet,
        domainName,
        issueName,
        suggestionName
      );

      if (!result.success) {
        throw new Error(result.error || 'Stone placement failed');
      }

      setTxHash(result.hash);

      if (onPlaced) {
        setTimeout(onPlaced, 1500);
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Stone placement failed');
    } finally {
      setIsPlacing(false);
    }
  };

  if (txHash) {
    return (
      <div className="bg-green-50 border border-green-200 rounded-lg p-4">
        <div className="flex items-center gap-2">
          <CheckCircleIcon className="h-5 w-5 text-green-600" />
          <span className="font-medium text-green-900">Stone Placed!</span>
        </div>
        <p className="text-sm text-green-800 mt-1">
          Your endorsement has been recorded
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-3">
          <p className="text-sm text-red-800">{error}</p>
        </div>
      )}

      {/* Current stone counts */}
      <div>
        <div className="text-sm font-medium text-gray-700 mb-3">
          Current Stones
        </div>
        <div className="grid grid-cols-3 gap-3">
          <div className="flex flex-col items-center gap-2 p-3 bg-green-50 border border-green-200 rounded-lg">
            <div className="w-6 h-6 bg-green-500 rounded-full" />
            <div className="text-lg font-bold text-green-900">
              {currentStones.green}
            </div>
            <div className="text-xs text-green-700">Green</div>
          </div>
          <div className="flex flex-col items-center gap-2 p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
            <div className="w-6 h-6 bg-yellow-500 rounded-full" />
            <div className="text-lg font-bold text-yellow-900">
              {currentStones.yellow}
            </div>
            <div className="text-xs text-yellow-700">Yellow</div>
          </div>
          <div className="flex flex-col items-center gap-2 p-3 bg-red-50 border border-red-200 rounded-lg">
            <div className="w-6 h-6 bg-red-500 rounded-full" />
            <div className="text-lg font-bold text-red-900">
              {currentStones.red}
            </div>
            <div className="text-xs text-red-700">Red</div>
          </div>
        </div>
      </div>

      {/* Zone info */}
      <div className="bg-gray-50 rounded-lg p-3 text-sm text-gray-600">
        <span className="font-medium">Zone:</span>{' '}
        {zone.charAt(0).toUpperCase() + zone.slice(1)} — determined by stone
        count relative to domain member threshold
      </div>

      {/* Place stone button */}
      <button
        onClick={handlePlaceStone}
        disabled={isPlacing}
        className="w-full px-4 py-3 bg-primary-600 text-white font-medium rounded-lg hover:bg-primary-700 transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
      >
        {isPlacing ? (
          <>
            <svg
              className="animate-spin h-5 w-5"
              viewBox="0 0 24 24"
              fill="none"
            >
              <circle
                className="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                strokeWidth="4"
              />
              <path
                className="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
              />
            </svg>
            Placing Stone...
          </>
        ) : (
          'Place Stone'
        )}
      </button>

      <p className="text-xs text-gray-500 text-center">
        Placing a stone endorses this suggestion. Earn rewards via VoteToEarn.
      </p>
    </div>
  );
}
