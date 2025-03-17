use anchor_lang::prelude::{borsh::BorshDeserialize, *};
use ethabi::{ParamType, Token};
use solana_program::program::get_return_data;

declare_id!("83zfZYacFrGq5eBnnp6EQPxapcpjpxdjAKpLavqtSJ32");

#[program]
pub mod hello_anchor {
    use anchor_lang::prelude::borsh::BorshSerialize;

    use super::*;

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

    pub fn set_data(ctx: Context<SetData>, value: String) -> Result<()> {
        ctx.accounts.data_account.data = value;
        Ok(())
    }

    pub fn get_data(ctx: Context<GetData>) -> Result<String> {
        Ok(ctx.accounts.data_account.data.clone())
    }

    pub fn set_evm(ctx: Context<SetEvm>, contract_address: Vec<u8>, value: String) -> Result<()> {
        do_set_svm(ctx.accounts.evm_executor.clone(), contract_address, value)
    }

    pub fn conditional_set_evm(ctx: Context<ConditionalSetEvm>, contract_address: Vec<u8>) -> Result<()> {
        if &get_evm(&ctx, &contract_address)? == &"evm" {
            return do_set_svm(ctx.accounts.evm_executor.clone(), contract_address.clone(), "svm".to_string())
        }

        do_set_svm(ctx.accounts.evm_executor.clone(), contract_address.clone(), "evm".to_string())
    }
}

pub fn get_evm(ctx: &Context<ConditionalSetEvm>, contract_address: &Vec<u8>) -> Result<String> {
    let contract_query = ContractQuery {
        contract_address: hex::encode(contract_address),
        calldata: hex::decode("3163d265").unwrap(),
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

fn do_set_svm<'a>(evm_executor: AccountInfo<'a>, contract_address: Vec<u8>, value: String) -> Result<()> {
    let discriminator = hex::decode("4ed3885e").unwrap();
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
    pub data_account: Account<'info, DataAccount>,
}

#[derive(Accounts)]
pub struct SetEvm<'info> {
    #[account(mut)]
    pub signer: Signer<'info>,
    /// CHECK: PoC only
    pub evm_executor: AccountInfo<'info>,
}

#[derive(Accounts)]
pub struct ConditionalSetEvm<'info> {
    #[account(mut)]
    pub signer: Signer<'info>,
    /// CHECK: PoC only
    pub evm_executor: AccountInfo<'info>,
}

#[account]
pub struct DataAccount {
    data: String,
}
