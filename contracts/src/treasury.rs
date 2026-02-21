use cosmwasm_std::{
    entry_point, to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult,
    Uint128,
};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

const STATE_KEY: &[u8] = b"state";

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct State {
    pub balance: Uint128,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct InstantiateMsg {}

fn save_state(storage: &mut dyn cosmwasm_std::Storage, state: &State) -> StdResult<()> {
    let data = cosmwasm_std::to_json_vec(state)?;
    storage.set(STATE_KEY, &data);
    Ok(())
}

fn load_state(storage: &dyn cosmwasm_std::Storage) -> StdResult<State> {
    let data = storage
        .get(STATE_KEY)
        .ok_or_else(|| cosmwasm_std::StdError::msg("state not found"))?;
    cosmwasm_std::from_json(data)
}

#[entry_point]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    _msg: InstantiateMsg,
) -> StdResult<Response> {
    let state = State {
        balance: Uint128::zero(),
    };
    save_state(deps.storage, &state)?;
    Ok(Response::default())
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum ExecuteMsg {
    Deposit { amount: Uint128 },
    Withdraw { amount: Uint128, recipient: String },
}

#[entry_point]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: ExecuteMsg,
) -> StdResult<Response> {
    let mut state = load_state(deps.storage)?;
    match msg {
        ExecuteMsg::Deposit { amount } => {
            state.balance = state.balance.checked_add(amount)?;
            save_state(deps.storage, &state)?;
            Ok(Response::new().add_attribute("action", "deposit"))
        }
        ExecuteMsg::Withdraw {
            amount,
            recipient: _,
        } => {
            if state.balance < amount {
                return Err(cosmwasm_std::StdError::msg("Insufficient funds"));
            }
            state.balance = state.balance.checked_sub(amount)?;
            save_state(deps.storage, &state)?;
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
    let state = load_state(deps.storage)?;
    match msg {
        QueryMsg::GetBalance {} => to_json_binary(&state.balance),
    }
}
