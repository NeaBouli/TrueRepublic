use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),
    #[error("Unauthorized")]
    Unauthorized {},
    #[error("Order not found")]
    OrderNotFound {},
    #[error("Order not open")]
    OrderNotOpen {},
    #[error("Not order owner")]
    NotOrderOwner {},
    #[error("Insufficient output")]
    InsufficientOutput {},
    #[error("Pool snapshot not found")]
    PoolSnapshotNotFound {},
}
