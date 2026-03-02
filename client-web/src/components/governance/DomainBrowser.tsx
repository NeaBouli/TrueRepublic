import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useGovernanceStore } from '@/stores/governanceStore';
import { Card } from '@/components/common/Card';
import { formatPnyx } from '@/utils/format';
import {
  UsersIcon,
  BanknotesIcon,
  ChevronRightIcon,
} from '@heroicons/react/24/outline';

export function DomainBrowser() {
  const navigate = useNavigate();
  const { domains, loadDomains, isLoading } = useGovernanceStore();

  useEffect(() => {
    loadDomains();
  }, [loadDomains]);

  const handleSelectDomain = (domainId: string) => {
    navigate(`/governance/domain/${domainId}`);
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-6xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold">Governance</h1>
              <p className="text-gray-600">
                Browse domains and participate in decisions
              </p>
            </div>
            <button
              onClick={() => navigate('/wallet')}
              className="text-sm text-primary-600 hover:text-primary-700"
            >
              Back to Wallet
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-4 py-8">
        <div className="mb-6">
          <h2 className="text-lg font-semibold mb-4">All Domains</h2>

          {isLoading && (
            <div className="text-center py-12 text-gray-500">
              Loading domains...
            </div>
          )}

          {!isLoading && domains.length === 0 && (
            <Card>
              <div className="text-center py-12 text-gray-500">
                <p className="mb-2">No domains found</p>
                <p className="text-sm">
                  Domains will appear here once created
                </p>
              </div>
            </Card>
          )}

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {domains.map((domain) => (
              <button
                key={domain.domainId}
                onClick={() => handleSelectDomain(domain.domainId)}
                className="text-left"
              >
                <Card className="cursor-pointer hover:shadow-lg transition-shadow h-full">
                  <div className="flex items-start justify-between mb-4">
                    <div>
                      <h3 className="text-lg font-semibold mb-1">
                        {domain.name}
                      </h3>
                      <div className="text-xs text-gray-500">
                        {domain.domainId}
                      </div>
                    </div>
                    <ChevronRightIcon className="h-5 w-5 text-gray-400" />
                  </div>

                  <div className="space-y-2">
                    <div className="flex items-center gap-2 text-sm text-gray-600">
                      <UsersIcon className="h-4 w-4" />
                      <span>{domain.memberCount} members</span>
                    </div>
                    <div className="flex items-center gap-2 text-sm text-gray-600">
                      <BanknotesIcon className="h-4 w-4" />
                      <span>{formatPnyx(domain.treasury)} PNYX</span>
                    </div>
                  </div>
                </Card>
              </button>
            ))}
          </div>
        </div>
      </main>
    </div>
  );
}
