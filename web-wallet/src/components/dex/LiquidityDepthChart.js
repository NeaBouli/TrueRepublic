import React, { useState, useEffect } from "react";
import { queryLiquidityDepth } from "../../services/api";

function impactColor(bps) {
  if (bps <= 10) return "text-green-400";
  if (bps <= 100) return "text-yellow-400";
  return "text-red-400";
}

export default function LiquidityDepthChart({ inputDenom, outputDenom }) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!inputDenom || !outputDenom || inputDenom === outputDenom) return;
    setLoading(true);
    queryLiquidityDepth(inputDenom, outputDenom)
      .then((d) => setData(d))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [inputDenom, outputDenom]);

  if (loading) {
    return (
      <div className="bg-dark-800 border border-dark-700 rounded-xl p-4">
        <p className="text-sm text-dark-500">Loading depth...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-dark-800 border border-dark-700 rounded-xl p-4">
        <p className="text-xs text-red-400">{error}</p>
      </div>
    );
  }

  if (!data?.levels?.length) return null;

  return (
    <div className="bg-dark-800 border border-dark-700 rounded-xl p-4">
      <h4 className="text-sm font-semibold text-dark-400 uppercase tracking-wider mb-3">
        Liquidity Depth: {inputDenom.toUpperCase()} {"\u2192"}{" "}
        {outputDenom.toUpperCase()}
      </h4>

      <table className="w-full text-sm">
        <thead>
          <tr className="text-dark-500 text-xs">
            <th className="text-left pb-2">Input</th>
            <th className="text-right pb-2">Output</th>
            <th className="text-right pb-2">Impact</th>
          </tr>
        </thead>
        <tbody>
          {data.levels.map((lv, i) => (
            <tr
              key={i}
              className="border-t border-dark-700"
            >
              <td className="py-1.5 text-dark-200">
                {Number(lv.input_amount).toLocaleString()}
              </td>
              <td className="py-1.5 text-right text-dark-200">
                {Number(lv.output_amount).toLocaleString()}
              </td>
              <td
                className={`py-1.5 text-right ${impactColor(
                  lv.price_impact_bps
                )}`}
              >
                {lv.price_impact_bps} bps
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
