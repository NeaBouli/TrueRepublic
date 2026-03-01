import React, { useState } from "react";
import { querySpotPrice } from "../../services/api";

export default function SpotPriceDisplay({ pools }) {
  const denoms = ["pnyx", ...(pools || []).map((p) => p.asset_denom)];
  const [inputDenom, setInputDenom] = useState("pnyx");
  const [outputDenom, setOutputDenom] = useState(denoms[1] || "atom");
  const [result, setResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleFetch = async () => {
    if (inputDenom === outputDenom) return;
    setLoading(true);
    setError(null);
    try {
      const data = await querySpotPrice(inputDenom, outputDenom);
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
        Spot Price
      </h4>

      <div className="flex gap-2 items-end">
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
        <button
          onClick={handleFetch}
          disabled={loading || inputDenom === outputDenom}
          className="px-4 py-1.5 text-sm bg-republic-600 text-white rounded-lg hover:bg-republic-700 transition-colors disabled:opacity-50"
        >
          {loading ? "..." : "Get Price"}
        </button>
      </div>

      {result && (
        <div className="space-y-1">
          <p className="text-sm text-dark-200">
            1 {result.input_symbol} ={" "}
            <span className="font-medium text-republic-400">
              {(Number(result.price_per_million) / 1_000_000).toFixed(6)}
            </span>{" "}
            {result.output_symbol}
          </p>
          <p className="text-xs text-dark-500">
            Route: {result.route?.map((d) => d.toUpperCase()).join(" \u2192 ")}
          </p>
        </div>
      )}

      {error && <p className="text-xs text-red-400">{error}</p>}
    </div>
  );
}
