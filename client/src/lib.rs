pub mod cli;
pub mod client;
pub mod event;
pub mod pad;
pub mod proto;
pub mod stream;

use std::io::{self, BufRead, IsTerminal};

use cli::{Command, Config};
use client::{ClientError, send, with_session};
use event::format_summary;
use proto::cursor::v1::CursorServiceClient;
use stream::Lines;
use tokio::sync::mpsc;

/// Logs in, runs the configured command, and logs out, printing the
/// command's output to stdout.
///
/// # Errors
///
/// Returns the errors of connecting, logging in, and the command.
pub async fn run(config: Config) -> Result<(), ClientError> {
    let channel = config.url.connect(config.ca_cert.as_ref()).await?;
    let client = CursorServiceClient::new(channel.clone());
    let username = config.credentials.username.clone();
    let command = config.command;

    let result = with_session(&channel, config.credentials, |authorization| async move {
        run_command(&client, &authorization, &username, command).await
    })
    .await;

    channel.close();
    channel.wait_idle().await;
    result
}

async fn run_command<T>(
    client: &CursorServiceClient<T>,
    authorization: &str,
    username: &str,
    command: Command,
) -> Result<(), ClientError>
where
    T: trevrpc::client::StreamingRpcTransport,
{
    let mut out = io::stdout();
    // the pad needs a terminal to draw on, and track and share one to read the mouse from
    let interactive = match command {
        Command::Send { .. } => false,
        Command::Watch => out.is_terminal(),
        Command::Track | Command::Share => out.is_terminal() && io::stdin().is_terminal(),
    };

    match command {
        Command::Send { x, y } => {
            let stream_id = send(client, authorization, x, y).await?;
            println!("{stream_id}");
        }
        Command::Watch if interactive => stream::watch_pad(client, authorization, username).await?,
        Command::Watch => stream::watch_lines(client, authorization, &mut out, ctrl_c()).await?,
        Command::Track => {
            let response = if interactive {
                stream::track_pad(client, authorization, username).await?
            } else {
                stream::track_lines(client, authorization, &mut stdin_lines(), ctrl_c()).await?
            };
            println!("{}", format_summary(&response));
        }
        Command::Share if interactive => stream::share_pad(client, authorization, username).await?,
        Command::Share => {
            stream::share_lines(
                client,
                authorization,
                &mut stdin_lines(),
                &mut out,
                ctrl_c(),
            )
            .await?;
        }
    }
    Ok(())
}

/// Completes on Ctrl-C, or never if it can't be caught.
async fn ctrl_c() {
    if tokio::signal::ctrl_c().await.is_err() {
        std::future::pending::<()>().await;
    }
}

/// Reads stdin's lines on a thread of their own, which, unlike the runtime's
/// blocking threads, doesn't keep the client from exiting while it waits for
/// input.
fn stdin_lines() -> Lines {
    let (sender, receiver) = mpsc::unbounded_channel();
    std::thread::spawn(move || {
        for line in io::stdin().lock().lines() {
            let failed = line.is_err();
            if sender.send(line).is_err() || failed {
                return;
            }
        }
    });
    receiver
}
