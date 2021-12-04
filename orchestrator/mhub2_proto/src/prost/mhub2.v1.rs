/// ExternalEventVoteRecord is an event that is pending of confirmation by 2/3 of
/// the signer set. The event is then attested and executed in the state machine
/// once the required threshold is met.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExternalEventVoteRecord {
    #[prost(message, optional, tag = "1")]
    pub event: ::core::option::Option<::prost_types::Any>,
    #[prost(string, repeated, tag = "2")]
    pub votes: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    #[prost(bool, tag = "3")]
    pub accepted: bool,
}
/// LatestBlockHeight defines the latest observed external block height
/// and the corresponding timestamp value in nanoseconds.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LatestBlockHeight {
    #[prost(uint64, tag = "1")]
    pub external_height: u64,
    #[prost(uint64, tag = "2")]
    pub cosmos_height: u64,
}
/// ExternalSigner represents a cosmos validator with its corresponding bridge
/// operator address and its staking consensus power.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExternalSigner {
    #[prost(uint64, tag = "1")]
    pub power: u64,
    #[prost(string, tag = "2")]
    pub external_address: ::prost::alloc::string::String,
}
/// SignerSetTx is the Bridge multisig set that relays
/// transactions the two chains. The staking validators keep external keys which
/// are used to check signatures in order to get significant gas
/// savings.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SignerSetTx {
    #[prost(uint64, tag = "1")]
    pub nonce: u64,
    #[prost(uint64, tag = "2")]
    pub height: u64,
    #[prost(message, repeated, tag = "3")]
    pub signers: ::prost::alloc::vec::Vec<ExternalSigner>,
    #[prost(uint64, tag = "4")]
    pub sequence: u64,
}
/// BatchTx represents a batch of transactions going from Cosmos to External Chain.
/// Batch txs are are identified by a unique hash and the token contract that is
/// shared by all the SendToExternal
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BatchTx {
    #[prost(uint64, tag = "1")]
    pub batch_nonce: u64,
    #[prost(uint64, tag = "2")]
    pub timeout: u64,
    #[prost(message, repeated, tag = "3")]
    pub transactions: ::prost::alloc::vec::Vec<SendToExternal>,
    #[prost(string, tag = "4")]
    pub external_token_id: ::prost::alloc::string::String,
    #[prost(uint64, tag = "5")]
    pub height: u64,
    #[prost(uint64, tag = "6")]
    pub sequence: u64,
}
/// SendToExternal represents an individual SendToExternal from Cosmos to
/// External chain
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SendToExternal {
    #[prost(uint64, tag = "1")]
    pub id: u64,
    #[prost(string, tag = "2")]
    pub sender: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub external_recipient: ::prost::alloc::string::String,
    #[prost(string, tag = "4")]
    pub chain_id: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "5")]
    pub token: ::core::option::Option<ExternalToken>,
    #[prost(message, optional, tag = "6")]
    pub fee: ::core::option::Option<ExternalToken>,
    #[prost(string, tag = "7")]
    pub tx_hash: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "8")]
    pub val_commission: ::core::option::Option<ExternalToken>,
    #[prost(uint64, tag = "9")]
    pub created_at: u64,
    #[prost(string, tag = "10")]
    pub refund_address: ::prost::alloc::string::String,
    #[prost(string, tag = "11")]
    pub refund_chain_id: ::prost::alloc::string::String,
}
/// ContractCallTx represents an individual arbitrary logic call transaction
/// from Cosmos to External.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ContractCallTx {
    #[prost(uint64, tag = "1")]
    pub invalidation_nonce: u64,
    #[prost(bytes = "vec", tag = "2")]
    pub invalidation_scope: ::prost::alloc::vec::Vec<u8>,
    #[prost(string, tag = "3")]
    pub address: ::prost::alloc::string::String,
    #[prost(bytes = "vec", tag = "4")]
    pub payload: ::prost::alloc::vec::Vec<u8>,
    #[prost(uint64, tag = "5")]
    pub timeout: u64,
    #[prost(message, repeated, tag = "6")]
    pub tokens: ::prost::alloc::vec::Vec<ExternalToken>,
    #[prost(message, repeated, tag = "7")]
    pub fees: ::prost::alloc::vec::Vec<ExternalToken>,
    #[prost(uint64, tag = "8")]
    pub height: u64,
    #[prost(uint64, tag = "9")]
    pub sequence: u64,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExternalToken {
    #[prost(uint64, tag = "1")]
    pub token_id: u64,
    #[prost(string, tag = "2")]
    pub external_token_id: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub amount: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TokenInfo {
    #[prost(uint64, tag = "1")]
    pub id: u64,
    #[prost(string, tag = "2")]
    pub denom: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub chain_id: ::prost::alloc::string::String,
    #[prost(string, tag = "4")]
    pub external_token_id: ::prost::alloc::string::String,
    #[prost(uint64, tag = "5")]
    pub external_decimals: u64,
    #[prost(bytes = "vec", tag = "6")]
    pub commission: ::prost::alloc::vec::Vec<u8>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TokenInfos {
    #[prost(message, repeated, tag = "1")]
    pub token_infos: ::prost::alloc::vec::Vec<TokenInfo>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct IdSet {
    #[prost(uint64, repeated, tag = "1")]
    pub ids: ::prost::alloc::vec::Vec<u64>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TxStatus {
    #[prost(string, tag = "1")]
    pub in_tx_hash: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub out_tx_hash: ::prost::alloc::string::String,
    #[prost(enumeration = "TxStatusType", tag = "3")]
    pub status: i32,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ColdStorageTransferProposal {
    #[prost(string, tag = "1")]
    pub chain_id: ::prost::alloc::string::String,
    #[prost(message, repeated, tag = "2")]
    pub amount: ::prost::alloc::vec::Vec<cosmos_sdk_proto::cosmos::base::v1beta1::Coin>,
}
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash, PartialOrd, Ord, ::prost::Enumeration)]
#[repr(i32)]
pub enum TxStatusType {
    TxStatusNotFound = 0,
    TxStatusDepositReceived = 1,
    TxStatusBatchCreated = 2,
    TxStatusBatchExecuted = 3,
    TxStatusRefunded = 4,
}
/// MsgSendToExternal submits a SendToExternal attempt to bridge an asset over to
/// External chain. The SendToExternal will be stored and then included in a batch and
/// then submitted to external chain.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgSendToExternal {
    #[prost(string, tag = "1")]
    pub sender: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub external_recipient: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "3")]
    pub amount: ::core::option::Option<cosmos_sdk_proto::cosmos::base::v1beta1::Coin>,
    #[prost(message, optional, tag = "4")]
    pub bridge_fee: ::core::option::Option<cosmos_sdk_proto::cosmos::base::v1beta1::Coin>,
    #[prost(string, tag = "5")]
    pub chain_id: ::prost::alloc::string::String,
}
/// MsgSendToExternalResponse returns the SendToExternal transaction ID which
/// will be included in the batch tx.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgSendToExternalResponse {
    #[prost(uint64, tag = "1")]
    pub id: u64,
}
/// MsgCancelSendToExternal allows the sender to cancel its own outgoing
/// SendToExternal tx and recieve a refund of the tokens and bridge fees. This tx
/// will only succeed if the SendToExternal tx hasn't been batched to be
/// processed and relayed to External.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgCancelSendToExternal {
    #[prost(uint64, tag = "1")]
    pub id: u64,
    #[prost(string, tag = "2")]
    pub sender: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgCancelSendToExternalResponse {}
