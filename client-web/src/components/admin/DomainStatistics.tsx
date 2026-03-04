import { useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useAdminStore } from '@/stores/adminStore';
import { Card } from '@/components/common/Card';
import { formatPnyx } from '@/utils/format';
import {
  UsersIcon,
  ChatBubbleLeftRightIcon,
  LightBulbIcon,
  BanknotesIcon,
  ShieldCheckIcon,
  KeyIcon,
} from '@heroicons/react/24/outline';

export function DomainStatistics() {
  const { domainId } = useParams<{ domainId: string }>();
  const { domainStats, loadDomainStats, isLoading } = useAdminStore();

  const stats = domainId ? domainStats[domainId] : null;

  useEffect(() => {
    if (domainId) {
      loadDomainStats(domainId);
    }
  }, [domainId, loadDomainStats]);

  if (isLoading && !stats) {
    return (
      <Card>
        <div className="text-center py-8 text-gray-500">
          Loading statistics...
        </div>
      </Card>
    );
  }

  if (!stats) {
    return (
      <Card>
        <div className="text-center py-8 text-gray-500">
          No statistics available
        </div>
      </Card>
    );
  }

  const statCards = [
    {
      label: 'Total Members',
      value: stats.totalMembers,
      icon: UsersIcon,
      color: 'bg-blue-100 text-blue-600',
    },
    {
      label: 'Total Issues',
      value: stats.totalIssues,
      icon: ChatBubbleLeftRightIcon,
      color: 'bg-green-100 text-green-600',
    },
    {
      label: 'Total Suggestions',
      value: stats.totalSuggestions,
      icon: LightBulbIcon,
      color: 'bg-yellow-100 text-yellow-600',
    },
    {
      label: 'Identity Commitments',
      value: stats.identityCommitments,
      icon: ShieldCheckIcon,
      color: 'bg-purple-100 text-purple-600',
    },
  ];

  return (
    <div className="space-y-6">
      {/* Treasury */}
      <Card>
        <div className="flex items-center gap-3 mb-2">
          <div className="p-3 bg-primary-100 rounded-lg">
            <BanknotesIcon className="h-6 w-6 text-primary-600" />
          </div>
          <div>
            <div className="text-sm text-gray-600">Treasury Balance</div>
            <div className="text-2xl font-bold">
              {formatPnyx(stats.treasuryBalance)} PNYX
            </div>
          </div>
        </div>
      </Card>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {statCards.map((stat) => {
          const Icon = stat.icon;
          return (
            <Card key={stat.label}>
              <div className="flex items-center gap-3">
                <div className={`p-3 rounded-lg ${stat.color}`}>
                  <Icon className="h-6 w-6" />
                </div>
                <div>
                  <div className="text-sm text-gray-600">{stat.label}</div>
                  <div className="text-2xl font-bold">{stat.value}</div>
                </div>
              </div>
            </Card>
          );
        })}
      </div>

      {/* ZKP Info */}
      <Card>
        <h3 className="text-lg font-semibold mb-4">ZKP / Anonymity State</h3>
        <div className="space-y-2 text-sm">
          <div className="flex justify-between">
            <span className="text-gray-600">Permission Register</span>
            <span className="font-medium">
              {stats.permissionRegCount} authorized keys
            </span>
          </div>
          <div className="flex justify-between">
            <span className="text-gray-600">Identity Commitments</span>
            <span className="font-medium">{stats.identityCommitments}</span>
          </div>
          <div className="flex justify-between items-start">
            <span className="text-gray-600">Merkle Root</span>
            <span className="font-mono text-xs text-gray-500 max-w-[200px] truncate">
              {stats.merkleRoot || 'Not set'}
            </span>
          </div>
        </div>
      </Card>
    </div>
  );
}
