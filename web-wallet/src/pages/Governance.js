import React, { useState, useEffect } from "react";
import { SigningStargateClient } from "@cosmjs/stargate";

const RPC_ENDPOINT = "https://rpc.truerepublic.network";
const CHAIN_ID = "truerepublic-1";

const chainConfig = {
    chainId: CHAIN_ID,
    chainName: "TrueRepublic",
    rpc: RPC_ENDPOINT,
    rest: "https://lcd.truerepublic.network",
    bip44: { coinType: 118 },
    bech32Config: {
        bech32PrefixAccAddr: "truerepublic",
        bech32PrefixAccPub: "truerepublicpub",
        bech32PrefixValAddr: "truerepublicvaloper",
        bech32PrefixValPub: "truerepublicvaloperpub",
        bech32PrefixConsAddr: "truerepublicvalcons",
        bech32PrefixConsPub: "truerepublicvalconspub",
    },
    currencies: [{ coinDenom: "PNYX", coinMinimalDenom: "pnyx", coinDecimals: 0 }],
    feeCurrencies: [{ coinDenom: "PNYX", coinMinimalDenom: "pnyx", coinDecimals: 0, gasPriceStep: { low: 0, average: 0, high: 0 } }],
    stakeCurrency: { coinDenom: "PNYX", coinMinimalDenom: "pnyx", coinDecimals: 0 },
};

function Governance() {
    const [wallet, setWallet] = useState(null);
    const [domains, setDomains] = useState([]);
    const [selectedDomain, setSelectedDomain] = useState("");
    const [title, setTitle] = useState("");
    const [description, setDescription] = useState("");
    const [vote, setVote] = useState("");
    const [selectedIssue, setSelectedIssue] = useState("");

    const connectWallet = async () => {
        if (!window.keplr) return alert("Keplr Wallet not installed!");
        await window.keplr.experimentalSuggestChain(chainConfig);
        await window.keplr.enable(CHAIN_ID);
        const offlineSigner = window.keplr.getOfflineSigner(CHAIN_ID);
        const accounts = await offlineSigner.getAccounts();
        setWallet(accounts[0].address);
    };

    const fetchDomains = async () => {
        try {
            const client = await SigningStargateClient.connect(RPC_ENDPOINT);
            const result = await client.queryAbci("custom/truedemocracy/domains", new Uint8Array());
            const decoded = new TextDecoder().decode(result.value);
            setDomains(JSON.parse(decoded));
        } catch (err) {
            console.error("Failed to fetch domains:", err);
        }
    };

    const submitProposal = async () => {
        if (!wallet || !selectedDomain || !title) return alert("Please fill all fields.");
        const offlineSigner = window.keplr.getOfflineSigner(CHAIN_ID);
        const client = await SigningStargateClient.connectWithSigner(RPC_ENDPOINT, offlineSigner);
        const msg = {
            typeUrl: "/truedemocracy.MsgSubmitProposal",
            value: {
                sender: wallet,
                domain_name: selectedDomain,
                issue_name: title,
                suggestion_name: description,
                creator: wallet,
                fee: [],
            },
        };
        const result = await client.signAndBroadcast(wallet, [msg], "auto");
        alert("Proposal submitted: " + result.transactionHash);
        fetchDomains();
    };

    useEffect(() => {
        fetchDomains();
    }, []);

    const currentDomain = domains.find(d => d.name === selectedDomain);
    const issues = currentDomain ? currentDomain.issues || [] : [];

    return (
        <div className="p-6 max-w-md mx-auto bg-gray-800 rounded-lg">
            <h2 className="text-2xl font-semibold mb-4">Governance</h2>
            <button onClick={connectWallet} className="px-4 py-2 bg-blue-500 rounded hover:bg-blue-600">Connect Wallet</button>
            {wallet && <p className="mt-2 text-sm text-gray-400">Connected: {wallet}</p>}

            <div className="mt-4">
                <h3 className="text-xl mb-2">Domains</h3>
                <select
                    value={selectedDomain}
                    onChange={(e) => setSelectedDomain(e.target.value)}
                    className="w-full p-2 mb-2 bg-gray-700 rounded"
                >
                    <option value="">Select Domain</option>
                    {domains.map(d => <option key={d.name} value={d.name}>{d.name}</option>)}
                </select>
            </div>

            {selectedDomain && (
                <>
                    <div className="mt-4">
                        <h3 className="text-xl mb-2">Issues</h3>
                        {issues.length === 0 ? (
                            <p className="text-gray-400">No issues in this domain.</p>
                        ) : (
                            <ul className="space-y-2">
                                {issues.map((issue, i) => (
                                    <li key={i} className="p-2 bg-gray-700 rounded">
                                        <strong>{issue.name}</strong> — {issue.stones} stones
                                        <ul className="ml-4 mt-1">
                                            {(issue.suggestions || []).map((s, j) => (
                                                <li key={j} className="text-sm text-gray-300">
                                                    {s.name} ({s.color}) — {s.stones} stones
                                                </li>
                                            ))}
                                        </ul>
                                    </li>
                                ))}
                            </ul>
                        )}
                    </div>
                    {wallet && (
                        <div className="mt-6">
                            <h3 className="text-xl mb-2">Submit Proposal</h3>
                            <input
                                type="text"
                                placeholder="Issue name"
                                value={title}
                                onChange={(e) => setTitle(e.target.value)}
                                className="w-full p-2 mb-2 bg-gray-700 rounded"
                            />
                            <input
                                type="text"
                                placeholder="Suggestion name"
                                value={description}
                                onChange={(e) => setDescription(e.target.value)}
                                className="w-full p-2 mb-2 bg-gray-700 rounded"
                            />
                            <button onClick={submitProposal} className="w-full px-4 py-2 bg-blue-500 rounded hover:bg-blue-600">Submit</button>
                        </div>
                    )}
                </>
            )}
        </div>
    );
}

export default Governance;