/// MsgRequestBatchTx requests a batch of transactions with a given coin
/// denomination to send across the bridge to external chain.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgRequestBatchTx {
    #[prost(string, tag = "1")]
    pub denom: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub signer: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgRequestBatchTxResponse {}
/// MsgSubmitExternalTxConfirmation submits an external signature for a given
/// validator
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgSubmitExternalTxConfirmation {
    /// TODO: can we make this take an array?
    #[prost(message, optional, tag = "1")]
    pub confirmation: ::core::option::Option<::prost_types::Any>,
    #[prost(string, tag = "2")]
    pub signer: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub chain_id: ::prost::alloc::string::String,
}
/// ContractCallTxConfirmation is a signature on behalf of a validator for a
/// ContractCallTx.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ContractCallTxConfirmation {
    #[prost(bytes = "vec", tag = "1")]
    pub invalidation_scope: ::prost::alloc::vec::Vec<u8>,
    #[prost(uint64, tag = "2")]
    pub invalidation_nonce: u64,
    #[prost(string, tag = "3")]
    pub external_signer: ::prost::alloc::string::String,
    #[prost(bytes = "vec", tag = "4")]
    pub signature: ::prost::alloc::vec::Vec<u8>,
}
/// BatchTxConfirmation is a signature on behalf of a validator for a BatchTx.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BatchTxConfirmation {
    #[prost(string, tag = "1")]
    pub external_token_id: ::prost::alloc::string::String,
    #[prost(uint64, tag = "2")]
    pub batch_nonce: u64,
    #[prost(string, tag = "3")]
    pub external_signer: ::prost::alloc::string::String,
    #[prost(bytes = "vec", tag = "4")]
    pub signature: ::prost::alloc::vec::Vec<u8>,
}
/// SignerSetTxConfirmation is a signature on behalf of a validator for a
/// SignerSetTx
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SignerSetTxConfirmation {
    #[prost(uint64, tag = "1")]
    pub signer_set_nonce: u64,
    #[prost(string, tag = "2")]
    pub external_signer: ::prost::alloc::string::String,
    #[prost(bytes = "vec", tag = "3")]
    pub signature: ::prost::alloc::vec::Vec<u8>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgSubmitTxConfirmationResponse {}
/// MsgSubmitExternalEvent
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgSubmitExternalEvent {
    #[prost(message, optional, tag = "1")]
    pub event: ::core::option::Option<::prost_types::Any>,
    #[prost(string, tag = "2")]
    pub signer: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgSubmitExternalEventResponse {}
/// MsgDelegateKey allows validators to delegate their voting responsibilities
/// to a given orchestrator address. This key is then used as an optional
/// authentication method for attesting events from External Chain.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgDelegateKeys {
    #[prost(string, tag = "1")]
    pub validator_address: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub orchestrator_address: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub external_address: ::prost::alloc::string::String,
    #[prost(bytes = "vec", tag = "4")]
    pub eth_signature: ::prost::alloc::vec::Vec<u8>,
    #[prost(string, tag = "5")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct MsgDelegateKeysResponse {}
/// DelegateKeysSignMsg defines the message structure an operator is expected to
/// sign when submitting a MsgDelegateKeys message. The resulting signature should
/// populate the eth_signature field.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DelegateKeysSignMsg {
    #[prost(string, tag = "1")]
    pub validator_address: ::prost::alloc::string::String,
    #[prost(uint64, tag = "2")]
    pub nonce: u64,
}
////////////
// Events //
////////////

/// SendToHubEvent is submitted when the SendToHubEvent is emitted by they
/// mhub2 contract.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SendToHubEvent {
    #[prost(uint64, tag = "1")]
    pub event_nonce: u64,
    #[prost(string, tag = "2")]
    pub external_coin_id: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub amount: ::prost::alloc::string::String,
    #[prost(string, tag = "4")]
    pub sender: ::prost::alloc::string::String,
    #[prost(string, tag = "5")]
    pub cosmos_receiver: ::prost::alloc::string::String,
    #[prost(uint64, tag = "6")]
    pub external_height: u64,
    #[prost(string, tag = "7")]
    pub tx_hash: ::prost::alloc::string::String,
}
/// TransferToChainEvent is submitted when the TransferToChainEvent is emitted by they
/// mhub2 contract.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TransferToChainEvent {
    #[prost(uint64, tag = "1")]
    pub event_nonce: u64,
    #[prost(string, tag = "2")]
    pub external_coin_id: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub amount: ::prost::alloc::string::String,
    #[prost(string, tag = "4")]
    pub fee: ::prost::alloc::string::String,
    #[prost(string, tag = "5")]
    pub sender: ::prost::alloc::string::String,
    #[prost(string, tag = "6")]
    pub receiver_chain_id: ::prost::alloc::string::String,
    #[prost(string, tag = "7")]
    pub external_receiver: ::prost::alloc::string::String,
    #[prost(uint64, tag = "8")]
    pub external_height: u64,
    #[prost(string, tag = "9")]
    pub tx_hash: ::prost::alloc::string::String,
}
/// BatchExecutedEvent claims that a batch of BatchTxExecuted operations on the
/// bridge contract was executed successfully
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BatchExecutedEvent {
    #[prost(string, tag = "1")]
    pub external_coin_id: ::prost::alloc::string::String,
    #[prost(uint64, tag = "2")]
    pub event_nonce: u64,
    #[prost(uint64, tag = "3")]
    pub external_height: u64,
    #[prost(uint64, tag = "4")]
    pub batch_nonce: u64,
    #[prost(string, tag = "5")]
    pub tx_hash: ::prost::alloc::string::String,
    #[prost(string, tag = "6")]
    pub fee_paid: ::prost::alloc::string::String,
    #[prost(string, tag = "7")]
    pub fee_payer: ::prost::alloc::string::String,
}
// ContractCallExecutedEvent describes a contract call that has been
// successfully executed on external chain.

