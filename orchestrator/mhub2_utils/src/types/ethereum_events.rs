//! This file parses the Gravity contract ethereum events. Note that there is no Ethereum ABI unpacking implementation. Instead each event
//! is parsed directly from it's binary representation. This is technical debt within this implementation. It's quite easy to parse any
//! individual event manually but a generic decoder can be quite challenging to implement. A proper implementation would probably closely
//! mirror Serde and perhaps even become a serde crate for Ethereum ABI decoding
//! For now reference the ABI encoding document here https://docs.soliditylang.org/en/v0.8.3/abi-spec.html

use super::ValsetMember;
use crate::error::GravityError;
use clarity::Address as EthAddress;
use deep_space::utils::bytes_to_hex_str;
use num256::Uint256;
use std::ops::Mul;
use web30::client::Web3;
use web30::types::Log;

/// A parsed struct representing the Ethereum event fired by the Gravity contract
/// when the validator set is updated.
#[derive(Serialize, Deserialize, Debug, Default, Clone, Eq, PartialEq, Hash)]
pub struct ValsetUpdatedEvent {
    pub valset_nonce: Uint256,
    pub event_nonce: Uint256,
    pub block_height: Uint256,
    pub members: Vec<ValsetMember>,
    pub tx_hash: String,
}

impl ValsetUpdatedEvent {
    /// This function is not an abi compatible bytes parser, but it's actually
    /// not hard at all to extract data like this by hand.
    pub fn from_log(input: &Log) -> Result<ValsetUpdatedEvent, GravityError> {
        // we have one indexed event so we should fined two indexes, one the event itself
        // and one the indexed nonce
        if input.topics.get(1).is_none() {
            return Err(GravityError::InvalidEventLogError(
                "Too few topics".to_string(),
            ));
        }
        let valset_nonce_data = &input.topics[1];
        let valset_nonce = Uint256::from_bytes_be(valset_nonce_data);
        if valset_nonce > u64::MAX.into() {
            return Err(GravityError::InvalidEventLogError(
                "Nonce overflow, probably incorrect parsing".to_string(),
            ));
        }
        let valset_nonce: u64 = valset_nonce.to_string().parse().unwrap();

        // first index is the event nonce, following two have event data we don't
        // care about, fourth index contains the length of the eth address array
        let index_start = 0;
        let index_end = index_start + 32;
        let nonce_data = &input.data[index_start..index_end];
        let event_nonce = Uint256::from_bytes_be(nonce_data);
        if event_nonce > u64::MAX.into() {
            return Err(GravityError::InvalidEventLogError(
                "Nonce overflow, probably incorrect parsing".to_string(),
            ));
        }
        let event_nonce: u64 = event_nonce.to_string().parse().unwrap();
        // first index is the event nonce, following two have event data we don't
        // care about, fourth index contains the length of the eth address array
        let index_start = 3 * 32;
        let index_end = index_start + 32;
        let eth_addresses_offset = index_start + 32;
        let len_eth_addresses = Uint256::from_bytes_be(&input.data[index_start..index_end]);
        if len_eth_addresses > usize::MAX.into() {
            return Err(GravityError::InvalidEventLogError(
                "Ethereum array len overflow, probably incorrect parsing".to_string(),
            ));
        }
        let len_eth_addresses: usize = len_eth_addresses.to_string().parse().unwrap();
        let index_start = (4 + len_eth_addresses) * 32;
        let index_end = index_start + 32;
        let powers_offset = index_start + 32;
        let len_powers = Uint256::from_bytes_be(&input.data[index_start..index_end]);
        if len_powers > usize::MAX.into() {
            return Err(GravityError::InvalidEventLogError(
                "Powers array len overflow, probably incorrect parsing".to_string(),
            ));
        }
        let len_powers: usize = len_eth_addresses.to_string().parse().unwrap();
        if len_powers != len_eth_addresses {
            return Err(GravityError::InvalidEventLogError(
                "Array len mismatch, probably incorrect parsing".to_string(),
            ));
        }

        let mut validators = Vec::new();
        for i in 0..len_eth_addresses {
            let power_start = (i * 32) + powers_offset;
            let power_end = power_start + 32;
            let address_start = (i * 32) + eth_addresses_offset;
            let address_end = address_start + 32;
            let power = Uint256::from_bytes_be(&input.data[power_start..power_end]);
            // an eth address at 20 bytes is 12 bytes shorter than the Uint256 it's stored in.
            let eth_address = EthAddress::from_slice(&input.data[address_start + 12..address_end]);
            if eth_address.is_err() {
                return Err(GravityError::InvalidEventLogError(
                    "Ethereum Address parsing error, probably incorrect parsing".to_string(),
                ));
            }
            let eth_address = Some(eth_address.unwrap());
            if power > u64::MAX.into() {
                return Err(GravityError::InvalidEventLogError(
                    "Power greater than u64::MAX, probably incorrect parsing".to_string(),
                ));
            }
            let power: u64 = power.to_string().parse().unwrap();
            validators.push(ValsetMember { power, eth_address })
        }
        let mut check = validators.clone();
        check.sort();
        check.reverse();
        // if the validator set is not sorted we're in a bad spot
        if validators != check {
            trace!(
                "Someone submitted an unsorted validator set, this means all updates will fail until someone feeds in this unsorted value by hand {:?} instead of {:?}",
                validators, check
            );
        }
        let block_height = if let Some(bn) = input.block_number.clone() {
            bn
        } else {
            return Err(GravityError::InvalidEventLogError(
                "Log does not have block number, we only search logs already in blocks?"
                    .to_string(),
            ));
        };

        Ok(ValsetUpdatedEvent {
            valset_nonce: valset_nonce.into(),
            event_nonce: event_nonce.into(),
            block_height,
            members: validators,
            tx_hash: format!(
                "0x{}",
                bytes_to_hex_str(input.transaction_hash.as_deref().unwrap())
            ),
        })
    }
    pub fn from_logs(input: &[Log]) -> Result<Vec<ValsetUpdatedEvent>, GravityError> {
        let mut res = Vec::new();
        for item in input {
            res.push(ValsetUpdatedEvent::from_log(item)?);
        }
        Ok(res)
    }
    /// returns all values in the array with event nonces greater
    /// than the provided value
    pub fn filter_by_event_nonce(event_nonce: u64, input: &[Self]) -> Vec<Self> {
        let mut ret = Vec::new();
        for item in input {
            if item.event_nonce > event_nonce.into() {
                ret.push(item.clone())
            }
        }
        ret
    }
}

