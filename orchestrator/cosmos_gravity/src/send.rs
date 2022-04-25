use bytes::BytesMut;
use clarity::Address as EthAddress;
use clarity::PrivateKey as EthPrivateKey;
use deep_space::address::Address;
use deep_space::error::CosmosGrpcError;
use deep_space::Contact;
use deep_space::Msg;
use mhub2_proto::mhub2 as proto;
use mhub2_proto::tx_committer::tx_committer_client::TxCommitterClient;
use mhub2_proto::tx_committer::{CommitTxReply, CommitTxRequest};
use prost::Message;
use prost_types::Any;
use std::time::Duration;
use tonic::transport::Channel;

pub const MEMO: &str = "Sent using Althea Orchestrator";
pub const TIMEOUT: Duration = Duration::from_secs(60);

/// Send a transaction updating the eth address for the sending
/// Cosmos address. The sending Cosmos address should be a validator
pub async fn update_gravity_delegate_addresses(
    tx_committer: &TxCommitterClient<Channel>,
    contact: &Contact,
    our_valoper_address: Address,
    delegate_eth_address: EthAddress,
    delegate_cosmos_address: Address,
    ethereum_key: EthPrivateKey,
    chain_id: String,
) -> Result<CommitTxReply, CosmosGrpcError> {
    let nonce = contact
        .get_account_info(our_valoper_address.into())
        .await?
        .sequence;

    let eth_sign_msg = proto::DelegateKeysSignMsg {
        validator_address: our_valoper_address.to_string().clone(),
        nonce,
    };

    let mut data = BytesMut::with_capacity(eth_sign_msg.encoded_len());
    Message::encode(&eth_sign_msg, &mut data).expect("encoding failed");

    let eth_signature = ethereum_key.sign_ethereum_msg(&data).to_bytes().to_vec();
    let msg = proto::MsgDelegateKeys {
        validator_address: our_valoper_address.to_string(),
        orchestrator_address: delegate_cosmos_address.to_string(),
        external_address: delegate_eth_address.to_string(),
        eth_signature,
        chain_id,
    };
    let msg = Msg::new("/mhub2.v1.MsgDelegateKeys", msg);
    send_messages(tx_committer, vec![msg]).await
}

pub async fn send_messages(
    tx_committer: &TxCommitterClient<Channel>,
    messages: Vec<Msg>,
) -> Result<CommitTxReply, CosmosGrpcError> {
    let response = tx_committer
        .clone()
        .commit_tx(CommitTxRequest {
            msgs: messages
                .iter()
                .map(|msg| {
                    let any_msg: Any = msg.clone().into();
                    let mut buf = Vec::new();
                    any_msg.encode(&mut buf).unwrap();
                    buf
                })
                .collect(),
        })
        .await?;

    Ok(response.into_inner())
}

pub async fn send_main_loop(
    client: &TxCommitterClient<Channel>,
    mut rx: tokio::sync::mpsc::Receiver<Vec<Msg>>,
) {
    while let Some(messages) = rx.recv().await {
        match send_messages(client, messages).await {
            Ok(res) => trace!("okay: {:?}", res),
            Err(err) => {
                if !err.to_string().contains("account sequence mismatch") {
                    error!("fail: {}", err)
                }
            }
        }
    }
}
