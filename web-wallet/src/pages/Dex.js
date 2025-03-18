import React, { useState } from "react";
import { SigningStargateClient } from "@cosmjs/stargate";

const RPC_ENDPOINT = "https://rpc.truerepublic.network";
const DEX_ADDRESS = "truerepublic1dex...";

function Dex() {
    const [wallet, setWallet] = useState(null);
    const [amount, setAmount] = useState("");
    const [fromAsset, setFromAsset] = useState("pnyx");
    const [toAsset, setToAsset] = useState("btc");

    const connectWallet = async () => {
        if (!window.keplr) return alert("Keplr Wallet not installed!");
        await window.keplr.enable("truerepublic-1");
        const offlineSigner = window.keplr.getOfflineSigner("truerepublic-1");
        const accounts = await offlineSigner.getAccounts();
        setWallet(accounts[0].address);
    };

    const swapTokens = async () => {
        if (!wallet || !amount) return alert("Please fill all fields.");
        const client = await SigningStargateClient.connectWithSigner(RPC_ENDPOINT, window.keplr.getOfflineSigner("truerepublic-1"));
        const msg = { swap: { from_asset: fromAsset, to_asset: toAsset, amount } };
        const result = await client.execute(wallet, DEX_ADDRESS, msg, "auto");
        alert("Swap successful: " + result.transactionHash);
    };

    return (
        <div className="p-6 max-w-md mx-auto bg-gray-800 rounded-lg">
            <h2 className="text-2xl font-semibold mb-4">DEX</h2>
            <button onClick={connectWallet} className="px-4 py-2 bg-blue-500 rounded hover:bg-blue-600">Connect Wallet</button>
            {wallet && (
                <>
                    <select
                        value={fromAsset}
                        onChange={(e) => setFromAsset(e.target.value)}
                        className="w-full p-2 mt-4 bg-gray-700 rounded"
                    >
                        <option value="pnyx">PNYX</option>
                        <option value="btc">BTC</option>
                        <option value="atom">ATOM</option>
                    </select>
                    <select
                        value={toAsset}
                        onChange={(e) => setToAsset(e.target.value)}
                        className="w-full p-2 mt-2 bg-gray-700 rounded"
                    >
                        <option value="btc">BTC</option>
                        <option value="pnyx">PNYX</option>
                        <option value="atom">ATOM</option>
                    </select>
                    <input
                        type="text"
                        placeholder="Amount"
                        value={amount}
                        onChange={(e) => setAmount(e.target.value)}
                        className="w-full p-2 mt-2 bg-gray-700 rounded"
                    />
                    <button onClick={swapTokens} className="w-full mt-4 px-4 py-2 bg-blue-500 rounded hover:bg-blue-600">Swap</button>
                </>
            )}
        </div>
    );
}

export default Dex;
