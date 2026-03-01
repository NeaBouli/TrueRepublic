use cosmwasm_std::Uint128;
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct InstantiateMsg {
    pub max_slippage_bps: u64,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum ExecuteMsg {
    PlaceLimitOrder {
        input_denom: String,
        input_amount: Uint128,
        output_denom: String,
        min_output: Uint128,
        expires_in_secs: u64,
    },
    CancelOrder {
        order_id: u64,
    },
    UpdatePoolSnapshot {
        asset_denom: String,
        pnyx_reserve: Uint128,
        asset_reserve: Uint128,
    },
    CheckAndExecute {
        order_ids: Vec<u64>,
    },
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
#[serde(rename_all = "snake_case")]
pub enum QueryMsg {
    Config {},
    Order {
        id: u64,
    },
    Orders {
        status: Option<String>,
    },
    PoolSnapshot {
        asset_denom: String,
    },
    EstimateOutput {
        input_denom: String,
        input_amount: Uint128,
        output_denom: String,
    },
}
