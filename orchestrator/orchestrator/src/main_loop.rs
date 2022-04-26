//! This file contains the main loops for two distinct functions that just happen to reside int his same binary for ease of use. The Ethereum Signer and the Ethereum Oracle are both roles in Gravity
//! that can only be run by a validator. This single binary the 'Orchestrator' runs not only these two rules but also the untrusted role of a relayer, that does not need any permissions and has it's
//! own crate and binary so that anyone may run it.

use crate::metrics;
use crate::{
    ethereum_event_watcher::check_for_events, metrics::metrics_main_loop,
    oracle_resync::get_last_checked_block,
};
use clarity::{address::Address as EthAddress, Uint256};
use clarity::{utils::bytes_to_hex_str, PrivateKey as EthPrivateKey};
use cosmos_gravity::send::send_messages;
use cosmos_gravity::{
    build,
    query::{
        get_oldest_unsigned_logic_call, get_oldest_unsigned_transaction_batch,
        get_oldest_unsigned_valsets,
    },
};
use deep_space::client::ChainStatus;
use deep_space::error::CosmosGrpcError;
use deep_space::{Address, Contact};
use ethereum_gravity::utils::get_gravity_id;
use mhub2_proto::mhub2::query_client::QueryClient as Mhub2QueryClient;
use mhub2_proto::tx_committer::tx_committer_client::TxCommitterClient;
use relayer::main_loop::relayer_main_loop;
use std::{
    net,
    time::{Duration, Instant},
};
use tokio::time::sleep as delay_for;
use tonic::transport::Channel;
use web30::client::Web3;

/// The execution speed governing all loops in this file
/// which is to say all loops started by Orchestrator main
/// loop except the relayer loop
pub const ETH_SIGNER_LOOP_SPEED: Duration = Duration::from_secs(11);
pub const ETH_ORACLE_LOOP_SPEED: Duration = Duration::from_secs(13);

/// This loop combines the three major roles required to make
/// up the 'Orchestrator', all three of these are async loops
/// meaning they will occupy the same thread, but since they do
/// very little actual cpu bound work and spend the vast majority
/// of all execution time sleeping this shouldn't be an issue at all.
pub async fn orchestrator_main_loop(
    cosmos_address: Address,
    ethereum_key: EthPrivateKey,
    web3: Web3,
    contact: Contact,
    grpc_client: Mhub2QueryClient<Channel>,
    grpc_tx_committer_client: &TxCommitterClient<Channel>,
    gravity_contract_address: EthAddress,
    chain_id: String,
    eth_fee_calculator_url: Option<String>,
    metrics_listen: &net::SocketAddr,
) {
    let a = eth_oracle_main_loop(
        cosmos_address.clone(),
        web3.clone(),
        contact.clone(),
        grpc_client.clone(),
        gravity_contract_address,
        &grpc_tx_committer_client,
        chain_id.clone(),
    );

    let b = eth_signer_main_loop(
        cosmos_address.clone(),
        ethereum_key,
        web3.clone(),
        contact.clone(),
        grpc_client.clone(),
        gravity_contract_address,
        &grpc_tx_committer_client,
        chain_id.clone(),
    );

    let c = relayer_main_loop(
        ethereum_key,
        web3,
        grpc_client.clone(),
        gravity_contract_address,
        chain_id.clone(),
        eth_fee_calculator_url,
        true,
    );

    let d = metrics_main_loop(metrics_listen);

    futures::future::join4(a, b, c, d).await;
}

const DELAY: Duration = Duration::from_secs(5);

