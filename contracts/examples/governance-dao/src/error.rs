use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),
    #[error("Unauthorized")]
    Unauthorized {},
    #[error("Already voted")]
    AlreadyVoted {},
    #[error("Proposal not found")]
    ProposalNotFound {},
    #[error("Proposal expired")]
    ProposalExpired {},
    #[error("Proposal not passed")]
    ProposalNotPassed {},
    #[error("Quorum not reached")]
    QuorumNotReached {},
    #[error("Already executed")]
    AlreadyExecuted {},
}
