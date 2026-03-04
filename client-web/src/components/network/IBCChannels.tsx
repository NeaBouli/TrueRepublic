import { useEffect } from 'react';
import { useNetworkStore } from '@/stores/networkStore';
import { Card } from '@/components/common/Card';
import {
  ArrowsRightLeftIcon,
  CheckCircleIcon,
  XCircleIcon,
} from '@heroicons/react/24/outline';

export function IBCChannels() {
  const { ibcChannels, loadIBCChannels, isLoading } = useNetworkStore();

  useEffect(() => {
    loadIBCChannels();
  }, [loadIBCChannels]);

  const getStateColor = (state: string) => {
    switch (state) {
      case 'STATE_OPEN':
        return 'bg-green-100 text-green-800';
      case 'STATE_CLOSED':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <Card>
      <div className="flex items-center gap-2 mb-6">
        <ArrowsRightLeftIcon className="h-6 w-6 text-primary-600" />
        <h2 className="text-xl font-bold">IBC Channels</h2>
      </div>

      {isLoading && ibcChannels.length === 0 && (
        <div className="text-center py-8 text-gray-500">
          Loading IBC channels...
        </div>
      )}

      {!isLoading && ibcChannels.length === 0 && (
        <div className="text-center py-8 text-gray-500">
          No IBC channels configured
        </div>
      )}

      <div className="space-y-3">
        {ibcChannels.map((channel) => {
          const isOpen = channel.state === 'STATE_OPEN';

          return (
            <div
              key={`${channel.port_id}-${channel.channel_id}`}
              className="flex items-center justify-between p-4 bg-gray-50 rounded-lg"
            >
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-2">
                  <div className="font-medium">
                    {channel.channel_id} ({channel.port_id})
                  </div>
                  <span
                    className={`px-2 py-0.5 text-xs font-medium rounded ${getStateColor(channel.state)}`}
                  >
                    {channel.state.replace('STATE_', '')}
                  </span>
                </div>

                <div className="text-sm text-gray-600">
                  <div>
                    Counterparty: {channel.counterparty.channel_id} (
                    {channel.counterparty.port_id})
                  </div>
                  {channel.connection_hops?.length > 0 && (
                    <div className="text-xs text-gray-500 mt-1">
                      Connection: {channel.connection_hops[0]}
                    </div>
                  )}
                </div>
              </div>

              <div className="flex-shrink-0 ml-4">
                {isOpen ? (
                  <CheckCircleIcon className="h-8 w-8 text-green-500" />
                ) : (
                  <XCircleIcon className="h-8 w-8 text-red-400" />
                )}
              </div>
            </div>
          );
        })}
      </div>
    </Card>
  );
}
