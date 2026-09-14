pub mod cli;
pub mod client;
pub mod proto;

use cli::Config;
use client::ClientError;

pub async fn run(config: Config) -> Result<u64, ClientError> {
    config.url.add(config.name, config.number).await
}
