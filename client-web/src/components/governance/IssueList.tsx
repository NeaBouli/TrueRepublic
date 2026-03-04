import { useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useGovernanceStore } from '@/stores/governanceStore';
import { useMembershipStore } from '@/stores/membershipStore';
import { useWalletStore } from '@/stores/walletStore';
import { useAdminStore } from '@/stores/adminStore';
import { Card } from '@/components/common/Card';
import { MembershipBadge } from '@/components/membership/MembershipBadge';
import {
  ArrowLeftIcon,
  ChatBubbleLeftRightIcon,
  ChevronRightIcon,
  ShieldCheckIcon,
} from '@heroicons/react/24/outline';

export function IssueList() {
  const navigate = useNavigate();
  const { domainId } = useParams<{ domainId: string }>();
  const { currentDomain, issues, selectDomain, isLoading } =
    useGovernanceStore();
  const { memberships } = useMembershipStore();
  const { currentWallet } = useWalletStore();
  const { isAdmin, checkAdmin } = useAdminStore();

  useEffect(() => {
    if (domainId) {
      selectDomain(domainId);
    }
  }, [domainId, selectDomain]);

  useEffect(() => {
    if (domainId && currentWallet) {
      checkAdmin(domainId, currentWallet.address);
    }
  }, [domainId, currentWallet, checkAdmin]);

  const handleSelectIssue = (issueId: string) => {
    navigate(`/governance/domain/${domainId}/issue/${issueId}`);
  };

  if (isLoading && !currentDomain) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-500">Loading...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-6xl mx-auto px-4 py-4">
          <button
            onClick={() => navigate('/governance')}
            className="flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-4"
          >
            <ArrowLeftIcon className="h-5 w-5" />
            Back to Domains
          </button>

          {currentDomain && (
            <div>
              <div className="flex items-center justify-between mb-2">
                <h1 className="text-2xl font-bold">{currentDomain.name}</h1>
                <div className="flex items-center gap-3">
                  <MembershipBadge domainId={domainId!} />
                  {domainId && isAdmin[domainId] && (
                    <button
                      onClick={() => navigate(`/admin/domain/${domainId}`)}
                      className="flex items-center gap-1.5 px-3 py-1 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 transition-colors"
                    >
                      <ShieldCheckIcon className="h-4 w-4" />
                      Admin
                    </button>
                  )}
                </div>
              </div>
              <div className="flex items-center gap-4">
                <p className="text-gray-600">
                  {currentDomain.memberCount} members
                </p>
                {domainId &&
                  !memberships[domainId]?.isMember && (
                    <button
                      onClick={() => navigate(`/onboard/${domainId}`)}
                      className="text-sm text-primary-600 hover:text-primary-700 font-medium"
                    >
                      Join Domain
                    </button>
                  )}
              </div>
            </div>
          )}
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-4 py-8">
        <div className="mb-6">
          <h2 className="text-lg font-semibold mb-4">Active Issues</h2>

          {isLoading && (
            <div className="text-center py-12 text-gray-500">
              Loading issues...
            </div>
          )}

          {!isLoading && issues.length === 0 && (
            <Card>
              <div className="text-center py-12 text-gray-500">
                <ChatBubbleLeftRightIcon className="h-12 w-12 mx-auto mb-3 text-gray-400" />
                <p className="mb-2">No issues yet</p>
                <p className="text-sm">
                  Issues will appear here once created
                </p>
              </div>
            </Card>
          )}

          <div className="space-y-4">
            {issues.map((issue) => (
              <button
                key={issue.issueId}
                onClick={() => handleSelectIssue(issue.issueId)}
                className="w-full text-left"
              >
                <Card className="cursor-pointer hover:shadow-lg transition-shadow">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <h3 className="text-lg font-semibold">
                          {issue.title}
                        </h3>
                        {issue.status === 'active' ? (
                          <span className="px-2 py-1 bg-green-100 text-green-800 text-xs font-medium rounded">
                            Active
                          </span>
                        ) : (
                          <span className="px-2 py-1 bg-gray-100 text-gray-800 text-xs font-medium rounded">
                            Closed
                          </span>
                        )}
                      </div>

                      <p className="text-gray-600 text-sm mb-3">
                        {issue.description}
                      </p>

                      <div className="text-xs text-gray-500">
                        Created{' '}
                        {new Date(issue.createdAt).toLocaleDateString()}
                      </div>
                    </div>

                    <ChevronRightIcon className="h-5 w-5 text-gray-400 flex-shrink-0 ml-4" />
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
