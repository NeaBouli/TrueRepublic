import React, { useState } from "react";
import MembershipStatus from "./zkp/MembershipStatus";

function SubmitProposalForm({ domainName, onSubmit }) {
  const [issueName, setIssueName] = useState("");
  const [suggestionName, setSuggestionName] = useState("");
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!issueName || !suggestionName) return;
    setSubmitting(true);
    try {
      await onSubmit(domainName, issueName, suggestionName);
      setIssueName("");
      setSuggestionName("");
    } catch (err) {
      alert("Failed to submit: " + err.message);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-3">
      <div>
        <label className="block text-xs font-medium text-dark-400 mb-1">
          Issue Name
        </label>
        <input
          type="text"
          value={issueName}
          onChange={(e) => setIssueName(e.target.value)}
          placeholder="e.g. Infrastructure Spending"
          className="w-full px-3 py-2 text-sm bg-dark-700 border border-dark-600 rounded-lg text-dark-200 placeholder-dark-500 focus:outline-none focus:border-republic-500"
        />
      </div>
      <div>
        <label className="block text-xs font-medium text-dark-400 mb-1">
          Suggestion
        </label>
        <input
          type="text"
          value={suggestionName}
          onChange={(e) => setSuggestionName(e.target.value)}
          placeholder="e.g. Increase budget by 10%"
          className="w-full px-3 py-2 text-sm bg-dark-700 border border-dark-600 rounded-lg text-dark-200 placeholder-dark-500 focus:outline-none focus:border-republic-500"
        />
      </div>
      <button
        type="submit"
        disabled={submitting || !issueName || !suggestionName}
        className="w-full px-4 py-2 text-sm font-medium bg-republic-600 text-white rounded-lg hover:bg-republic-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {submitting ? "Submitting..." : "Submit Proposal"}
      </button>
    </form>
  );
}

export default function DomainInfo({ domain, domainName, connected, onSubmitProposal }) {
  if (!domainName) {
    return (
      <div className="text-dark-400 text-sm">
        <p>Select a domain to see details and submit proposals.</p>
      </div>
    );
  }

  const issues = domain?.issues || [];
  const totalStones = issues.reduce((sum, i) => sum + (i.stones || 0), 0);
  const totalSuggestions = issues.reduce(
    (sum, i) => sum + (i.suggestions?.length || 0),
    0
  );

  return (
    <div className="space-y-6">
      {/* Domain stats */}
      <div>
        <h2 className="text-sm font-semibold text-dark-400 uppercase tracking-wider mb-3">
          Domain Info
        </h2>
        <div className="bg-dark-800 border border-dark-700 rounded-xl p-4 space-y-3">
          <div className="flex justify-between text-sm">
            <span className="text-dark-400">Issues</span>
            <span className="text-dark-200 font-medium">{issues.length}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-dark-400">Suggestions</span>
            <span className="text-dark-200 font-medium">{totalSuggestions}</span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-dark-400">Total Stones</span>
            <span className="text-dark-200 font-medium">{totalStones}</span>
          </div>
        </div>
      </div>

      {/* ZKP Membership */}
      <div>
        <h2 className="text-sm font-semibold text-dark-400 uppercase tracking-wider mb-3">
          ZKP Membership
        </h2>
        <MembershipStatus domainName={domainName} />
      </div>

      {/* Submit proposal */}
      {connected ? (
        <div>
          <h2 className="text-sm font-semibold text-dark-400 uppercase tracking-wider mb-3">
            Submit Proposal
          </h2>
          <div className="bg-dark-800 border border-dark-700 rounded-xl p-4">
            <SubmitProposalForm
              domainName={domainName}
              onSubmit={onSubmitProposal}
            />
          </div>
        </div>
      ) : (
        <div className="bg-dark-800 border border-dark-700 rounded-xl p-4 text-center">
          <p className="text-sm text-dark-400">
            Connect your wallet to submit proposals and vote.
          </p>
        </div>
      )}
    </div>
  );
}
