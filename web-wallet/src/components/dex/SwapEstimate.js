import React, { useState } from "react";
import { queryEstimateSwap } from "../../services/api";

export default function SwapEstimate({ pools }) {
  const denoms = ["pnyx", ...(pools || []).map((p) => p.asset_denom)];
  const [inputDenom, setInputDenom] = useState("pnyx");
  const [outputDenom, setOutputDenom] = useState(denoms[1] || "atom");
  const [amount, setAmount] = useState("");
  const [result, setResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleEstimate = async () => {
    if (!amount || inputDenom === outputDenom) return;
    setLoading(true);
    setError(null);
    try {
      const data = await queryEstimateSwap(inputDenom, amount, outputDenom);
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
        Swap Estimate
      </h4>

      <div className="space-y-2">
        <div className="flex gap-2">
          <div className="flex-1">
            <label className="block text-xs text-dark-500 mb-1">From</label>
            <select
              value={inputDenom}
              onChange={(e) => setInputDenom(e.target.value)}
              className="w-full px-3 py-1.5 text-sm bg-dark-700 border border-dark-600 rounded-lg text-dark-200 focus:outline-none focus:border-republic-500"
            >
              {denoms.map((d) => (
                <option key={d} value={d}>
                  {d.toUpperCase()}
                </option>
              ))}
            </select>
          </div>
          <div className="flex-1">
            <label className="block text-xs text-dark-500 mb-1">To</label>
            <select
              value={outputDenom}
              onChange={(e) => setOutputDenom(e.target.value)}
              className="w-full px-3 py-1.5 text-sm bg-dark-700 border border-dark-600 rounded-lg text-dark-200 focus:outline-none focus:border-republic-500"
            >
              {denoms.map((d) => (
                <option key={d} value={d}>
                  {d.toUpperCase()}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div>
          <label className="block text-xs text-dark-500 mb-1">Amount</label>
          <input
            type="number"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            placeholder="0"
            className="w-full px-3 py-1.5 text-sm bg-dark-700 border border-dark-600 rounded-lg text-dark-200 placeholder-dark-500 focus:outline-none focus:border-republic-500"
          />
        </div>

        <button
          onClick={handleEstimate}
          disabled={loading || !amount || inputDenom === outputDenom}
          className="w-full px-3 py-2 text-sm font-medium bg-republic-600 text-white rounded-lg hover:bg-republic-700 transition-colors disabled:opacity-50"
        >
          {loading ? "Estimating..." : "Estimate Output"}
        </button>
      </div>

      {result && (
        <div className="space-y-1.5">
          <div className="flex justify-between text-sm">
            <span className="text-dark-400">Expected Output</span>
            <span className="text-dark-200 font-medium">
              {Number(result.expected_output).toLocaleString()}{" "}
              {outputDenom.toUpperCase()}
            </span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-dark-400">Route</span>
            <span className="text-dark-300 text-xs">
              {result.route_symbols?.join(" \u2192 ")}
            </span>
          </div>
          <div className="flex justify-between text-sm">
            <span className="text-dark-400">Hops</span>
            <span className="text-dark-200">{result.hops}</span>
          </div>
          {result.hops > 1 && (
            <p className="text-xs text-yellow-400">
              Cross-asset swap via PNYX hub (2 hops, higher fees)
            </p>
          )}
        </div>
      )}

      {error && <p className="text-xs text-red-400">{error}</p>}
    </div>
  );
}
