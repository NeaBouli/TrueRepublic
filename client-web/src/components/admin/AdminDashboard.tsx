import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useWalletStore } from '@/stores/walletStore';
import { useAdminStore } from '@/stores/adminStore';
import { Card } from '@/components/common/Card';
import { MemberManagement } from './MemberManagement';
import { DomainStatistics } from './DomainStatistics';
import {
  ArrowLeftIcon,
  ShieldCheckIcon,
  LinkIcon,
  ClipboardDocumentIcon,
} from '@heroicons/react/24/outline';

export function AdminDashboard() {
  const navigate = useNavigate();
  const { domainId } = useParams<{ domainId: string }>();
  const { currentWallet } = useWalletStore();
  const { isAdmin, checkAdmin } = useAdminStore();

  const [activeTab, setActiveTab] = useState<'stats' | 'members'>('stats');
  const [inviteLink, setInviteLink] = useState('');
  const [copied, setCopied] = useState(false);

  const isUserAdmin = domainId ? isAdmin[domainId] : false;

  useEffect(() => {
    if (domainId && currentWallet) {
      checkAdmin(domainId, currentWallet.address);
    }
  }, [domainId, currentWallet, checkAdmin]);

  const handleGenerateInvite = () => {
    if (!domainId) return;
    const code = Math.random().toString(36).substring(2, 15);
    setInviteLink(`truerepublic://join/domain/${domainId}?invite=${code}`);
    setCopied(false);
  };

  const handleCopyInvite = async () => {
    if (inviteLink) {
      await navigator.clipboard.writeText(inviteLink);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  if (!currentWallet) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <Card className="max-w-md w-full text-center">
          <p className="text-gray-600 mb-4">Please connect your wallet</p>
        </Card>
      </div>
    );
  }

  if (isUserAdmin === false) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <Card className="max-w-md w-full text-center">
          <ShieldCheckIcon className="h-16 w-16 text-red-400 mx-auto mb-4" />
          <h2 className="text-2xl font-bold mb-2">Access Denied</h2>
          <p className="text-gray-600 mb-6">
            You are not an admin of this domain
          </p>
          <button
            onClick={() => navigate('/governance')}
            className="btn btn-primary w-full"
          >
            Back to Domains
          </button>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-6xl mx-auto px-4 py-4">
          <button
            onClick={() => navigate(`/governance/domain/${domainId}`)}
            className="flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-4"
          >
            <ArrowLeftIcon className="h-5 w-5" />
            Back to Domain
          </button>

          <div className="flex items-center justify-between">
            <div>
              <div className="flex items-center gap-2">
                <ShieldCheckIcon className="h-8 w-8 text-primary-600" />
                <h1 className="text-2xl font-bold">Domain Admin</h1>
              </div>
              <p className="text-gray-600 mt-1">Domain: {domainId}</p>
            </div>

            <button
              onClick={handleGenerateInvite}
              className="flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
            >
              <LinkIcon className="h-5 w-5" />
              Generate Invite
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-4 py-8">
        {/* Invite Link */}
        {inviteLink && (
          <Card className="mb-6">
            <div className="flex items-center justify-between">
              <div className="flex-1 mr-4">
                <div className="text-sm text-gray-600 mb-1">Invite Link</div>
                <code className="text-xs font-mono break-all">
                  {inviteLink}
                </code>
              </div>
              <button
                onClick={handleCopyInvite}
                className="flex items-center gap-1.5 px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition-colors"
              >
                <ClipboardDocumentIcon className="h-4 w-4" />
                {copied ? 'Copied!' : 'Copy'}
              </button>
            </div>
          </Card>
        )}

        {/* Tabs */}
        <div className="flex gap-4 mb-6">
          <button
            onClick={() => setActiveTab('stats')}
            className={`px-4 py-2 rounded-lg font-medium transition-colors ${
              activeTab === 'stats'
                ? 'bg-primary-600 text-white'
                : 'bg-white text-gray-700 hover:bg-gray-50'
            }`}
          >
            Statistics
          </button>
          <button
            onClick={() => setActiveTab('members')}
            className={`px-4 py-2 rounded-lg font-medium transition-colors ${
              activeTab === 'members'
                ? 'bg-primary-600 text-white'
                : 'bg-white text-gray-700 hover:bg-gray-50'
            }`}
          >
            Members
          </button>
        </div>

        {/* Content */}
        {activeTab === 'stats' && <DomainStatistics />}
        {activeTab === 'members' && <MemberManagement />}
      </main>
    </div>
  );
}
