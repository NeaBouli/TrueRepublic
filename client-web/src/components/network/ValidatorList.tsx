import { useEffect } from 'react';
import { useNetworkStore } from '@/stores/networkStore';
import { Card } from '@/components/common/Card';
import { formatAddress, formatPnyx } from '@/utils/format';
import {
  ServerIcon,
  CheckCircleIcon,
  XCircleIcon,
} from '@heroicons/react/24/outline';

export function ValidatorList() {
  const { validators, loadValidators, isLoading } = useNetworkStore();

  useEffect(() => {
    loadValidators();
  }, [loadValidators]);

  return (
    <Card>
      <div className="flex items-center gap-2 mb-6">
        <ServerIcon className="h-6 w-6 text-primary-600" />
        <h2 className="text-xl font-bold">Validators</h2>
      </div>

      {isLoading && validators.length === 0 && (
        <div className="text-center py-8 text-gray-500">
          Loading validators...
        </div>
      )}

      {!isLoading && validators.length === 0 && (
        <div className="text-center py-8 text-gray-500">
          No validators found
        </div>
      )}

      <div className="space-y-3">
        {validators.map((validator) => {
          // Stake is sdk.Coins — find PNYX amount
          const pnyxStake = validator.stake?.find(
            (c) => c.denom === 'pnyx' || c.denom === 'upnyx'
          );
          const stakeDisplay = pnyxStake
            ? formatPnyx(pnyxStake.amount)
            : '0.00';

          return (
            <div
              key={validator.operator_addr}
              className="flex items-center justify-between p-4 bg-gray-50 rounded-lg"
            >
              <div className="flex-1">
                <div className="flex items-center gap-2 mb-1">
                  <div className="font-medium font-mono text-sm">
                    {formatAddress(validator.operator_addr, 16)}
                  </div>
                  {validator.jailed ? (
                    <span className="px-2 py-0.5 bg-red-100 text-red-800 text-xs font-medium rounded">
                      Jailed
                    </span>
                  ) : (
                    <span className="px-2 py-0.5 bg-green-100 text-green-800 text-xs font-medium rounded">
                      Active
                    </span>
                  )}
                </div>

                <div className="flex items-center gap-4 text-sm text-gray-600">
                  <span>Stake: {stakeDisplay} PNYX</span>
                  <span>Power: {validator.power}</span>
                  <span>Domains: {validator.domains?.length || 0}</span>
                  {validator.missed_blocks > 0 && (
                    <span className="text-yellow-600">
                      Missed: {validator.missed_blocks}
                    </span>
                  )}
                </div>
              </div>

              <div className="flex-shrink-0 ml-4">
                {validator.jailed ? (
                  <XCircleIcon className="h-8 w-8 text-red-400" />
                ) : (
                  <CheckCircleIcon className="h-8 w-8 text-green-500" />
                )}
              </div>
            </div>
          );
        })}
      </div>
    </Card>
  );
}
