import React, { useState } from "react";

function SuggestionBar({ suggestion, maxStones }) {
  const percentage = maxStones > 0 ? (suggestion.stones / maxStones) * 100 : 0;
  const colorMap = {
    green: "bg-green-500",
    red: "bg-red-500",
    blue: "bg-blue-500",
    yellow: "bg-yellow-500",
    purple: "bg-purple-500",
    orange: "bg-orange-500",
  };
  const barColor = colorMap[suggestion.color] || "bg-republic-500";

  return (
    <div className="group">
      <div className="flex items-center justify-between text-sm mb-1">
        <span className="text-dark-200">{suggestion.name}</span>
        <span className="text-dark-400 text-xs">
          {suggestion.stones} stones
        </span>
      </div>
      <div className="h-2 bg-dark-700 rounded-full overflow-hidden">
        <div
          className={`h-full rounded-full transition-all duration-500 ${barColor}`}
          style={{ width: `${Math.max(percentage, 2)}%` }}
        />
      </div>
    </div>
  );
}

function IssueCard({ issue, domainName, onVote, connected }) {
  const [expanded, setExpanded] = useState(false);
  const [voteTarget, setVoteTarget] = useState("");
  const [stoneCount, setStoneCount] = useState("1");

  const suggestions = issue.suggestions || [];
  const maxStones = Math.max(...suggestions.map((s) => s.stones || 0), 1);

  return (
    <div className="bg-dark-800 border border-dark-700 rounded-xl p-4 transition-colors hover:border-dark-600">
      <div
        className="flex items-start justify-between cursor-pointer"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex-1">
          <h3 className="text-base font-medium text-dark-100">{issue.name}</h3>
          <div className="flex items-center gap-3 mt-1 text-xs text-dark-400">
            <span>{issue.stones || 0} total stones</span>
            <span>{suggestions.length} suggestions</span>
          </div>
        </div>
        <span className="text-dark-500 text-sm ml-2">
          {expanded ? "\u25B2" : "\u25BC"}
        </span>
      </div>

      {expanded && (
        <div className="mt-4 space-y-3">
          {suggestions.length === 0 ? (
            <p className="text-sm text-dark-400">No suggestions yet.</p>
          ) : (
            suggestions.map((s, i) => (
              <SuggestionBar key={i} suggestion={s} maxStones={maxStones} />
            ))
          )}

          {connected && onVote && (
            <div className="mt-4 pt-3 border-t border-dark-700">
              <div className="flex gap-2">
                <select
                  value={voteTarget}
                  onChange={(e) => setVoteTarget(e.target.value)}
                  className="flex-1 px-3 py-1.5 text-sm bg-dark-700 border border-dark-600 rounded-lg text-dark-200 focus:outline-none focus:border-republic-500"
                >
                  <option value="">Select suggestion</option>
                  {suggestions.map((s, i) => (
                    <option key={i} value={s.name}>
                      {s.name}
                    </option>
                  ))}
                </select>
                <input
                  type="number"
                  min="1"
                  value={stoneCount}
                  onChange={(e) => setStoneCount(e.target.value)}
                  className="w-20 px-3 py-1.5 text-sm bg-dark-700 border border-dark-600 rounded-lg text-dark-200 focus:outline-none focus:border-republic-500"
                  placeholder="Stones"
                />
                <button
                  onClick={() => {
                    if (voteTarget && stoneCount) {
                      onVote(domainName, issue.name, voteTarget, stoneCount);
                    }
                  }}
                  className="px-4 py-1.5 text-sm bg-republic-600 text-white rounded-lg hover:bg-republic-700 transition-colors"
                >
                  Vote
                </button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default function ProposalFeed({
  domain,
  domainName,
  onVote,
  connected,
}) {
  if (!domainName) {
    return (
      <div className="flex items-center justify-center h-64 text-dark-400">
        <div className="text-center">
          <div className="text-4xl mb-3">&#127963;</div>
          <p className="text-lg font-medium">Select a domain</p>
          <p className="text-sm mt-1">
            Choose a governance domain from the left to view issues
          </p>
        </div>
      </div>
    );
  }

  const issues = domain?.issues || [];

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-xl font-semibold text-dark-100">{domainName}</h2>
        <span className="text-sm text-dark-400">
          {issues.length} {issues.length === 1 ? "issue" : "issues"}
        </span>
      </div>

      {issues.length === 0 ? (
        <div className="text-center py-12 text-dark-400">
          <p>No issues in this domain yet.</p>
          <p className="text-sm mt-1">Be the first to submit a proposal!</p>
        </div>
      ) : (
        <div className="space-y-3">
          {issues.map((issue, i) => (
            <IssueCard
              key={i}
              issue={issue}
              domainName={domainName}
              onVote={onVote}
              connected={connected}
            />
          ))}
        </div>
      )}
    </div>
  );
}
