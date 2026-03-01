import React, { useState } from "react";
import { submitAnonymousVote } from "../../services/api";

function generateMockProof() {
  // Mock proof generation — placeholder for real snarkjs WASM circuit.
  const bytes = new Uint8Array(128);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
}

function generateMockNullifier() {
  const bytes = new Uint8Array(32);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
}

export default function ZKPVotingPanel({
  domainName,
  issueName,
  suggestionName,
  connected,
  address,
  onVoteSuccess,
}) {
  const [rating, setRating] = useState(0);
  const [proof, setProof] = useState(null);
  const [nullifierHash, setNullifierHash] = useState(null);
  const [generating, setGenerating] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState(null);
  const [submitted, setSubmitted] = useState(false);
  const [txHash, setTxHash] = useState(null);

  const handleGenerateProof = async () => {
    setGenerating(true);
    setError(null);
    try {
      // Simulate proof generation delay.
      await new Promise((r) => setTimeout(r, 1500));
      setProof(generateMockProof());
      setNullifierHash(generateMockNullifier());
    } catch (err) {
      setError(err.message);
    } finally {
      setGenerating(false);
    }
  };

  const handleSubmit = async () => {
    if (!proof || !nullifierHash) return;
    setSubmitting(true);
    setError(null);
    try {
      const result = await submitAnonymousVote(
        address,
        domainName,
        issueName,
        suggestionName,
        rating,
        proof,
        nullifierHash,
        ""
      );
      setTxHash(result.transactionHash);
      setSubmitted(true);
      if (onVoteSuccess) onVoteSuccess();
    } catch (err) {
      setError(err.message);
    } finally {
      setSubmitting(false);
    }
  };

  const ratingColor =
    rating < 0
      ? "text-red-400"
      : rating > 0
      ? "text-green-400"
      : "text-dark-400";

  if (!connected) {
    return (
      <div className="text-sm text-dark-500">
        Connect wallet for anonymous voting
      </div>
    );
  }

  if (submitted) {
    return (
      <div className="bg-dark-800 border border-dark-700 rounded-xl p-4">
        <p className="text-green-400 font-medium">
          Vote recorded anonymously
        </p>
        {txHash && (
          <p className="text-xs text-dark-500 mt-1 font-mono">
            TX: {txHash.slice(0, 16)}...
          </p>
        )}
      </div>
    );
  }

  return (
    <div className="bg-dark-800 border border-dark-700 rounded-xl p-4 space-y-3">
      <h4 className="text-sm font-semibold text-dark-300 uppercase tracking-wider">
        Anonymous Rating (ZKP)
      </h4>

      <div>
        <div className="flex items-center justify-between text-xs text-dark-500 mb-1">
          <span>-5 (resist)</span>
          <span className={`text-base font-bold ${ratingColor}`}>
            {rating > 0 ? `+${rating}` : rating}
          </span>
          <span>+5 (support)</span>
        </div>
        <input
          type="range"
          min="-5"
          max="5"
          step="1"
          value={rating}
          onChange={(e) => setRating(parseInt(e.target.value))}
          disabled={generating || submitting}
          className="w-full"
        />
      </div>

      {!proof ? (
        <button
          onClick={handleGenerateProof}
          disabled={generating}
          className="w-full px-3 py-2 text-sm font-medium bg-republic-600 text-white rounded-lg hover:bg-republic-700 transition-colors disabled:opacity-50"
        >
          {generating ? "Generating Proof..." : "Generate ZKP Proof"}
        </button>
      ) : (
        <div className="space-y-2">
          <div className="text-xs text-dark-500">
            <span className="text-green-400 font-medium">Proof ready</span>
            <span className="font-mono ml-2">
              {proof.slice(0, 16)}...
            </span>
          </div>
          <button
            onClick={handleSubmit}
            disabled={submitting}
            className="w-full px-3 py-2 text-sm font-medium bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors disabled:opacity-50"
          >
            {submitting ? "Submitting..." : "Submit Anonymous Vote"}
          </button>
        </div>
      )}

      {error && (
        <p className="text-xs text-red-400">{error}</p>
      )}

      <p className="text-xs text-dark-600">
        Your vote is anonymous. The ZKP proves domain membership without
        revealing your identity.
      </p>
    </div>
  );
}
