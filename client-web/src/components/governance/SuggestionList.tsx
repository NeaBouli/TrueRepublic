import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useGovernanceStore } from '@/stores/governanceStore';
import { Card } from '@/components/common/Card';
import { VotingPanel } from '@/components/zkp/VotingPanel';
import type { Suggestion } from '@/types/governance';
import {
  ArrowLeftIcon,
  StarIcon,
  UserIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline';
import { formatAddress } from '@/utils/format';

function getZoneBorderColor(zone: Suggestion['zone']): string {
  switch (zone) {
    case 'green':
      return 'border-l-green-500';
    case 'yellow':
      return 'border-l-yellow-500';
    case 'red':
      return 'border-l-red-500';
    default:
      return 'border-l-gray-300';
  }
}

function getZoneBadgeClass(zone: Suggestion['zone']): string {
  switch (zone) {
    case 'green':
      return 'bg-green-100 text-green-800';
    case 'yellow':
      return 'bg-yellow-100 text-yellow-800';
    case 'red':
      return 'bg-red-100 text-red-800';
    default:
      return 'bg-gray-100 text-gray-800';
  }
}

function SuggestionCard({
  suggestion,
  onVote,
}: {
  suggestion: Suggestion;
  onVote: (suggestionId: string) => void;
}) {
  return (
    <Card className={`border-l-4 ${getZoneBorderColor(suggestion.zone)}`}>
      <div className="mb-4">
        <div className="flex items-start justify-between mb-2">
          <h3 className="text-lg font-semibold flex-1">
            {suggestion.title}
          </h3>
          <div className="flex items-center gap-1 text-yellow-600">
            <StarIcon className="h-5 w-5 fill-current" />
            <span className="font-bold">
              {suggestion.avgRating.toFixed(1)}
            </span>
          </div>
        </div>

        <p className="text-gray-600 text-sm mb-3">
          {suggestion.description}
        </p>

        <div className="flex items-center gap-2 text-xs text-gray-500 mb-3">
          <UserIcon className="h-4 w-4" />
          <span>by {formatAddress(suggestion.creator, 6)}</span>
        </div>
      </div>

      {/* Stones Display */}
      <div className="flex items-center gap-4 mb-3">
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 bg-green-500 rounded-full" />
          <span className="text-sm font-medium">
            {suggestion.greenStones}
          </span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 bg-yellow-500 rounded-full" />
          <span className="text-sm font-medium">
            {suggestion.yellowStones}
          </span>
        </div>
        <div className="flex items-center gap-1">
          <div className="w-3 h-3 bg-red-500 rounded-full" />
          <span className="text-sm font-medium">
            {suggestion.redStones}
          </span>
        </div>
      </div>

      {/* Rating Stats + Vote Button */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3 text-xs text-gray-500">
          <span>{suggestion.ratingCount} ratings</span>
          <span>&middot;</span>
          <span
            className={`px-2 py-0.5 rounded text-xs font-medium ${getZoneBadgeClass(suggestion.zone)}`}
          >
            {suggestion.zone.charAt(0).toUpperCase() +
              suggestion.zone.slice(1)}
          </span>
        </div>

        <button
          onClick={() => onVote(suggestion.suggestionId)}
          className="px-4 py-2 bg-primary-600 text-white text-sm font-medium rounded-lg hover:bg-primary-700 transition-colors"
        >
          Vote
        </button>
      </div>
    </Card>
  );
}

export function SuggestionList() {
  const navigate = useNavigate();
  const { domainId, issueId } = useParams<{
    domainId: string;
    issueId: string;
  }>();
  const { currentDomain, currentIssue, suggestions, selectIssue, isLoading } =
    useGovernanceStore();
  const [selectedForVoting, setSelectedForVoting] = useState<string | null>(
    null
  );

  useEffect(() => {
    if (domainId && issueId) {
      selectIssue(domainId, issueId);
    }
  }, [domainId, issueId, selectIssue]);

  const selectedSuggestion = selectedForVoting
    ? suggestions.find((s) => s.suggestionId === selectedForVoting)
    : null;

  if (isLoading && !currentIssue) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-gray-500">Loading...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-6xl mx-auto px-4 py-4">
          <button
            onClick={() => navigate(`/governance/domain/${domainId}`)}
            className="flex items-center gap-2 text-gray-600 hover:text-gray-900 mb-4"
          >
            <ArrowLeftIcon className="h-5 w-5" />
            Back to Issues
          </button>

          {currentDomain && currentIssue && (
            <div>
              <div className="text-sm text-gray-600 mb-1">
                {currentDomain.name}
              </div>
              <h1 className="text-2xl font-bold">{currentIssue.title}</h1>
              <p className="text-gray-600 mt-2">
                {currentIssue.description}
              </p>
            </div>
          )}
        </div>
      </header>

      <main className="max-w-6xl mx-auto px-4 py-8">
        <div className="mb-6">
          <h2 className="text-lg font-semibold mb-4">Suggestions</h2>

          {isLoading && (
            <div className="text-center py-12 text-gray-500">
              Loading suggestions...
            </div>
          )}

          {!isLoading && suggestions.length === 0 && (
            <Card>
              <div className="text-center py-12 text-gray-500">
                <p className="mb-2">No suggestions yet</p>
                <p className="text-sm">Be the first to make a suggestion</p>
              </div>
            </Card>
          )}

          <div className="space-y-4">
            {[...suggestions]
              .sort((a, b) => b.avgRating - a.avgRating)
              .map((suggestion) => (
                <SuggestionCard
                  key={suggestion.suggestionId}
                  suggestion={suggestion}
                  onVote={setSelectedForVoting}
                />
              ))}
          </div>
        </div>
      </main>

      {/* Voting Modal */}
      {selectedForVoting && selectedSuggestion && domainId && issueId && (
        <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
          <div className="bg-white rounded-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
            <div className="sticky top-0 bg-white border-b border-gray-200 px-6 py-4 flex items-center justify-between rounded-t-xl">
              <h3 className="text-lg font-semibold">Anonymous Voting</h3>
              <button
                onClick={() => setSelectedForVoting(null)}
                className="p-2 hover:bg-gray-100 rounded-lg transition-colors"
              >
                <XMarkIcon className="h-5 w-5 text-gray-600" />
              </button>
            </div>
            <div className="p-6">
              <VotingPanel
                suggestion={selectedSuggestion}
                domainId={domainId}
                issueName={issueId}
                onVoteSubmitted={() => {
                  setSelectedForVoting(null);
                }}
              />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
