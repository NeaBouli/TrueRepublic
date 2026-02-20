import React, { useState, useEffect } from "react";
import Header from "../components/Header";
import useWallet from "../hooks/useWallet";
import { fetchPools, swapTokens } from "../services/api";

function Dex() {
  const wallet = useWallet();
  const [amount, setAmount] = useState("");
  const [fromAsset, setFromAsset] = useState("pnyx");
  const [toAsset, setToAsset] = useState("atom");
  const [pools, setPools] = useState([]);
  const [swapping, setSwapping] = useState(false);

  const loadPools = async () => {
    try {
      const data = await fetchPools();
      setPools(data || []);
    } catch (err) {
      console.error("Failed to fetch pools:", err);
    }
  };

  useEffect(() => {
    loadPools();
  }, []);

  const handleSwap = async () => {
    if (!amount) return alert("Please enter an amount.");
    if (fromAsset === toAsset) return alert("From and To assets must differ.");
    setSwapping(true);
    try {
      const result = await swapTokens(wallet.address, fromAsset, amount, toAsset);
      alert("Swap successful: " + result.transactionHash);
      setAmount("");
      loadPools();
      wallet.refreshBalance();
    } catch (err) {
      alert("Error: " + err.message);
    } finally {
      setSwapping(false);
    }
  };

  return (
    <div className="min-h-screen bg-dark-900 text-dark-50">
      <header className="border-b border-dark-700 bg-dark-850">
        <Header
          address={wallet.address}
          onConnect={wallet.connect}
          onDisconnect={wallet.disconnect}
          loading={wallet.loading}
        />
      </header>

      <div className="max-w-lg mx-auto p-6 mt-8">
        <h2 className="text-2xl font-semibold mb-6">DEX</h2>

        {/* Pools */}
        {pools.length > 0 && (
          <div className="bg-dark-800 border border-dark-700 rounded-xl p-5 mb-6">
            <h3 className="text-sm font-semibold text-dark-400 uppercase tracking-wider mb-3">
              Liquidity Pools
            </h3>
            <div className="space-y-2">
              {pools.map((p, i) => (
                <div
                  key={i}
                  className="flex justify-between items-center text-sm py-2 border-b border-dark-700 last:border-0"
                >
                  <span className="text-dark-200 font-medium">
                    PNYX / {p.asset_denom?.toUpperCase()}
                  </span>
                  <div className="text-right text-dark-400 text-xs">
                    <div>
                      {p.pnyx_reserve} / {p.asset_reserve}
                    </div>
                    <div>Burned: {p.total_burned}</div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Swap form */}
        {wallet.connected ? (
          <div className="bg-dark-800 border border-dark-700 rounded-xl p-5">
            <h3 className="text-lg font-medium mb-4">Swap Tokens</h3>
            <div className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-dark-400 mb-1">
                  From
                </label>
                <select
                  value={fromAsset}
                  onChange={(e) => setFromAsset(e.target.value)}
                  className="w-full px-3 py-2 text-sm bg-dark-700 border border-dark-600 rounded-lg text-dark-200 focus:outline-none focus:border-republic-500"
                >
                  <option value="pnyx">PNYX</option>
                  <option value="atom">ATOM</option>
                </select>
              </div>
              <div className="text-center text-dark-500">&#8595;</div>
              <div>
                <label className="block text-xs font-medium text-dark-400 mb-1">
                  To
                </label>
                <select
                  value={toAsset}
                  onChange={(e) => setToAsset(e.target.value)}
                  className="w-full px-3 py-2 text-sm bg-dark-700 border border-dark-600 rounded-lg text-dark-200 focus:outline-none focus:border-republic-500"
                >
                  <option value="atom">ATOM</option>
                  <option value="pnyx">PNYX</option>
                </select>
              </div>
              <div>
                <label className="block text-xs font-medium text-dark-400 mb-1">
                  Amount
                </label>
                <input
                  type="number"
                  placeholder="0"
                  value={amount}
                  onChange={(e) => setAmount(e.target.value)}
                  className="w-full px-3 py-2 text-sm bg-dark-700 border border-dark-600 rounded-lg text-dark-200 placeholder-dark-500 focus:outline-none focus:border-republic-500"
                />
              </div>
              <p className="text-xs text-dark-500">
                Fee: 0.3% swap fee. 1% burn on PNYX output.
              </p>
              <button
                onClick={handleSwap}
                disabled={swapping || !amount}
                className="w-full px-4 py-2.5 text-sm font-medium bg-republic-600 text-white rounded-lg hover:bg-republic-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {swapping ? "Swapping..." : "Swap"}
              </button>
            </div>
          </div>
        ) : (
          <div className="bg-dark-800 border border-dark-700 rounded-xl p-8 text-center">
            <div className="text-4xl mb-3">&#128260;</div>
            <p className="text-dark-300">Connect your wallet to swap tokens</p>
          </div>
        )}
      </div>
    </div>
  );
}

export default Dex;
