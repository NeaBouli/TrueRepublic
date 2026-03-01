use cosmwasm_std::Uint128;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct InstantiateMsg {
    pub domain_name: String,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum ExecuteMsg {
    CreateSchedule {
        beneficiary: String,
        total_amount: Uint128,
        start_time: u64,
        cliff_duration_secs: u64,
        vesting_duration_secs: u64,
    },
    Claim {
        schedule_id: u64,
    },
    CancelSchedule {
        schedule_id: u64,
    },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {
    Config {},
    Schedule { id: u64 },
    SchedulesByBeneficiary { beneficiary: String },
    AllSchedules {},
    TreasuryBalance {},
}
