use cosmwasm_std::{Addr, Uint128};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Config {
    pub admin: Addr,
    pub domain_name: String,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct VestingSchedule {
    pub id: u64,
    pub beneficiary: Addr,
    pub total_amount: Uint128,
    pub released_amount: Uint128,
    pub start_time: u64,
    pub cliff_time: u64,
    pub end_time: u64,
    pub cancelled: bool,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct VestingState {
    pub config: Config,
    pub schedules: Vec<VestingSchedule>,
    pub next_schedule_id: u64,
}

pub fn compute_vested_amount(schedule: &VestingSchedule, current_time: u64) -> Uint128 {
    if schedule.cancelled || current_time < schedule.cliff_time {
        return Uint128::zero();
    }
    if current_time >= schedule.end_time {
        return schedule.total_amount;
    }
    let elapsed = current_time - schedule.start_time;
    let total_duration = schedule.end_time - schedule.start_time;
    schedule
        .total_amount
        .multiply_ratio(elapsed as u128, total_duration as u128)
}
