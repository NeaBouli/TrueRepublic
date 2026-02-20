import React, { useState } from "react";
import Header from "../components/Header";
import useWallet from "../hooks/useWallet";
import { sendTokens } from "../services/api";

function Wallet() {
  const wallet = useWallet();
  const [recipient, setRecipient] = useState("");
  const [amount, setAmount] = useState("");
  const [sending, setSending] = useState(false);

  const handleSend = async () => {
    if (!recipient || !amount) return alert("Please fill all fields.");
    setSending(true);
    try {
      const result = await sendTokens(wallet.address, recipient, amount);
      alert("Transaction successful: " + result.transactionHash);
      setRecipient("");
      setAmount("");
      wallet.refreshBalance();
    } catch (err) {
      alert("Error: " + err.message);
    } finally {
      setSending(false);
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
        <h2 className="text-2xl font-semibold mb-6">Wallet</h2>

        {wallet.error && (
          <div className="mb-4 p-3 bg-red-900/30 border border-red-700 rounded-lg text-sm text-red-300">
            {wallet.error}
          </div>
        )}

        {wallet.connected ? (
          <div className="space-y-6">
            {/* Balance card */}
            <div className="bg-dark-800 border border-dark-700 rounded-xl p-5">
              <div className="text-sm text-dark-400 mb-1">Address</div>
              <div className="font-mono text-sm text-republic-300 break-all">
                {wallet.address}
              </div>
              <div className="mt-4 text-sm text-dark-400 mb-1">Balance</div>
              <div className="text-3xl font-bold text-dark-100">
                {wallet.balance ? `${wallet.balance.amount} PNYX` : "Loading..."}
              </div>
            </div>

            {/* Send form */}
            <div className="bg-dark-800 border border-dark-700 rounded-xl p-5">
              <h3 className="text-lg font-medium mb-4">Send PNYX</h3>
              <div className="space-y-3">
                <div>
                  <label className="block text-xs font-medium text-dark-400 mb-1">
                    Recipient
                  </label>
                  <input
                    type="text"
                    placeholder="truerepublic1..."
                    value={recipient}
                    onChange={(e) => setRecipient(e.target.value)}
                    className="w-full px-3 py-2 text-sm bg-dark-700 border border-dark-600 rounded-lg text-dark-200 placeholder-dark-500 focus:outline-none focus:border-republic-500"
                  />
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
                <button
                  onClick={handleSend}
                  disabled={sending || !recipient || !amount}
                  className="w-full px-4 py-2.5 text-sm font-medium bg-republic-600 text-white rounded-lg hover:bg-republic-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {sending ? "Sending..." : "Send PNYX"}
                </button>
              </div>
            </div>
          </div>
        ) : (
          <div className="bg-dark-800 border border-dark-700 rounded-xl p-8 text-center">
            <div className="text-4xl mb-3">&#128176;</div>
            <p className="text-dark-300">Connect your wallet to get started</p>
          </div>
        )}
      </div>
    </div>
  );
}

export default Wallet;
