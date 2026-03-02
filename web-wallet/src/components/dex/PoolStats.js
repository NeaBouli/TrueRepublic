import React, { useState, useEffect } from "react";
import { queryPoolStats } from "../../services/api";
import pnyxIcon from "../../assets/images/pnyx-icon.png";

export default function PoolStats({ assetDenom }) {
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!assetDenom) return;
    setLoading(true);
    queryPoolStats(assetDenom)
      .then((data) => setStats(data))
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
  }, [assetDenom]);

  if (loading) {
    return (
      <div className="bg-dark-800 border border-dark-700 rounded-xl p-4">
        <p className="text-sm text-dark-500">Loading pool stats...</p>
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

  if (!stats) return null;

  const spotPrice = stats.spot_price_per_million
    ? (Number(stats.spot_price_per_million) / 1_000_000).toFixed(6)
    : "N/A";

  const rows = [
    ["Pool", `PNYX / ${stats.asset_symbol || stats.asset_denom}`],
    ["Swap Count", stats.swap_count],
    ["Volume (PNYX)", stats.total_volume_pnyx],
    ["Fees Earned", stats.total_fees_earned],
    ["Total Burned", stats.total_burned],
    ["PNYX Reserve", stats.pnyx_reserve],
    ["Asset Reserve", stats.asset_reserve],
    ["Spot Price", spotPrice],
    ["Total Shares", stats.total_shares],
  ];

  return (
    <div className="bg-dark-800 border border-dark-700 rounded-xl p-4">
      <h4 className="text-sm font-semibold text-dark-400 uppercase tracking-wider mb-3 flex items-center gap-2">
        <img src={pnyxIcon} alt="PNYX" className="w-5 h-5 rounded-full" />
        {stats.asset_symbol || assetDenom} Pool Stats
      </h4>
      <div className="space-y-1.5">
        {rows.map(([label, value]) => (
          <div key={label} className="flex justify-between text-sm">
            <span className="text-dark-400">{label}</span>
            <span className="text-dark-200">{value}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
