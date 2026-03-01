use cosmwasm_std::{Addr, Uint128};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Config {
    pub admin: Addr,
    pub max_slippage_bps: u64,
    pub fee_bps: u64,
    pub burn_bps: u64,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct LimitOrder {
    pub id: u64,
    pub owner: Addr,
    pub input_denom: String,
    pub input_amount: Uint128,
    pub output_denom: String,
    pub min_output: Uint128,
    pub created_at: u64,
    pub expires_at: u64,
    pub status: OrderStatus,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum OrderStatus {
    Open,
    Filled,
    Cancelled,
    Expired,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct PoolSnapshot {
    pub asset_denom: String,
    pub pnyx_reserve: Uint128,
    pub asset_reserve: Uint128,
    pub last_updated: u64,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct BotState {
    pub config: Config,
    pub orders: Vec<LimitOrder>,
    pub next_order_id: u64,
    pub snapshots: Vec<PoolSnapshot>,
}
