//! Orchestrator is a sort of specialized relayer for Althea-Gravity that runs on every validator.
//! Things this binary is responsible for
//!   * Performing all the Ethereum signing required to submit updates and generate batches
//!   * Progressing the validator set update generation process.
//!   * Observing events on the Ethereum chain and submitting oracle messages for validator consensus
//! Things this binary needs
//!   * Access to the validators signing Ethereum key
//!   * Access to the validators Cosmos key
//!   * Access to an Cosmos chain RPC server
//!   * Access to an Ethereum chain RPC server

#[macro_use]
extern crate serde_derive;
#[macro_use]
extern crate lazy_static;
#[macro_use]
extern crate log;

mod ethereum_event_watcher;
mod get_with_retry;
mod main_loop;
mod metrics;
mod oracle_resync;

use crate::main_loop::orchestrator_main_loop;
use clarity::Address as EthAddress;
use clarity::PrivateKey as EthPrivateKey;
use deep_space::Address;
use docopt::Docopt;
use env_logger::Env;
use main_loop::{ETH_ORACLE_LOOP_SPEED, ETH_SIGNER_LOOP_SPEED};
use mhub2_proto::tx_committer::AddressRequest;
use mhub2_utils::connection_prep::create_rpc_connections;
use mhub2_utils::connection_prep::{check_delegate_addresses, wait_for_cosmos_node_ready};
use relayer::main_loop::LOOP_SPEED as RELAYER_LOOP_SPEED;
use std::cmp::min;

#[derive(Debug, Deserialize)]
struct Args {
    flag_ethereum_key: String,
    flag_cosmos_grpc: String,
    flag_ethereum_rpc: String,
    flag_contract_address: String,
    flag_chain_id: String,
    flag_metrics_listen: String,
    flag_committer_grpc: Option<String>,
    flag_eth_fee_calculator_url: Option<String>,
}

lazy_static! {
    pub static ref USAGE: String = format!(
    "Usage: {} [--eth-fee-calculator-url=<furl>] [--committer-grpc=<url>] --chain-id=<id> --ethereum-key=<key> --cosmos-grpc=<url> --ethereum-rpc=<url> --contract-address=<addr> --metrics-listen=<addr>
    Options:
        -h --help                           Show this screen.
        --ethereum-key=<ekey>               The Ethereum private key of the validator
        --cosmos-grpc=<gurl>                The Cosmos gRPC url, usually the validator
        --committer-grpc=<gurl>             The Committer gRPC url [default: http://localhost:7070]
        --address-prefix=<prefix>           The prefix for addresses on this Cosmos chain
        --ethereum-rpc=<eurl>               The Ethereum RPC url, should be a self hosted node
        --contract-address=<addr>           The Ethereum contract address for Gravity, this is temporary
        --metrics-listen=<addr>             The address metrics server listens on [default: 127.0.0.1:3000]. 
    About:
        The Validator companion binary for Minter Hub 2. This must be run by all Minter Hub 2 chain validators
        and is a mix of a relayer + oracle + external signing infrastructure
        Written By: {}
        Version {}",
        env!("CARGO_PKG_NAME"),
        env!("CARGO_PKG_AUTHORS"),
        env!("CARGO_PKG_VERSION"),
    );
}

#[actix_rt::main]
async fn main() {
    env_logger::Builder::from_env(Env::default().default_filter_or("info")).init();
    // On Linux static builds we need to probe ssl certs path to be able to
    // do TLS stuff.
    openssl_probe::init_ssl_cert_env_vars();

    let args: Args = Docopt::new(USAGE.as_str())
        .and_then(|d| d.deserialize())
        .unwrap_or_else(|e| e.exit());
    let ethereum_key: EthPrivateKey = args
        .flag_ethereum_key
        .parse()
        .expect("Invalid Ethereum private key!");
    let contract_address: EthAddress = args
        .flag_contract_address
        .parse()
        .expect("Invalid contract address!");
    let metrics_listen = args
        .flag_metrics_listen
        .parse()
        .expect("Invalid metrics listen address!");

    let chain_id = args.flag_chain_id;
    let eth_fee_calculator_url = args.flag_eth_fee_calculator_url;

    let timeout = min(
        min(ETH_SIGNER_LOOP_SPEED, ETH_ORACLE_LOOP_SPEED),
        RELAYER_LOOP_SPEED,
    );

    trace!("Probing RPC connections");

    let committer_url = if args.flag_committer_grpc.is_some() {
        args.flag_committer_grpc.unwrap()
    } else {
        "http://localhost:7070".into()
    };

    // probe all rpc connections and see if they are valid
    let connections = create_rpc_connections(
        "hub".into(),
        Some(args.flag_cosmos_grpc),
        Some(committer_url),
        Some(args.flag_ethereum_rpc),
        timeout,
    )
    .await;

    let mut grpc = connections.grpc.clone().unwrap();
    let contact = connections.contact.clone().unwrap();
    let mut grpc_committer = connections.grpc_committer.clone().unwrap();

    let public_eth_key = ethereum_key
        .to_public_key()
        .expect("Invalid Ethereum Private Key!");

    let our_cosmos_address = Address::from_bech32(
        grpc_committer
            .address(AddressRequest {})
            .await
            .unwrap()
            .into_inner()
            .address,
    )
    .unwrap();

    info!("Starting Gravity Validator companion binary Relayer + Oracle + Eth Signer");
    info!(
        "Ethereum Address: {} Cosmos Address {}",
        public_eth_key, our_cosmos_address
    );

    // check if the cosmos node is syncing, if so wait for it
    // we can't move any steps above this because they may fail on an incorrect
    // historic chain state while syncing occurs
    wait_for_cosmos_node_ready(&contact).await;

    // check if the delegate addresses are correctly configured
    check_delegate_addresses(
        &mut grpc,
        public_eth_key,
        our_cosmos_address,
        &contact.get_prefix(),
        chain_id.clone(),
    )
    .await;

    orchestrator_main_loop(
        our_cosmos_address,
        ethereum_key,
        connections.web3.unwrap(),
        connections.contact.unwrap(),
        connections.grpc.unwrap(),
        &connections.grpc_committer.unwrap(),
        contract_address,
        chain_id.clone(),
        eth_fee_calculator_url.clone(),
        &metrics_listen,
    )
    .await;
}
