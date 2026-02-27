package truedemocracy

// CosmWasm custom query and message bindings for the truedemocracy module.
// These allow smart contracts to read domain/governance state and submit
// governance actions (stone voting, election votes) via custom wasm messages.

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// --- Custom Query Types ---

// WasmCustomQuery is the top-level query envelope sent by contracts.
type WasmCustomQuery struct {
	Domain          *WasmQueryDomain          `json:"domain,omitempty"`
	DomainMembers   *WasmQueryDomainMembers   `json:"domain_members,omitempty"`
	Issue           *WasmQueryIssue           `json:"issue,omitempty"`
	Suggestion      *WasmQuerySuggestion      `json:"suggestion,omitempty"`
	PurgeSchedule   *WasmQueryPurgeSchedule   `json:"purge_schedule,omitempty"`
	Nullifier       *WasmQueryNullifier       `json:"nullifier,omitempty"`
	DomainTreasury  *WasmQueryDomainTreasury  `json:"domain_treasury,omitempty"`
}

type WasmQueryDomain struct {
	Name string `json:"name"`
}

type WasmQueryDomainMembers struct {
	DomainName string `json:"domain_name"`
}

type WasmQueryIssue struct {
	DomainName string `json:"domain_name"`
	IssueName  string `json:"issue_name"`
}

type WasmQuerySuggestion struct {
	DomainName     string `json:"domain_name"`
	IssueName      string `json:"issue_name"`
	SuggestionName string `json:"suggestion_name"`
}

type WasmQueryPurgeSchedule struct {
	DomainName string `json:"domain_name"`
}

type WasmQueryNullifier struct {
	DomainName   string `json:"domain_name"`
	NullifierHex string `json:"nullifier_hex"`
}

type WasmQueryDomainTreasury struct {
	DomainName string `json:"domain_name"`
}

// --- Custom Query Response Types ---

type WasmDomainResponse struct {
	Name         string   `json:"name"`
	Admin        string   `json:"admin"`
	MemberCount  int      `json:"member_count"`
	Treasury     string   `json:"treasury"`
	IssueCount   int      `json:"issue_count"`
	MerkleRoot   string   `json:"merkle_root,omitempty"`
	TotalPayouts int64    `json:"total_payouts"`
	Options      WasmDomainOptionsResponse `json:"options"`
}

type WasmDomainOptionsResponse struct {
	AdminElectable   bool   `json:"admin_electable"`
	AnyoneCanJoin    bool   `json:"anyone_can_join"`
	OnlyAdminIssues  bool   `json:"only_admin_issues"`
	CoinBurnRequired bool   `json:"coin_burn_required"`
	VotingMode       int    `json:"voting_mode"`
}

type WasmDomainMembersResponse struct {
	DomainName string   `json:"domain_name"`
	Members    []string `json:"members"`
}

type WasmIssueResponse struct {
	Name           string                     `json:"name"`
	Stones         int                        `json:"stones"`
	SuggestionCount int                       `json:"suggestion_count"`
	Suggestions    []WasmSuggestionBrief      `json:"suggestions"`
	CreationDate   int64                      `json:"creation_date"`
	ExternalLink   string                     `json:"external_link,omitempty"`
}

type WasmSuggestionBrief struct {
	Name    string `json:"name"`
	Creator string `json:"creator"`
	Stones  int    `json:"stones"`
	Color   string `json:"color"`
	Score   int    `json:"score"`
}

type WasmSuggestionResponse struct {
	Name         string `json:"name"`
	Creator      string `json:"creator"`
	Stones       int    `json:"stones"`
	Color        string `json:"color"`
	RatingCount  int    `json:"rating_count"`
	Score        int    `json:"score"`
	DwellTime    int64  `json:"dwell_time"`
	CreationDate int64  `json:"creation_date"`
	ExternalLink string `json:"external_link,omitempty"`
	DeleteVotes  int    `json:"delete_votes"`
}

