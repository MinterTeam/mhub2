use mhub2_proto::mhub2::query_client::QueryClient as GravityQueryClient;
use mhub2_utils::{error::GravityError, types::Valset};
use tonic::transport::Channel;

/// This function finds the latest valset on the Gravity contract by looking back through the event
/// history and finding the most recent ValsetUpdatedEvent. Most of the time this will be very fast
/// as the latest update will be in recent blockchain history and the search moves from the present
/// backwards in time. In the case that the validator set has not been updated for a very long time
/// this will take longer.
pub async fn find_latest_valset(
    grpc_client: &mut GravityQueryClient<Channel>,
    chain_id: String,
) -> Result<Valset, GravityError> {
    let cosmos_chain_valset =
        cosmos_gravity::query::get_last_observed_valset(grpc_client, chain_id.clone()).await?;

    if cosmos_chain_valset.is_none() {
        return Err(GravityError::ValsetNotFoundError);
    }

    Ok(cosmos_chain_valset.unwrap())
}
