import { useState, useEffect, useCallback } from 'react';
import { useIdentityStore } from '@/stores/identityStore';
import { ZKPService } from '@/services/zkp';
import { Card } from '@/components/common/Card';
import { Button } from '@/components/common/Button';
import { ProofProgress } from './ProofProgress';
import { IdentitySetup } from './IdentitySetup';
import type { Suggestion } from '@/types/governance';
import type { ProofGenerationStatus } from '@/types/zkp';
import { RATING_MIN, RATING_MAX } from '@/types/zkp';
import { DEFAULT_CHAIN } from '@/config/chains';
import { ShieldCheckIcon } from '@heroicons/react/24/outline';

interface VotingPanelProps {
  suggestion: Suggestion;
  domainId: string;
  issueName: string;
  onVoteSubmitted?: () => void;
}

const zkpService = new ZKPService(DEFAULT_CHAIN);

export function VotingPanel({
  suggestion,
  domainId,
  issueName,
  onVoteSubmitted,
}: VotingPanelProps) {
  const { identity, hasIdentity } = useIdentityStore();

  const [rating, setRating] = useState(0);
  const [isGenerating, setIsGenerating] = useState(false);
  const [proofStatus, setProofStatus] = useState<ProofGenerationStatus>({
    step: 'idle',
    progress: 0,
    message: 'Ready to generate proof',
  });
  const [alreadyVoted, setAlreadyVoted] = useState(false);
  const [showIdentitySetup, setShowIdentitySetup] = useState(false);

  const checkIfAlreadyVoted = useCallback(async () => {
    if (!identity) return;

    try {
      const extNullifier = zkpService.computeExternalNullifier(
        domainId,
        issueName,
        suggestion.suggestionId
      );
      const nullifierHash = zkpService.computeNullifierHash(
        identity.secret,
        extNullifier
      );
      const used = await zkpService.isNullifierUsed(domainId, nullifierHash);
      setAlreadyVoted(used);
    } catch {
      // Best-effort check; node may be offline
    }
  }, [identity, domainId, issueName, suggestion.suggestionId]);

  useEffect(() => {
    if (hasIdentity && identity) {
      checkIfAlreadyVoted();
    }
  }, [hasIdentity, identity, checkIfAlreadyVoted]);

  const handleVote = async () => {
    if (!identity) return;

    setIsGenerating(true);

    try {
      // Initialize ZKP service
      await zkpService.initialize((status) => {
        setProofStatus(status);
      });

      // Fetch Merkle proof using identity commitment
      const merkleProof = await zkpService.fetchMerkleProof(
        domainId,
        identity.commitment
      );

      if (!merkleProof) {
        throw new Error(
          'Not a member of this domain, or commitment not registered'
        );
      }

      // Compute external nullifier for this vote context
      const externalNullifier = zkpService.computeExternalNullifier(
        domainId,
        issueName,
        suggestion.suggestionId
      );

      // Generate Groth16 proof
      const proof = await zkpService.generateProof({
        identitySecret: identity.secret,
        merkleRoot: merkleProof.root,
        merkleProof,
        externalNullifier,
        rating,
        domainName: domainId,
        issueName,
        suggestionName: suggestion.suggestionId,
      });

      // TODO: Submit MsgRateWithProof to blockchain via SigningStargateClient
      // const voteParams: VoteWithProofParams = {
      //   domain_name: domainId,
      //   issue_name: issueName,
      //   suggestion_name: suggestion.suggestionId,
      //   rating,
      //   proof: proof.proof,
      //   nullifier_hash: proof.nullifierHash,
      //   merkle_root: proof.merkleRoot,
      // };
      void proof; // suppress unused lint for now

      onVoteSubmitted?.();
    } catch (err: unknown) {
      const message =
        err instanceof Error ? err.message : 'Proof generation failed';
      setProofStatus({
        step: 'error',
        progress: 0,
        message: 'Failed to generate proof',
        error: message,
      });
    }
  };

  if (!hasIdentity || showIdentitySetup) {
    return (
      <IdentitySetup
        onComplete={() => setShowIdentitySetup(false)}
      />
    );
  }

  if (alreadyVoted) {
    return (
      <Card>
        <div className="text-center py-8">
          <ShieldCheckIcon className="h-12 w-12 text-green-600 mx-auto mb-3" />
          <h3 className="text-lg font-semibold mb-2">Already Voted</h3>
          <p className="text-gray-600">
            You have already voted on this suggestion anonymously
          </p>
        </div>
      </Card>
    );
  }

  if (isGenerating) {
    return (
      <Card>
        <h3 className="text-lg font-semibold mb-6">
          Generating Anonymous Vote
        </h3>
        <ProofProgress status={proofStatus} />

        {proofStatus.step === 'error' && (
          <Button
            onClick={() => {
              setIsGenerating(false);
              setProofStatus({
                step: 'idle',
                progress: 0,
                message: 'Ready to generate proof',
              });
            }}
            variant="secondary"
            className="w-full mt-6"
          >
            Try Again
          </Button>
        )}
      </Card>
    );
  }

  return (
    <Card>
      <h3 className="text-lg font-semibold mb-4">Vote Anonymously</h3>

      <div className="mb-6">
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Your Rating ({RATING_MIN} to +{RATING_MAX})
        </label>
        <div className="flex items-center gap-4">
          <input
            type="range"
            min={RATING_MIN}
            max={RATING_MAX}
            value={rating}
            onChange={(e) => setRating(parseInt(e.target.value, 10))}
            className="flex-1 h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer"
          />
          <div className="w-12 text-center">
            <span
              className={`text-2xl font-bold ${
                rating > 0
                  ? 'text-green-600'
                  : rating < 0
                    ? 'text-red-600'
                    : 'text-gray-600'
              }`}
            >
              {rating > 0 ? `+${rating}` : rating}
            </span>
          </div>
        </div>
        <div className="flex justify-between text-xs text-gray-500 mt-1">
          <span>Strong Opposition ({RATING_MIN})</span>
          <span>Strong Support (+{RATING_MAX})</span>
        </div>
      </div>

      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
        <div className="text-sm text-blue-900">
          <strong>Anonymous Voting</strong>
          <p className="mt-1 text-blue-800">
            Your vote will be cryptographically anonymous. No one can link
            it to your identity.
          </p>
        </div>
      </div>

      <Button
        onClick={handleVote}
        className="w-full flex items-center justify-center gap-2"
      >
        <ShieldCheckIcon className="h-5 w-5" />
        Submit Anonymous Vote
      </Button>

      <div className="mt-3 text-center">
        <button
          onClick={() => setShowIdentitySetup(true)}
          className="text-sm text-gray-600 hover:text-gray-900"
        >
          Manage Identity
        </button>
      </div>
    </Card>
  );
}
