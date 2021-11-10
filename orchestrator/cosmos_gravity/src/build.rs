use clarity::PrivateKey as EthPrivateKey;
use deep_space::private_key::PrivateKey as CosmosPrivateKey;
use deep_space::utils::bytes_to_hex_str;
use deep_space::Contact;
use deep_space::Msg;
use ethereum_gravity::utils::downcast_uint256;
use mhub2_proto::mhub2 as proto;
use mhub2_proto::ToAny;
use mhub2_utils::message_signatures::{
    encode_logic_call_confirm, encode_tx_batch_confirm, encode_valset_confirm,
};
use mhub2_utils::types::*;

pub fn signer_set_tx_confirmation_messages(
    contact: &Contact,
    ethereum_key: EthPrivateKey,
    valsets: Vec<Valset>,
    cosmos_key: CosmosPrivateKey,
    gravity_id: String,
    chain_id: String,
) -> Vec<Msg> {
    let cosmos_address = cosmos_key.to_address(&contact.get_prefix()).unwrap();
    let ethereum_address = ethereum_key.to_public_key().unwrap();

    let mut msgs = Vec::new();
    for valset in valsets {
        let data = encode_valset_confirm(gravity_id.clone(), valset.clone());
        let signature = ethereum_key.sign_ethereum_msg(&data);
        let confirmation = proto::SignerSetTxConfirmation {
            external_signer: ethereum_address.to_string(),
            signer_set_nonce: valset.nonce,
            signature: signature.to_bytes().to_vec(),
        };
        let msg = proto::MsgSubmitExternalTxConfirmation {
            signer: cosmos_address.to_string(),
            confirmation: confirmation.to_any(),
            chain_id: chain_id.clone(),
        };
        let msg = Msg::new("/mhub2.v1.MsgSubmitExternalTxConfirmation", msg);
        msgs.push(msg);
    }
    msgs
}

pub fn batch_tx_confirmation_messages(
    contact: &Contact,
    ethereum_key: EthPrivateKey,
    batches: Vec<TransactionBatch>,
    cosmos_key: CosmosPrivateKey,
    gravity_id: String,
    chain_id: String,
) -> Vec<Msg> {
    let cosmos_address = cosmos_key.to_address(&contact.get_prefix()).unwrap();
    let ethereum_address = ethereum_key.to_public_key().unwrap();

    let mut msgs = Vec::new();
    for batch in batches {
        let data = encode_tx_batch_confirm(gravity_id.clone(), batch.clone());
        let signature = ethereum_key.sign_ethereum_msg(&data);
        let confirmation = proto::BatchTxConfirmation {
            external_token_id: batch.token_contract.to_string(),
            batch_nonce: batch.nonce,
            external_signer: ethereum_address.to_string(),
            signature: signature.to_bytes().to_vec(),
        };
        let msg = proto::MsgSubmitExternalEvent {
            signer: cosmos_address.to_string(),
            event: confirmation.to_any(),
            chain_id: chain_id.clone(),
        };
        let msg = Msg::new("/mhub2.v1.MsgSubmitExternalTxConfirmation", msg);
        msgs.push(msg);
    }
    msgs
}

pub fn contract_call_tx_confirmation_messages(
    contact: &Contact,
    ethereum_key: EthPrivateKey,
    logic_calls: Vec<LogicCall>,
    cosmos_key: CosmosPrivateKey,
    gravity_id: String,
    chain_id: String,
) -> Vec<Msg> {
    let cosmos_address = cosmos_key.to_address(&contact.get_prefix()).unwrap();
    let ethereum_address = ethereum_key.to_public_key().unwrap();

    let mut msgs = Vec::new();
    for logic_call in logic_calls {
        let data = encode_logic_call_confirm(gravity_id.clone(), logic_call.clone());
        let signature = ethereum_key.sign_ethereum_msg(&data);
        let confirmation = proto::ContractCallTxConfirmation {
            external_signer: ethereum_address.to_string(),
            signature: signature.to_bytes().to_vec(),
            invalidation_scope: bytes_to_hex_str(&logic_call.invalidation_id)
                .as_bytes()
                .to_vec(),
            invalidation_nonce: logic_call.invalidation_nonce,
        };
        let msg = proto::MsgSubmitExternalTxConfirmation {
            signer: cosmos_address.to_string(),
            confirmation: confirmation.to_any(),
            chain_id: chain_id.clone(),
        };
        let msg = Msg::new("/mhub2.v1.MsgSubmitExternalTxConfirmation", msg);
        msgs.push(msg);
    }
    msgs
}