/// This function is responsible for making sure that Ethereum events are retrieved from the Ethereum blockchain
/// and ferried over to Cosmos where they will be used to issue tokens or process batches.
pub async fn eth_oracle_main_loop(
    our_cosmos_address: Address,
    web3: Web3,
    contact: Contact,
    grpc_client: Mhub2QueryClient<Channel>,
    gravity_contract_address: EthAddress,
    tx_committer_client: &TxCommitterClient<Channel>,
    chain_id: String,
) {
    let long_timeout_web30 = Web3::new(&web3.get_url(), Duration::from_secs(120));
    let mut last_checked_block: Uint256 = get_last_checked_block(
        grpc_client.clone(),
        our_cosmos_address,
        gravity_contract_address,
        &long_timeout_web30,
        chain_id.clone(),
    )
    .await;
    info!("Oracle resync complete, Oracle now operational");
    let mut grpc_client = grpc_client;

    loop {
        let loop_start = Instant::now();

        let latest_eth_block = web3.eth_block_number().await;
        let latest_cosmos_block = contact.get_chain_status().await;
        match (latest_eth_block, latest_cosmos_block) {
            (Ok(latest_eth_block), Ok(ChainStatus::Moving { block_height })) => {
                metrics::set_cosmos_block_height(block_height.clone());
                metrics::set_ethereum_block_height(latest_eth_block.clone());
                trace!(
                    "Latest Eth block {} Latest Cosmos block {}",
                    latest_eth_block,
                    block_height,
                );
            }
            (Ok(_latest_eth_block), Ok(ChainStatus::Syncing)) => {
                warn!("Cosmos node syncing, Eth oracle paused");
                delay_for(DELAY).await;
                continue;
            }
            (Ok(_latest_eth_block), Ok(ChainStatus::WaitingToStart)) => {
                warn!("Cosmos node syncing waiting for chain start, Eth oracle paused");
                delay_for(DELAY).await;
                continue;
            }
            (Ok(_), Err(_)) => {
                metrics::COSMOS_UNAVAILABLE.inc();
                warn!("Could not contact Cosmos grpc, trying again");
                delay_for(DELAY).await;
                continue;
            }
            (Err(_), Ok(_)) => {
                metrics::ETHEREUM_UNAVAILABLE.inc();
                warn!("Could not contact Eth node, trying again");
                delay_for(DELAY).await;
                continue;
            }
            (Err(_), Err(_)) => {
                metrics::COSMOS_UNAVAILABLE.inc();
                metrics::ETHEREUM_UNAVAILABLE.inc();
                error!("Could not reach Ethereum or Cosmos rpc!");
                delay_for(DELAY).await;
                continue;
            }
        }

        // Relays events from Ethereum -> Cosmos
        match check_for_events(
            &web3,
            &contact,
            &mut grpc_client,
            gravity_contract_address,
            our_cosmos_address,
            last_checked_block.clone(),
            &tx_committer_client.clone(),
            chain_id.clone(),
        )
        .await
        {
            Ok(new_block) => last_checked_block = new_block,
            Err(e) => {
                metrics::ETHEREUM_EVENT_CHECK_FAILURES.inc();
                error!("Failed to get events for block range, Check your Eth node and Cosmos gRPC {:?}", e);
                if let mhub2_utils::error::GravityError::CosmosGrpcError(err) = e {
                    if let CosmosGrpcError::TransactionFailed { tx: _, time: _ } = err {
                        delay_for(Duration::from_secs(10)).await;
                    }
                }
            }
        }

        // a bit of logic that tires to keep things running every LOOP_SPEED seconds exactly
        // this is not required for any specific reason. In fact we expect and plan for
        // the timing being off significantly
        let elapsed = Instant::now() - loop_start;
        if elapsed < ETH_ORACLE_LOOP_SPEED {
            delay_for(ETH_ORACLE_LOOP_SPEED - elapsed).await;
        }
    }
}

