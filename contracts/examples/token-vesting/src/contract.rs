use cosmwasm_std::{
    entry_point, to_json_binary, Binary, CosmosMsg, Deps, DepsMut, Env, MessageInfo, QueryRequest,
    Response, StdResult, Uint128,
};
use truerepublic_bindings::{DomainTreasuryResponse, TrueRepublicMsg, TrueRepublicQuery};

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{compute_vested_amount, Config, VestingSchedule, VestingState};

const STATE_KEY: &[u8] = b"vesting_state";

fn save_state(storage: &mut dyn cosmwasm_std::Storage, state: &VestingState) -> StdResult<()> {
    storage.set(STATE_KEY, &cosmwasm_std::to_json_vec(state)?);
    Ok(())
}

fn load_state(storage: &dyn cosmwasm_std::Storage) -> StdResult<VestingState> {
    let data = storage
        .get(STATE_KEY)
        .ok_or_else(|| cosmwasm_std::StdError::msg("not found: state"))?;
    cosmwasm_std::from_json(data)
}

#[entry_point]
pub fn instantiate(
    deps: DepsMut<TrueRepublicQuery>,
    _env: Env,
    info: MessageInfo,
    msg: InstantiateMsg,
) -> Result<Response<TrueRepublicMsg>, ContractError> {
    let state = VestingState {
        config: Config {
            admin: info.sender,
            domain_name: msg.domain_name,
        },
        schedules: vec![],
        next_schedule_id: 1,
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
        ExecuteMsg::CreateSchedule {
            beneficiary,
            total_amount,
            start_time,
            cliff_duration_secs,
            vesting_duration_secs,
        } => {
            let mut state = load_state(deps.storage)?;
            if info.sender != state.config.admin {
                return Err(ContractError::Unauthorized {});
            }
            let beneficiary_addr = deps.api.addr_validate(&beneficiary)?;
            let schedule = VestingSchedule {
                id: state.next_schedule_id,
                beneficiary: beneficiary_addr,
                total_amount,
                released_amount: Uint128::zero(),
                start_time,
                cliff_time: start_time + cliff_duration_secs,
                end_time: start_time + vesting_duration_secs,
                cancelled: false,
            };
            state.schedules.push(schedule);
            state.next_schedule_id += 1;
            save_state(deps.storage, &state)?;
            Ok(Response::new().add_attribute("action", "create_schedule"))
        }
        ExecuteMsg::Claim { schedule_id } => {
            let mut state = load_state(deps.storage)?;
            let schedule = state
                .schedules
                .iter_mut()
                .find(|s| s.id == schedule_id)
                .ok_or(ContractError::ScheduleNotFound {})?;
            if schedule.beneficiary != info.sender {
                return Err(ContractError::NotBeneficiary {});
            }
            if schedule.cancelled {
                return Err(ContractError::AlreadyCancelled {});
            }
            let vested = compute_vested_amount(schedule, env.block.time.seconds());
            let claimable = vested
                .checked_sub(schedule.released_amount)
                .map_err(|_| ContractError::NothingToClaim {})?;
            if claimable.is_zero() {
                return Err(ContractError::NothingToClaim {});
            }
            schedule.released_amount += claimable;
            let withdraw_msg = CosmosMsg::Custom(TrueRepublicMsg::WithdrawFromDomain {
                domain_name: state.config.domain_name.clone(),
                recipient: schedule.beneficiary.to_string(),
                amount: format!("{}pnyx", claimable),
            });
            save_state(deps.storage, &state)?;
            Ok(Response::new()
                .add_message(withdraw_msg)
                .add_attribute("action", "claim")
                .add_attribute("amount", claimable.to_string()))
        }
        ExecuteMsg::CancelSchedule { schedule_id } => {
            let mut state = load_state(deps.storage)?;
            if info.sender != state.config.admin {
                return Err(ContractError::Unauthorized {});
            }
            let schedule = state
                .schedules
                .iter_mut()
                .find(|s| s.id == schedule_id)
                .ok_or(ContractError::ScheduleNotFound {})?;
            if schedule.cancelled {
                return Err(ContractError::AlreadyCancelled {});
            }
            schedule.cancelled = true;
            save_state(deps.storage, &state)?;
            Ok(Response::new().add_attribute("action", "cancel_schedule"))
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
        QueryMsg::Schedule { id } => {
            let state = load_state(deps.storage)?;
            let schedule = state
                .schedules
                .iter()
                .find(|s| s.id == id)
                .ok_or_else(|| cosmwasm_std::StdError::msg("not found: schedule"))?;
            to_json_binary(schedule)
        }
        QueryMsg::SchedulesByBeneficiary { beneficiary } => {
            let state = load_state(deps.storage)?;
            let filtered: Vec<&VestingSchedule> = state
                .schedules
                .iter()
                .filter(|s| s.beneficiary.as_str() == beneficiary)
                .collect();
            to_json_binary(&filtered)
        }
        QueryMsg::AllSchedules {} => {
            let state = load_state(deps.storage)?;
            to_json_binary(&state.schedules)
        }
        QueryMsg::TreasuryBalance {} => {
            let state = load_state(deps.storage)?;
            let treasury: DomainTreasuryResponse =
                deps.querier
                    .query(&QueryRequest::Custom(TrueRepublicQuery::DomainTreasury {
                        domain_name: state.config.domain_name.clone(),
                    }))?;
            to_json_binary(&treasury)
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::{message_info, mock_env, MockApi};
    use cosmwasm_std::{Addr, Timestamp};
    use truerepublic_testing_utils::mock_dependencies_with_truerepublic;

    fn beneficiary_addr() -> Addr {
        MockApi::default().addr_make("beneficiary1")
    }

    fn env_at_time(secs: u64) -> Env {
        let mut env = mock_env();
        env.block.time = Timestamp::from_seconds(secs);
        env
    }

    #[test]
    fn test_create_schedule() {
        let mut deps = mock_dependencies_with_truerepublic(&[]);
        let admin = message_info(&Addr::unchecked("admin"), &[]);
        instantiate(
            deps.as_mut(),
            mock_env(),
            admin.clone(),
            InstantiateMsg {
                domain_name: "TestDomain".to_string(),
            },
        )
        .unwrap();

        let ben = beneficiary_addr();
        let msg = ExecuteMsg::CreateSchedule {
            beneficiary: ben.to_string(),
            total_amount: Uint128::new(100_000),
            start_time: 1000,
            cliff_duration_secs: 3600,
            vesting_duration_secs: 36000,
        };
        execute(deps.as_mut(), mock_env(), admin, msg).unwrap();

        let res = query(deps.as_ref(), mock_env(), QueryMsg::Schedule { id: 1 }).unwrap();
        let schedule: VestingSchedule = cosmwasm_std::from_json(res).unwrap();
        assert_eq!(schedule.total_amount, Uint128::new(100_000));
        assert_eq!(schedule.cliff_time, 4600); // 1000 + 3600
        assert_eq!(schedule.end_time, 37000); // 1000 + 36000
    }

    #[test]
    fn test_claim_before_cliff_fails() {
        let mut deps = mock_dependencies_with_truerepublic(&[]);
        let admin = message_info(&Addr::unchecked("admin"), &[]);
        instantiate(
            deps.as_mut(),
            mock_env(),
            admin.clone(),
            InstantiateMsg {
                domain_name: "TestDomain".to_string(),
            },
        )
        .unwrap();

        let ben = beneficiary_addr();
        let msg = ExecuteMsg::CreateSchedule {
            beneficiary: ben.to_string(),
            total_amount: Uint128::new(100_000),
            start_time: 1000,
            cliff_duration_secs: 3600,
            vesting_duration_secs: 36000,
        };
        execute(deps.as_mut(), mock_env(), admin, msg).unwrap();

        // Try to claim before cliff (at time 2000, cliff is 4600)
        let claim_msg = ExecuteMsg::Claim { schedule_id: 1 };
        let err = execute(
            deps.as_mut(),
            env_at_time(2000),
            message_info(&ben, &[]),
            claim_msg,
        )
        .unwrap_err();
        assert!(format!("{}", err).contains("Nothing to claim"));
    }

    #[test]
    fn test_claim_partial_vesting() {
        let mut deps = mock_dependencies_with_truerepublic(&[]);
        let admin = message_info(&Addr::unchecked("admin"), &[]);
        instantiate(
            deps.as_mut(),
            mock_env(),
            admin.clone(),
            InstantiateMsg {
                domain_name: "TestDomain".to_string(),
            },
        )
        .unwrap();

        let ben = beneficiary_addr();
        let msg = ExecuteMsg::CreateSchedule {
            beneficiary: ben.to_string(),
            total_amount: Uint128::new(100_000),
            start_time: 0,
            cliff_duration_secs: 0, // no cliff
            vesting_duration_secs: 10_000,
        };
        execute(deps.as_mut(), mock_env(), admin, msg).unwrap();

        // Claim at 50% through vesting period
        let claim_msg = ExecuteMsg::Claim { schedule_id: 1 };
        let res = execute(
            deps.as_mut(),
            env_at_time(5000),
            message_info(&ben, &[]),
            claim_msg,
        )
        .unwrap();
        assert_eq!(res.messages.len(), 1); // WithdrawFromDomain emitted
        assert!(res
            .attributes
            .iter()
            .any(|a| a.key == "amount" && a.value == "50000"));
    }

    #[test]
    fn test_claim_after_full_vesting() {
        let mut deps = mock_dependencies_with_truerepublic(&[]);
        let admin = message_info(&Addr::unchecked("admin"), &[]);
        instantiate(
            deps.as_mut(),
            mock_env(),
            admin.clone(),
            InstantiateMsg {
                domain_name: "TestDomain".to_string(),
            },
        )
        .unwrap();

        let ben = beneficiary_addr();
        let msg = ExecuteMsg::CreateSchedule {
            beneficiary: ben.to_string(),
            total_amount: Uint128::new(100_000),
            start_time: 0,
            cliff_duration_secs: 0,
            vesting_duration_secs: 10_000,
        };
        execute(deps.as_mut(), mock_env(), admin, msg).unwrap();

        let claim_msg = ExecuteMsg::Claim { schedule_id: 1 };
        let res = execute(
            deps.as_mut(),
            env_at_time(20_000),
            message_info(&ben, &[]),
            claim_msg,
        )
        .unwrap();
        assert!(res
            .attributes
            .iter()
            .any(|a| a.key == "amount" && a.value == "100000"));
    }

    #[test]
    fn test_cancel_prevents_further_claims() {
        let mut deps = mock_dependencies_with_truerepublic(&[]);
        let admin = message_info(&Addr::unchecked("admin"), &[]);
        instantiate(
            deps.as_mut(),
            mock_env(),
            admin.clone(),
            InstantiateMsg {
                domain_name: "TestDomain".to_string(),
            },
        )
        .unwrap();

        let ben = beneficiary_addr();
        let msg = ExecuteMsg::CreateSchedule {
            beneficiary: ben.to_string(),
            total_amount: Uint128::new(100_000),
            start_time: 0,
            cliff_duration_secs: 0,
            vesting_duration_secs: 10_000,
        };
        execute(deps.as_mut(), mock_env(), admin.clone(), msg).unwrap();

        // Cancel
        execute(
            deps.as_mut(),
            mock_env(),
            admin,
            ExecuteMsg::CancelSchedule { schedule_id: 1 },
        )
        .unwrap();

        // Try to claim
        let err = execute(
            deps.as_mut(),
            env_at_time(5000),
            message_info(&ben, &[]),
            ExecuteMsg::Claim { schedule_id: 1 },
        )
        .unwrap_err();
        assert!(format!("{}", err).contains("cancelled"));
    }

    #[test]
    fn test_query_treasury() {
        let mut deps = mock_dependencies_with_truerepublic(&[]);
        let admin = message_info(&Addr::unchecked("admin"), &[]);
        instantiate(
            deps.as_mut(),
            mock_env(),
            admin,
            InstantiateMsg {
                domain_name: "TestDomain".to_string(),
            },
        )
        .unwrap();

        let res = query(deps.as_ref(), mock_env(), QueryMsg::TreasuryBalance {}).unwrap();
        let treasury: DomainTreasuryResponse = cosmwasm_std::from_json(res).unwrap();
        assert_eq!(treasury.domain_name, "TestDomain");
        assert_eq!(treasury.amount, "1000000pnyx");
    }
}