/// A parsed struct representing the Ethereum event fired by the Gravity contract when
/// a transaction batch is executed.
#[derive(Serialize, Deserialize, Debug, Default, Clone, Eq, PartialEq, Hash)]
pub struct TransactionBatchExecutedEvent {
    /// the nonce attached to the transaction batch that follows
    /// it throughout it's lifecycle
    pub batch_nonce: Uint256,
    /// The block height this event occurred at
    pub block_height: Uint256,
    /// The ERC20 token contract address for the batch executed, since batches are uniform
    /// in token type there is only one
    pub erc20: EthAddress,
    /// the event nonce representing a unique ordering of events coming out
    /// of the Gravity solidity contract. Ensuring that these events can only be played
    /// back in order
    pub event_nonce: Uint256,
    /// Hash of the transaction in which event has occurred
    pub tx_hash: String,

    pub tx_sender: String,
    pub fee: Uint256,
}

impl TransactionBatchExecutedEvent {
    pub async fn from_log(
        input: &Log,
        web3: &Web3,
    ) -> Result<TransactionBatchExecutedEvent, GravityError> {
        if let (Some(batch_nonce_data), Some(erc20_data)) =
            (input.topics.get(1), input.topics.get(2))
        {
            let batch_nonce = Uint256::from_bytes_be(batch_nonce_data);
            let erc20 = EthAddress::from_slice(&erc20_data[12..32])?;
            let event_nonce = Uint256::from_bytes_be(&input.data);
            let block_height = if let Some(bn) = input.block_number.clone() {
                bn
            } else {
                return Err(GravityError::InvalidEventLogError(
                    "Log does not have block number, we only search logs already in blocks?"
                        .to_string(),
                ));
            };
            if event_nonce > u64::MAX.into()
                || batch_nonce > u64::MAX.into()
                || block_height > u64::MAX.into()
            {
                Err(GravityError::InvalidEventLogError(
                    "Event nonce overflow, probably incorrect parsing".to_string(),
                ))
            } else {
                let tx = web3
                    .eth_get_transaction_by_hash(Uint256::from_bytes_be(
                        input.transaction_hash.as_deref().unwrap().as_slice(),
                    ))
                    .await
                    .unwrap()
                    .unwrap();
                let tx_receipt = web3
                    .eth_get_transaction_receipt(Uint256::from_bytes_be(
                        input.transaction_hash.as_deref().unwrap().as_slice(),
                    ))
                    .await
                    .unwrap()
                    .unwrap();

                Ok(TransactionBatchExecutedEvent {
                    batch_nonce,
                    block_height,
                    erc20,
                    event_nonce,
                    tx_hash: format!(
                        "0x{}",
                        bytes_to_hex_str(input.transaction_hash.as_deref().unwrap())
                    ),
                    tx_sender: tx_receipt.from.to_string(),
                    fee: tx.gas_price.mul(tx_receipt.gas_used),
                })
            }
        } else {
            Err(GravityError::InvalidEventLogError(
                "Too few topics".to_string(),
            ))
        }
    }
    pub async fn from_logs(
        input: &[Log],
        web3: &Web3,
    ) -> Result<Vec<TransactionBatchExecutedEvent>, GravityError> {
        let mut res = Vec::new();
        for item in input {
            res.push(TransactionBatchExecutedEvent::from_log(item, web3).await?);
        }
        Ok(res)
    }
    /// returns all values in the array with event nonces greater
    /// than the provided value
    pub fn filter_by_event_nonce(event_nonce: u64, input: &[Self]) -> Vec<Self> {
        let mut ret = Vec::new();
        for item in input {
            if item.event_nonce > event_nonce.into() {
                ret.push(item.clone())
            }
        }
        ret
    }
}

