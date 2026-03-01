use cosmwasm_std::{
    entry_point, to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult,
    Uint128,
};
use serde::{Deserialize, Serialize};
use truerepublic_bindings::{TrueRepublicMsg, TrueRepublicQuery};

use crate::amm;
use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{BotState, Config, LimitOrder, OrderStatus, PoolSnapshot};

const STATE_KEY: &[u8] = b"bot_state";

fn save_state(storage: &mut dyn cosmwasm_std::Storage, state: &BotState) -> StdResult<()> {
    storage.set(STATE_KEY, &cosmwasm_std::to_json_vec(state)?);
    Ok(())
}

fn load_state(storage: &dyn cosmwasm_std::Storage) -> StdResult<BotState> {
    let data = storage
        .get(STATE_KEY)
        .ok_or_else(|| cosmwasm_std::StdError::msg("not found: state"))?;
    cosmwasm_std::from_json(data)
}

#[derive(Serialize, Deserialize)]
struct EstimateResponse {
    output: Uint128,
    burn: Uint128,
}

#[entry_point]
pub fn instantiate(
    deps: DepsMut<TrueRepublicQuery>,
    _env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response<TrueRepublicMsg>, ContractError> {
    let state = BotState {
        config: Config {
            admin: info.sender,
            max_slippage_bps: msg.max_slippage_bps,
            fee_bps: amm::SWAP_FEE_BPS,
            burn_bps: amm::BURN_BPS,
        },
        orders: vec![],
        next_order_id: 1,
        snapshots: vec![],
    };
    save_state(deps.storage, &state)?;
    Ok(Response::new().add_attribute("method", "instantiate"))
}

#[entry_point]
pub fn execute(
    deps: DepsMut<TrueRepublicQuery>,
    env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> Result<Response<TrueRepublicMsg>, ContractError> {
    match msg {
        ExecuteMsg::PlaceLimitOrder {
            input_denom,
            input_amount,
            output_denom,
            min_output,
            expires_in_secs,
        } => {
            let mut state = load_state(deps.storage)?;
            let order = LimitOrder {
                id: state.next_order_id,
                owner: info.sender,
                input_denom,
                input_amount,
                output_denom,
                min_output,
                created_at: env.block.time.seconds(),
                expires_at: env.block.time.seconds() + expires_in_secs,
                status: OrderStatus::Open,
            };
            state.orders.push(order);
            state.next_order_id += 1;
            save_state(deps.storage, &state)?;
            Ok(Response::new().add_attribute("action", "place_limit_order"))
        }
        ExecuteMsg::CancelOrder { order_id } => {
            let mut state = load_state(deps.storage)?;
            let order = state
                .orders
                .iter_mut()
                .find(|o| o.id == order_id)
                .ok_or(ContractError::OrderNotFound {})?;
            if order.owner != info.sender {
                return Err(ContractError::NotOrderOwner {});
            }
            if order.status != OrderStatus::Open {
                return Err(ContractError::OrderNotOpen {});
            }
            order.status = OrderStatus::Cancelled;
            save_state(deps.storage, &state)?;
            Ok(Response::new().add_attribute("action", "cancel_order"))
        }
        ExecuteMsg::UpdatePoolSnapshot {
            asset_denom,
            pnyx_reserve,
            asset_reserve,
        } => {
            let mut state = load_state(deps.storage)?;
            if info.sender != state.config.admin {
                return Err(ContractError::Unauthorized {});
            }
            let snapshot = PoolSnapshot {
                asset_denom: asset_denom.clone(),
                pnyx_reserve,
                asset_reserve,
                last_updated: env.block.time.seconds(),
            };
            if let Some(existing) = state
                .snapshots
                .iter_mut()
                .find(|s| s.asset_denom == asset_denom)
            {
                *existing = snapshot;
            } else {
                state.snapshots.push(snapshot);
            }
            save_state(deps.storage, &state)?;
            Ok(Response::new().add_attribute("action", "update_snapshot"))
        }
        ExecuteMsg::CheckAndExecute { order_ids } => {
            let mut state = load_state(deps.storage)?;
            let mut filled = 0u64;
            for order in state.orders.iter_mut() {
                if !order_ids.contains(&order.id) || order.status != OrderStatus::Open {
                    continue;
                }
                if env.block.time.seconds() > order.expires_at {
                    order.status = OrderStatus::Expired;
                    continue;
                }
                // Find pool snapshot for the output denom
                let asset_denom = if order.input_denom == "pnyx" {
                    &order.output_denom
                } else {
                    &order.input_denom
                };
                let snapshot = state
                    .snapshots
                    .iter()
                    .find(|s| s.asset_denom == *asset_denom);
                if let Some(snap) = snapshot {
                    let output_is_pnyx = order.output_denom == "pnyx";
                    let (in_reserve, out_reserve) = if order.input_denom == "pnyx" {
                        (snap.pnyx_reserve, snap.asset_reserve)
                    } else {
                        (snap.asset_reserve, snap.pnyx_reserve)
                    };
                    let (output, _) = amm::compute_swap_output(
                        in_reserve,
                        out_reserve,
                        order.input_amount,
                        output_is_pnyx,
                    );
                    if output >= order.min_output {
                        order.status = OrderStatus::Filled;
                        filled += 1;
                    }
                }
            }
            save_state(deps.storage, &state)?;
            Ok(Response::new()
                .add_attribute("action", "check_and_execute")
                .add_attribute("filled", filled.to_string()))
        }
    }
}

#[entry_point]
pub fn query(deps: Deps<TrueRepublicQuery>, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    match msg {
        QueryMsg::Config {} => {
            let state = load_state(deps.storage)?;
            to_json_binary(&state.config)
        }
        QueryMsg::Order { id } => {
            let state = load_state(deps.storage)?;
            let order = state
                .orders
                .iter()
                .find(|o| o.id == id)
                .ok_or_else(|| cosmwasm_std::StdError::msg("not found: order"))?;
            to_json_binary(order)
        }
        QueryMsg::Orders { status } => {
            let state = load_state(deps.storage)?;
            let filtered: Vec<&LimitOrder> = if let Some(s) = status {
                state
                    .orders
                    .iter()
                    .filter(|o| format!("{:?}", o.status).to_lowercase() == s)
                    .collect()
            } else {
                state.orders.iter().collect()
            };
            to_json_binary(&filtered)
        }
        QueryMsg::PoolSnapshot { asset_denom } => {
            let state = load_state(deps.storage)?;
            let snap = state
                .snapshots
                .iter()
                .find(|s| s.asset_denom == asset_denom)
                .ok_or_else(|| cosmwasm_std::StdError::msg("not found: snapshot"))?;
            to_json_binary(snap)
        }
        QueryMsg::EstimateOutput {
            input_denom,
            input_amount,
            output_denom,
        } => {
            let state = load_state(deps.storage)?;
            let asset_denom = if input_denom == "pnyx" {
                &output_denom
            } else {
                &input_denom
            };
            let snap = state
                .snapshots
                .iter()
                .find(|s| s.asset_denom == *asset_denom)
                .ok_or_else(|| cosmwasm_std::StdError::msg("not found: snapshot"))?;
            let output_is_pnyx = output_denom == "pnyx";
            let (in_reserve, out_reserve) = if input_denom == "pnyx" {
                (snap.pnyx_reserve, snap.asset_reserve)
            } else {
                (snap.asset_reserve, snap.pnyx_reserve)
            };
            let (output, burn) =
                amm::compute_swap_output(in_reserve, out_reserve, input_amount, output_is_pnyx);
            to_json_binary(&EstimateResponse { output, burn })
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::{message_info, mock_env};
    use cosmwasm_std::Addr;
    use truerepublic_testing_utils::mock_dependencies_with_truerepublic;

    #[test]
    fn test_instantiate() {
        let mut deps = mock_dependencies_with_truerepublic(&[]);
        let info = message_info(&Addr::unchecked("admin"), &[]);
        let msg = InstantiateMsg {
            max_slippage_bps: 100,
        };
        let res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();
        assert_eq!(res.attributes[0].value, "instantiate");
    }

    #[test]
    fn test_place_and_cancel_order() {
        let mut deps = mock_dependencies_with_truerepublic(&[]);
        let info = message_info(&Addr::unchecked("trader"), &[]);
        instantiate(
            deps.as_mut(),
            mock_env(),
            message_info(&Addr::unchecked("admin"), &[]),
            InstantiateMsg {
                max_slippage_bps: 100,
            },
        )
        .unwrap();

        let place_msg = ExecuteMsg::PlaceLimitOrder {
            input_denom: "pnyx".to_string(),
            input_amount: Uint128::new(1000),
            output_denom: "atom".to_string(),
            min_output: Uint128::new(900),
            expires_in_secs: 3600,
        };
        execute(deps.as_mut(), mock_env(), info.clone(), place_msg).unwrap();

        let cancel_msg = ExecuteMsg::CancelOrder { order_id: 1 };
        execute(deps.as_mut(), mock_env(), info, cancel_msg).unwrap();

        let res = query(deps.as_ref(), mock_env(), QueryMsg::Order { id: 1 }).unwrap();
        let order: LimitOrder = cosmwasm_std::from_json(res).unwrap();
        assert_eq!(order.status, OrderStatus::Cancelled);
    }

    #[test]
    fn test_fill_order_when_price_met() {
        let mut deps = mock_dependencies_with_truerepublic(&[]);
        let admin = message_info(&Addr::unchecked("admin"), &[]);
        instantiate(
            deps.as_mut(),
            mock_env(),
            admin.clone(),
            InstantiateMsg {
                max_slippage_bps: 100,
            },
        )
        .unwrap();

        // Place order
        let place_msg = ExecuteMsg::PlaceLimitOrder {
            input_denom: "pnyx".to_string(),
            input_amount: Uint128::new(1000),
            output_denom: "atom".to_string(),
            min_output: Uint128::new(900),
            expires_in_secs: 3600,
        };
        execute(
            deps.as_mut(),
            mock_env(),
            message_info(&Addr::unchecked("trader"), &[]),
            place_msg,
        )
        .unwrap();

        // Update pool snapshot (1M/1M = ~996 output for 1000 input)
        let snap_msg = ExecuteMsg::UpdatePoolSnapshot {
            asset_denom: "atom".to_string(),
            pnyx_reserve: Uint128::new(1_000_000),
            asset_reserve: Uint128::new(1_000_000),
        };
        execute(deps.as_mut(), mock_env(), admin.clone(), snap_msg).unwrap();

        // Check and execute
        let check_msg = ExecuteMsg::CheckAndExecute { order_ids: vec![1] };
        let res = execute(deps.as_mut(), mock_env(), admin, check_msg).unwrap();
        assert!(res
            .attributes
            .iter()
            .any(|a| a.key == "filled" && a.value == "1"));
    }
}
