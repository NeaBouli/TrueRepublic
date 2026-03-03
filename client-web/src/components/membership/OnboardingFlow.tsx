import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useWalletStore } from '@/stores/walletStore';
import { useIdentityStore } from '@/stores/identityStore';
import { useMembershipStore } from '@/stores/membershipStore';
import { MembershipService } from '@/services/membership';
import { WalletService } from '@/services/wallet';
import { Card } from '@/components/common/Card';
import { Button } from '@/components/common/Button';
import { DEFAULT_CHAIN } from '@/config/chains';
import {
  ArrowLeftIcon,
  CheckCircleIcon,
  ClockIcon,
  XCircleIcon,
  ShieldCheckIcon,
} from '@heroicons/react/24/outline';

export function OnboardingFlow() {
  const navigate = useNavigate();
  const { domainId } = useParams<{ domainId: string }>();
  const { currentWallet, password } = useWalletStore();
  const { identity, hasIdentity, createIdentity } = useIdentityStore();
  const { memberships, loadMembership } = useMembershipStore();

  const [step, setStep] = useState<
    'identity' | 'submit' | 'waiting' | 'complete'
  >('identity');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');
  const [txHash, setTxHash] = useState('');

  const membership = domainId ? memberships[domainId] : null;

  useEffect(() => {
    if (!domainId || !currentWallet) return;
    loadMembership(domainId, currentWallet.address);
  }, [domainId, currentWallet, loadMembership]);

  useEffect(() => {
    if (!hasIdentity) {
      setStep('identity');
    } else if (membership?.isMember && membership?.hasIdentityCommitment) {
      setStep('complete');
    } else if (membership?.isMember) {
      // Member but no identity commitment registered yet
      setStep('submit');
    } else {
      setStep('waiting');
    }
  }, [hasIdentity, membership]);

  // Poll for membership approval
  useEffect(() => {
    if (step !== 'waiting' || !domainId || !currentWallet) return;

    const interval = setInterval(() => {
      loadMembership(domainId, currentWallet.address);
    }, 5000);

    return () => clearInterval(interval);
  }, [step, domainId, currentWallet, loadMembership]);

  const handleCreateIdentity = () => {
    createIdentity();
  };

  const handleRegisterIdentity = async () => {
    if (!currentWallet || !password || !identity || !domainId) return;

    setIsSubmitting(true);
    setError('');

    try {
      const membershipService = new MembershipService(DEFAULT_CHAIN);
      const wallet = await WalletService.getWalletForSigning(
        currentWallet.address,
        password
      );

      const result = await membershipService.registerIdentity(
        wallet,
        domainId,
        identity.commitment
      );

      if (!result.success) {
        throw new Error(result.error || 'Identity registration failed');
      }

      setTxHash(result.hash);
      await loadMembership(domainId, currentWallet.address);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Registration failed');
    } finally {
      setIsSubmitting(false);
    }
  };

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
              <ShieldCheckIcon className="h-16 w-16 text-primary-600 mx-auto mb-4" />
              <h2 className="text-2xl font-bold mb-2">
                Anonymous Identity Required
              </h2>
              <p className="text-gray-600">
                Create an anonymous identity to join this domain
              </p>
            </div>

            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
              <h3 className="font-semibold text-blue-900 mb-2">
                Why do you need this?
              </h3>
              <p className="text-sm text-blue-800">
                Your identity commitment will be stored in the domain's Merkle
                tree, allowing you to vote anonymously while proving you're a
                member.
              </p>
            </div>

            <Button onClick={handleCreateIdentity} className="w-full">
              Create Anonymous Identity
            </Button>
          </Card>
        )}

        {/* Submit Step — register identity commitment on-chain */}
        {step === 'submit' && (
          <Card>
            <h2 className="text-2xl font-bold mb-6">
              Register Identity Commitment
            </h2>

            {error && (
              <div className="bg-red-50 border border-red-200 rounded-lg p-3 mb-4 flex items-start gap-2">
                <XCircleIcon className="h-5 w-5 text-red-600 flex-shrink-0 mt-0.5" />
                <p className="text-sm text-red-800">{error}</p>
              </div>
            )}

            <div className="space-y-4 mb-6">
              <div className="flex items-start gap-3">
                <CheckCircleIcon className="h-6 w-6 text-green-600 flex-shrink-0" />
                <div className="flex-1">
                  <h3 className="font-semibold">Membership Approved</h3>
                  <p className="text-sm text-gray-600">
                    You are a verified member of this domain
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <div className="flex-shrink-0 w-6 h-6 bg-primary-100 text-primary-700 rounded-full flex items-center justify-center text-xs font-bold">
                  2
                </div>
                <div className="flex-1">
                  <h3 className="font-semibold">Register ZKP Identity</h3>
                  <p className="text-sm text-gray-600">
                    Submit your identity commitment to the Merkle tree for
                    anonymous voting
                  </p>
                </div>
              </div>
            </div>

            <Button
              onClick={handleRegisterIdentity}
              isLoading={isSubmitting}
              className="w-full"
            >
              Register Identity Commitment
            </Button>
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

            {txHash && (
              <div className="bg-gray-50 rounded-lg p-4 mb-6">
                <div className="text-xs text-gray-600 mb-1">
                  Transaction Hash
                </div>
                <code className="text-xs font-mono break-all text-gray-700">
                  {txHash}
                </code>
              </div>
            )}

            <div className="space-y-4 mb-6">
              <div className="flex items-start gap-3">
                <CheckCircleIcon className="h-6 w-6 text-green-600 flex-shrink-0" />
                <div className="flex-1">
                  <h3 className="font-semibold">Step 1: Request Submitted</h3>
                  <p className="text-sm text-gray-600">
                    Your onboarding request is on the blockchain
                  </p>
                </div>
              </div>

              <div className="flex items-start gap-3">
                <ClockIcon className="h-6 w-6 text-yellow-600 flex-shrink-0 animate-pulse" />
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
