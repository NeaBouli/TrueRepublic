import React, { useState, useEffect } from "react";
import { queryZKPState } from "../../services/api";

export default function MembershipStatus({ domainName }) {
  const [state, setState] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!domainName) return;
    setLoading(true);
    setError(null);
    queryZKPState(domainName)
      .then((data) => setState(data))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [domainName]);

  if (!domainName) return null;

  if (loading) {
    return (
      <div className="bg-dark-800 border border-dark-700 rounded-xl p-4">
        <p className="text-sm text-dark-500">Loading ZKP state...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-dark-800 border border-dark-700 rounded-xl p-4">
        <p className="text-xs text-dark-500">{error}</p>
      </div>
    );
  }

  if (!state) return null;

  return (
    <div className="bg-dark-800 border border-dark-700 rounded-xl p-4 space-y-2">
      <h4 className="text-sm font-semibold text-dark-400 uppercase tracking-wider">
        ZKP Membership
      </h4>

      <div className="space-y-1.5 text-sm">
        <div className="flex justify-between">
          <span className="text-dark-400">Merkle Root</span>
          <span className="text-dark-200 font-mono text-xs">
            {state.merkle_root
              ? state.merkle_root.slice(0, 16) + "..."
              : "none"}
          </span>
        </div>
        <div className="flex justify-between">
          <span className="text-dark-400">Commitments</span>
          <span className="text-dark-200">{state.commitment_count}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-dark-400">Members</span>
          <span className="text-dark-200">{state.member_count}</span>
        </div>
        <div className="flex justify-between">
          <span className="text-dark-400">VK Initialized</span>
          <span className={state.vk_initialized ? "text-green-400" : "text-red-400"}>
            {state.vk_initialized ? "Yes" : "No"}
          </span>
        </div>
        <div className="flex justify-between">
          <span className="text-dark-400">Root History</span>
          <span className="text-dark-200">
            {state.merkle_root_history?.length || 0} entries
          </span>
        </div>
      </div>
    </div>
  );
}
