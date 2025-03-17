use anchor_lang::prelude::{borsh::{BorshSerialize, BorshDeserialize}, *};
use ethabi::{ParamType, Token};
use solana_program::program::get_return_data;

declare_id!("83zfZYacFrGq5eBnnp6EQPxapcpjpxdjAKpLavqtSJ32");

#[program]
pub mod hello_anchor {
    use super::*;

    pub fn set_data(ctx: Context<SetData>, value: String) -> Result<()> {
        ctx.accounts.data_account.data = value;
        Ok(())
    }

    pub fn get_data(ctx: Context<GetData>) -> Result<String> {
        if ctx.accounts.data_account.lamports() == 0 {
            return Ok("".to_string());
        }
        let account_data = DataAccount::try_deserialize(&mut &ctx.accounts.data_account.data.borrow()[..])?;
        Ok(account_data.data)
    }

    pub fn set_evm(ctx: Context<SetEvm>, contract_address: Vec<u8>) -> Result<()> {
        if &get_evm(&ctx, &contract_address)? == &"evm" {
            return do_set_evm(ctx.accounts.evm_executor.clone(), contract_address.clone(), "svm".to_string())
        }
        do_set_evm(ctx.accounts.evm_executor.clone(), contract_address.clone(), "evm".to_string())
    }
}

pub const EVM: u8 = 2;
pub const CROSS_VM_QUERY: u8 = 0;
pub const CROSS_VM_TX: u8 = 1;

#[derive(Debug, Clone, PartialEq, BorshDeserialize, BorshSerialize)]
pub struct MsgExecuteEvmContract {
    pub sender: String,
    pub contract_address: Vec<u8>,
    pub calldata: Vec<u8>,
    pub input_amount: Vec<u8>,
}

#[derive(Debug, Clone, PartialEq, BorshDeserialize, BorshSerialize)]
pub struct ContractQuery {
    pub contract_address: String,
    pub calldata: Vec<u8>,
}

pub fn get_evm(ctx: &Context<SetEvm>, contract_address: &Vec<u8>) -> Result<String> {
    let contract_query = ContractQuery {
        contract_address: hex::encode(contract_address),
        calldata: hex::decode("3bc5de30").unwrap(),
    };
    let ix = solana_program::instruction::Instruction {
        program_id: *ctx.accounts.evm_executor.key,
        accounts: vec![],
        data: [
            vec![EVM],
            vec![CROSS_VM_QUERY],
            contract_query.try_to_vec().unwrap(),
        ]
        .concat(),
    };
    solana_program::program::invoke_signed(&ix, &[], &[])?;
    let (_, return_data) = get_return_data().unwrap();
    let token = ethabi::decode(&[ParamType::String], &return_data).unwrap();
    Ok(token.get(0).unwrap().clone().into_string().unwrap())
}

fn do_set_evm<'a>(evm_executor: AccountInfo<'a>, contract_address: Vec<u8>, value: String) -> Result<()> {
    let discriminator = hex::decode("47064d6a").unwrap();
    let token = Token::String(value);
    let encoded_value = ethabi::encode(&[token]);
    let calldata = [discriminator, encoded_value].concat().to_vec();
    let execute_evm = MsgExecuteEvmContract {
        sender: "".to_string(),
        contract_address: contract_address.to_vec(),
        calldata,
        input_amount: vec![],
    };

    let ix = solana_program::instruction::Instruction {
        program_id: *evm_executor.key,
        accounts: vec![],
        data: [
            vec![EVM],
            vec![CROSS_VM_TX],
            execute_evm.try_to_vec().unwrap(),
        ]
        .concat(),
    };
    solana_program::program::invoke_signed(&ix, &[], &[])?;
    Ok(())
}

#[derive(Accounts)]
pub struct SetData<'info> {
    #[account(init_if_needed, seeds=[b"data"], bump, payer = signer, space = 100)]
    pub data_account: Account<'info, DataAccount>,
    #[account(mut)]
    pub signer: Signer<'info>,
    pub system_program: Program<'info, System>,
}

#[derive(Accounts)]
pub struct GetData<'info> {
    #[account(seeds=[b"data"], bump)]
    /// CHECK: PoC Only
    pub data_account: AccountInfo<'info>,
}

#[derive(Accounts)]
pub struct SetEvm<'info> {
    #[account(mut)]
    pub signer: Signer<'info>,
    /// CHECK: PoC only
    pub evm_executor: AccountInfo<'info>,
}

#[account]
pub struct DataAccount {
    data: String,
}
