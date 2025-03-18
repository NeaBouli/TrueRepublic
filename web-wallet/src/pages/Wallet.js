import React, { useState, useEffect } from "react";
import { SigningStargateClient } from "@cosmjs/stargate";

const RPC_ENDPOINT = "https://rpc.truerepublic.network";

function Wallet() {
    const [wallet, setWallet] = useState(null);
    const [balance, setBalance] = useState("Loading...");
    const [recipient, setRecipient] = useState("");
    const [amount, setAmount] = useState("");

    const connectWallet = async () => {
        if (!window.keplr) return alert("Keplr Wallet not installed!");
        try {
            await window.keplr.enable("truerepublic-1");
            const offlineSigner = window.keplr.getOfflineSigner("truerepublic-1");
            const accounts = await offlineSigner.getAccounts();
            setWallet(accounts[0].address);
            updateBalance(accounts[0].address);
        } catch (error) {
            setBalance("Error: " + error.message);
        }
    };

    const updateBalance = async (address) => {
        try {
            const client = await SigningStargateClient.connect(RPC_ENDPOINT);
            const balance = await client.getBalance(address, "pnyx");
            setBalance(`${balance.amount} PNYX`);
        } catch (error) {
            setBalance("Error: " + error.message);
        }
    };

    const sendPNYX = async () => {
        if (!wallet || !recipient || !amount) return alert("Please fill all fields.");
        try {
            const client = await SigningStargateClient.connectWithSigner(RPC_ENDPOINT, window.keplr.getOfflineSigner("truerepublic-1"));
            const result = await client.sendTokens(wallet, recipient, [{ denom: "pnyx", amount }], "auto");
            alert("Transaction successful: " + result.transactionHash);
            updateBalance(wallet);
        } catch (error) {
            alert("Error: " + error.message);
        }
    };

    useEffect(() => {
        if (wallet) {
            const interval = setInterval(() => updateBalance(wallet), 5000);
            return () => clearInterval(interval);
        }
    }, [wallet]);

    return (
        <div className="p-6 max-w-md mx-auto bg-gray-800 rounded-lg">
            <h2 className="text-2xl font-semibold mb-4">Wallet</h2>
            <button onClick={connectWallet} className="px-4 py-2 bg-blue-500 rounded hover:bg-blue-600">Connect Wallet</button>
            {wallet && (
                <>
                    <p className="mt-4">Address: <span className="text-blue-400">{wallet}</span></p>
                    <p className="mt-2">Balance: <span className="text-green-400">{balance}</span></p>
                    <input
                        type="text"
                        placeholder="Recipient Address"
                        value={recipient}
                        onChange={(e) => setRecipient(e.target.value)}
                        className="w-full p-2 mt-4 bg-gray-700 rounded"
                    />
                    <input
                        type="text"
                        placeholder="Amount (PNYX)"
                        value={amount}
                        onChange={(e) => setAmount(e.target.value)}
                        className="w-full p-2 mt-2 bg-gray-700 rounded"
                    />
                    <button onClick={sendPNYX} className="w-full mt-4 px-4 py-2 bg-blue-500 rounded hover:bg-blue-600">Send PNYX</button>
                </>
            )}
        </div>
    );
}

export default Wallet;
