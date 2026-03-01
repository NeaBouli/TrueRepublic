import React, { useState } from "react";
import { queryNullifier } from "../../services/api";

export default function NullifierStatus({ domainName }) {
  const [hash, setHash] = useState("");
  const [result, setResult] = useState(null);
  const [checking, setChecking] = useState(false);

  const handleCheck = async () => {
    if (!hash.trim() || !domainName) return;
    setChecking(true);
    setResult(null);
    try {
      const data = await queryNullifier(domainName, hash.trim());
      setResult(data);
    } catch (err) {
      setResult({ error: err.message });
    } finally {
      setChecking(false);
    }
  };

  return (
    <div className="bg-dark-800 border border-dark-700 rounded-xl p-4 space-y-3">
      <h4 className="text-sm font-semibold text-dark-400 uppercase tracking-wider">
        Nullifier Check
      </h4>

      <div className="flex gap-2">
        <input
          type="text"
          value={hash}
          onChange={(e) => setHash(e.target.value)}
          placeholder="Nullifier hash (hex)"
          className="flex-1 px-3 py-1.5 text-sm bg-dark-700 border border-dark-600 rounded-lg text-dark-200 placeholder-dark-500 focus:outline-none focus:border-republic-500 font-mono"
        />
        <button
          onClick={handleCheck}
          disabled={checking || !hash.trim()}
          className="px-4 py-1.5 text-sm bg-republic-600 text-white rounded-lg hover:bg-republic-700 transition-colors disabled:opacity-50"
        >
          {checking ? "..." : "Check"}
        </button>
      </div>

      {result && !result.error && (
        <div
          className={`text-sm ${
            result.used ? "text-red-400" : "text-green-400"
          }`}
        >
          {result.used
            ? "Already used \u2014 vote was cast"
            : "Available \u2014 not yet used"}
        </div>
      )}

      {result?.error && (
        <p className="text-xs text-red-400">{result.error}</p>
      )}
    </div>
  );
}