type WasmPurgeScheduleResponse struct {
	DomainName       string `json:"domain_name"`
	NextPurgeTime    int64  `json:"next_purge_time"`
	PurgeInterval    int64  `json:"purge_interval"`
	AnnouncementLead int64  `json:"announcement_lead"`
}

type WasmNullifierResponse struct {
	Used bool `json:"used"`
}

type WasmDomainTreasuryResponse struct {
	DomainName string `json:"domain_name"`
	Amount     string `json:"amount"` // e.g. "500000pnyx"
}

// --- Custom Query Handler ---

// CustomQueryHandler returns a query handler function for CosmWasm contracts
// to read truedemocracy state. The returned function matches the signature
// expected by wasmd's QueryPlugins.Custom field.
func CustomQueryHandler(keeper Keeper) func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	return func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
		var query WasmCustomQuery
		if err := json.Unmarshal(request, &query); err != nil {
			return nil, fmt.Errorf("invalid truedemocracy query: %w", err)
		}

		switch {
		case query.Domain != nil:
			return handleQueryDomain(ctx, keeper, query.Domain)
		case query.DomainMembers != nil:
			return handleQueryDomainMembers(ctx, keeper, query.DomainMembers)
		case query.Issue != nil:
			return handleQueryIssue(ctx, keeper, query.Issue)
		case query.Suggestion != nil:
			return handleQuerySuggestion(ctx, keeper, query.Suggestion)
		case query.PurgeSchedule != nil:
			return handleQueryPurgeSchedule(ctx, keeper, query.PurgeSchedule)
		case query.Nullifier != nil:
			return handleQueryNullifier(ctx, keeper, query.Nullifier)
		case query.DomainTreasury != nil:
			return handleQueryDomainTreasury(ctx, keeper, query.DomainTreasury)
		default:
			return nil, fmt.Errorf("unknown truedemocracy query")
		}
	}
}

func handleQueryDomain(ctx sdk.Context, keeper Keeper, req *WasmQueryDomain) ([]byte, error) {
	domain, found := keeper.GetDomain(ctx, req.Name)
	if !found {
		return nil, fmt.Errorf("domain not found: %s", req.Name)
	}

	resp := WasmDomainResponse{
		Name:         domain.Name,
		Admin:        domain.Admin.String(),
		MemberCount:  len(domain.Members),
		Treasury:     domain.Treasury.String(),
		IssueCount:   len(domain.Issues),
		MerkleRoot:   domain.MerkleRoot,
		TotalPayouts: domain.TotalPayouts,
		Options: WasmDomainOptionsResponse{
			AdminElectable:   domain.Options.AdminElectable,
			AnyoneCanJoin:    domain.Options.AnyoneCanJoin,
			OnlyAdminIssues:  domain.Options.OnlyAdminIssues,
			CoinBurnRequired: domain.Options.CoinBurnRequired,
			VotingMode:       int(domain.Options.VotingMode),
		},
	}
	return json.Marshal(resp)
}

func handleQueryDomainMembers(ctx sdk.Context, keeper Keeper, req *WasmQueryDomainMembers) ([]byte, error) {
	domain, found := keeper.GetDomain(ctx, req.DomainName)
	if !found {
		return nil, fmt.Errorf("domain not found: %s", req.DomainName)
	}

	resp := WasmDomainMembersResponse{
		DomainName: domain.Name,
		Members:    domain.Members,
	}
	return json.Marshal(resp)
}

