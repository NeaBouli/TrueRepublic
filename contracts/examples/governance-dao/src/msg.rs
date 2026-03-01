use crate::state::ProposalAction;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct InstantiateMsg {
    pub domain_name: String,
    pub quorum_bps: u64,
    pub threshold_bps: u64,
    pub voting_period_secs: u64,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum ExecuteMsg {
    CreateProposal {
        title: String,
        description: String,
        action: ProposalAction,
    },
    Vote {
        proposal_id: u64,
        approve: bool,
    },
    ExecuteProposal {
        proposal_id: u64,
    },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {
    Config {},
    Proposal { id: u64 },
    Proposals {},
    DomainInfo {},
}
