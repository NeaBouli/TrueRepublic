use cosmwasm_std::{entry_point, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult, to_binary};
use serde::{Deserialize, Serialize};

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct Proposal {
    pub id: u64,
    pub title: String,
    pub description: String,
    pub votes: Vec<i8>, // Systemic Consensing: -5 bis +5
    pub executed: bool,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct State {
    pub proposals: Vec<Proposal>,
    pub next_id: u64,
    pub key_pairs: Vec<KeyPair>, // Anonymity: Globale/Domain-spezifische Keys
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct KeyPair {
    pub owner: String,
    pub public_key: String, // Platzhalter für echte Schlüssel
}

#[entry_point]
pub fn instantiate(deps: DepsMut, _env: Env, info: MessageInfo, _msg: ()) -> StdResult<Response> {
    deps.storage.save(b"state", &State { 
        proposals: vec![], 
        next_id: 1,
        key_pairs: vec![KeyPair { owner: info.sender.to_string(), public_key: "pk_placeholder".to_string() }]
    })?;
    Ok(Response::default())
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub enum ExecuteMsg {
    SubmitProposal { title: String, description: String },
    Vote { proposal_id: u64, vote: i8, public_key: String },
}

#[entry_point]
pub fn execute(deps: DepsMut, _env: Env, info: MessageInfo, msg: ExecuteMsg) -> StdResult<Response> {
    let mut state = deps.storage.load::<State>(b"state")?;
    match msg {
        ExecuteMsg::SubmitProposal { title, description } => {
            state.proposals.push(Proposal {
                id: state.next_id,
                title,
                description,
                votes: vec![],
                executed: false,
            });
            state.next_id += 1;
            deps.storage.save(b"state", &state)?;
            Ok(Response::new().add_attribute("action", "submit_proposal"))
        }
        ExecuteMsg::Vote { proposal_id, vote, public_key } => {
            if vote < -5 || vote > 5 {
                return Err(cosmwasm_std::StdError::generic_err("Vote must be between -5 and 5"));
            }
            let key_exists = state.key_pairs.iter().any(|kp| kp.owner == info.sender.to_string() && kp.public_key == public_key);
            if !key_exists {
                return Err(cosmwasm_std::StdError::generic_err("Invalid key pair"));
            }
            let proposal = state.proposals.iter_mut().find(|p| p.id == proposal_id)
                .ok_or(cosmwasm_std::StdError::generic_err("Proposal not found"))?;
            proposal.votes.push(vote);
            deps.storage.save(b"state", &state)?;
            Ok(Response::new().add_attribute("action", "vote"))
        }
    }
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub enum QueryMsg {
    GetProposals {},
}

#[entry_point]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    let state = deps.storage.load::<State>(b"state")?;
    match msg {
        QueryMsg::GetProposals {} => to_binary(&state.proposals),
    }
}
