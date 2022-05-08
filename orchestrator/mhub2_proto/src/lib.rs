//! This crate provides Mhub2 proto definitions in Rust and also re-exports cosmos_sdk_proto for use by downstream
//! crates. By default around a dozen proto files are generated and places into the prost folder. We could then proceed
//! to fix up all these files and use them as the required dependencies to the Mhub file, but we chose instead to replace
//! those paths with references ot upstream cosmos-sdk-proto and delete the other files. This reduces cruft in this repo even
//! if it does make for a somewhat more confusing proto generation process.

pub use cosmos_sdk_proto;
pub mod mhub2 {
    include!("../../mhub2_proto/src/prost/mhub2.v1.rs");
}
pub mod tx_committer {
    include!("../../mhub2_proto/src/prost/tx_committer.rs");
}

use bytes::BytesMut;
use prost::Message;
use prost_types::Any;

pub trait ToAny {
    fn to_any(&self) -> Option<prost_types::Any>
    where
        Self: prost::Message;
}

impl ToAny for mhub2::BatchExecutedEvent {
    fn to_any(&self) -> Option<prost_types::Any> {
        let mut buf = BytesMut::with_capacity(self.encoded_len());
        self.encode(&mut buf).expect("encoding failed");
        Some(Any {
            type_url: "/mhub2.v1.BatchExecutedEvent".into(),
            value: buf.to_vec(),
        })
    }
}

impl ToAny for mhub2::BatchTxConfirmation {
    fn to_any(&self) -> Option<prost_types::Any> {
        let mut buf = BytesMut::with_capacity(self.encoded_len());
        self.encode(&mut buf).expect("encoding failed");
        Some(Any {
            type_url: "/mhub2.v1.BatchTxConfirmation".into(),
            value: buf.to_vec(),
        })
    }
}

impl ToAny for mhub2::ContractCallExecutedEvent {
    fn to_any(&self) -> Option<prost_types::Any> {
        let mut buf = BytesMut::with_capacity(self.encoded_len());
        self.encode(&mut buf).expect("encoding failed");
        Some(Any {
            type_url: "/mhub2.v1.ContractCallExecutedEvent".into(),
            value: buf.to_vec(),
        })
    }
}

impl ToAny for mhub2::ContractCallTxConfirmation {
    fn to_any(&self) -> Option<prost_types::Any> {
        let mut buf = BytesMut::with_capacity(self.encoded_len());
        self.encode(&mut buf).expect("encoding failed");
        Some(Any {
            type_url: "/mhub2.v1.ContractCallTxConfirmation".into(),
            value: buf.to_vec(),
        })
    }
}

impl ToAny for mhub2::SendToHubEvent {
    fn to_any(&self) -> Option<prost_types::Any> {
        let mut buf = BytesMut::with_capacity(self.encoded_len());
        self.encode(&mut buf).expect("encoding failed");
        Some(Any {
            type_url: "/mhub2.v1.SendToHubEvent".into(),
            value: buf.to_vec(),
        })
    }
}

impl ToAny for mhub2::TransferToChainEvent {
    fn to_any(&self) -> Option<prost_types::Any> {
        let mut buf = BytesMut::with_capacity(self.encoded_len());
        self.encode(&mut buf).expect("encoding failed");
        Some(Any {
            type_url: "/mhub2.v1.TransferToChainEvent".into(),
            value: buf.to_vec(),
        })
    }
}

impl ToAny for mhub2::SignerSetTxExecutedEvent {
    fn to_any(&self) -> Option<prost_types::Any> {
        let mut buf = BytesMut::with_capacity(self.encoded_len());
        self.encode(&mut buf).expect("encoding failed");
        Some(Any {
            type_url: "/mhub2.v1.SignerSetTxExecutedEvent".into(),
            value: buf.to_vec(),
        })
    }
}

impl ToAny for mhub2::SignerSetTxConfirmation {
    fn to_any(&self) -> Option<prost_types::Any> {
        let mut buf = BytesMut::with_capacity(self.encoded_len());
        self.encode(&mut buf).expect("encoding failed");
        Some(Any {
            type_url: "/mhub2.v1.SignerSetTxConfirmation".into(),
            value: buf.to_vec(),
        })
    }
}
