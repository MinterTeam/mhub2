#![warn(clippy::all)]
#![allow(clippy::pedantic)]

extern crate serde;
extern crate serde_json;
#[macro_use]
extern crate serde_derive;
extern crate awc;
extern crate clarity;
extern crate futures;
extern crate num256;
extern crate tokio;
#[macro_use]
extern crate log;
#[macro_use]
extern crate lazy_static;

pub mod amm;
pub mod client;
mod erc20_utils;
mod event_utils;
pub mod jsonrpc;
pub mod types;

pub use event_utils::address_to_event;
