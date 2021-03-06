use clarity::Address as EthAddress;
use num256::Uint256;
mod batches;
mod ethereum_events;
mod logic_call;
mod signatures;
mod valsets;
use crate::error::GravityError;

pub use batches::*;
pub use ethereum_events::*;
pub use logic_call::*;
pub use signatures::*;
pub use valsets::*;

#[derive(Serialize, Deserialize, Debug, Default, Clone, Eq, PartialEq, Hash)]
pub struct Erc20Token {
    pub amount: Uint256,
    #[serde(rename = "contract")]
    pub token_contract_address: EthAddress,
}

impl Erc20Token {
    pub fn from_proto(input: mhub2_proto::mhub2::ExternalToken) -> Result<Self, GravityError> {
        Ok(Erc20Token {
            amount: input.amount.parse()?,
            token_contract_address: EthAddress::parse_and_validate(&input.external_token_id)?,
        })
    }
}
