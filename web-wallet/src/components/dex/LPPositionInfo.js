import React, { useState } from "react";
import { queryLPPosition } from "../../services/api";

export default function LPPositionInfo({ pools }) {
  const [assetDenom, setAssetDenom] = useState(
    pools?.[0]?.asset_denom || "atom"
  );
  const [shares, setShares] = useState("");
  const [result, setResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleCalculate = async () => {
    if (!shares || !assetDenom) return;
    setLoading(true);
    setError(null);
    try {
      const data = await queryLPPosition(assetDenom, shares);
      setResult(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-dark-800 border border-dark-700 rounded-xl p-4 space-y-3">
      <h4 className="text-sm font-semibold text-dark-400 uppercase tracking-wider">
        LP Position
      </h4>

      <div className="flex gap-2 items-end">
        <div className="flex-1">
          <label className="block text-xs text-dark-500 mb-1">Pool</label>
          <select
            value={assetDenom}
            onChange={(e) => setAssetDenom(e.target.value)}
            className="w-full px-3 py-1.5 text-sm bg-dark-700 border border-dark-600 rounded-lg text-dark-200 focus:outline-none focus:border-republic-500"
          >
            {(pools || []).map((p) => (
              <option key={p.asset_denom} value={p.asset_denom}>
                PNYX / {p.asset_denom?.toUpperCase()}
              </option>
            ))}
          </select>
        </div>
        <div className="flex-1">
          <label className="block text-xs text-dark-500 mb-1">Shares</label>
          <input
            type="number"
            value={shares}
            onChange={(e) => setShares(e.target.value)}
            placeholder="0"
            className="w-full px-3 py-1.5 text-sm bg-dark-700 border border-dark-600 rounded-lg text-dark-200 placeholder-dark-500 focus:outline-none focus:border-republic-500"
          />
        </div>
        <button
          onClick={handleCalculate}
          disabled={loading || !shares}
          className="px-4 py-1.5 text-sm bg-republic-600 text-white rounded-lg hover:bg-republic-700 transition-colors disabled:opacity-50"
        >
          {loading ? "..." : "Calculate"}
        </button>
      </div>

      {result && (
        <div className="space-y-1.5 text-sm">
          <div className="flex justify-between">
            <span className="text-dark-400">PNYX Value</span>
            <span className="text-dark-200">{result.pnyx_value}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-dark-400">Asset Value</span>
            <span className="text-dark-200">{result.asset_value}</span>
          </div>
          <div className="flex justify-between">
            <span className="text-dark-400">Pool Share</span>
            <span className="text-dark-200">
              {(result.share_of_pool_bps / 100).toFixed(2)}%
            </span>
          </div>
        </div>
      )}

      {error && <p className="text-xs text-red-400">{error}</p>}
    </div>
  );
}
