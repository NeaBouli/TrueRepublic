use cosmwasm_std::Addr;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Config {
    pub admin: Addr,
    pub domain_name: String,
    pub quorum_bps: u64,
    pub threshold_bps: u64,
    pub voting_period_secs: u64,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Proposal {
    pub id: u64,
    pub proposer: Addr,
    pub title: String,
    pub description: String,
    pub action: ProposalAction,
    pub votes_for: u64,
    pub votes_against: u64,
    pub voters: Vec<String>,
    pub open_until: u64,
    pub executed: bool,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum ProposalAction {
    PlaceStoneOnIssue {
        issue_name: String,
    },
    PlaceStoneOnSuggestion {
        issue_name: String,
        suggestion_name: String,
    },
    DepositToDomain {
        amount: String,
    },
    WithdrawFromDomain {
        recipient: String,
        amount: String,
    },
    CastElectionVote {
        issue_name: String,
        candidate_name: String,
        choice: i32,
    },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct DaoState {
    pub config: Config,
    pub proposals: Vec<Proposal>,
    pub next_id: u64,
}
