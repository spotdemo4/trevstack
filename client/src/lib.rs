pub mod cli;
pub mod client;
pub mod proto;

use cli::Config;
use client::ClientError;

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
