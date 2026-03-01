use cosmwasm_std::{
    entry_point, to_json_binary, Binary, CosmosMsg, Deps, DepsMut, Env, MessageInfo, QueryRequest,
    Response, StdResult,
};
use truerepublic_bindings::{TrueRepublicMsg, TrueRepublicQuery};

use crate::error::ContractError;
use crate::msg::{ExecuteMsg, InstantiateMsg, QueryMsg};
use crate::state::{Config, DaoState, Proposal, ProposalAction};

const STATE_KEY: &[u8] = b"dao_state";

fn save_state(storage: &mut dyn cosmwasm_std::Storage, state: &DaoState) -> StdResult<()> {
    storage.set(STATE_KEY, &cosmwasm_std::to_json_vec(state)?);
    Ok(())
}

fn load_state(storage: &dyn cosmwasm_std::Storage) -> StdResult<DaoState> {
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
    // Verify domain exists via custom query
    let _domain: truerepublic_bindings::DomainResponse =
        deps.querier
            .query(&QueryRequest::Custom(TrueRepublicQuery::Domain {
                name: msg.domain_name.clone(),
            }))?;

    let state = DaoState {
        config: Config {
            admin: info.sender.clone(),
            domain_name: msg.domain_name,
            quorum_bps: msg.quorum_bps,
            threshold_bps: msg.threshold_bps,
            voting_period_secs: msg.voting_period_secs,
        },
        proposals: vec![],
        next_id: 1,
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
        ExecuteMsg::CreateProposal {
            title,
            description,
            action,
        } => {
            let mut state = load_state(deps.storage)?;
            let proposal = Proposal {
                id: state.next_id,
                proposer: info.sender,
                title,
                description,
                action,
                votes_for: 0,
                votes_against: 0,
                voters: vec![],
                open_until: env.block.time.seconds() + state.config.voting_period_secs,
                executed: false,
            };
            state.proposals.push(proposal);
            state.next_id += 1;
            save_state(deps.storage, &state)?;
            Ok(Response::new().add_attribute("action", "create_proposal"))
        }
        ExecuteMsg::Vote {
            proposal_id,
            approve,
        } => {
            let mut state = load_state(deps.storage)?;
            let proposal = state
                .proposals
                .iter_mut()
                .find(|p| p.id == proposal_id)
                .ok_or(ContractError::ProposalNotFound {})?;

            if env.block.time.seconds() > proposal.open_until {
                return Err(ContractError::ProposalExpired {});
            }
            let voter = info.sender.to_string();
            if proposal.voters.contains(&voter) {
                return Err(ContractError::AlreadyVoted {});
            }
            proposal.voters.push(voter);
            if approve {
                proposal.votes_for += 1;
            } else {
                proposal.votes_against += 1;
            }
            save_state(deps.storage, &state)?;
            Ok(Response::new().add_attribute("action", "vote"))
        }
        ExecuteMsg::ExecuteProposal { proposal_id } => {
            let mut state = load_state(deps.storage)?;
            let proposal = state
                .proposals
                .iter_mut()
                .find(|p| p.id == proposal_id)
                .ok_or(ContractError::ProposalNotFound {})?;

            if proposal.executed {
                return Err(ContractError::AlreadyExecuted {});
            }

            // Check quorum: total_votes * 10000 / member_count >= quorum_bps
            let domain: truerepublic_bindings::DomainResponse =
                deps.querier
                    .query(&QueryRequest::Custom(TrueRepublicQuery::Domain {
                        name: state.config.domain_name.clone(),
                    }))?;
            let total_votes = proposal.votes_for + proposal.votes_against;
            if domain.member_count > 0
                && total_votes * 10000 / (domain.member_count as u64) < state.config.quorum_bps
            {
                return Err(ContractError::QuorumNotReached {});
            }

            // Check threshold
            if total_votes > 0
                && proposal.votes_for * 10000 / total_votes < state.config.threshold_bps
            {
                return Err(ContractError::ProposalNotPassed {});
            }

            // Convert action to custom message
            let cosmos_msg = match &proposal.action {
                ProposalAction::PlaceStoneOnIssue { issue_name } => {
                    CosmosMsg::Custom(TrueRepublicMsg::PlaceStoneOnIssue {
                        domain_name: state.config.domain_name.clone(),
                        issue_name: issue_name.clone(),
                    })
                }
                ProposalAction::PlaceStoneOnSuggestion {
                    issue_name,
                    suggestion_name,
                } => CosmosMsg::Custom(TrueRepublicMsg::PlaceStoneOnSuggestion {
                    domain_name: state.config.domain_name.clone(),
                    issue_name: issue_name.clone(),
                    suggestion_name: suggestion_name.clone(),
                }),
                ProposalAction::DepositToDomain { amount } => {
                    CosmosMsg::Custom(TrueRepublicMsg::DepositToDomain {
                        domain_name: state.config.domain_name.clone(),
                        amount: amount.clone(),
                    })
                }
                ProposalAction::WithdrawFromDomain { recipient, amount } => {
                    CosmosMsg::Custom(TrueRepublicMsg::WithdrawFromDomain {
                        domain_name: state.config.domain_name.clone(),
                        recipient: recipient.clone(),
                        amount: amount.clone(),
                    })
                }
                ProposalAction::CastElectionVote {
                    issue_name,
                    candidate_name,
                    choice,
                } => CosmosMsg::Custom(TrueRepublicMsg::CastElectionVote {
                    domain_name: state.config.domain_name.clone(),
                    issue_name: issue_name.clone(),
                    candidate_name: candidate_name.clone(),
                    choice: *choice,
                }),
            };

            proposal.executed = true;
            save_state(deps.storage, &state)?;
            Ok(Response::new()
                .add_message(cosmos_msg)
                .add_attribute("action", "execute_proposal"))
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
        QueryMsg::Proposal { id } => {
            let state = load_state(deps.storage)?;
            let proposal = state
                .proposals
                .iter()
                .find(|p| p.id == id)
                .ok_or_else(|| cosmwasm_std::StdError::msg("not found: proposal"))?;
            to_json_binary(proposal)
        }
        QueryMsg::Proposals {} => {
            let state = load_state(deps.storage)?;
            to_json_binary(&state.proposals)
        }
        QueryMsg::DomainInfo {} => {
            let state = load_state(deps.storage)?;
            let domain: truerepublic_bindings::DomainResponse =
                deps.querier
                    .query(&QueryRequest::Custom(TrueRepublicQuery::Domain {
                        name: state.config.domain_name.clone(),
                    }))?;
            to_json_binary(&domain)
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
        let info = message_info(&Addr::unchecked("creator"), &[]);
        let msg = InstantiateMsg {
            domain_name: "TestDomain".to_string(),
            quorum_bps: 5000,
            threshold_bps: 6000,
            voting_period_secs: 86400,
        };
        let res = instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();
        assert_eq!(res.attributes[0].value, "instantiate");
    }

    #[test]
    fn test_create_and_vote() {
        let mut deps = mock_dependencies_with_truerepublic(&[]);
        let info = message_info(&Addr::unchecked("creator"), &[]);
        let msg = InstantiateMsg {
            domain_name: "TestDomain".to_string(),
            quorum_bps: 0, // no quorum for test
            threshold_bps: 5000,
            voting_period_secs: 86400,
        };
        instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        // Create proposal
        let create_msg = ExecuteMsg::CreateProposal {
            title: "Fund climate action".to_string(),
            description: "Deposit 1000 PNYX".to_string(),
            action: ProposalAction::PlaceStoneOnIssue {
                issue_name: "Climate".to_string(),
            },
        };
        execute(deps.as_mut(), mock_env(), info.clone(), create_msg).unwrap();

        // Vote
        let vote_msg = ExecuteMsg::Vote {
            proposal_id: 1,
            approve: true,
        };
        execute(deps.as_mut(), mock_env(), info, vote_msg).unwrap();

        // Query proposal
        let res = query(deps.as_ref(), mock_env(), QueryMsg::Proposal { id: 1 }).unwrap();
        let proposal: Proposal = cosmwasm_std::from_json(res).unwrap();
        assert_eq!(proposal.votes_for, 1);
        assert_eq!(proposal.votes_against, 0);
    }

    #[test]
    fn test_execute_proposal_emits_custom_msg() {
        let mut deps = mock_dependencies_with_truerepublic(&[]);
        let info = message_info(&Addr::unchecked("creator"), &[]);
        let msg = InstantiateMsg {
            domain_name: "TestDomain".to_string(),
            quorum_bps: 0,
            threshold_bps: 5000,
            voting_period_secs: 86400,
        };
        instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        let create_msg = ExecuteMsg::CreateProposal {
            title: "Place stone".to_string(),
            description: "Support climate".to_string(),
            action: ProposalAction::PlaceStoneOnIssue {
                issue_name: "Climate".to_string(),
            },
        };
        execute(deps.as_mut(), mock_env(), info.clone(), create_msg).unwrap();

        let vote_msg = ExecuteMsg::Vote {
            proposal_id: 1,
            approve: true,
        };
        execute(deps.as_mut(), mock_env(), info.clone(), vote_msg).unwrap();

        let exec_msg = ExecuteMsg::ExecuteProposal { proposal_id: 1 };
        let res = execute(deps.as_mut(), mock_env(), info, exec_msg).unwrap();
        assert_eq!(res.messages.len(), 1);
    }

    #[test]
    fn test_duplicate_vote_rejected() {
        let mut deps = mock_dependencies_with_truerepublic(&[]);
        let info = message_info(&Addr::unchecked("creator"), &[]);
        let msg = InstantiateMsg {
            domain_name: "TestDomain".to_string(),
            quorum_bps: 0,
            threshold_bps: 5000,
            voting_period_secs: 86400,
        };
        instantiate(deps.as_mut(), mock_env(), info.clone(), msg).unwrap();

        let create_msg = ExecuteMsg::CreateProposal {
            title: "Test".to_string(),
            description: "Test".to_string(),
            action: ProposalAction::PlaceStoneOnIssue {
                issue_name: "Test".to_string(),
            },
        };
        execute(deps.as_mut(), mock_env(), info.clone(), create_msg).unwrap();

        let vote_msg = ExecuteMsg::Vote {
            proposal_id: 1,
            approve: true,
        };
        execute(deps.as_mut(), mock_env(), info.clone(), vote_msg.clone()).unwrap();
        let err = execute(deps.as_mut(), mock_env(), info, vote_msg).unwrap_err();
        assert!(format!("{}", err).contains("Already voted"));
    }

    #[test]
    fn test_query_domain_info() {
        let mut deps = mock_dependencies_with_truerepublic(&[]);
        let info = message_info(&Addr::unchecked("creator"), &[]);
        let msg = InstantiateMsg {
            domain_name: "TestDomain".to_string(),
            quorum_bps: 5000,
            threshold_bps: 6000,
            voting_period_secs: 86400,
        };
        instantiate(deps.as_mut(), mock_env(), info, msg).unwrap();

        let res = query(deps.as_ref(), mock_env(), QueryMsg::DomainInfo {}).unwrap();
        let domain: truerepublic_bindings::DomainResponse = cosmwasm_std::from_json(res).unwrap();
        assert_eq!(domain.name, "TestDomain");
        assert_eq!(domain.member_count, 10);
    }
}
