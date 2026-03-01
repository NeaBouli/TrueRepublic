use cosmwasm_std::CustomMsg;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

/// Custom messages that a CosmWasm contract can send to the
/// TrueRepublic blockchain via `CosmosMsg::Custom`.
///
/// Each variant maps to a handler in the Go `WasmMsgEncoder`.
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum TrueRepublicMsg {
    PlaceStoneOnIssue {
        domain_name: String,
        issue_name: String,
    },
    PlaceStoneOnSuggestion {
        domain_name: String,
        issue_name: String,
        suggestion_name: String,
    },
    CastElectionVote {
        domain_name: String,
        issue_name: String,
        candidate_name: String,
        choice: i32,
    },
    DepositToDomain {
        domain_name: String,
        amount: String,
    },
    WithdrawFromDomain {
        domain_name: String,
        recipient: String,
        amount: String,
    },
}

impl CustomMsg for TrueRepublicMsg {}