pub fn ethereum_event_messages(
    contact: &Contact,
    cosmos_key: CosmosPrivateKey,
    transfers: Vec<TransferToChainEvent>,
    batches: Vec<TransactionBatchExecutedEvent>,
    logic_calls: Vec<LogicCallExecutedEvent>,
    valsets: Vec<ValsetUpdatedEvent>,
    chain_id: String,
) -> Vec<Msg> {
    let cosmos_address = cosmos_key.to_address(&contact.get_prefix()).unwrap();

    // This sorts oracle messages by event nonce before submitting them. It's not a pretty implementation because
    // we're missing an intermediary layer of abstraction. We could implement 'EventTrait' and then implement sort
    // for it, but then when we go to transform 'EventTrait' objects into GravityMsg enum values we'll have all sorts
    // of issues extracting the inner object from the TraitObject. Likewise we could implement sort of GravityMsg but that
    // would require a truly horrendous (nearly 100 line) match statement to deal with all combinations. That match statement
    // could be reduced by adding two traits to sort against but really this is the easiest option.
    //
    // We index the events by event nonce in an unordered hashmap and then play them back in order into a vec
    let mut unordered_msgs = std::collections::HashMap::new();
    for transfer in transfers {
        let event = proto::TransferToChainEvent {
            event_nonce: downcast_uint256(transfer.event_nonce.clone()).unwrap(),
            external_height: downcast_uint256(transfer.block_height).unwrap(),
            external_coin_id: transfer.erc20.to_string(),
            amount: transfer.amount.to_string(),
            receiver_chain_id: transfer.destination_chain,
            external_receiver: transfer.destination.to_string(),
            sender: transfer.sender.to_string(),
            tx_hash: transfer.tx_hash,
            fee: transfer.fee.to_string(),
        };
        let msg = proto::MsgSubmitExternalEvent {
            signer: cosmos_address.to_string(),
            event: event.to_any(),
            chain_id: chain_id.clone(),
        };
        let msg = Msg::new("/mhub2.v1.MsgSubmitExternalEvent", msg);
        unordered_msgs.insert(transfer.event_nonce, msg);
    }
    for batch in batches {
        let event = proto::BatchExecutedEvent {
            event_nonce: downcast_uint256(batch.event_nonce.clone()).unwrap(),
            batch_nonce: downcast_uint256(batch.batch_nonce.clone()).unwrap(),
            external_height: downcast_uint256(batch.block_height).unwrap(),
            external_coin_id: batch.erc20.to_string(),
            tx_hash: batch.tx_hash,
            fee_paid: batch.fee.to_string(),
            fee_payer: batch.tx_sender,
        };
        let msg = proto::MsgSubmitExternalEvent {
            signer: cosmos_address.to_string(),
            event: event.to_any(),
            chain_id: chain_id.clone(),
        };
        let msg = Msg::new("/mhub2.v1.MsgSubmitExternalEvent", msg);
        unordered_msgs.insert(batch.event_nonce, msg);
    }
    for logic_call in logic_calls {
        let event = proto::ContractCallExecutedEvent {
            event_nonce: downcast_uint256(logic_call.event_nonce.clone()).unwrap(),
            external_height: downcast_uint256(logic_call.block_height).unwrap(),
            invalidation_scope: logic_call.invalidation_id,
            invalidation_nonce: downcast_uint256(logic_call.invalidation_nonce).unwrap(),
            tx_hash: logic_call.tx_hash,
            return_data: logic_call.return_data,
        };
        let msg = proto::MsgSubmitExternalEvent {
            signer: cosmos_address.to_string(),
            event: event.to_any(),
            chain_id: chain_id.clone(),
        };
        let msg = Msg::new("/mhub2.v1.MsgSubmitExternalEvent", msg);
        unordered_msgs.insert(logic_call.event_nonce, msg);
    }
    for valset in valsets {
        let event = proto::SignerSetTxExecutedEvent {
            event_nonce: downcast_uint256(valset.event_nonce.clone()).unwrap(),
            signer_set_tx_nonce: downcast_uint256(valset.valset_nonce.clone()).unwrap(),
            external_height: downcast_uint256(valset.block_height).unwrap(),
            members: valset.members.iter().map(|v| v.into()).collect(),
            tx_hash: valset.tx_hash,
        };
        let msg = proto::MsgSubmitExternalEvent {
            signer: cosmos_address.to_string(),
            event: event.to_any(),
            chain_id: chain_id.clone(),
        };
        let msg = Msg::new("/mhub2.v1.MsgSubmitExternalEvent", msg);
        unordered_msgs.insert(valset.event_nonce, msg);
    }

    let mut keys = Vec::new();
    for (key, _) in unordered_msgs.iter() {
        keys.push(key.clone());
    }
    keys.sort();

    let mut msgs = Vec::new();
    for i in keys.iter() {
        msgs.push(unordered_msgs.remove_entry(&i).unwrap().1);
    }

    msgs
}
