use cosmwasm_std::StdError;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),
    #[error("Unauthorized")]
    Unauthorized {},
    #[error("Schedule not found")]
    ScheduleNotFound {},
    #[error("Nothing to claim")]
    NothingToClaim {},
    #[error("Already cancelled")]
    AlreadyCancelled {},
    #[error("Not beneficiary")]
    NotBeneficiary {},
}
