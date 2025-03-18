import React, { useState, useEffect } from "react";
import { SigningStargateClient } from "@cosmjs/stargate";

const RPC_ENDPOINT = "https://rpc.truerepublic.network";
const GOVERNANCE_ADDRESS = "truerepublic1gov...";

function Governance() {
    const [wallet, setWallet] = useState(null);
    const [proposals, setProposals] = useState([]);
    const [title, setTitle] = useState("");
    const [description, setDescription] = useState("");
    const [selectedProposal, setSelectedProposal] = useState("");
    const [vote, setVote] = useState("");

    const connectWallet = async () => {
        if (!window.keplr) return alert("Keplr Wallet not installed!");
        await window.keplr.enable("truerepublic-1");
        const offlineSigner = window.keplr.getOfflineSigner("truerepublic-1");
        const accounts = await offlineSigner.getAccounts();
        setWallet(accounts[0].address);
        fetchProposals();
    };

    const fetchProposals = async () => {
        const client = await SigningStargateClient.connect(RPC_ENDPOINT);
        const result = await client.queryContractSmart(GOVERNANCE_ADDRESS, { get_proposals: {} });
        setProposals(result);
    };

    const submitProposal = async () => {
        if (!wallet || !title || !description) return alert("Please fill all fields.");
        const client = await SigningStargateClient.connectWithSigner(RPC_ENDPOINT, window.keplr.getOfflineSigner("truerepublic-1"));
        const msg = { submit_proposal: { title, description } };
        const result = await client.execute(wallet, GOVERNANCE_ADDRESS, msg, "auto");
        alert("Proposal submitted: " + result.transactionHash);
        fetchProposals();
    };

    const voteProposal = async () => {
        if (!wallet || !selectedProposal || !vote) return alert("Please select proposal and vote.");
        const client = await SigningStargateClient.connectWithSigner(RPC_ENDPOINT, window.keplr.getOfflineSigner("truerepublic-1"));
        const msg = { vote: { proposal_id: Number(selectedProposal), vote: Number(vote) } };
        const result = await client.execute(wallet, GOVERNANCE_ADDRESS, msg, "auto");
        alert("Vote submitted: " + result.transactionHash);
        fetchProposals();
    };

    useEffect(() => {
        if (wallet) fetchProposals();
    }, [wallet]);

    return (
        <div className="p-6 max-w-md mx-auto bg-gray-800 rounded-lg">
            <h2 className="text-2xl font-semibold mb-4">Governance</h2>
            <button onClick={connectWallet} className="px-4 py-2 bg-blue-500 rounded hover:bg-blue-600">Connect Wallet</button>
            {wallet && (
                <>
                    <div className="mt-6">
                        <h3 className="text-xl mb-2">Submit Proposal</h3>
                        <input
                            type="text"
                            placeholder="Title"
                            value={title}
                            onChange={(e) => setTitle(e.target.value)}
                            className="w-full p-2 mb-2 bg-gray-700 rounded"
                        />
                        <textarea
                            placeholder="Description"
                            value={description}
                            onChange={(e) => setDescription(e.target.value)}
                            className="w-full p-2 mb-2 bg-gray-700 rounded"
                        />
                        <button onClick={submitProposal} className="w-full px-4 py-2 bg-blue-500 rounded hover:bg-blue-600">Submit</button>
                    </div>
                    <div className="mt-6">
                        <h3 className="text-xl mb-2">Vote</h3>
                        <select
                            value={selectedProposal}
                            onChange={(e) => setSelectedProposal(e.target.value)}
                            className="w-full p-2 mb-2 bg-gray-700 rounded"
                        >
                            <option value="">Select Proposal</option>
                            {proposals.map(p => <option key={p.id} value={p.id}>{p.title}</option>)}
                        </select>
                        <input
                            type="number"
                            min="-5"
                            max="5"
                            placeholder="Vote (-5 to +5)"
                            value={vote}
                            onChange={(e) => setVote(e.target.value)}
                            className="w-full p-2 mb-2 bg-gray-700 rounded"
                        />
                        <button onClick={voteProposal} className="w-full px-4 py-2 bg-blue-500 rounded hover:bg-blue-600">Vote</button>
                    </div>
                </>
            )}
        </div>
    );
}

export default Governance;
