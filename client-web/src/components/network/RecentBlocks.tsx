import { useEffect } from 'react';
import { useNetworkStore } from '@/stores/networkStore';
import { Card } from '@/components/common/Card';
import { formatAddress } from '@/utils/format';
import { CubeIcon, ClockIcon } from '@heroicons/react/24/outline';

export function RecentBlocks() {
  const { recentBlocks, loadRecentBlocks, isLoading } = useNetworkStore();

  useEffect(() => {
    loadRecentBlocks();

    const interval = setInterval(loadRecentBlocks, 10000);
    return () => clearInterval(interval);
  }, [loadRecentBlocks]);

  return (
    <Card>
      <div className="flex items-center gap-2 mb-6">
        <CubeIcon className="h-6 w-6 text-primary-600" />
        <h2 className="text-xl font-bold">Recent Blocks</h2>
      </div>

      {isLoading && recentBlocks.length === 0 && (
        <div className="text-center py-8 text-gray-500">
          Loading blocks...
        </div>
      )}

      {!isLoading && recentBlocks.length === 0 && (
        <div className="text-center py-8 text-gray-500">No blocks found</div>
      )}

      <div className="space-y-2">
        {recentBlocks.map((block) => (
          <div
            key={block.height}
            className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
          >
            <div className="flex items-center gap-4">
              <div className="text-center min-w-[80px]">
                <div className="text-xs text-gray-500">Height</div>
                <div className="font-bold">{block.height.toLocaleString()}</div>
              </div>

              <div className="border-l border-gray-300 pl-4">
                <div className="text-sm text-gray-600">
                  Proposer: {formatAddress(block.proposer, 8)}
                </div>
                <div className="text-xs text-gray-400 font-mono">
                  {block.hash.substring(0, 16)}...
                </div>
              </div>
            </div>

            <div className="text-right">
              <div className="text-sm font-medium">
                {block.txCount} {block.txCount === 1 ? 'tx' : 'txs'}
              </div>
              <div className="flex items-center gap-1 text-xs text-gray-500">
                <ClockIcon className="h-3 w-3" />
                {new Date(block.time).toLocaleTimeString()}
              </div>
            </div>
          </div>
        ))}
      </div>
    </Card>
  );
}
