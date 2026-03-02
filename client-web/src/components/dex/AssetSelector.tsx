import { useState, useEffect } from 'react';
import { useDEXStore } from '@/stores/dexStore';
import { ChevronDownIcon, CheckIcon } from '@heroicons/react/24/outline';

interface AssetSelectorProps {
  selected: string | null;
  onSelect: (denom: string) => void;
  exclude?: string;
  label?: string;
}

export function AssetSelector({
  selected,
  onSelect,
  exclude,
  label = 'Select Asset',
}: AssetSelectorProps) {
  const { assets, loadAssets } = useDEXStore();
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    if (assets.length === 0) {
      loadAssets();
    }
  }, [assets.length, loadAssets]);

  const selectedAsset = assets.find((a) => a.ibc_denom === selected);
  const availableAssets = assets.filter(
    (a) => a.trading_enabled && (!exclude || a.ibc_denom !== exclude)
  );

  return (
    <div className="relative">
      <label className="block text-sm font-medium text-gray-700 mb-1">
        {label}
      </label>

      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="w-full flex items-center justify-between px-4 py-3 bg-white border border-gray-300 rounded-lg hover:border-gray-400 transition-colors"
      >
        {selectedAsset ? (
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 bg-primary-100 rounded-full flex items-center justify-center">
              <span className="font-bold text-primary-700 text-xs">
                {selectedAsset.symbol.slice(0, 4)}
              </span>
            </div>
            <div className="text-left">
              <div className="font-medium">{selectedAsset.symbol}</div>
              <div className="text-xs text-gray-500">{selectedAsset.name}</div>
            </div>
          </div>
        ) : (
          <span className="text-gray-500">Select an asset</span>
        )}
        <ChevronDownIcon className="h-5 w-5 text-gray-400" />
      </button>

      {isOpen && (
        <>
          <div
            className="fixed inset-0 z-10"
            onClick={() => setIsOpen(false)}
          />
          <div className="absolute z-20 w-full mt-2 bg-white border border-gray-200 rounded-lg shadow-lg max-h-64 overflow-y-auto">
            {availableAssets.map((asset) => (
              <button
                key={asset.ibc_denom}
                type="button"
                onClick={() => {
                  onSelect(asset.ibc_denom);
                  setIsOpen(false);
                }}
                className="w-full flex items-center justify-between px-4 py-3 hover:bg-gray-50 transition-colors"
              >
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 bg-primary-100 rounded-full flex items-center justify-center">
                    <span className="font-bold text-primary-700 text-xs">
                      {asset.symbol.slice(0, 4)}
                    </span>
                  </div>
                  <div className="text-left">
                    <div className="font-medium">{asset.symbol}</div>
                    <div className="text-xs text-gray-500">{asset.name}</div>
                  </div>
                </div>
                {selected === asset.ibc_denom && (
                  <CheckIcon className="h-5 w-5 text-primary-600" />
                )}
              </button>
            ))}
            {availableAssets.length === 0 && (
              <div className="px-4 py-8 text-center text-gray-500 text-sm">
                No assets available
              </div>
            )}
          </div>
        </>
      )}
    </div>
  );
}
