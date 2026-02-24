package truedemocracy

import "testing"

// ---------- ComputeSuggestionScore ----------

func TestComputeSuggestionScore(t *testing.T) {
	t.Run("whitepaper example ice cream", func(t *testing.T) {
		// WP §3.2: Fritz(-3) + Anna(4) + George(3) = 4
		s := Suggestion{
			Name: "Ice cream",
			Ratings: []Rating{
				{DomainPubKeyHex: "fritz", Value: -3},
				{DomainPubKeyHex: "anna", Value: 4},
				{DomainPubKeyHex: "george", Value: 3},
			},
		}
		score := ComputeSuggestionScore(s)
		if score != 4 {
			t.Errorf("ice cream score = %d, want 4", score)
		}
	})

	t.Run("whitepaper example waffles", func(t *testing.T) {
		// WP §3.2: Fritz(0) + Anna(4) + George(2) = 6
		s := Suggestion{
			Name: "Waffles",
			Ratings: []Rating{
				{DomainPubKeyHex: "fritz", Value: 0},
				{DomainPubKeyHex: "anna", Value: 4},
				{DomainPubKeyHex: "george", Value: 2},
			},
		}
		score := ComputeSuggestionScore(s)
		if score != 6 {
			t.Errorf("waffles score = %d, want 6", score)
		}
	})

	t.Run("whitepaper example beer", func(t *testing.T) {
		// WP §3.2: Fritz(1) + Anna(5) + George(-5) = 1
		s := Suggestion{
			Name: "Beer",
			Ratings: []Rating{
				{DomainPubKeyHex: "fritz", Value: 1},
				{DomainPubKeyHex: "anna", Value: 5},
				{DomainPubKeyHex: "george", Value: -5},
			},
		}
		score := ComputeSuggestionScore(s)
		if score != 1 {
			t.Errorf("beer score = %d, want 1", score)
		}
	})

	t.Run("whitepaper example default nothing", func(t *testing.T) {
		// WP §3.2: Fritz(-5) + Anna(-5) + George(-4) = -14
		s := Suggestion{
			Name: "Default (nothing)",
			Ratings: []Rating{
				{DomainPubKeyHex: "fritz", Value: -5},
				{DomainPubKeyHex: "anna", Value: -5},
				{DomainPubKeyHex: "george", Value: -4},
			},
		}
		score := ComputeSuggestionScore(s)
		if score != -14 {
			t.Errorf("default score = %d, want -14", score)
		}
	})

	t.Run("no ratings", func(t *testing.T) {
		s := Suggestion{Name: "Empty", Ratings: nil}
		if score := ComputeSuggestionScore(s); score != 0 {
			t.Errorf("empty score = %d, want 0", score)
		}
	})

	t.Run("all negative", func(t *testing.T) {
		s := Suggestion{
			Name: "Bad",
			Ratings: []Rating{
				{Value: -5}, {Value: -3}, {Value: -4},
			},
		}
		if score := ComputeSuggestionScore(s); score != -12 {
			t.Errorf("all negative score = %d, want -12", score)
		}
	})

	t.Run("all positive", func(t *testing.T) {
		s := Suggestion{
			Name: "Great",
			Ratings: []Rating{
				{Value: 5}, {Value: 4}, {Value: 3},
			},
		}
		if score := ComputeSuggestionScore(s); score != 12 {
			t.Errorf("all positive score = %d, want 12", score)
		}
	})

	t.Run("single rating", func(t *testing.T) {
		s := Suggestion{
			Name:    "Solo",
			Ratings: []Rating{{Value: -2}},
		}
		if score := ComputeSuggestionScore(s); score != -2 {
			t.Errorf("single score = %d, want -2", score)
		}
	})
}

// ---------- RankSuggestionsByScore ----------

func TestRankSuggestionsByScore(t *testing.T) {
	t.Run("whitepaper full example", func(t *testing.T) {
		// WP §3.2: Waffles(6) > Ice cream(4) > Beer(1) > Default(-14)
		suggestions := []Suggestion{
			{
				Name: "Ice cream",
				Ratings: []Rating{
					{Value: -3}, {Value: 4}, {Value: 3},
				},
			},
			{
				Name: "Waffles",
				Ratings: []Rating{
					{Value: 0}, {Value: 4}, {Value: 2},
				},
			},
			{
				Name: "Beer",
				Ratings: []Rating{
					{Value: 1}, {Value: 5}, {Value: -5},
				},
			},
			{
				Name: "Default (nothing)",
				Ratings: []Rating{
					{Value: -5}, {Value: -5}, {Value: -4},
				},
			},
		}

		ranked := RankSuggestionsByScore(suggestions)
		if len(ranked) != 4 {
			t.Fatalf("ranked len = %d, want 4", len(ranked))
		}

		expected := []struct {
			name  string
			score int
		}{
			{"Waffles", 6},
			{"Ice cream", 4},
			{"Beer", 1},
			{"Default (nothing)", -14},
		}
		for i, e := range expected {
			if ranked[i].Name != e.name || ranked[i].Score != e.score {
				t.Errorf("rank[%d] = {%q, %d}, want {%q, %d}",
					i, ranked[i].Name, ranked[i].Score, e.name, e.score)
			}
		}
	})

	t.Run("tie broken by stones", func(t *testing.T) {
		suggestions := []Suggestion{
			{Name: "A", Stones: 2, Ratings: []Rating{{Value: 3}}},
			{Name: "B", Stones: 5, Ratings: []Rating{{Value: 3}}},
		}
		ranked := RankSuggestionsByScore(suggestions)
		if ranked[0].Name != "B" {
			t.Errorf("expected B first (more stones), got %q", ranked[0].Name)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		ranked := RankSuggestionsByScore(nil)
		if len(ranked) != 0 {
			t.Errorf("ranked len = %d, want 0", len(ranked))
		}
	})

	t.Run("rating count populated", func(t *testing.T) {
		suggestions := []Suggestion{
			{Name: "X", Ratings: []Rating{{Value: 1}, {Value: 2}, {Value: 3}}},
		}
		ranked := RankSuggestionsByScore(suggestions)
		if ranked[0].Count != 3 {
			t.Errorf("count = %d, want 3", ranked[0].Count)
		}
	})
}

// ---------- FindConsensusWinner ----------

func TestFindConsensusWinner(t *testing.T) {
	t.Run("whitepaper winner is waffles", func(t *testing.T) {
		suggestions := []Suggestion{
			{Name: "Ice cream", Ratings: []Rating{{Value: -3}, {Value: 4}, {Value: 3}}},
			{Name: "Waffles", Ratings: []Rating{{Value: 0}, {Value: 4}, {Value: 2}}},
			{Name: "Beer", Ratings: []Rating{{Value: 1}, {Value: 5}, {Value: -5}}},
		}
		winner, score := FindConsensusWinner(suggestions)
		if winner != "Waffles" || score != 6 {
			t.Errorf("winner = %q (%d), want Waffles (6)", winner, score)
		}
	})

	t.Run("no suggestions", func(t *testing.T) {
		winner, score := FindConsensusWinner(nil)
		if winner != "" || score != 0 {
			t.Errorf("empty winner = %q (%d), want '' (0)", winner, score)
		}
	})

	t.Run("no ratings returns no winner", func(t *testing.T) {
		suggestions := []Suggestion{
			{Name: "A", Ratings: nil},
			{Name: "B", Ratings: []Rating{}},
		}
		winner, _ := FindConsensusWinner(suggestions)
		if winner != "" {
			t.Errorf("no-ratings winner = %q, want ''", winner)
		}
	})
}