/// A parsed struct representing the Ethereum event fired when someone makes a deposit
/// on the Gravity contract
#[derive(Serialize, Deserialize, Debug, Clone, Eq, PartialEq, Hash)]
pub struct TransferToChainEvent {
    /// The token contract address for the deposit
    pub erc20: EthAddress,
    /// The Ethereum Sender
    pub sender: EthAddress,
    /// The destination chain
    pub destination_chain: String,
    /// The Cosmos destination
    pub destination: EthAddress,
    /// The amount of the erc20 token that is being sent
    pub amount: Uint256,
    pub fee: Uint256,
    /// The transaction's nonce, used to make sure there can be no accidental duplication
    pub event_nonce: Uint256,
    /// The block height this event occurred at
    pub block_height: Uint256,
    /// Hash of the transaction in which event has occurred
    pub tx_hash: String,
}

impl TransferToChainEvent {
    pub fn from_log(input: &Log) -> Result<TransferToChainEvent, GravityError> {
        let topics = (
            input.topics.get(1),
            input.topics.get(2),
            input.topics.get(3),
        );
        if let (Some(erc20_data), Some(sender_data), Some(destination_chain_data)) = topics {
            let erc20 = EthAddress::from_slice(&erc20_data[12..32])?;
            let sender = EthAddress::from_slice(&sender_data[12..32])?;
            let destination_chain = String::from_utf8_lossy(&destination_chain_data)
                .trim_end_matches('\x00')
                .to_string();
            let destination = EthAddress::from_slice(&input.data[12..32]).unwrap();
            let amount = Uint256::from_bytes_be(&input.data[32..64]);
            let fee = Uint256::from_bytes_be(&input.data[64..96]);
            let event_nonce = Uint256::from_bytes_be(&input.data[96..]);
            let block_height = if let Some(bn) = input.block_number.clone() {
                bn
            } else {
                return Err(GravityError::InvalidEventLogError(
                    "Log does not have block number, we only search logs already in blocks?"
                        .to_string(),
                ));
            };
            if event_nonce > u64::MAX.into() || block_height > u64::MAX.into() {
                Err(GravityError::InvalidEventLogError(
                    "Event nonce overflow, probably incorrect parsing".to_string(),
                ))
            } else {
                Ok(TransferToChainEvent {
                    erc20,
                    sender,
                    destination_chain,
                    destination,
                    amount,
                    fee,
                    event_nonce,
                    block_height,
                    tx_hash: format!(
                        "0x{}",
                        bytes_to_hex_str(input.transaction_hash.as_deref().unwrap())
                    ),
                })
            }
        } else {
            Err(GravityError::InvalidEventLogError(
                "Too few topics".to_string(),
            ))
        }
    }
    pub fn from_logs(input: &[Log]) -> Result<Vec<TransferToChainEvent>, GravityError> {
        let mut res = Vec::new();
        for item in input {
            res.push(Self::from_log(item)?);
        }
        Ok(res)
    }
    /// returns all values in the array with event nonces greater
    /// than the provided value
    pub fn filter_by_event_nonce(event_nonce: u64, input: &[Self]) -> Vec<Self> {
        let mut ret = Vec::new();
        for item in input {
            if item.event_nonce > event_nonce.into() {
                ret.push(item.clone())
            }
        }
        ret
    }
}

