import { useState, useEffect, useCallback } from "react";
import { CHAIN_ID, chainConfig, getBalance } from "../services/api";

export default function useWallet() {
  const [address, setAddress] = useState(null);
  const [balance, setBalance] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const updateBalance = useCallback(async (addr) => {
    try {
      const bal = await getBalance(addr);
      setBalance(bal);
    } catch (err) {
      console.error("Failed to fetch balance:", err);
    }
  }, []);

  const connect = useCallback(async () => {
    if (!window.keplr) {
      setError("Keplr Wallet not installed! Please install the Keplr browser extension.");
      return;
    }
    setLoading(true);
    setError(null);
    try {
      await window.keplr.experimentalSuggestChain(chainConfig);
      await window.keplr.enable(CHAIN_ID);
      const offlineSigner = window.keplr.getOfflineSigner(CHAIN_ID);
      const accounts = await offlineSigner.getAccounts();
      const addr = accounts[0].address;
      setAddress(addr);
      await updateBalance(addr);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [updateBalance]);

  const disconnect = useCallback(() => {
    setAddress(null);
    setBalance(null);
    setError(null);
  }, []);

  // Auto-refresh balance every 10 seconds
  useEffect(() => {
    if (!address) return;
    const interval = setInterval(() => updateBalance(address), 10000);
    return () => clearInterval(interval);
  }, [address, updateBalance]);

  // Listen for Keplr account changes
  useEffect(() => {
    const handler = () => {
      if (address) connect();
    };
    window.addEventListener("keplr_keystorechange", handler);
    return () => window.removeEventListener("keplr_keystorechange", handler);
  }, [address, connect]);

  return {
    address,
    balance,
    loading,
    error,
    connected: !!address,
    connect,
    disconnect,
    refreshBalance: () => address && updateBalance(address),
  };
}
