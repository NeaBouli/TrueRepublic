package truedemocracy

import (
	"encoding/json"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// --- Custom Query Handler Tests ---

func TestWasmQueryDomain(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "TestDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 500)))

	handler := CustomQueryHandler(k)

	t.Run("found", func(t *testing.T) {
		reqBytes, _ := json.Marshal(WasmCustomQuery{
			Domain: &WasmQueryDomain{Name: "TestDomain"},
		})
		respBytes, err := handler(ctx, reqBytes)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var resp WasmDomainResponse
		if err := json.Unmarshal(respBytes, &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp.Name != "TestDomain" {
			t.Errorf("name = %q, want TestDomain", resp.Name)
		}
		if resp.MemberCount != 1 {
			t.Errorf("member_count = %d, want 1", resp.MemberCount)
		}
		if resp.IssueCount != 0 {
			t.Errorf("issue_count = %d, want 0", resp.IssueCount)
		}
	})

	t.Run("not found", func(t *testing.T) {
		reqBytes, _ := json.Marshal(WasmCustomQuery{
			Domain: &WasmQueryDomain{Name: "NoSuchDomain"},
		})
		_, err := handler(ctx, reqBytes)
		if err == nil {
			t.Fatal("expected error for missing domain")
		}
	})
}

func TestWasmQueryDomainMembers(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "MembersDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100)))

	// Add extra members.
	domain, _ := k.GetDomain(ctx, "MembersDomain")
	domain.Members = append(domain.Members, "alice", "bob")
	st := ctx.KVStore(k.StoreKey)
	st.Set([]byte("domain:MembersDomain"), k.cdc.MustMarshalLengthPrefixed(&domain))

	handler := CustomQueryHandler(k)

	t.Run("returns all members", func(t *testing.T) {
		reqBytes, _ := json.Marshal(WasmCustomQuery{
			DomainMembers: &WasmQueryDomainMembers{DomainName: "MembersDomain"},
		})
		respBytes, err := handler(ctx, reqBytes)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var resp WasmDomainMembersResponse
		if err := json.Unmarshal(respBytes, &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if len(resp.Members) != 3 {
			t.Errorf("members count = %d, want 3", len(resp.Members))
		}
	})

	t.Run("domain not found", func(t *testing.T) {
		reqBytes, _ := json.Marshal(WasmCustomQuery{
			DomainMembers: &WasmQueryDomainMembers{DomainName: "Missing"},
		})
		_, err := handler(ctx, reqBytes)
		if err == nil {
			t.Fatal("expected error for missing domain")
		}
	})
}

func TestWasmQueryIssue(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "IssueDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100)))

	// Add issue with suggestions.
	domain, _ := k.GetDomain(ctx, "IssueDomain")
	domain.Issues = []Issue{
		{
			Name:         "Climate",
			Stones:       5,
			CreationDate: 1000,
			Suggestions: []Suggestion{
				{Name: "SolarPlan", Creator: "alice", Stones: 3, Color: "green", Ratings: []Rating{{Value: 4}}, CreationDate: 1001},
				{Name: "WindPlan", Creator: "bob", Stones: 1, Color: "yellow", Ratings: nil, CreationDate: 1002},
			},
		},
	}
	st := ctx.KVStore(k.StoreKey)
	st.Set([]byte("domain:IssueDomain"), k.cdc.MustMarshalLengthPrefixed(&domain))

	handler := CustomQueryHandler(k)

	t.Run("found with suggestions", func(t *testing.T) {
		reqBytes, _ := json.Marshal(WasmCustomQuery{
			Issue: &WasmQueryIssue{DomainName: "IssueDomain", IssueName: "Climate"},
		})
		respBytes, err := handler(ctx, reqBytes)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var resp WasmIssueResponse
		if err := json.Unmarshal(respBytes, &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if resp.Name != "Climate" {
			t.Errorf("name = %q, want Climate", resp.Name)
		}
		if resp.Stones != 5 {
			t.Errorf("stones = %d, want 5", resp.Stones)
		}
		if resp.SuggestionCount != 2 {
			t.Errorf("suggestion_count = %d, want 2", resp.SuggestionCount)
		}
		if len(resp.Suggestions) != 2 {
			t.Fatalf("suggestions len = %d, want 2", len(resp.Suggestions))
		}
		if resp.Suggestions[0].Name != "SolarPlan" {
			t.Errorf("first suggestion = %q, want SolarPlan", resp.Suggestions[0].Name)
		}
	})

	t.Run("issue not found", func(t *testing.T) {
		reqBytes, _ := json.Marshal(WasmCustomQuery{
			Issue: &WasmQueryIssue{DomainName: "IssueDomain", IssueName: "NoIssue"},
		})
		_, err := handler(ctx, reqBytes)
		if err == nil {
			t.Fatal("expected error for missing issue")
		}
	})

	t.Run("domain not found", func(t *testing.T) {
		reqBytes, _ := json.Marshal(WasmCustomQuery{
			Issue: &WasmQueryIssue{DomainName: "NoSuchDomain", IssueName: "Climate"},
		})
		_, err := handler(ctx, reqBytes)
		if err == nil {
			t.Fatal("expected error for missing domain")
		}
	})
}

func TestWasmQuerySuggestion(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "SugDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100)))

	domain, _ := k.GetDomain(ctx, "SugDomain")
	domain.Issues = []Issue{
		{
			Name: "Energy",
			Suggestions: []Suggestion{
				{
					Name:         "Nuclear",
					Creator:      "charlie",
					Stones:       7,
					Color:        "green",
					Ratings:      []Rating{{Value: 3}, {Value: -2}},
					DwellTime:    86400,
					CreationDate: 2000,
					ExternalLink: "https://example.com",
					DeleteVotes:  1,
				},
			},
		},
	}
	st := ctx.KVStore(k.StoreKey)
	st.Set([]byte("domain:SugDomain"), k.cdc.MustMarshalLengthPrefixed(&domain))

	handler := CustomQueryHandler(k)

	t.Run("found", func(t *testing.T) {
		reqBytes, _ := json.Marshal(WasmCustomQuery{
			Suggestion: &WasmQuerySuggestion{
				DomainName:     "SugDomain",
				IssueName:      "Energy",
				SuggestionName: "Nuclear",
			},
		})
		respBytes, err := handler(ctx, reqBytes)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var resp WasmSuggestionResponse
		if err := json.Unmarshal(respBytes, &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if resp.Name != "Nuclear" {
			t.Errorf("name = %q, want Nuclear", resp.Name)
		}
		if resp.Creator != "charlie" {
			t.Errorf("creator = %q, want charlie", resp.Creator)
		}
		if resp.RatingCount != 2 {
			t.Errorf("rating_count = %d, want 2", resp.RatingCount)
		}
		if resp.DwellTime != 86400 {
			t.Errorf("dwell_time = %d, want 86400", resp.DwellTime)
		}
		if resp.DeleteVotes != 1 {
			t.Errorf("delete_votes = %d, want 1", resp.DeleteVotes)
		}
	})

	t.Run("suggestion not found", func(t *testing.T) {
		reqBytes, _ := json.Marshal(WasmCustomQuery{
			Suggestion: &WasmQuerySuggestion{
				DomainName:     "SugDomain",
				IssueName:      "Energy",
				SuggestionName: "NoSuch",
			},
		})
		_, err := handler(ctx, reqBytes)
		if err == nil {
			t.Fatal("expected error for missing suggestion")
		}
	})

	t.Run("issue not found", func(t *testing.T) {
		reqBytes, _ := json.Marshal(WasmCustomQuery{
			Suggestion: &WasmQuerySuggestion{
				DomainName:     "SugDomain",
				IssueName:      "NoIssue",
				SuggestionName: "Nuclear",
			},
		})
		_, err := handler(ctx, reqBytes)
		if err == nil {
			t.Fatal("expected error for missing issue")
		}
	})
}

func TestWasmQueryPurgeSchedule(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "PurgeDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100)))

	handler := CustomQueryHandler(k)

	t.Run("found", func(t *testing.T) {
		reqBytes, _ := json.Marshal(WasmCustomQuery{
			PurgeSchedule: &WasmQueryPurgeSchedule{DomainName: "PurgeDomain"},
		})
		respBytes, err := handler(ctx, reqBytes)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var resp WasmPurgeScheduleResponse
		if err := json.Unmarshal(respBytes, &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if resp.DomainName != "PurgeDomain" {
			t.Errorf("domain_name = %q, want PurgeDomain", resp.DomainName)
		}
		if resp.PurgeInterval <= 0 {
			t.Errorf("purge_interval = %d, want > 0", resp.PurgeInterval)
		}
		if resp.AnnouncementLead <= 0 {
			t.Errorf("announcement_lead = %d, want > 0", resp.AnnouncementLead)
		}
	})

	t.Run("not found", func(t *testing.T) {
		reqBytes, _ := json.Marshal(WasmCustomQuery{
			PurgeSchedule: &WasmQueryPurgeSchedule{DomainName: "NoDomain"},
		})
		_, err := handler(ctx, reqBytes)
		if err == nil {
			t.Fatal("expected error for missing purge schedule")
		}
	})
}

func TestWasmQueryNullifier(t *testing.T) {
	k, ctx := setupKeeper(t)
	k.CreateDomain(ctx, "NullDomain", sdk.AccAddress("admin1"), sdk.NewCoins(sdk.NewInt64Coin("pnyx", 100)))

	// Mark one nullifier as used.
	k.SetNullifierUsed(ctx, "NullDomain", "aabbccdd", 100)

	handler := CustomQueryHandler(k)

	t.Run("used nullifier", func(t *testing.T) {
		reqBytes, _ := json.Marshal(WasmCustomQuery{
			Nullifier: &WasmQueryNullifier{DomainName: "NullDomain", NullifierHex: "aabbccdd"},
		})
		respBytes, err := handler(ctx, reqBytes)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var resp WasmNullifierResponse
		if err := json.Unmarshal(respBytes, &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if !resp.Used {
			t.Error("nullifier should be used")
		}
	})

	t.Run("unused nullifier", func(t *testing.T) {
		reqBytes, _ := json.Marshal(WasmCustomQuery{
			Nullifier: &WasmQueryNullifier{DomainName: "NullDomain", NullifierHex: "deadbeef"},
		})
		respBytes, err := handler(ctx, reqBytes)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var resp WasmNullifierResponse
		if err := json.Unmarshal(respBytes, &resp); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if resp.Used {
			t.Error("nullifier should be unused")
		}
	})
}

func TestWasmQueryInvalidJSON(t *testing.T) {
	k, ctx := setupKeeper(t)
	handler := CustomQueryHandler(k)

	_, err := handler(ctx, []byte(`{bad json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestWasmQueryUnknownType(t *testing.T) {
	k, ctx := setupKeeper(t)
	handler := CustomQueryHandler(k)

	// Empty query — no field set.
	reqBytes, _ := json.Marshal(WasmCustomQuery{})
	_, err := handler(ctx, reqBytes)
	if err == nil {
		t.Fatal("expected error for unknown query type")
	}
}

// --- Custom Message Encoder Tests ---

func TestWasmMsgPlaceStoneOnIssue(t *testing.T) {
	encoder := CustomMessageEncoder()
	sender := sdk.AccAddress("contract1")

	msgBytes, _ := json.Marshal(WasmCustomMsg{
		PlaceStoneOnIssue: &WasmMsgPlaceStoneOnIssue{
			DomainName: "TestDomain",
			IssueName:  "Climate",
		},
	})

	msgs, err := encoder(sender, msgBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("msgs len = %d, want 1", len(msgs))
	}
	m, ok := msgs[0].(*MsgPlaceStoneOnIssue)
	if !ok {
		t.Fatalf("wrong msg type: %T", msgs[0])
	}
	if m.DomainName != "TestDomain" {
		t.Errorf("domain = %q, want TestDomain", m.DomainName)
	}
	if m.IssueName != "Climate" {
		t.Errorf("issue = %q, want Climate", m.IssueName)
	}
	if m.MemberAddr != sender.String() {
		t.Errorf("member_addr = %q, want %q", m.MemberAddr, sender.String())
	}
}

func TestWasmMsgPlaceStoneOnSuggestion(t *testing.T) {
	encoder := CustomMessageEncoder()
	sender := sdk.AccAddress("contract2")

	msgBytes, _ := json.Marshal(WasmCustomMsg{
		PlaceStoneOnSuggestion: &WasmMsgPlaceStoneOnSuggestion{
			DomainName:     "TestDomain",
			IssueName:      "Climate",
			SuggestionName: "SolarPlan",
		},
	})

	msgs, err := encoder(sender, msgBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("msgs len = %d, want 1", len(msgs))
	}
	m, ok := msgs[0].(*MsgPlaceStoneOnSuggestion)
	if !ok {
		t.Fatalf("wrong msg type: %T", msgs[0])
	}
	if m.SuggestionName != "SolarPlan" {
		t.Errorf("suggestion = %q, want SolarPlan", m.SuggestionName)
	}
	if m.MemberAddr != sender.String() {
		t.Errorf("member_addr = %q, want %q", m.MemberAddr, sender.String())
	}
}

func TestWasmMsgCastElectionVote(t *testing.T) {
	encoder := CustomMessageEncoder()
	sender := sdk.AccAddress("contract3")

	msgBytes, _ := json.Marshal(WasmCustomMsg{
		CastElectionVote: &WasmMsgCastElectionVote{
			DomainName:    "TestDomain",
			IssueName:     "Election",
			CandidateName: "alice",
			Choice:        0, // Approve
		},
	})

	msgs, err := encoder(sender, msgBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("msgs len = %d, want 1", len(msgs))
	}
	m, ok := msgs[0].(*MsgCastElectionVote)
	if !ok {
		t.Fatalf("wrong msg type: %T", msgs[0])
	}
	if m.CandidateName != "alice" {
		t.Errorf("candidate = %q, want alice", m.CandidateName)
	}
	if m.VoterAddr != sender.String() {
		t.Errorf("voter_addr = %q, want %q", m.VoterAddr, sender.String())
	}
	if m.Choice != 0 {
		t.Errorf("choice = %d, want 0", m.Choice)
	}
}

func TestWasmMsgEncoderInvalidJSON(t *testing.T) {
	encoder := CustomMessageEncoder()
	sender := sdk.AccAddress("contract1")

	_, err := encoder(sender, []byte(`{bad`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestWasmMsgEncoderUnknownType(t *testing.T) {
	encoder := CustomMessageEncoder()
	sender := sdk.AccAddress("contract1")

	msgBytes, _ := json.Marshal(WasmCustomMsg{})
	_, err := encoder(sender, msgBytes)
	if err == nil {
		t.Fatal("expected error for unknown message type")
	}
}