/// A parsed struct representing the Ethereum event fired when someone uses the Gravity
/// contract to deploy a new ERC20 contract representing a Cosmos asset
#[derive(Serialize, Deserialize, Debug, Default, Clone, Eq, PartialEq, Hash)]
pub struct LogicCallExecutedEvent {
    pub invalidation_id: Vec<u8>,
    pub invalidation_nonce: Uint256,
    pub return_data: Vec<u8>,
    pub event_nonce: Uint256,
    pub block_height: Uint256,
    pub tx_hash: String,
}

impl LogicCallExecutedEvent {
    pub fn from_log(input: &Log) -> Result<LogicCallExecutedEvent, GravityError> {
        let event_nonce = Uint256::from_bytes_be(&input.data[64..96]);

        if event_nonce > u64::MAX.into() {
            Err(GravityError::InvalidEventLogError(
                "Event nonce overflow, probably incorrect parsing".to_string(),
            ))
        } else {
            Ok(LogicCallExecutedEvent {
                invalidation_id: input.data[..32].into(),
                invalidation_nonce: Uint256::from_bytes_be(&input.data[32..64]),
                return_data: input.data[96..].into(),
                event_nonce,
                block_height: Default::default(),
                tx_hash: "".to_string(),
            })
        }
    }
    pub fn from_logs(input: &[Log]) -> Result<Vec<LogicCallExecutedEvent>, GravityError> {
        let mut res = Vec::new();
        for item in input {
            res.push(LogicCallExecutedEvent::from_log(item)?);
        }
        Ok(res)
    }
    /// returns all values in the array with event nonces greater
    /// than the provided value
    pub fn filter_by_event_nonce(event_nonce: u64, input: &[Self]) -> Vec<Self> {
        let mut ret = Vec::new();
        for item in input {
            if item.event_nonce > event_nonce.into() {
                ret.push(item.clone())
            }
        }
        ret
    }
}

/// Function used for debug printing hex dumps
/// of ethereum events
fn _debug_print_data(input: &[u8]) {
    let count = input.len() / 32;
    println!("data hex dump");
    for i in 0..count {
        println!("0x{}", bytes_to_hex_str(&input[(i * 32)..((i * 32) + 32)]))
    }
    println!("end dump");
}
