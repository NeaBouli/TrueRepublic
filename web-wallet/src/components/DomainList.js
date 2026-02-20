import React from "react";

export default function DomainList({
  domains,
  selectedDomain,
  onSelectDomain,
  loading,
}) {
  return (
    <div>
      <h2 className="text-sm font-semibold text-dark-400 uppercase tracking-wider mb-3">
        Domains
      </h2>

      {loading ? (
        <div className="space-y-2">
          {[1, 2, 3].map((i) => (
            <div
              key={i}
              className="h-12 bg-dark-700 rounded-lg animate-pulse"
            />
          ))}
        </div>
      ) : domains.length === 0 ? (
        <p className="text-sm text-dark-400">No domains found.</p>
      ) : (
        <ul className="space-y-1">
          {domains.map((domain) => {
            const issueCount = domain.issues ? domain.issues.length : 0;
            const isSelected = selectedDomain === domain.name;
            return (
              <li key={domain.name}>
                <button
                  onClick={() => onSelectDomain(domain.name)}
                  className={`w-full text-left px-3 py-2.5 rounded-lg transition-colors ${
                    isSelected
                      ? "bg-republic-600/20 text-republic-300 border border-republic-600/30"
                      : "hover:bg-dark-700 text-dark-200"
                  }`}
                >
                  <div className="font-medium text-sm">{domain.name}</div>
                  <div className="text-xs text-dark-400 mt-0.5">
                    {issueCount} {issueCount === 1 ? "issue" : "issues"}
                  </div>
                </button>
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}
