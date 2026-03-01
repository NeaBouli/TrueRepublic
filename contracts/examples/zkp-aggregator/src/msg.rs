use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct InstantiateMsg {
    pub domain_name: String,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum ExecuteMsg {
    CreateRound {
        issue_name: String,
        suggestion_names: Vec<String>,
        duration_secs: u64,
    },
    SubmitVote {
        round_id: u64,
        nullifier_hex: String,
        ratings: Vec<i64>,
    },
    FinalizeRound {
        round_id: u64,
    },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {
    Config {},
    Round { id: u64 },
    ActiveRounds {},
    RoundResults { id: u64 },
}