/// The eth_signer simply signs off on any batches or validator sets provided by the validator
/// since these are provided directly by a trusted Cosmsos node they can simply be assumed to be
/// valid and signed off on.
pub async fn eth_signer_main_loop(
    our_cosmos_address: Address,
    ethereum_key: EthPrivateKey,
    web3: Web3,
    contact: Contact,
    grpc_client: Mhub2QueryClient<Channel>,
    contract_address: EthAddress,
    tx_committer_client: &TxCommitterClient<Channel>,
    chain_id: String,
) {
    let our_ethereum_address = ethereum_key.to_public_key().unwrap();
    let mut grpc_client = grpc_client;

    let gravity_id = get_gravity_id(contract_address, our_ethereum_address, &web3).await;
    if gravity_id.is_err() {
        error!("Failed to get GravityID, check your Eth node");
        return;
    }
    let gravity_id = gravity_id.unwrap();
    let gravity_id = String::from_utf8(gravity_id).expect("Invalid GravityID");

    loop {
        let loop_start = Instant::now();

        let latest_eth_block = web3.eth_block_number().await;
        let latest_cosmos_block = contact.get_chain_status().await;
        match (latest_eth_block, latest_cosmos_block) {
            (Ok(latest_eth_block), Ok(ChainStatus::Moving { block_height })) => {
                metrics::set_cosmos_block_height(block_height.clone());
                metrics::set_ethereum_block_height(latest_eth_block.clone());
                trace!(
                    "Latest Eth block {} Latest Cosmos block {}",
                    latest_eth_block,
                    block_height,
                );
            }
            (Ok(_latest_eth_block), Ok(ChainStatus::Syncing)) => {
                warn!("Cosmos node syncing, Eth signer paused");
                delay_for(DELAY).await;
                continue;
            }
            (Ok(_latest_eth_block), Ok(ChainStatus::WaitingToStart)) => {
                warn!("Cosmos node syncing waiting for chain start, Eth signer paused");
                delay_for(DELAY).await;
                continue;
            }
            (Ok(_), Err(_)) => {
                metrics::COSMOS_UNAVAILABLE.inc();
                warn!("Could not contact Cosmos grpc, trying again");
                delay_for(DELAY).await;
                continue;
            }
            (Err(_), Ok(_)) => {
                metrics::ETHEREUM_UNAVAILABLE.inc();
                warn!("Could not contact Eth node, trying again");
                delay_for(DELAY).await;
                continue;
            }
            (Err(_), Err(_)) => {
                metrics::COSMOS_UNAVAILABLE.inc();
                metrics::ETHEREUM_UNAVAILABLE.inc();
                error!("Could not reach Ethereum or Cosmos rpc!");
                delay_for(DELAY).await;
                continue;
            }
        }

        // sign the last unsigned valsets
        match get_oldest_unsigned_valsets(&mut grpc_client, our_cosmos_address, chain_id.clone())
            .await
        {
            Ok(valsets) => {
                if valsets.is_empty() {
                    trace!("No validator sets to sign, node is caught up!")
                } else {
                    info!(
                        "Sending {} valset confirms starting with {}",
                        valsets.len(),
                        valsets[0].nonce
                    );
                    let messages = build::signer_set_tx_confirmation_messages(
                        ethereum_key,
                        valsets,
                        our_cosmos_address,
                        gravity_id.clone(),
                        chain_id.clone(),
                    );
                    send_messages(tx_committer_client, messages)
                        .await
                        .expect("Could not send messages");
                }
            }
            Err(e) => {
                metrics::UNSIGNED_VALSET_FAILURES.inc();
                info!(
                    "Failed to get unsigned valsets, check your Cosmos gRPC {:?}",
                    e
                );
            }
        }

        // sign the last unsigned batch, TODO check if we already have signed this
        match get_oldest_unsigned_transaction_batch(
            &mut grpc_client,
            our_cosmos_address,
            chain_id.clone(),
        )
        .await
        {
            Ok(Some(last_unsigned_batch)) => {
                info!(
                    "Sending batch confirm for {}:{} fees {} timeout {}",
                    last_unsigned_batch.token_contract,
                    last_unsigned_batch.nonce,
                    last_unsigned_batch.total_fee.amount,
                    last_unsigned_batch.batch_timeout,
                );
                let transaction_batches = vec![last_unsigned_batch];
                let messages = build::batch_tx_confirmation_messages(
                    ethereum_key,
                    transaction_batches,
                    our_cosmos_address,
                    gravity_id.clone(),
                    chain_id.clone(),
                );
                send_messages(tx_committer_client, messages)
                    .await
                    .expect("Could not send messages");
            }
            Ok(None) => info!("No unsigned batches! Everything good!"),
            Err(e) => {
                metrics::UNSIGNED_BATCH_FAILURES.inc();
                info!(
                    "Failed to get unsigned Batches, check your Cosmos gRPC {:?}",
                    e
                );
            }
        }

        let logic_calls =
            get_oldest_unsigned_logic_call(&mut grpc_client, our_cosmos_address, chain_id.clone())
                .await;
        if let Ok(logic_calls) = logic_calls {
            for logic_call in logic_calls {
                info!(
                    "Sending Logic call confirm for {}:{}",
                    bytes_to_hex_str(&logic_call.invalidation_id),
                    logic_call.invalidation_nonce
                );
                let logic_calls = vec![logic_call];
                let messages = build::contract_call_tx_confirmation_messages(
                    ethereum_key,
                    logic_calls,
                    our_cosmos_address,
                    gravity_id.clone(),
                    chain_id.clone(),
                );
                send_messages(tx_committer_client, messages)
                    .await
                    .expect("Could not send messages");
            }
        } else if let Err(e) = logic_calls {
            metrics::UNSIGNED_LOGIC_CALL_FAILURES.inc();
            info!(
                "Failed to get unsigned Logic Calls, check your Cosmos gRPC {:?}",
                e
            );
        }

        // a bit of logic that tires to keep things running every LOOP_SPEED seconds exactly
        // this is not required for any specific reason. In fact we expect and plan for
        // the timing being off significantly
        let elapsed = Instant::now() - loop_start;
        if elapsed < ETH_SIGNER_LOOP_SPEED {
            delay_for(ETH_SIGNER_LOOP_SPEED - elapsed).await;
        }
    }
}
