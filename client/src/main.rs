use std::{env, process::ExitCode};

use clap::Parser;
use trevstack_client::{cli::Cli, run};

#[tokio::main]
async fn main() -> ExitCode {
    let config = match Cli::parse().into_config(|name| env::var(name).ok()) {
        Ok(config) => config,
        Err(error) => {
            eprintln!("client: {error}");
            return ExitCode::FAILURE;
        }
    };

    match run(config).await {
        Ok(sum) => {
            println!("{sum}");
            ExitCode::SUCCESS
        }
        Err(error) => {
            eprintln!("client: {error}");
            ExitCode::FAILURE
        }
    }
}
