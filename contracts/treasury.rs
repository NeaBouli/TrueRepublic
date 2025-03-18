use cosmwasm_std::{entry_point, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult, to_binary, Uint128};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct State {
    pub balance: Uint128,
}

#[entry_point]
pub fn instantiate(deps: DepsMut, _env: Env, _info: MessageInfo, _msg: ()) -> StdResult<Response> {
    deps.storage.save(b"state", &State { balance: Uint128::zero() })?;
    Ok(Response::default())
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub enum ExecuteMsg {
    Deposit { amount: Uint128 },
    Withdraw { amount: Uint128, recipient: String },
}

#[entry_point]
pub fn execute(deps: DepsMut, _env: Env, info: MessageInfo, msg: ExecuteMsg) -> StdResult<Response> {
    let mut state = deps.storage.load::<State>(b"state")?;
    match msg {
        ExecuteMsg::Deposit { amount } => {
            state.balance = state.balance.checked_add(amount)?;
            deps.storage.save(b"state", &state)?;
            Ok(Response::new().add_attribute("action", "deposit"))
        }
        ExecuteMsg::Withdraw { amount, recipient } => {
            if state.balance < amount {
                return Err(cosmwasm_std::StdError::generic_err("Insufficient funds"));
            }
            state.balance = state.balance.checked_sub(amount)?;
            deps.storage.save(b"state", &state)?;
            Ok(Response::new().add_attribute("action", "withdraw"))
        }
    }
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub enum QueryMsg {
    GetBalance {},
}

#[entry_point]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    let state = deps.storage.load::<State>(b"state")?;
    match msg {
        QueryMsg::GetBalance {} => to_binary(&state.balance),
    }
}
