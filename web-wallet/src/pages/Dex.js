import React, { useState, useEffect } from "react";
import { SigningStargateClient } from "@cosmjs/stargate";

const RPC_ENDPOINT = "https://rpc.truerepublic.network";
const CHAIN_ID = "truerepublic-1";

function Dex() {
    const [wallet, setWallet] = useState(null);
    const [amount, setAmount] = useState("");
    const [fromAsset, setFromAsset] = useState("pnyx");
    const [toAsset, setToAsset] = useState("atom");
    const [pools, setPools] = useState([]);

    const connectWallet = async () => {
        if (!window.keplr) return alert("Keplr Wallet not installed!");
        await window.keplr.enable(CHAIN_ID);
        const offlineSigner = window.keplr.getOfflineSigner(CHAIN_ID);
        const accounts = await offlineSigner.getAccounts();
        setWallet(accounts[0].address);
    };

    const fetchPools = async () => {
        try {
            const client = await SigningStargateClient.connect(RPC_ENDPOINT);
            const result = await client.queryAbci("custom/dex/pools", new Uint8Array());
            const decoded = new TextDecoder().decode(result.value);
            setPools(JSON.parse(decoded));
        } catch (err) {
            console.error("Failed to fetch pools:", err);
        }
    };

    const swapTokens = async () => {
        if (!wallet || !amount) return alert("Please fill all fields.");
        if (fromAsset === toAsset) return alert("From and To assets must be different.");
        const offlineSigner = window.keplr.getOfflineSigner(CHAIN_ID);
        const client = await SigningStargateClient.connectWithSigner(RPC_ENDPOINT, offlineSigner);
        const msg = {
            typeUrl: "/dex.MsgSwap",
            value: {
                sender: wallet,
                input_denom: fromAsset,
                input_amt: Number(amount),
                output_denom: toAsset,
            },
        };
        const result = await client.signAndBroadcast(wallet, [msg], "auto");
        alert("Swap successful: " + result.transactionHash);
        fetchPools();
    };

    useEffect(() => {
        fetchPools();
    }, []);

    return (
        <div className="p-6 max-w-md mx-auto bg-gray-800 rounded-lg">
            <h2 className="text-2xl font-semibold mb-4">DEX</h2>
            <button onClick={connectWallet} className="px-4 py-2 bg-blue-500 rounded hover:bg-blue-600">Connect Wallet</button>
            {wallet && <p className="mt-2 text-sm text-gray-400">Connected: {wallet}</p>}

            {pools.length > 0 && (
                <div className="mt-4">
                    <h3 className="text-xl mb-2">Pools</h3>
                    <ul className="space-y-1">
                        {pools.map((p, i) => (
                            <li key={i} className="text-sm text-gray-300">
                                PNYX/{p.asset_denom}: {p.pnyx_reserve} / {p.asset_reserve} (burned: {p.total_burned})
                            </li>
                        ))}
                    </ul>
                </div>
            )}

            {wallet && (
                <div className="mt-6">
                    <h3 className="text-xl mb-2">Swap</h3>
                    <label className="text-sm text-gray-400">From</label>
                    <select
                        value={fromAsset}
                        onChange={(e) => setFromAsset(e.target.value)}
                        className="w-full p-2 mt-1 mb-2 bg-gray-700 rounded"
                    >
                        <option value="pnyx">PNYX</option>
                        <option value="atom">ATOM</option>
                    </select>
                    <label className="text-sm text-gray-400">To</label>
                    <select
                        value={toAsset}
                        onChange={(e) => setToAsset(e.target.value)}
                        className="w-full p-2 mt-1 mb-2 bg-gray-700 rounded"
                    >
                        <option value="atom">ATOM</option>
                        <option value="pnyx">PNYX</option>
                    </select>
                    <input
                        type="number"
                        placeholder="Amount"
                        value={amount}
                        onChange={(e) => setAmount(e.target.value)}
                        className="w-full p-2 mt-2 bg-gray-700 rounded"
                    />
                    <p className="text-xs text-gray-500 mt-1">Fee: 0.3% swap fee. 1% burn on PNYX output.</p>
                    <button onClick={swapTokens} className="w-full mt-4 px-4 py-2 bg-blue-500 rounded hover:bg-blue-600">Swap</button>
                </div>
            )}
        </div>
    );
}

export default Dex;
