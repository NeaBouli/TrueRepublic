use cosmwasm_std::{
    entry_point, to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult,
};
use schemars::JsonSchema;
use serde::{Deserialize, Serialize};

const STATE_KEY: &[u8] = b"state";

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct Proposal {
    pub id: u64,
    pub title: String,
    pub description: String,
    pub votes: Vec<i8>, // Systemic Consensing: -5 to +5
    pub executed: bool,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct State {
    pub proposals: Vec<Proposal>,
    pub next_id: u64,
    pub key_pairs: Vec<KeyPair>,
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub struct KeyPair {
    pub owner: String,
    pub public_key: String,
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
        .ok_or_else(|| cosmwasm_std::StdError::not_found("state"))?;
    cosmwasm_std::from_json(data)
}

#[entry_point]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    _msg: InstantiateMsg,
) -> StdResult<Response> {
    let state = State {
        proposals: vec![],
        next_id: 1,
        key_pairs: vec![KeyPair {
            owner: info.sender.to_string(),
            public_key: "pk_placeholder".to_string(),
        }],
    };
    save_state(deps.storage, &state)?;
    Ok(Response::default())
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum ExecuteMsg {
    SubmitProposal {
        title: String,
        description: String,
    },
    Vote {
        proposal_id: u64,
        vote: i8,
        public_key: String,
    },
}

#[entry_point]
pub fn execute(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: ExecuteMsg,
) -> StdResult<Response> {
    let mut state = load_state(deps.storage)?;
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
            save_state(deps.storage, &state)?;
            Ok(Response::new().add_attribute("action", "submit_proposal"))
        }
        ExecuteMsg::Vote {
            proposal_id,
            vote,
            public_key,
        } => {
            if !(-5..=5).contains(&vote) {
                return Err(cosmwasm_std::StdError::generic_err(
                    "Vote must be between -5 and 5",
                ));
            }
            let sender = info.sender.as_str();
            let key_exists = state
                .key_pairs
                .iter()
                .any(|kp| kp.owner == sender && kp.public_key == public_key);
            if !key_exists {
                return Err(cosmwasm_std::StdError::generic_err("Invalid key pair"));
            }
            let proposal = state
                .proposals
                .iter_mut()
                .find(|p| p.id == proposal_id)
                .ok_or(cosmwasm_std::StdError::generic_err("Proposal not found"))?;
            proposal.votes.push(vote);
            save_state(deps.storage, &state)?;
            Ok(Response::new().add_attribute("action", "vote"))
        }
    }
}

#[derive(Serialize, Deserialize, Clone, Debug, PartialEq, JsonSchema)]
pub enum QueryMsg {
    GetProposals {},
}

#[entry_point]
pub fn query(deps: Deps, _env: Env, msg: QueryMsg) -> StdResult<Binary> {
    let state = load_state(deps.storage)?;
    match msg {
        QueryMsg::GetProposals {} => to_json_binary(&state.proposals),
    }
}
