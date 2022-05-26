use solana_program::{
    account_info::{next_account_info, AccountInfo},
    entrypoint,
    entrypoint::ProgramResult,
    program::invoke,
    pubkey::Pubkey,
    system_instruction,
    system_program,
};

entrypoint!(process_instruction);

fn process_instruction(
    program_id: &Pubkey,
    accounts: &[AccountInfo],
    instruction_data: &[u8],
) -> ProgramResult {
    let account_info_iter = &mut accounts.iter();

    let payer = next_account_info(account_info_iter)?;
    let recipient = next_account_info(account_info_iter)?;
    // The system program is a required account to invoke a system
    // instruction, even though we don't use it directly.
    let system_account = next_account_info(account_info_iter)?;

    assert!(payer.is_writable);
    assert!(payer.is_signer);
    assert!(recipient.is_writable);
    assert!(system_program::check_id(system_account.key));

    let lamports = 1000000;

    invoke(
        &system_instruction::transfer(payer.key, recipient.key, lamports),
        &[payer.clone(), recipient.clone(), system_account.clone()],
    )
}