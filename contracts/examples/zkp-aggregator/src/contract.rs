use cosmwasm_std::{
    entry_point, to_json_binary, Binary, CosmosMsg, Deps, DepsMut, Env, MessageInfo, QueryRequest,
    Response, StdResult,
};
use truerepublic_bindings::{NullifierResponse, TrueRepublicMsg, TrueRepublicQuery};

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{AggregatorState, Config, VotingRound};

const STATE_KEY: &[u8] = b"aggregator_state";

fn save_state(storage: &mut dyn cosmwasm_std::Storage, state: &AggregatorState) -> StdResult<()> {
    storage.set(STATE_KEY, &cosmwasm_std::to_json_vec(state)?);
    Ok(())
}

fn load_state(storage: &dyn cosmwasm_std::Storage) -> StdResult<AggregatorState> {
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
    let state = AggregatorState {
        config: Config {
            admin: info.sender,
            domain_name: msg.domain_name,
        },
        rounds: vec![],
        next_round_id: 1,
        used_nullifiers: vec![],
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
        ExecuteMsg::CreateRound {
            issue_name,
            suggestion_names,
            duration_secs,
        } => {
            let mut state = load_state(deps.storage)?;
            if info.sender != state.config.admin {
                return Err(ContractError::Unauthorized {});
            }
            let round = VotingRound {
                id: state.next_round_id,
                issue_name,
                suggestions: suggestion_names.clone(),
                scores: vec![0; suggestion_names.len()],
                participant_count: 0,
                started_at: env.block.time.seconds(),
                ends_at: env.block.time.seconds() + duration_secs,
                finalized: false,
            };
            state.rounds.push(round);
            state.next_round_id += 1;
            save_state(deps.storage, &state)?;
            Ok(Response::new().add_attribute("action", "create_round"))
        }
        ExecuteMsg::SubmitVote {
            round_id,
            nullifier_hex,
            ratings,
        } => {
            let mut state = load_state(deps.storage)?;

            // Check on-chain nullifier
            let nullifier_resp: NullifierResponse =
                deps.querier
                    .query(&QueryRequest::Custom(TrueRepublicQuery::Nullifier {
                        domain_name: state.config.domain_name.clone(),
                        nullifier_hex: nullifier_hex.clone(),
                    }))?;
            if nullifier_resp.used {
                return Err(ContractError::NullifierAlreadyUsed {});
            }

            // Check local nullifier for this round
            if state
                .used_nullifiers
                .iter()
                .any(|(rid, h)| *rid == round_id && *h == nullifier_hex)
            {
                return Err(ContractError::NullifierAlreadyUsed {});
            }

            let round = state
                .rounds
                .iter_mut()
                .find(|r| r.id == round_id)
                .ok_or(ContractError::RoundNotFound {})?;

            if round.finalized {
                return Err(ContractError::RoundFinalized {});
            }
            if env.block.time.seconds() > round.ends_at {
                return Err(ContractError::RoundNotActive {});
            }
            if ratings.len() != round.suggestions.len() {
                return Err(ContractError::InvalidRatingsLength {});
            }
            for r in &ratings {
                if *r < -5 || *r > 5 {
                    return Err(ContractError::RatingOutOfRange {});
                }
            }

            // Aggregate scores
            for (i, rating) in ratings.iter().enumerate() {
                round.scores[i] += rating;
            }
            round.participant_count += 1;
            state.used_nullifiers.push((round_id, nullifier_hex));

            // Emit stone placement messages for positive ratings
            let mut msgs: Vec<CosmosMsg<TrueRepublicMsg>> = vec![];
            for (i, rating) in ratings.iter().enumerate() {
                if *rating > 0 {
                    msgs.push(CosmosMsg::Custom(TrueRepublicMsg::PlaceStoneOnSuggestion {
                        domain_name: state.config.domain_name.clone(),
                        issue_name: round.issue_name.clone(),
                        suggestion_name: round.suggestions[i].clone(),
                    }));
                }
            }

            save_state(deps.storage, &state)?;
            let mut resp = Response::new().add_attribute("action", "submit_vote");
            for m in msgs {
                resp = resp.add_message(m);
            }
            Ok(resp)
        }
        ExecuteMsg::FinalizeRound { round_id } => {
            let mut state = load_state(deps.storage)?;
            if info.sender != state.config.admin {
                return Err(ContractError::Unauthorized {});
            }
            let round = state
                .rounds
                .iter_mut()
                .find(|r| r.id == round_id)
                .ok_or(ContractError::RoundNotFound {})?;
            round.finalized = true;
            save_state(deps.storage, &state)?;
            Ok(Response::new().add_attribute("action", "finalize_round"))
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
        QueryMsg::Round { id } => {
            let state = load_state(deps.storage)?;
            let round = state
                .rounds
                .iter()
                .find(|r| r.id == id)
                .ok_or_else(|| cosmwasm_std::StdError::msg("not found: round"))?;
            to_json_binary(round)
        }
        QueryMsg::ActiveRounds {} => {
            let state = load_state(deps.storage)?;
            let active: Vec<&VotingRound> = state.rounds.iter().filter(|r| !r.finalized).collect();
            to_json_binary(&active)
        }
        QueryMsg::RoundResults { id } => {
            let state = load_state(deps.storage)?;
            let round = state
                .rounds
                .iter()
                .find(|r| r.id == id)
                .ok_or_else(|| cosmwasm_std::StdError::msg("not found: round"))?;
            to_json_binary(round)
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
    fn test_create_round() {
        let mut deps = mock_dependencies_with_truerepublic(&[]);
        let info = message_info(&Addr::unchecked("admin"), &[]);
        instantiate(
            deps.as_mut(),
            mock_env(),
            info.clone(),
            InstantiateMsg {
                domain_name: "TestDomain".to_string(),
            },
        )
        .unwrap();

        let msg = ExecuteMsg::CreateRound {
            issue_name: "Climate".to_string(),
            suggestion_names: vec!["GreenDeal".to_string(), "CarbonTax".to_string()],
            duration_secs: 86400,
        };
        execute(deps.as_mut(), mock_env(), info, msg).unwrap();

        let res = query(deps.as_ref(), mock_env(), QueryMsg::Round { id: 1 }).unwrap();
        let round: VotingRound = cosmwasm_std::from_json(res).unwrap();
        assert_eq!(round.suggestions.len(), 2);
        assert_eq!(round.scores, vec![0, 0]);
    }

    #[test]
    fn test_submit_vote_aggregates_scores() {
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

        let create_msg = ExecuteMsg::CreateRound {
            issue_name: "Climate".to_string(),
            suggestion_names: vec!["A".to_string(), "B".to_string()],
            duration_secs: 86400,
        };
        execute(deps.as_mut(), mock_env(), admin, create_msg).unwrap();

        // Submit votes from different users
        let vote1 = ExecuteMsg::SubmitVote {
            round_id: 1,
            nullifier_hex: "abc123".to_string(),
            ratings: vec![3, -2],
        };
        execute(
            deps.as_mut(),
            mock_env(),
            message_info(&Addr::unchecked("voter1"), &[]),
            vote1,
        )
        .unwrap();

        let vote2 = ExecuteMsg::SubmitVote {
            round_id: 1,
            nullifier_hex: "def456".to_string(),
            ratings: vec![1, 4],
        };
        execute(
            deps.as_mut(),
            mock_env(),
            message_info(&Addr::unchecked("voter2"), &[]),
            vote2,
        )
        .unwrap();

        let res = query(deps.as_ref(), mock_env(), QueryMsg::Round { id: 1 }).unwrap();
        let round: VotingRound = cosmwasm_std::from_json(res).unwrap();
        assert_eq!(round.scores, vec![4, 2]); // 3+1=4, -2+4=2
        assert_eq!(round.participant_count, 2);
    }

    #[test]
    fn test_duplicate_nullifier_rejected() {
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

        let create_msg = ExecuteMsg::CreateRound {
            issue_name: "Test".to_string(),
            suggestion_names: vec!["A".to_string()],
            duration_secs: 86400,
        };
        execute(deps.as_mut(), mock_env(), admin, create_msg).unwrap();

        let vote = ExecuteMsg::SubmitVote {
            round_id: 1,
            nullifier_hex: "same_nullifier".to_string(),
            ratings: vec![3],
        };
        execute(
            deps.as_mut(),
            mock_env(),
            message_info(&Addr::unchecked("v1"), &[]),
            vote.clone(),
        )
        .unwrap();
        let err = execute(
            deps.as_mut(),
            mock_env(),
            message_info(&Addr::unchecked("v2"), &[]),
            vote,
        )
        .unwrap_err();
        assert!(format!("{}", err).contains("Nullifier already used"));
    }

    #[test]
    fn test_submit_vote_emits_stone_messages() {
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

        let create_msg = ExecuteMsg::CreateRound {
            issue_name: "Climate".to_string(),
            suggestion_names: vec!["A".to_string(), "B".to_string(), "C".to_string()],
            duration_secs: 86400,
        };
        execute(deps.as_mut(), mock_env(), admin, create_msg).unwrap();

        // Positive ratings on A (+3) and C (+1), negative on B (-2)
        let vote = ExecuteMsg::SubmitVote {
            round_id: 1,
            nullifier_hex: "unique".to_string(),
            ratings: vec![3, -2, 1],
        };
        let res = execute(
            deps.as_mut(),
            mock_env(),
            message_info(&Addr::unchecked("voter"), &[]),
            vote,
        )
        .unwrap();
        // Should emit 2 stone messages (A and C have positive ratings)
        assert_eq!(res.messages.len(), 2);
    }

    #[test]
    fn test_rating_out_of_range() {
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

        let create_msg = ExecuteMsg::CreateRound {
            issue_name: "Test".to_string(),
            suggestion_names: vec!["A".to_string()],
            duration_secs: 86400,
        };
        execute(deps.as_mut(), mock_env(), admin, create_msg).unwrap();

        let vote = ExecuteMsg::SubmitVote {
            round_id: 1,
            nullifier_hex: "test".to_string(),
            ratings: vec![6], // out of range
        };
        let err = execute(
            deps.as_mut(),
            mock_env(),
            message_info(&Addr::unchecked("voter"), &[]),
            vote,
        )
        .unwrap_err();
        assert!(format!("{}", err).contains("Rating out of range"));
    }
}
