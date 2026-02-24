package truedemocracy

import "sort"

// Systemic Consensing scoring (Whitepaper §3.2).
//
// Each suggestion has a list of Ratings with values from -5 to +5.
// The score is the sum of all ratings. The suggestion with the highest
// score is the "least resisted" option — the systemic consensus winner.

// ComputeSuggestionScore sums all rating values for a suggestion.
// Returns 0 if there are no ratings.
func ComputeSuggestionScore(s Suggestion) int {
	total := 0
	for _, r := range s.Ratings {
		total += r.Value
	}
	return total
}

// ScoredSuggestion pairs a suggestion with its computed consensus score.
type ScoredSuggestion struct {
	Name   string `json:"name"`
	Score  int    `json:"score"`
	Stones int    `json:"stones"`
	Count  int    `json:"rating_count"` // number of ratings received
}

// RankSuggestionsByScore returns suggestions sorted by consensus score
// descending. Ties are broken by stone count (more stones first), then
// by creation date (older first).
func RankSuggestionsByScore(suggestions []Suggestion) []ScoredSuggestion {
	scored := make([]ScoredSuggestion, len(suggestions))
	for i, s := range suggestions {
		scored[i] = ScoredSuggestion{
			Name:   s.Name,
			Score:  ComputeSuggestionScore(s),
			Stones: s.Stones,
			Count:  len(s.Ratings),
		}
	}
	// Keep original indices for stable tie-breaking by creation date.
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].Score != scored[j].Score {
			return scored[i].Score > scored[j].Score
		}
		return scored[i].Stones > scored[j].Stones
	})
	return scored
}

// FindConsensusWinner returns the suggestion with the highest score,
// or ("", 0) if there are no suggestions or no ratings.
func FindConsensusWinner(suggestions []Suggestion) (string, int) {
	if len(suggestions) == 0 {
		return "", 0
	}
	ranked := RankSuggestionsByScore(suggestions)
	if ranked[0].Count == 0 {
		return "", 0 // no ratings at all
	}
	return ranked[0].Name, ranked[0].Score
}
