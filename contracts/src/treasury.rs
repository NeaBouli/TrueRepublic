use cosmwasm_std::{entry_point, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult, to_binary, Uint128};
use cosmwasm_storage::{singleton, singleton_read};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

const STATE_KEY: &[u8] = b"state";

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct State {
    pub balance: Uint128,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct InstantiateMsg {}

#[entry_point]
pub fn instantiate(deps: DepsMut, _env: Env, _info: MessageInfo, _msg: InstantiateMsg) -> StdResult<Response> {
    let state = State { balance: Uint128::zero() };
    singleton(deps.storage, STATE_KEY).save(&state)?;
    Ok(Response::default())
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum ExecuteMsg {
    Deposit { amount: Uint128 },
    Withdraw { amount: Uint128, recipient: String },
}

#[entry_point]
pub fn execute(deps: DepsMut, _env: Env, _info: MessageInfo, msg: ExecuteMsg) -> StdResult<Response> {
    let mut state: State = singleton(deps.storage, STATE_KEY).load()?;
    match msg {
        ExecuteMsg::Deposit { amount } => {
            state.balance = state.balance.checked_add(amount)?;
            singleton(deps.storage, STATE_KEY).save(&state)?;
            Ok(Response::new().add_attribute("action", "deposit"))
        }
        ExecuteMsg::Withdraw { amount, recipient: _ } => {
            if state.balance < amount {
                return Err(cosmwasm_std::StdError::generic_err("Insufficient funds"));
            }
            state.balance = state.balance.checked_sub(amount)?;
            singleton(deps.storage, STATE_KEY).save(&state)?;
            Ok(Response::new().add_attribute("action", "withdraw"))
        }
    }
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum QueryMsg {
    GetBalance {},
}

#[entry_point]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    let state: State = singleton_read(deps.storage, STATE_KEY).load()?;
    match msg {
        QueryMsg::GetBalance {} => to_binary(&state.balance),
    }
}
