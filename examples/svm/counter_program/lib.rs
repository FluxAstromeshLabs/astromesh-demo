use anchor_lang::prelude::*;
// This is your program's public key and it will update
// automatically when you build the project.
declare_id!("EUkzMcNLDVysYtJv8mE6wCbnyuya67jNrtB5MrHG3Xja");

// this program supports creating student, student will upload their exam to have total score updated
// this is for demostrating purpose so no RBAC added

#[program]
mod counter {
    use super::*;
    pub fn count(ctx: Context<Count>) -> Result<()> {
        ctx.accounts.counter.value += 1;
        Ok(())
    }
}

#[derive(Accounts)]
pub struct Count<'info> {
    // initialize the signers' PDA account for this program, so 1 wallet => 1 program account, something same as PDA
    // you don't need an extra wallet just to interact with this program => better UX
    #[account(init_if_needed, seeds = [b"counter"], bump, payer = signer, space = 8+8)]
    pub counter: Account<'info, Counter>,
    #[account(mut)]
    pub signer: Signer<'info>,
    pub system_program: Program<'info, System>,
}

#[account]
pub struct Counter {
    pub value: u64,
}