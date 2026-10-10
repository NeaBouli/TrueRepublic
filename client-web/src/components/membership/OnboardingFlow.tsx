import { useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useWalletStore } from '@/stores/walletStore';
import { IDENTITY_CREATION_DISABLED, useIdentityStore } from '@/stores/identityStore';
import { useMembershipStore } from '@/stores/membershipStore';
import { PREVIEW_IDENTITY_REGISTRATION_DISABLED } from '@/services/membership';
import { Card } from '@/components/common/Card';

import { Button } from '@/components/common/Button';
import {
  ArrowLeftIcon,
  CheckCircleIcon,
  ClockIcon,
  XCircleIcon,
  ShieldExclamationIcon,
} from '@heroicons/react/24/outline';

export function OnboardingFlow() {
  const navigate = useNavigate();
  const { domainId } = useParams<{ domainId: string }>();
  const { currentWallet } = useWalletStore();
  const { identity, hasIdentity } = useIdentityStore();
  const { memberships, loadMembership } = useMembershipStore();

  const membership = domainId ? memberships[domainId] : null;
  const step = !hasIdentity
    ? 'identity'
    : membership?.isMember && membership.hasIdentityCommitment
      ? 'complete'
      : membership?.isMember
        ? 'submit'
        : 'waiting';

  useEffect(() => {
    if (!domainId || !currentWallet) return;
    loadMembership(domainId, currentWallet.address, identity?.commitment);
  }, [domainId, currentWallet, identity?.commitment, loadMembership]);

  // Poll for membership approval
  useEffect(() => {
    if (step !== 'waiting' || !domainId || !currentWallet) return;

    const interval = setInterval(() => {
      loadMembership(domainId, currentWallet.address, identity?.commitment);
    }, 5000);

    return () => clearInterval(interval);
  }, [step, domainId, currentWallet, identity?.commitment, loadMembership]);

  if (!domainId) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <Card className="max-w-md w-full text-center">
          <XCircleIcon className="h-16 w-16 text-red-600 mx-auto mb-4" />
          <h2 className="text-2xl font-bold mb-2">Invalid Domain</h2>
          <Button
            onClick={() => navigate('/governance')}
            className="w-full mt-4"
          >
            Browse Domains
          </Button>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-2xl mx-auto px-4 py-4">
          <button
            onClick={() => navigate(-1)}
            className="flex items-center gap-2 text-gray-600 hover:text-gray-900"
          >
            <ArrowLeftIcon className="h-5 w-5" />
            Back
          </button>
        </div>
      </header>

      <main className="max-w-2xl mx-auto px-4 py-8">
        {/* Identity Step */}
        {step === 'identity' && (
          <Card>
            <div className="text-center mb-6">
              <ShieldExclamationIcon className="h-16 w-16 text-yellow-600 mx-auto mb-4" aria-hidden="true" />
              <h2 className="text-2xl font-bold mb-2">
                Identity Creation Unavailable
              </h2>
              <p className="text-gray-600">
                This wallet has no preview identity for this domain
              </p>
            </div>

            <div
              className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6"
              data-testid="identity-creation-disabled-notice"
            >
              <p className="text-sm text-yellow-900 break-words">
                {IDENTITY_CREATION_DISABLED} Joining with an anonymous identity
                is not available in this client.
              </p>
            </div>

            <Button
              type="button"
              disabled
              aria-describedby="onboarding-identity-creation-disabled-reason"
              className="w-full min-h-[44px] disabled:bg-gray-200 disabled:text-gray-600"
            >
              Identity Creation Disabled in Preview
            </Button>
            <p id="onboarding-identity-creation-disabled-reason" className="sr-only">
              {IDENTITY_CREATION_DISABLED}
            </p>
          </Card>
        )}

        {/* Submit Step — identity registration is disabled in the preview (issue #309) */}
        {step === 'submit' && (
          <Card>
            <h2 className="text-2xl font-bold mb-6">
              Identity Registration Unavailable
            </h2>

            <div className="space-y-4 mb-6">
              <div className="flex items-start gap-3">
                <CheckCircleIcon className="h-6 w-6 text-green-600 shrink-0" />
                <div className="flex-1">
                  <h3 className="font-semibold">Membership Approved</h3>
                  <p className="text-sm text-gray-600">
                    You are a verified member of this domain
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <div className="shrink-0 w-6 h-6 bg-gray-100 text-gray-600 rounded-full flex items-center justify-center text-xs font-bold">
                  2
                </div>
                <div className="flex-1">
                  <h3 className="font-semibold">Identity Registration</h3>
                  <p className="text-sm text-gray-600">
                    Not available in this preview
                  </p>
                </div>
              </div>
            </div>

            <div
              className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6"
              data-testid="identity-registration-disabled-notice"
            >
              <p className="text-sm text-yellow-900 break-words">
                {PREVIEW_IDENTITY_REGISTRATION_DISABLED} No transaction is sent,
                nothing is registered on-chain and anonymous voting is not
                available in this client.
              </p>
            </div>

            <Button
              type="button"
              disabled
              aria-describedby="identity-registration-disabled-reason"
              className="w-full min-h-[44px] disabled:bg-gray-200 disabled:text-gray-600"
            >
              Registration Disabled in Preview
            </Button>
            <p id="identity-registration-disabled-reason" className="sr-only">
              {PREVIEW_IDENTITY_REGISTRATION_DISABLED}
            </p>
          </Card>
        )}

        {/* Waiting Step — not yet a member, waiting for admin */}
        {step === 'waiting' && (
          <Card>
            <div className="text-center mb-6">
              <ClockIcon className="h-16 w-16 text-yellow-600 mx-auto mb-4 animate-pulse" />
              <h2 className="text-2xl font-bold mb-2">
                Waiting for Verification
              </h2>
              <p className="text-gray-600">
                Your membership request has been submitted
              </p>
            </div>

            <div className="space-y-4 mb-6">
              <div className="flex items-start gap-3">
                <CheckCircleIcon className="h-6 w-6 text-green-600 shrink-0" />
                <div className="flex-1">
                  <h3 className="font-semibold">Step 1: Request Submitted</h3>
                  <p className="text-sm text-gray-600">
                    Your onboarding request is on the blockchain
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <ClockIcon className="h-6 w-6 text-yellow-600 shrink-0 animate-pulse" />
                <div className="flex-1">
                  <h3 className="font-semibold">Step 2: Waiting for Admin</h3>
                  <p className="text-sm text-gray-600">
                    The domain admin will verify your request shortly
                  </p>
                </div>
              </div>
            </div>

            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <p className="text-sm text-blue-900">
                <strong>What's happening?</strong>
              </p>
              <p className="text-sm text-blue-800 mt-1">
                The domain admin needs to approve your membership via
                MsgApproveOnboarding. This ensures only legitimate members join.
              </p>
            </div>
          </Card>
        )}

        {/* Complete Step */}
        {step === 'complete' && (
          <Card>
            <div className="text-center mb-6">
              <CheckCircleIcon className="h-16 w-16 text-green-600 mx-auto mb-4" />
              <h2 className="text-2xl font-bold mb-2">Membership Active!</h2>
              <p className="text-gray-600">
                You can now vote anonymously in this domain
              </p>
            </div>

            <div className="bg-green-50 border border-green-200 rounded-lg p-4 mb-6">
              <h3 className="font-semibold text-green-900 mb-2">
                You're all set!
              </h3>
              <ul className="text-sm text-green-800 space-y-1">
                <li>Your identity is in the domain Merkle tree</li>
                <li>You can vote anonymously on suggestions</li>
                <li>No one can link your votes to you</li>
              </ul>
            </div>

            <div className="space-y-3">
              <Button
                onClick={() => navigate(`/governance/domain/${domainId}`)}
                className="w-full"
              >
                Browse Domain Issues
              </Button>
              <Button
                variant="secondary"
                onClick={() => navigate('/governance')}
                className="w-full"
              >
                Back to Domains
              </Button>
            </div>
          </Card>
        )}
      </main>
    </div>
  );
}