func handleQueryIssue(ctx sdk.Context, keeper Keeper, req *WasmQueryIssue) ([]byte, error) {
	domain, found := keeper.GetDomain(ctx, req.DomainName)
	if !found {
		return nil, fmt.Errorf("domain not found: %s", req.DomainName)
	}

	for _, issue := range domain.Issues {
		if issue.Name == req.IssueName {
			var suggestions []WasmSuggestionBrief
			for _, s := range issue.Suggestions {
				score := ComputeSuggestionScore(s)
				suggestions = append(suggestions, WasmSuggestionBrief{
					Name:    s.Name,
					Creator: s.Creator,
					Stones:  s.Stones,
					Color:   s.Color,
					Score:   score,
				})
			}
			if suggestions == nil {
				suggestions = []WasmSuggestionBrief{}
			}
			resp := WasmIssueResponse{
				Name:            issue.Name,
				Stones:          issue.Stones,
				SuggestionCount: len(issue.Suggestions),
				Suggestions:     suggestions,
				CreationDate:    issue.CreationDate,
				ExternalLink:    issue.ExternalLink,
			}
			return json.Marshal(resp)
		}
	}

	return nil, fmt.Errorf("issue not found: %s in domain %s", req.IssueName, req.DomainName)
}

func handleQuerySuggestion(ctx sdk.Context, keeper Keeper, req *WasmQuerySuggestion) ([]byte, error) {
	domain, found := keeper.GetDomain(ctx, req.DomainName)
	if !found {
		return nil, fmt.Errorf("domain not found: %s", req.DomainName)
	}

	for _, issue := range domain.Issues {
		if issue.Name == req.IssueName {
			for _, s := range issue.Suggestions {
				if s.Name == req.SuggestionName {
					score := ComputeSuggestionScore(s)
					resp := WasmSuggestionResponse{
						Name:         s.Name,
						Creator:      s.Creator,
						Stones:       s.Stones,
						Color:        s.Color,
						RatingCount:  len(s.Ratings),
						Score:        score,
						DwellTime:    s.DwellTime,
						CreationDate: s.CreationDate,
						ExternalLink: s.ExternalLink,
						DeleteVotes:  s.DeleteVotes,
					}
					return json.Marshal(resp)
				}
			}
			return nil, fmt.Errorf("suggestion not found: %s in issue %s", req.SuggestionName, req.IssueName)
		}
	}

	return nil, fmt.Errorf("issue not found: %s in domain %s", req.IssueName, req.DomainName)
}

func handleQueryPurgeSchedule(ctx sdk.Context, keeper Keeper, req *WasmQueryPurgeSchedule) ([]byte, error) {
	schedule, found := keeper.GetBigPurgeSchedule(ctx, req.DomainName)
	if !found {
		return nil, fmt.Errorf("purge schedule not found for domain: %s", req.DomainName)
	}

	resp := WasmPurgeScheduleResponse{
		DomainName:       schedule.DomainName,
		NextPurgeTime:    schedule.NextPurgeTime,
		PurgeInterval:    schedule.PurgeInterval,
		AnnouncementLead: schedule.AnnouncementLead,
	}
	return json.Marshal(resp)
}

func handleQueryNullifier(ctx sdk.Context, keeper Keeper, req *WasmQueryNullifier) ([]byte, error) {
	used := keeper.IsNullifierUsed(ctx, req.DomainName, req.NullifierHex)
	resp := WasmNullifierResponse{Used: used}
	return json.Marshal(resp)
}

func handleQueryDomainTreasury(ctx sdk.Context, keeper Keeper, req *WasmQueryDomainTreasury) ([]byte, error) {
	domain, found := keeper.GetDomain(ctx, req.DomainName)
	if !found {
		return nil, fmt.Errorf("domain not found: %s", req.DomainName)
	}

	resp := WasmDomainTreasuryResponse{
		DomainName: domain.Name,
		Amount:     domain.Treasury.String(),
	}
	return json.Marshal(resp)
}

// --- Custom Message Types ---

// WasmCustomMsg is the top-level message envelope sent by contracts.
type WasmCustomMsg struct {
	PlaceStoneOnIssue      *WasmMsgPlaceStoneOnIssue      `json:"place_stone_on_issue,omitempty"`
	PlaceStoneOnSuggestion *WasmMsgPlaceStoneOnSuggestion `json:"place_stone_on_suggestion,omitempty"`
	CastElectionVote       *WasmMsgCastElectionVote       `json:"cast_election_vote,omitempty"`
	DepositToDomain        *WasmMsgDepositToDomain        `json:"deposit_to_domain,omitempty"`
	WithdrawFromDomain     *WasmMsgWithdrawFromDomain     `json:"withdraw_from_domain,omitempty"`
}

