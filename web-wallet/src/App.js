import React, { useState, useEffect, useCallback } from "react";
import ThreeColumnLayout from "./components/ThreeColumnLayout";
import Header from "./components/Header";
import DomainList from "./components/DomainList";
import ProposalFeed from "./components/ProposalFeed";
import DomainInfo from "./components/DomainInfo";
import useWallet from "./hooks/useWallet";
import { fetchDomains, submitProposal, castVote } from "./services/api";

function App() {
  const wallet = useWallet();
  const [domains, setDomains] = useState([]);
  const [selectedDomain, setSelectedDomain] = useState("");
  const [domainsLoading, setDomainsLoading] = useState(true);

  const loadDomains = useCallback(async () => {
    setDomainsLoading(true);
    try {
      const data = await fetchDomains();
      setDomains(data || []);
    } catch (err) {
      console.error("Failed to fetch domains:", err);
      setDomains([]);
    } finally {
      setDomainsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadDomains();
  }, [loadDomains]);

  const currentDomain = domains.find((d) => d.name === selectedDomain);

  const handleSubmitProposal = async (domainName, issueName, suggestionName) => {
    const result = await submitProposal(
      wallet.address,
      domainName,
      issueName,
      suggestionName
    );
    alert("Proposal submitted! TX: " + result.transactionHash);
    loadDomains();
  };

  const handleVote = async (domainName, issueName, suggestionName, stones) => {
    const result = await castVote(
      wallet.address,
      domainName,
      issueName,
      suggestionName,
      stones
    );
    alert("Vote cast! TX: " + result.transactionHash);
    loadDomains();
  };

  return (
    <ThreeColumnLayout
      header={
        <Header
          address={wallet.address}
          onConnect={wallet.connect}
          onDisconnect={wallet.disconnect}
          loading={wallet.loading}
        />
      }
      left={
        <DomainList
          domains={domains}
          selectedDomain={selectedDomain}
          onSelectDomain={setSelectedDomain}
          loading={domainsLoading}
        />
      }
      center={
        <ProposalFeed
          domain={currentDomain}
          domainName={selectedDomain}
          onVote={handleVote}
          connected={wallet.connected}
          address={wallet.address}
        />
      }
      right={
        <DomainInfo
          domain={currentDomain}
          domainName={selectedDomain}
          connected={wallet.connected}
          onSubmitProposal={handleSubmitProposal}
        />
      }
    />
  );
}

export default App;
