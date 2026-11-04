pub mod cli;
pub mod client;
pub mod proto;

use cli::Config;
use client::ClientError;

/// Sends the configured name and number to the server, returning the new sum.
///
/// # Errors
///
/// Returns the errors of [`ServiceUrl::add`](client::ServiceUrl::add).
pub async fn run(config: Config) -> Result<u64, ClientError> {
    config
        .url
        .add(
            config.ca_cert.as_deref(),
            config.credentials,
            config.name,
            config.number,
        )
        .await
}
