import React from "react";
import { Link, useLocation } from "react-router-dom";

const navItems = [
  { path: "/", label: "Governance" },
  { path: "/wallet", label: "Wallet" },
  { path: "/dex", label: "DEX" },
];

export default function Header({ address, onConnect, onDisconnect, loading }) {
  const location = useLocation();

  const shortAddress = address
    ? `${address.slice(0, 12)}...${address.slice(-6)}`
    : null;

  return (
    <div className="flex items-center justify-between px-4 py-3">
      {/* Logo + Nav */}
      <div className="flex items-center gap-6">
        <Link to="/" className="flex items-center gap-2">
          <img src="/logo.svg" alt="TrueRepublic" className="h-8" />
          <span className="text-lg font-bold text-dark-50 hidden sm:inline">
            TrueRepublic
          </span>
        </Link>

        <nav className="flex gap-1">
          {navItems.map((item) => (
            <Link
              key={item.path}
              to={item.path}
              className={`px-3 py-1.5 rounded-md text-sm font-medium transition-colors ${
                location.pathname === item.path
                  ? "bg-republic-600 text-white"
                  : "text-dark-300 hover:text-white hover:bg-dark-700"
              }`}
            >
              {item.label}
            </Link>
          ))}
        </nav>
      </div>

      {/* Wallet connection */}
      <div className="flex items-center gap-3">
        {address ? (
          <>
            <span className="text-sm text-dark-300 hidden sm:inline font-mono">
              {shortAddress}
            </span>
            <button
              onClick={onDisconnect}
              className="px-3 py-1.5 text-sm rounded-md bg-dark-700 text-dark-300 hover:bg-dark-600 hover:text-white transition-colors"
            >
              Disconnect
            </button>
          </>
        ) : (
          <button
            onClick={onConnect}
            disabled={loading}
            className="px-4 py-1.5 text-sm rounded-md bg-republic-600 text-white hover:bg-republic-700 transition-colors disabled:opacity-50"
          >
            {loading ? "Connecting..." : "Connect Wallet"}
          </button>
        )}
      </div>
    </div>
  );
}