type WasmMsgPlaceStoneOnIssue struct {
	DomainName string `json:"domain_name"`
	IssueName  string `json:"issue_name"`
}

type WasmMsgPlaceStoneOnSuggestion struct {
	DomainName     string `json:"domain_name"`
	IssueName      string `json:"issue_name"`
	SuggestionName string `json:"suggestion_name"`
}

type WasmMsgCastElectionVote struct {
	DomainName    string `json:"domain_name"`
	IssueName     string `json:"issue_name"`
	CandidateName string `json:"candidate_name"`
	Choice        int    `json:"choice"` // 0=Approve, 1=Abstain
}

type WasmMsgDepositToDomain struct {
	DomainName string `json:"domain_name"`
	Amount     string `json:"amount"` // e.g. "100pnyx"
}

type WasmMsgWithdrawFromDomain struct {
	DomainName string `json:"domain_name"`
	Recipient  string `json:"recipient"` // bech32
	Amount     string `json:"amount"`    // e.g. "100pnyx"
}

// --- Custom Message Encoder ---

// CustomMessageEncoder returns a message encoder function for CosmWasm contracts
// to submit governance actions. The returned function matches the signature
// expected by wasmd's MessageEncoders.Custom field.
func CustomMessageEncoder() func(sender sdk.AccAddress, msg json.RawMessage) ([]sdk.Msg, error) {
	return func(sender sdk.AccAddress, msg json.RawMessage) ([]sdk.Msg, error) {
		var customMsg WasmCustomMsg
		if err := json.Unmarshal(msg, &customMsg); err != nil {
			return nil, fmt.Errorf("invalid truedemocracy message: %w", err)
		}

		switch {
		case customMsg.PlaceStoneOnIssue != nil:
			m := customMsg.PlaceStoneOnIssue
			return []sdk.Msg{&MsgPlaceStoneOnIssue{
				DomainName: m.DomainName,
				IssueName:  m.IssueName,
				MemberAddr: sender.String(),
			}}, nil

		case customMsg.PlaceStoneOnSuggestion != nil:
			m := customMsg.PlaceStoneOnSuggestion
			return []sdk.Msg{&MsgPlaceStoneOnSuggestion{
				DomainName:     m.DomainName,
				IssueName:      m.IssueName,
				SuggestionName: m.SuggestionName,
				MemberAddr:     sender.String(),
			}}, nil

		case customMsg.CastElectionVote != nil:
			m := customMsg.CastElectionVote
			return []sdk.Msg{&MsgCastElectionVote{
				DomainName:    m.DomainName,
				IssueName:     m.IssueName,
				CandidateName: m.CandidateName,
				VoterAddr:     sender.String(),
				Choice:        int32(m.Choice),
			}}, nil

		case customMsg.DepositToDomain != nil:
			m := customMsg.DepositToDomain
			coin, err := sdk.ParseCoinNormalized(m.Amount)
			if err != nil {
				return nil, fmt.Errorf("invalid deposit amount: %w", err)
			}
			return []sdk.Msg{&MsgDepositToDomain{
				Sender:     sender,
				DomainName: m.DomainName,
				Amount:     coin,
			}}, nil

		case customMsg.WithdrawFromDomain != nil:
			m := customMsg.WithdrawFromDomain
			coin, err := sdk.ParseCoinNormalized(m.Amount)
			if err != nil {
				return nil, fmt.Errorf("invalid withdraw amount: %w", err)
			}
			return []sdk.Msg{&MsgWithdrawFromDomain{
				Sender:     sender,
				DomainName: m.DomainName,
				Recipient:  m.Recipient,
				Amount:     coin,
			}}, nil

		default:
			return nil, fmt.Errorf("unknown truedemocracy message")
		}
	}
}
