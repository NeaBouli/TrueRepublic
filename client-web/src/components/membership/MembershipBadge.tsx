import { useEffect } from 'react';
import { useWalletStore } from '@/stores/walletStore';
import { useMembershipStore } from '@/stores/membershipStore';
import { useIdentityStore } from '@/stores/identityStore';
import {
  CheckCircleIcon,
  ClockIcon,
  XCircleIcon,
} from '@heroicons/react/24/outline';

interface MembershipBadgeProps {
  domainId: string;
}

export function MembershipBadge({ domainId }: MembershipBadgeProps) {
  const { currentWallet } = useWalletStore();
  const { memberships, loadMembership } = useMembershipStore();
  const { identity } = useIdentityStore();

  const membership = memberships[domainId];

  useEffect(() => {
    if (currentWallet && domainId) {
      loadMembership(domainId, currentWallet.address, identity?.commitment);
    }
  }, [domainId, currentWallet, identity?.commitment, loadMembership]);

  if (!membership || !membership.isMember) {
    return (
      <span className="inline-flex items-center gap-1 px-2 py-1 bg-gray-100 text-gray-700 text-xs font-medium rounded">
        <XCircleIcon className="h-4 w-4" />
        Not a Member
      </span>
    );
  }

  if (membership.isMember && membership.hasIdentityCommitment) {
    return (
      <span className="inline-flex items-center gap-1 px-2 py-1 bg-green-100 text-green-800 text-xs font-medium rounded">
        <CheckCircleIcon className="h-4 w-4" />
        Member
      </span>
    );
  }

  if (membership.hasIdentityCommitment === null) {
    return (
      <span className="inline-flex items-center gap-1 px-2 py-1 bg-gray-100 text-gray-700 text-xs font-medium rounded">
        <ClockIcon className="h-4 w-4" />
        Member · Identity Unknown
      </span>
    );
  }

  return (
    <span className="inline-flex items-center gap-1 px-2 py-1 bg-yellow-100 text-yellow-800 text-xs font-medium rounded animate-pulse">
      <ClockIcon className="h-4 w-4" />
      Pending Setup
    </span>
  );
}
