use cosmwasm_std::Addr;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Config {
    pub admin: Addr,
    pub domain_name: String,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct VotingRound {
    pub id: u64,
    pub issue_name: String,
    pub suggestions: Vec<String>,
    pub scores: Vec<i64>,
    pub participant_count: u64,
    pub started_at: u64,
    pub ends_at: u64,
    pub finalized: bool,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct AggregatorState {
    pub config: Config,
    pub rounds: Vec<VotingRound>,
    pub next_round_id: u64,
    pub used_nullifiers: Vec<(u64, String)>, // (round_id, nullifier_hex)
}
