import { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useMembershipStore } from '@/stores/membershipStore';
import { useGovernanceStore } from '@/stores/governanceStore';
import { Card } from '@/components/common/Card';
import { Button } from '@/components/common/Button';
import {
  EnvelopeIcon,
  UsersIcon,
  BanknotesIcon,
} from '@heroicons/react/24/outline';
import { formatPnyx } from '@/utils/format';

export function InviteHandler() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { currentInvite, parseInvite, clearInvite } = useMembershipStore();
  const { currentDomain, selectDomain } = useGovernanceStore();

  useEffect(() => {
    const inviteParam = searchParams.get('invite');
    const domainParam = searchParams.get('domain');

    if (inviteParam && domainParam) {
      const link = `truerepublic://join/domain/${domainParam}?invite=${inviteParam}`;
      parseInvite(link);
    }
  }, [searchParams, parseInvite]);

  useEffect(() => {
    if (currentInvite) {
      selectDomain(currentInvite.domainId);
    }
  }, [currentInvite, selectDomain]);

  const handleJoin = () => {
    if (currentInvite) {
      navigate(`/onboard/${currentInvite.domainId}`);
    }
  };

  const handleDecline = () => {
    clearInvite();
    navigate('/governance');
  };

  if (!currentInvite) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <Card className="max-w-md w-full text-center">
          <EnvelopeIcon className="h-16 w-16 text-gray-400 mx-auto mb-4" />
          <h2 className="text-2xl font-bold mb-2">Invalid Invite</h2>
          <p className="text-gray-600 mb-6">
            This invite link is not valid or has expired
          </p>
          <Button onClick={() => navigate('/governance')} className="w-full">
            Browse Domains
          </Button>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <Card className="max-w-md w-full">
        <div className="text-center mb-6">
          <EnvelopeIcon className="h-16 w-16 text-primary-600 mx-auto mb-4" />
          <h2 className="text-2xl font-bold mb-2">You're Invited!</h2>
          <p className="text-gray-600">
            Join {currentInvite.domainName}
          </p>
        </div>

        {currentDomain && (
          <div className="bg-gray-50 rounded-lg p-4 mb-6 space-y-2">
            <div className="flex items-center gap-2 text-sm">
              <UsersIcon className="h-4 w-4 text-gray-600" />
              <span className="text-gray-700">
                {currentDomain.memberCount} members
              </span>
            </div>
            <div className="flex items-center gap-2 text-sm">
              <BanknotesIcon className="h-4 w-4 text-gray-600" />
              <span className="text-gray-700">
                {formatPnyx(currentDomain.treasury)} PNYX treasury
              </span>
            </div>
          </div>
        )}

        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
          <h3 className="font-semibold text-blue-900 mb-2">
            What happens next?
          </h3>
          <ol className="text-sm text-blue-800 space-y-1">
            <li>1. Submit onboarding request (domain key registration)</li>
            <li>2. Admin verifies and approves you</li>
            <li>3. Register your ZKP identity commitment</li>
            <li>4. You can vote anonymously!</li>
          </ol>
        </div>

        <div className="space-y-3">
          <Button onClick={handleJoin} className="w-full">
            Join Domain
          </Button>
          <Button
            variant="secondary"
            onClick={handleDecline}
            className="w-full"
          >
            Decline
          </Button>
        </div>
      </Card>
    </div>
  );
}