/// NOTE: bytes.HexBytes is supposed to "help" with json encoding/decoding
/// investigate?
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ContractCallExecutedEvent {
    #[prost(uint64, tag = "1")]
    pub event_nonce: u64,
    #[prost(bytes = "vec", tag = "2")]
    pub invalidation_scope: ::prost::alloc::vec::Vec<u8>,
    #[prost(uint64, tag = "3")]
    pub invalidation_nonce: u64,
    #[prost(bytes = "vec", tag = "4")]
    pub return_data: ::prost::alloc::vec::Vec<u8>,
    #[prost(uint64, tag = "5")]
    pub external_height: u64,
    #[prost(string, tag = "6")]
    pub tx_hash: ::prost::alloc::string::String,
}
/// This informs the Cosmos module that a validator
/// set has been updated.
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SignerSetTxExecutedEvent {
    #[prost(uint64, tag = "1")]
    pub event_nonce: u64,
    #[prost(uint64, tag = "2")]
    pub signer_set_tx_nonce: u64,
    #[prost(uint64, tag = "3")]
    pub external_height: u64,
    #[prost(message, repeated, tag = "4")]
    pub members: ::prost::alloc::vec::Vec<ExternalSigner>,
    #[prost(string, tag = "5")]
    pub tx_hash: ::prost::alloc::string::String,
}
#[doc = r" Generated client implementations."]
pub mod msg_client {
    #![allow(unused_variables, dead_code, missing_docs)]
    use tonic::codegen::*;
    #[doc = " Msg defines the state transitions possible within Mhub2"]
    pub struct MsgClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl MsgClient<tonic::transport::Channel> {
        #[doc = r" Attempt to create a new client by connecting to a given endpoint."]
        pub async fn connect<D>(dst: D) -> Result<Self, tonic::transport::Error>
        where
            D: std::convert::TryInto<tonic::transport::Endpoint>,
            D::Error: Into<StdError>,
        {
            let conn = tonic::transport::Endpoint::new(dst)?.connect().await?;
            Ok(Self::new(conn))
        }
    }
    impl<T> MsgClient<T>
    where
        T: tonic::client::GrpcService<tonic::body::BoxBody>,
        T::ResponseBody: Body + HttpBody + Send + 'static,
        T::Error: Into<StdError>,
        <T::ResponseBody as HttpBody>::Error: Into<StdError> + Send,
    {
        pub fn new(inner: T) -> Self {
            let inner = tonic::client::Grpc::new(inner);
            Self { inner }
        }
        pub fn with_interceptor(inner: T, interceptor: impl Into<tonic::Interceptor>) -> Self {
            let inner = tonic::client::Grpc::with_interceptor(inner, interceptor);
            Self { inner }
        }
        pub async fn send_to_external(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgSendToExternal>,
        ) -> Result<tonic::Response<super::MsgSendToExternalResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Msg/SendToExternal");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn cancel_send_to_external(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgCancelSendToExternal>,
        ) -> Result<tonic::Response<super::MsgCancelSendToExternalResponse>, tonic::Status>
        {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Msg/CancelSendToExternal");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn request_batch_tx(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgRequestBatchTx>,
        ) -> Result<tonic::Response<super::MsgRequestBatchTxResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Msg/RequestBatchTx");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn submit_tx_confirmation(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgSubmitExternalTxConfirmation>,
        ) -> Result<tonic::Response<super::MsgSubmitTxConfirmationResponse>, tonic::Status>
        {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Msg/SubmitTxConfirmation");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn submit_external_event(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgSubmitExternalEvent>,
        ) -> Result<tonic::Response<super::MsgSubmitExternalEventResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Msg/SubmitExternalEvent");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn set_delegate_keys(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgDelegateKeys>,
        ) -> Result<tonic::Response<super::MsgDelegateKeysResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Msg/SetDelegateKeys");
            self.inner.unary(request.into_request(), path, codec).await
        }
    }
    impl<T: Clone> Clone for MsgClient<T> {
        fn clone(&self) -> Self {
            Self {
                inner: self.inner.clone(),
            }
        }
    }
    impl<T> std::fmt::Debug for MsgClient<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "MsgClient {{ ... }}")
        }
    }
}
/// Params represent the Mhub2 genesis and store parameters
/// gravity_id:
/// a random 32 byte value to prevent signature reuse, for example if the
/// cosmos validators decided to use the same Ethereum keys for another chain
/// also running Mhub2 we would not want it to be possible to play a deposit
/// from chain A back on chain B's Mhub2. This value IS USED ON ETHEREUM so
/// it must be set in your genesis.json before launch and not changed after
/// deploying Mhub2
///
/// contract_hash:
/// the code hash of a known good version of the Mhub2 contract
/// solidity code. This can be used to verify the correct version
/// of the contract has been deployed. This is a reference value for
/// goernance action only it is never read by any Mhub2 code
///
/// bridge_ethereum_address:
/// is address of the bridge contract on the Ethereum side, this is a
/// reference value for governance only and is not actually used by any
/// Mhub2 code
///
/// bridge_chain_id:
/// the unique identifier of the Ethereum chain, this is a reference value
/// only and is not actually used by any Mhub2 code
///
/// These reference values may be used by future Mhub2 client implemetnations
/// to allow for saftey features or convenience features like the Mhub2 address
/// in your relayer. A relayer would require a configured Mhub2 address if
/// governance had not set the address on the chain it was relaying for.
///
/// signed_signer_set_txs_window
/// signed_batches_window
/// signed_ethereum_signatures_window
///
/// These values represent the time in blocks that a validator has to submit
/// a signature for a batch or valset, or to submit a ethereum_signature for a
/// particular attestation nonce. In the case of attestations this clock starts
/// when the attestation is created, but only allows for slashing once the event
/// has passed
///
/// target_eth_tx_timeout:
///
/// This is the 'target' value for when ethereum transactions time out, this is a target
/// because Ethereum is a probabilistic chain and you can't say for sure what the
/// block frequency is ahead of time.
///
/// average_block_time
/// average_ethereum_block_time
///
/// These values are the average Cosmos block time and Ethereum block time
/// respectively and they are used to compute what the target batch timeout is. It
/// is important that governance updates these in case of any major, prolonged
/// change in the time it takes to produce a block
///
/// slash_fraction_signer_set_tx
/// slash_fraction_batch
/// slash_fraction_ethereum_signature
/// slash_fraction_conflicting_ethereum_signature
///
/// The slashing fractions for the various Mhub2 related slashing conditions.
/// The first three refer to not submitting a particular message, the third for
/// submitting a different ethereum_signature for the same Ethereum event
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct Params {
    #[prost(string, tag = "1")]
    pub gravity_id: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub contract_source_hash: ::prost::alloc::string::String,
    #[prost(string, tag = "4")]
    pub bridge_ethereum_address: ::prost::alloc::string::String,
    #[prost(uint64, tag = "5")]
    pub bridge_chain_id: u64,
    #[prost(uint64, tag = "6")]
    pub signed_signer_set_txs_window: u64,
    #[prost(uint64, tag = "7")]
    pub signed_batches_window: u64,
    #[prost(uint64, tag = "8")]
    pub ethereum_signatures_window: u64,
    #[prost(uint64, tag = "10")]
    pub target_eth_tx_timeout: u64,
    #[prost(uint64, tag = "11")]
    pub average_block_time: u64,
    #[prost(uint64, tag = "12")]
    pub average_ethereum_block_time: u64,
    #[prost(uint64, tag = "13")]
    pub average_bsc_block_time: u64,
    /// TODO: slash fraction for contract call txs too
    #[prost(bytes = "vec", tag = "14")]
    pub slash_fraction_signer_set_tx: ::prost::alloc::vec::Vec<u8>,
    #[prost(bytes = "vec", tag = "15")]
    pub slash_fraction_batch: ::prost::alloc::vec::Vec<u8>,
    #[prost(bytes = "vec", tag = "16")]
    pub slash_fraction_ethereum_signature: ::prost::alloc::vec::Vec<u8>,
    #[prost(bytes = "vec", tag = "17")]
    pub slash_fraction_conflicting_ethereum_signature: ::prost::alloc::vec::Vec<u8>,
    #[prost(uint64, tag = "18")]
    pub unbond_slashing_signer_set_txs_window: u64,
    #[prost(string, repeated, tag = "19")]
    pub chains: ::prost::alloc::vec::Vec<::prost::alloc::string::String>,
    #[prost(uint64, tag = "20")]
    pub outgoing_tx_timeout: u64,
}
/// GenesisState struct
/// TODO: this need to be audited and potentially simplified using the new
/// interfaces
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct GenesisState {
    #[prost(message, optional, tag = "1")]
    pub params: ::core::option::Option<Params>,
    #[prost(message, repeated, tag = "5")]
    pub external_states: ::prost::alloc::vec::Vec<ExternalState>,
    #[prost(message, optional, tag = "6")]
    pub token_infos: ::core::option::Option<TokenInfos>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExternalState {
    #[prost(string, tag = "1")]
    pub chain_id: ::prost::alloc::string::String,
    #[prost(message, repeated, tag = "2")]
    pub external_event_vote_records: ::prost::alloc::vec::Vec<ExternalEventVoteRecord>,
    #[prost(message, repeated, tag = "3")]
    pub delegate_keys: ::prost::alloc::vec::Vec<MsgDelegateKeys>,
    #[prost(message, repeated, tag = "4")]
    pub unbatched_send_to_external_txs: ::prost::alloc::vec::Vec<SendToExternal>,
    #[prost(uint64, tag = "5")]
    pub last_observed_event_nonce: u64,
    #[prost(message, repeated, tag = "6")]
    pub outgoing_txs: ::prost::alloc::vec::Vec<::prost_types::Any>,
    #[prost(message, repeated, tag = "7")]
    pub confirmations: ::prost::alloc::vec::Vec<::prost_types::Any>,
    #[prost(uint64, tag = "8")]
    pub sequence: u64,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TokenInfosRequest {}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TokenInfosResponse {
    #[prost(message, optional, tag = "1")]
    pub list: ::core::option::Option<TokenInfos>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TransactionStatusRequest {
    #[prost(string, tag = "1")]
    pub tx_hash: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct TransactionStatusResponse {
    #[prost(message, optional, tag = "1")]
    pub status: ::core::option::Option<TxStatus>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DiscountForHolderRequest {
    #[prost(string, tag = "1")]
    pub address: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DiscountForHolderResponse {
    #[prost(bytes = "vec", tag = "1")]
    pub discount: ::prost::alloc::vec::Vec<u8>,
}
///  rpc Params
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ParamsRequest {}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ParamsResponse {
    #[prost(message, optional, tag = "1")]
    pub params: ::core::option::Option<Params>,
}
///  rpc SignerSetTx
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SignerSetTxRequest {
    #[prost(uint64, tag = "1")]
    pub signer_set_nonce: u64,
    #[prost(string, tag = "2")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LatestSignerSetTxRequest {
    #[prost(string, tag = "1")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LastObservedSignerSetTxRequest {
    #[prost(string, tag = "1")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SignerSetTxResponse {
    #[prost(message, optional, tag = "1")]
    pub signer_set: ::core::option::Option<SignerSetTx>,
}
///  rpc BatchTx
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BatchTxRequest {
    #[prost(string, tag = "1")]
    pub external_token_id: ::prost::alloc::string::String,
    #[prost(uint64, tag = "2")]
    pub batch_nonce: u64,
    #[prost(string, tag = "3")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BatchTxResponse {
    #[prost(message, optional, tag = "1")]
    pub batch: ::core::option::Option<BatchTx>,
}
///  rpc ContractCallTx
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ContractCallTxRequest {
    #[prost(bytes = "vec", tag = "1")]
    pub invalidation_scope: ::prost::alloc::vec::Vec<u8>,
    #[prost(uint64, tag = "2")]
    pub invalidation_nonce: u64,
    #[prost(string, tag = "3")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ContractCallTxResponse {
    #[prost(message, optional, tag = "1")]
    pub logic_call: ::core::option::Option<ContractCallTx>,
}
/// rpc SignerSetTxConfirmations
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SignerSetTxConfirmationsRequest {
    #[prost(uint64, tag = "1")]
    pub signer_set_nonce: u64,
    #[prost(string, tag = "2")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SignerSetTxConfirmationsResponse {
    #[prost(message, repeated, tag = "1")]
    pub signatures: ::prost::alloc::vec::Vec<SignerSetTxConfirmation>,
}
///  rpc SignerSetTxs
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SignerSetTxsRequest {
    #[prost(message, optional, tag = "1")]
    pub pagination: ::core::option::Option<cosmos_sdk_proto::cosmos::base::query::v1beta1::PageRequest>,
    #[prost(string, tag = "2")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct SignerSetTxsResponse {
    #[prost(message, repeated, tag = "1")]
    pub signer_sets: ::prost::alloc::vec::Vec<SignerSetTx>,
    #[prost(message, optional, tag = "2")]
    pub pagination:
        ::core::option::Option<cosmos_sdk_proto::cosmos::base::query::v1beta1::PageResponse>,
}
///  rpc BatchTxs
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BatchTxsRequest {
    #[prost(message, optional, tag = "1")]
    pub pagination: ::core::option::Option<cosmos_sdk_proto::cosmos::base::query::v1beta1::PageRequest>,
    #[prost(string, tag = "2")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BatchTxsResponse {
    #[prost(message, repeated, tag = "1")]
    pub batches: ::prost::alloc::vec::Vec<BatchTx>,
    #[prost(message, optional, tag = "2")]
    pub pagination:
        ::core::option::Option<cosmos_sdk_proto::cosmos::base::query::v1beta1::PageResponse>,
}
///  rpc ContractCallTxs
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ContractCallTxsRequest {
    #[prost(message, optional, tag = "1")]
    pub pagination: ::core::option::Option<cosmos_sdk_proto::cosmos::base::query::v1beta1::PageRequest>,
    #[prost(string, tag = "2")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ContractCallTxsResponse {
    #[prost(message, repeated, tag = "1")]
    pub calls: ::prost::alloc::vec::Vec<ContractCallTx>,
    #[prost(message, optional, tag = "2")]
    pub pagination:
        ::core::option::Option<cosmos_sdk_proto::cosmos::base::query::v1beta1::PageResponse>,
}
// NOTE(levi) pending queries: this is my address; what do I need to sign??
// why orchestrator key? hot, signing thing all the time so validator key can be
// safer

/// rpc UnsignedSignerSetTxs
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UnsignedSignerSetTxsRequest {
    /// NOTE: this is an sdk.AccAddress and can represent either the
    /// orchestartor address or the cooresponding validator address
    #[prost(string, tag = "1")]
    pub address: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UnsignedSignerSetTxsResponse {
    #[prost(message, repeated, tag = "1")]
    pub signer_sets: ::prost::alloc::vec::Vec<SignerSetTx>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UnsignedBatchTxsRequest {
    /// NOTE: this is an sdk.AccAddress and can represent either the
    /// orchestrator address or the cooresponding validator address
    #[prost(string, tag = "1")]
    pub address: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UnsignedBatchTxsResponse {
    /// Note these are returned with the signature empty
    #[prost(message, repeated, tag = "1")]
    pub batches: ::prost::alloc::vec::Vec<BatchTx>,
}
///  rpc UnsignedContractCallTxs
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UnsignedContractCallTxsRequest {
    #[prost(string, tag = "1")]
    pub address: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UnsignedContractCallTxsResponse {
    #[prost(message, repeated, tag = "1")]
    pub calls: ::prost::alloc::vec::Vec<ContractCallTx>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BatchTxFeesRequest {
    #[prost(string, tag = "1")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BatchTxFeesResponse {
    #[prost(message, repeated, tag = "1")]
    pub fees: ::prost::alloc::vec::Vec<cosmos_sdk_proto::cosmos::base::v1beta1::Coin>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ContractCallTxConfirmationsRequest {
    #[prost(bytes = "vec", tag = "1")]
    pub invalidation_scope: ::prost::alloc::vec::Vec<u8>,
    #[prost(uint64, tag = "2")]
    pub invalidation_nonce: u64,
    #[prost(string, tag = "3")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ContractCallTxConfirmationsResponse {
    #[prost(message, repeated, tag = "1")]
    pub signatures: ::prost::alloc::vec::Vec<ContractCallTxConfirmation>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BatchTxConfirmationsRequest {
    #[prost(uint64, tag = "1")]
    pub batch_nonce: u64,
    #[prost(string, tag = "2")]
    pub external_token_id: ::prost::alloc::string::String,
    #[prost(string, tag = "3")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BatchTxConfirmationsResponse {
    #[prost(message, repeated, tag = "1")]
    pub signatures: ::prost::alloc::vec::Vec<BatchTxConfirmation>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LastSubmittedExternalEventRequest {
    #[prost(string, tag = "1")]
    pub address: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct LastSubmittedExternalEventResponse {
    #[prost(uint64, tag = "1")]
    pub event_nonce: u64,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExternalIdToDenomRequest {
    #[prost(string, tag = "1")]
    pub external_id: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct ExternalIdToDenomResponse {
    #[prost(string, tag = "1")]
    pub denom: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DenomToExternalIdRequest {
    #[prost(string, tag = "1")]
    pub denom: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DenomToExternalIdResponse {
    #[prost(string, tag = "1")]
    pub external_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DelegateKeysByValidatorRequest {
    #[prost(string, tag = "1")]
    pub validator_address: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DelegateKeysByValidatorResponse {
    #[prost(string, tag = "1")]
    pub eth_address: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub orchestrator_address: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DelegateKeysByExternalSignerRequest {
    #[prost(string, tag = "1")]
    pub external_signer: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DelegateKeysByExternalSignerResponse {
    #[prost(string, tag = "1")]
    pub validator_address: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub orchestrator_address: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DelegateKeysByOrchestratorRequest {
    #[prost(string, tag = "1")]
    pub orchestrator_address: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DelegateKeysByOrchestratorResponse {
    #[prost(string, tag = "1")]
    pub validator_address: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub external_signer: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DelegateKeysRequest {
    #[prost(string, tag = "1")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct DelegateKeysResponse {
    #[prost(message, repeated, tag = "1")]
    pub delegate_keys: ::prost::alloc::vec::Vec<MsgDelegateKeys>,
}
/// NOTE: if there is no sender address, return all
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BatchedSendToExternalsRequest {
    #[prost(string, tag = "1")]
    pub sender_address: ::prost::alloc::string::String,
    /// todo: figure out how to paginate given n Batches with m Send To Externals
    ///  cosmos.base.query.v1beta1.PageRequest pagination = 2;
    #[prost(string, tag = "2")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct BatchedSendToExternalsResponse {
    #[prost(message, repeated, tag = "1")]
    pub send_to_externals: ::prost::alloc::vec::Vec<SendToExternal>,
    ///  cosmos.base.query.v1beta1.PageResponse pagination = 2;
    #[prost(string, tag = "2")]
    pub chain_id: ::prost::alloc::string::String,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UnbatchedSendToExternalsRequest {
    #[prost(string, tag = "1")]
    pub sender_address: ::prost::alloc::string::String,
    #[prost(string, tag = "2")]
    pub chain_id: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "3")]
    pub pagination: ::core::option::Option<cosmos_sdk_proto::cosmos::base::query::v1beta1::PageRequest>,
}
#[derive(Clone, PartialEq, ::prost::Message)]
pub struct UnbatchedSendToExternalsResponse {
    #[prost(message, repeated, tag = "1")]
    pub send_to_externals: ::prost::alloc::vec::Vec<SendToExternal>,
    #[prost(string, tag = "2")]
    pub chain_id: ::prost::alloc::string::String,
    #[prost(message, optional, tag = "3")]
    pub pagination:
        ::core::option::Option<cosmos_sdk_proto::cosmos::base::query::v1beta1::PageResponse>,
}
#[doc = r" Generated client implementations."]
pub mod query_client {
    #![allow(unused_variables, dead_code, missing_docs)]
    use tonic::codegen::*;
    #[doc = " Query defines the gRPC querier service"]
    pub struct QueryClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl QueryClient<tonic::transport::Channel> {
        #[doc = r" Attempt to create a new client by connecting to a given endpoint."]
        pub async fn connect<D>(dst: D) -> Result<Self, tonic::transport::Error>
        where
            D: std::convert::TryInto<tonic::transport::Endpoint>,
            D::Error: Into<StdError>,
        {
            let conn = tonic::transport::Endpoint::new(dst)?.connect().await?;
            Ok(Self::new(conn))
        }
    }
    impl<T> QueryClient<T>
    where
        T: tonic::client::GrpcService<tonic::body::BoxBody>,
        T::ResponseBody: Body + HttpBody + Send + 'static,
        T::Error: Into<StdError>,
        <T::ResponseBody as HttpBody>::Error: Into<StdError> + Send,
    {
        pub fn new(inner: T) -> Self {
            let inner = tonic::client::Grpc::new(inner);
            Self { inner }
        }
        pub fn with_interceptor(inner: T, interceptor: impl Into<tonic::Interceptor>) -> Self {
            let inner = tonic::client::Grpc::with_interceptor(inner, interceptor);
            Self { inner }
        }
        #[doc = " Module parameters query"]
        pub async fn params(
            &mut self,
            request: impl tonic::IntoRequest<super::ParamsRequest>,
        ) -> Result<tonic::Response<super::ParamsResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Query/Params");
            self.inner.unary(request.into_request(), path, codec).await
        }
        #[doc = " get info on individual outgoing data"]
        pub async fn signer_set_tx(
            &mut self,
            request: impl tonic::IntoRequest<super::SignerSetTxRequest>,
        ) -> Result<tonic::Response<super::SignerSetTxResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Query/SignerSetTx");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn latest_signer_set_tx(
            &mut self,
            request: impl tonic::IntoRequest<super::LatestSignerSetTxRequest>,
        ) -> Result<tonic::Response<super::SignerSetTxResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Query/LatestSignerSetTx");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn last_observed_signer_set_tx(
            &mut self,
            request: impl tonic::IntoRequest<super::LastObservedSignerSetTxRequest>,
        ) -> Result<tonic::Response<super::SignerSetTxResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path =
                http::uri::PathAndQuery::from_static("/mhub2.v1.Query/LastObservedSignerSetTx");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn batch_tx(
            &mut self,
            request: impl tonic::IntoRequest<super::BatchTxRequest>,
        ) -> Result<tonic::Response<super::BatchTxResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Query/BatchTx");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn contract_call_tx(
            &mut self,
            request: impl tonic::IntoRequest<super::ContractCallTxRequest>,
        ) -> Result<tonic::Response<super::ContractCallTxResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Query/ContractCallTx");
            self.inner.unary(request.into_request(), path, codec).await
        }
        #[doc = " get collections of outgoing traffic from the bridge"]
        pub async fn signer_set_txs(
            &mut self,
            request: impl tonic::IntoRequest<super::SignerSetTxsRequest>,
        ) -> Result<tonic::Response<super::SignerSetTxsResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Query/SignerSetTxs");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn batch_txs(
            &mut self,
            request: impl tonic::IntoRequest<super::BatchTxsRequest>,
        ) -> Result<tonic::Response<super::BatchTxsResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Query/BatchTxs");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn contract_call_txs(
            &mut self,
            request: impl tonic::IntoRequest<super::ContractCallTxsRequest>,
        ) -> Result<tonic::Response<super::ContractCallTxsResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Query/ContractCallTxs");
            self.inner.unary(request.into_request(), path, codec).await
        }
        #[doc = " TODO: can/should we group these into one endpoint?"]
        pub async fn signer_set_tx_confirmations(
            &mut self,
            request: impl tonic::IntoRequest<super::SignerSetTxConfirmationsRequest>,
        ) -> Result<tonic::Response<super::SignerSetTxConfirmationsResponse>, tonic::Status>
        {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path =
                http::uri::PathAndQuery::from_static("/mhub2.v1.Query/SignerSetTxConfirmations");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn batch_tx_confirmations(
            &mut self,
            request: impl tonic::IntoRequest<super::BatchTxConfirmationsRequest>,
        ) -> Result<tonic::Response<super::BatchTxConfirmationsResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Query/BatchTxConfirmations");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn contract_call_tx_confirmations(
            &mut self,
            request: impl tonic::IntoRequest<super::ContractCallTxConfirmationsRequest>,
        ) -> Result<tonic::Response<super::ContractCallTxConfirmationsResponse>, tonic::Status>
        {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path =
                http::uri::PathAndQuery::from_static("/mhub2.v1.Query/ContractCallTxConfirmations");
            self.inner.unary(request.into_request(), path, codec).await
        }
        #[doc = " pending external signature queries for orchestrators to figure out which"]
        #[doc = " signatures they are missing"]
        #[doc = " TODO: can/should we group this into one endpoint?"]
        pub async fn unsigned_signer_set_txs(
            &mut self,
            request: impl tonic::IntoRequest<super::UnsignedSignerSetTxsRequest>,
        ) -> Result<tonic::Response<super::UnsignedSignerSetTxsResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Query/UnsignedSignerSetTxs");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn unsigned_batch_txs(
            &mut self,
            request: impl tonic::IntoRequest<super::UnsignedBatchTxsRequest>,
        ) -> Result<tonic::Response<super::UnsignedBatchTxsResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Query/UnsignedBatchTxs");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn unsigned_contract_call_txs(
            &mut self,
            request: impl tonic::IntoRequest<super::UnsignedContractCallTxsRequest>,
        ) -> Result<tonic::Response<super::UnsignedContractCallTxsResponse>, tonic::Status>
        {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path =
                http::uri::PathAndQuery::from_static("/mhub2.v1.Query/UnsignedContractCallTxs");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn last_submitted_external_event(
            &mut self,
            request: impl tonic::IntoRequest<super::LastSubmittedExternalEventRequest>,
        ) -> Result<tonic::Response<super::LastSubmittedExternalEventResponse>, tonic::Status>
        {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path =
                http::uri::PathAndQuery::from_static("/mhub2.v1.Query/LastSubmittedExternalEvent");
            self.inner.unary(request.into_request(), path, codec).await
        }
        #[doc = " Queries the fees for all pending batches, results are returned in sdk.Coin"]
        #[doc = " (fee_amount_int)(contract_address) style"]
        pub async fn batch_tx_fees(
            &mut self,
            request: impl tonic::IntoRequest<super::BatchTxFeesRequest>,
        ) -> Result<tonic::Response<super::BatchTxFeesResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Query/BatchTxFees");
            self.inner.unary(request.into_request(), path, codec).await
        }
        #[doc = " Query for info about denoms tracked by mhub2"]
        pub async fn external_id_to_denom(
            &mut self,
            request: impl tonic::IntoRequest<super::ExternalIdToDenomRequest>,
        ) -> Result<tonic::Response<super::ExternalIdToDenomResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Query/ExternalIdToDenom");
            self.inner.unary(request.into_request(), path, codec).await
        }
        #[doc = " Query for info about denoms tracked by mhub2"]
        pub async fn denom_to_external_id(
            &mut self,
            request: impl tonic::IntoRequest<super::DenomToExternalIdRequest>,
        ) -> Result<tonic::Response<super::DenomToExternalIdResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Query/DenomToExternalId");
            self.inner.unary(request.into_request(), path, codec).await
        }
        #[doc = " Query for batch send to externals"]
        pub async fn batched_send_to_externals(
            &mut self,
            request: impl tonic::IntoRequest<super::BatchedSendToExternalsRequest>,
        ) -> Result<tonic::Response<super::BatchedSendToExternalsResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path =
                http::uri::PathAndQuery::from_static("/mhub2.v1.Query/BatchedSendToExternals");
            self.inner.unary(request.into_request(), path, codec).await
        }
        #[doc = " Query for unbatched send to externals"]
        pub async fn unbatched_send_to_externals(
            &mut self,
            request: impl tonic::IntoRequest<super::UnbatchedSendToExternalsRequest>,
        ) -> Result<tonic::Response<super::UnbatchedSendToExternalsResponse>, tonic::Status>
        {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path =
                http::uri::PathAndQuery::from_static("/mhub2.v1.Query/UnbatchedSendToExternals");
            self.inner.unary(request.into_request(), path, codec).await
        }
        #[doc = " delegate keys"]
        pub async fn delegate_keys_by_validator(
            &mut self,
            request: impl tonic::IntoRequest<super::DelegateKeysByValidatorRequest>,
        ) -> Result<tonic::Response<super::DelegateKeysByValidatorResponse>, tonic::Status>
        {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path =
                http::uri::PathAndQuery::from_static("/mhub2.v1.Query/DelegateKeysByValidator");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn delegate_keys_by_external_signer(
            &mut self,
            request: impl tonic::IntoRequest<super::DelegateKeysByExternalSignerRequest>,
        ) -> Result<tonic::Response<super::DelegateKeysByExternalSignerResponse>, tonic::Status>
        {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/mhub2.v1.Query/DelegateKeysByExternalSigner",
            );
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn delegate_keys_by_orchestrator(
            &mut self,
            request: impl tonic::IntoRequest<super::DelegateKeysByOrchestratorRequest>,
        ) -> Result<tonic::Response<super::DelegateKeysByOrchestratorResponse>, tonic::Status>
        {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path =
                http::uri::PathAndQuery::from_static("/mhub2.v1.Query/DelegateKeysByOrchestrator");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn delegate_keys(
            &mut self,
            request: impl tonic::IntoRequest<super::DelegateKeysRequest>,
        ) -> Result<tonic::Response<super::DelegateKeysResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Query/DelegateKeys");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn token_infos(
            &mut self,
            request: impl tonic::IntoRequest<super::TokenInfosRequest>,
        ) -> Result<tonic::Response<super::TokenInfosResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Query/TokenInfos");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn transaction_status(
            &mut self,
            request: impl tonic::IntoRequest<super::TransactionStatusRequest>,
        ) -> Result<tonic::Response<super::TransactionStatusResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Query/TransactionStatus");
            self.inner.unary(request.into_request(), path, codec).await
        }
        pub async fn discount_for_holder(
            &mut self,
            request: impl tonic::IntoRequest<super::DiscountForHolderRequest>,
        ) -> Result<tonic::Response<super::DiscountForHolderResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::new(
                    tonic::Code::Unknown,
                    format!("Service was not ready: {}", e.into()),
                )
            })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/mhub2.v1.Query/DiscountForHolder");
            self.inner.unary(request.into_request(), path, codec).await
        }
    }
    impl<T: Clone> Clone for QueryClient<T> {
        fn clone(&self) -> Self {
            Self {
                inner: self.inner.clone(),
            }
        }
    }
    impl<T> std::fmt::Debug for QueryClient<T> {
        fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
            write!(f, "QueryClient {{ ... }}")
        }
    }
}
