use cosmwasm_std::CustomQuery;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

/// Custom queries supported by the TrueRepublic blockchain.
///
/// Each variant maps to a field in the Go `WasmCustomQuery` struct.
/// Serde's default externally-tagged enum representation produces
/// `{"variant_name": { ...fields... }}`, which matches the Go JSON
/// format where only one `omitempty` pointer field is non-null.
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum TrueRepublicQuery {
    Domain {
        name: String,
    },
    DomainMembers {
        domain_name: String,
    },
    Issue {
        domain_name: String,
        issue_name: String,
    },
    Suggestion {
        domain_name: String,
        issue_name: String,
        suggestion_name: String,
    },
    PurgeSchedule {
        domain_name: String,
    },
    Nullifier {
        domain_name: String,
        nullifier_hex: String,
    },
    DomainTreasury {
        domain_name: String,
    },
}

impl CustomQuery for TrueRepublicQuery {}

// ---------------------------------------------------------------------------
// Response types
// ---------------------------------------------------------------------------

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct DomainResponse {
    pub name: String,
    pub admin: String,
    pub member_count: i64,
    pub treasury: String,
    pub issue_count: i64,
    pub merkle_root: Option<String>,
    pub total_payouts: i64,
    pub options: DomainOptionsResponse,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct DomainOptionsResponse {
    pub admin_electable: bool,
    pub anyone_can_join: bool,
    pub only_admin_issues: bool,
    pub coin_burn_required: bool,
    pub voting_mode: i32,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct DomainMembersResponse {
    pub domain_name: String,
    pub members: Vec<String>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct IssueResponse {
    pub name: String,
    pub stones: i64,
    pub suggestion_count: i64,
    pub suggestions: Vec<SuggestionBrief>,
    pub creation_date: i64,
    pub external_link: Option<String>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct SuggestionBrief {
    pub name: String,
    pub creator: String,
    pub stones: i64,
    pub color: String,
    pub score: i64,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct SuggestionResponse {
    pub name: String,
    pub creator: String,
    pub stones: i64,
    pub color: String,
    pub rating_count: i64,
    pub score: i64,
    pub dwell_time: i64,
    pub creation_date: i64,
    pub external_link: Option<String>,
    pub delete_votes: i64,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct PurgeScheduleResponse {
    pub domain_name: String,
    pub next_purge_time: i64,
    pub purge_interval: i64,
    pub announcement_lead: i64,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct NullifierResponse {
    pub used: bool,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct DomainTreasuryResponse {
    pub domain_name: String,
    pub amount: String,
}
