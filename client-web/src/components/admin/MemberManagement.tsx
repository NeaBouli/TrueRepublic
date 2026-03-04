import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { useWalletStore } from '@/stores/walletStore';
import { useAdminStore } from '@/stores/adminStore';
import { AdminService } from '@/services/admin';
import { WalletService } from '@/services/wallet';
import { Card } from '@/components/common/Card';
import { Button } from '@/components/common/Button';
import { Input } from '@/components/common/Input';
import { DEFAULT_CHAIN } from '@/config/chains';
import { formatAddress } from '@/utils/format';
import {
  CheckCircleIcon,
  UserPlusIcon,
  UsersIcon,
} from '@heroicons/react/24/outline';

export function MemberManagement() {
  const { domainId } = useParams<{ domainId: string }>();
  const { currentWallet, password } = useWalletStore();
  const { domainMembers, loadDomainMembers } = useAdminStore();

  const [newMemberAddress, setNewMemberAddress] = useState('');
  const [isAdding, setIsAdding] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  const members = domainId ? domainMembers[domainId] || [] : [];

  useEffect(() => {
    if (domainId) {
      loadDomainMembers(domainId);
    }
  }, [domainId, loadDomainMembers]);

  const handleAddMember = async () => {
    if (!currentWallet || !password || !domainId || !newMemberAddress) return;

    setIsAdding(true);
    setError('');
    setSuccess('');

    try {
      const adminService = new AdminService(DEFAULT_CHAIN);
      const wallet = await WalletService.getWalletForSigning(
        currentWallet.address,
        password
      );

      const result = await adminService.addMember(wallet, {
        domain_name: domainId,
        new_member: newMemberAddress,
      });

      if (!result.success) {
        throw new Error(result.error || 'Failed to add member');
      }

      setSuccess('Member added successfully!');
      setNewMemberAddress('');

      // Reload members
      setTimeout(() => {
        loadDomainMembers(domainId);
      }, 1500);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to add member');
    } finally {
      setIsAdding(false);
    }
  };

  return (
    <div className="space-y-6">
      {/* Add Member Form */}
      <Card>
        <div className="flex items-center gap-2 mb-4">
          <UserPlusIcon className="h-6 w-6 text-primary-600" />
          <h3 className="text-lg font-semibold">Add Member</h3>
        </div>

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4">
            <p className="text-sm text-red-800">{error}</p>
          </div>
        )}

        {success && (
          <div className="bg-green-50 border border-green-200 rounded-lg p-3 mb-4">
            <div className="flex items-center gap-2">
              <CheckCircleIcon className="h-5 w-5 text-green-600" />
              <p className="text-sm text-green-800">{success}</p>
            </div>
          </div>
        )}

        <div className="space-y-4">
          <Input
            label="Member Address"
            value={newMemberAddress}
            onChange={(e) => setNewMemberAddress(e.target.value)}
            placeholder="true1..."
          />

          <Button
            onClick={handleAddMember}
            isLoading={isAdding}
            disabled={!newMemberAddress}
            className="w-full"
          >
            Add Member
          </Button>
        </div>
      </Card>

      {/* Member List */}
      <Card>
        <div className="flex items-center gap-2 mb-4">
          <UsersIcon className="h-6 w-6 text-primary-600" />
          <h3 className="text-lg font-semibold">
            Domain Members ({members.length})
          </h3>
        </div>

        {members.length === 0 && (
          <div className="text-center py-8 text-gray-500">No members yet</div>
        )}

        <div className="space-y-2">
          {members.map((member) => (
            <div
              key={member.address}
              className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
            >
              <div className="flex-1">
                <div className="font-medium text-sm font-mono">
                  {formatAddress(member.address, 12)}
                </div>
              </div>
              <div className="flex items-center gap-2">
                {member.inPermissionReg && (
                  <span className="px-2 py-1 bg-green-100 text-green-800 text-xs font-medium rounded">
                    Authorized
                  </span>
                )}
                {member.hasIdentityCommitment && (
                  <span className="px-2 py-1 bg-blue-100 text-blue-800 text-xs font-medium rounded">
                    Identity
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}
