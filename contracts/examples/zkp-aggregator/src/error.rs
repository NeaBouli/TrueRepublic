use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),
    #[error("Unauthorized")]
    Unauthorized {},
    #[error("Round not found")]
    RoundNotFound {},
    #[error("Round finalized")]
    RoundFinalized {},
    #[error("Round not active")]
    RoundNotActive {},
    #[error("Nullifier already used")]
    NullifierAlreadyUsed {},
    #[error("Invalid ratings length")]
    InvalidRatingsLength {},
    #[error("Rating out of range")]
    RatingOutOfRange {},
}
