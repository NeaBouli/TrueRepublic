import { useEffect } from 'react';
import { useNetworkStore } from '@/stores/networkStore';
import { Card } from '@/components/common/Card';
import {
  ServerIcon,
  CubeIcon,
  ClockIcon,
  SignalIcon,
} from '@heroicons/react/24/outline';

export function NetworkOverview() {
  const { networkInfo, validators, loadNetworkInfo } = useNetworkStore();

  useEffect(() => {
    loadNetworkInfo();

    const interval = setInterval(loadNetworkInfo, 10000);
    return () => clearInterval(interval);
  }, [loadNetworkInfo]);

  if (!networkInfo) {
    return (
      <Card>
        <div className="text-center py-8 text-gray-500">
          Loading network info...
        </div>
      </Card>
    );
  }

  const statCards = [
    {
      label: 'Latest Block',
      value: networkInfo.latestBlockHeight.toLocaleString(),
      icon: CubeIcon,
      color: 'bg-blue-100 text-blue-600',
    },
    {
      label: 'Validators',
      value: validators.length || networkInfo.totalValidators,
      icon: ServerIcon,
      color: 'bg-green-100 text-green-600',
    },
    {
      label: 'Chain ID',
      value: networkInfo.chainId,
      icon: SignalIcon,
      color: 'bg-purple-100 text-purple-600',
    },
    {
      label: 'Block Time',
      value: new Date(networkInfo.latestBlockTime).toLocaleTimeString(),
      icon: ClockIcon,
      color: 'bg-yellow-100 text-yellow-600',
    },
  ];

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      {statCards.map((stat) => {
        const Icon = stat.icon;
        return (
          <Card key={stat.label}>
            <div className="flex items-center gap-3">
              <div className={`p-3 rounded-lg ${stat.color}`}>
                <Icon className="h-6 w-6" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="text-sm text-gray-600">{stat.label}</div>
                <div className="text-lg font-bold truncate">{stat.value}</div>
              </div>
            </div>
          </Card>
        );
      })}
    </div>
  );
}
