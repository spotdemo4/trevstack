use std::path::PathBuf;

use clap::{Parser, Subcommand};
use thiserror::Error;

use crate::client::{Credentials, ServiceUrl, ServiceUrlError};

const DEFAULT_URL: &str = "https://localhost:8080";
const URL_ENV: &str = "STACK_URL";
const CA_CERT_ENV: &str = "STACK_CA_CERT";
const USERNAME_ENV: &str = "STACK_USERNAME";
const PASSWORD_ENV: &str = "STACK_PASSWORD";

// the doc comments are the --help text, where backticks would be shown as-is
#[allow(clippy::doc_markdown)]
#[derive(Debug, Parser)]
#[command(
    version,
    about = "Send, stream and watch cursor positions on the Cursor service"
)]
pub struct Cli {
    /// Server URL [env: STACK_URL]
    #[arg(long, global = true)]
    url: Option<String>,

    /// PEM file with certificates to trust in addition to the bundled roots
    /// [env: STACK_CA_CERT]
    #[arg(long, global = true)]
    ca_cert: Option<PathBuf>,

    /// Account username [env: STACK_USERNAME]
    #[arg(long, global = true)]
    username: Option<String>,

    /// Account password; prefer the environment variable to keep it out of the
    /// process list [env: STACK_PASSWORD]
    #[arg(long, global = true)]
    password: Option<String>,

    #[command(subcommand)]
    command: Command,
}

// the doc comments are the --help text, where backticks would be shown as-is
#[allow(clippy::doc_markdown)]
#[derive(Clone, Copy, Debug, PartialEq, Subcommand)]
pub enum Command {
    /// Record one position and print its stream ID
    Send {
        /// Horizontal position, from 0 (left edge) to 1 (right edge)
        #[arg(value_parser = parse_coordinate)]
        x: f64,
        /// Vertical position, from 0 (top edge) to 1 (bottom edge)
        #[arg(value_parser = parse_coordinate)]
        y: f64,
    },
    /// Show every user's cursor as it moves; prints one line per event when
    /// stdout is not a terminal
    Watch,
    /// Stream your mouse position over the terminal, then print the stream's
    /// summary; reads "x y" lines from stdin when not run in a terminal
    Track,
    /// Stream your mouse position over the terminal while showing everyone
    /// else's; reads "x y" lines from stdin and prints one line per event when
    /// not run in a terminal
    Share,
}

/// Parses a coordinate: a finite number from 0 to 1.
///
/// # Errors
///
/// Returns an error if `input` is not a number from 0 to 1.
pub fn parse_coordinate(input: &str) -> Result<f64, String> {
    let value: f64 = input
        .trim()
        .parse()
        .map_err(|_| format!("{input:?} is not a number"))?;
    if (0.0..=1.0).contains(&value) {
        Ok(value)
    } else {
        Err(format!("{input} is not from 0 to 1"))
    }
}

#[derive(Debug)]
pub struct Config {
    pub url: ServiceUrl,
    pub ca_cert: Option<PathBuf>,
    pub credentials: Credentials,
    pub command: Command,
}

#[derive(Debug, Error)]
pub enum ConfigError {
    #[error(transparent)]
    Url(#[from] ServiceUrlError),
    #[error("a username is required; pass --username or set {USERNAME_ENV}")]
    MissingUsername,
    #[error("a password is required; pass --password or set {PASSWORD_ENV}")]
    MissingPassword,
}

impl Cli {
    /// Builds the configuration, falling back to the environment variables
    /// read by `get_env` for options that weren't passed.
    ///
    /// # Errors
    ///
    /// Returns an error if the URL is invalid, or the username or password is
    /// missing.
    pub fn into_config(
        self,
        get_env: impl Fn(&str) -> Option<String>,
    ) -> Result<Config, ConfigError> {
        let url = self
            .url
            .or_else(|| get_env(URL_ENV))
            .unwrap_or_else(|| DEFAULT_URL.to_owned());
        let ca_cert = self
            .ca_cert
            .or_else(|| get_env(CA_CERT_ENV).map(PathBuf::from));
        let username = self
            .username
            .or_else(|| get_env(USERNAME_ENV))
            .ok_or(ConfigError::MissingUsername)?;
        let password = self
            .password
            .or_else(|| get_env(PASSWORD_ENV))
            .ok_or(ConfigError::MissingPassword)?;

        Ok(Config {
            url: url.parse()?,
            ca_cert,
            credentials: Credentials { username, password },
            command: self.command,
        })
    }
}

#[cfg(test)]
mod tests {
    use clap::Parser;

    use std::path::Path;

    use super::{
        CA_CERT_ENV, Cli, Command, ConfigError, DEFAULT_URL, PASSWORD_ENV, URL_ENV, USERNAME_ENV,
        parse_coordinate,
    };

    fn credentials_env(name: &str) -> Option<String> {
        match name {
            USERNAME_ENV => Some("env_user".to_owned()),
            PASSWORD_ENV => Some("env-password".to_owned()),
            _ => None,
        }
    }

    #[test]
    fn uses_default_url() {
        let config = Cli::try_parse_from(["client", "send", "0.25", "0.75"])
            .unwrap()
            .into_config(credentials_env)
            .unwrap();

        assert_eq!(config.url.to_string(), DEFAULT_URL);
        assert_eq!(config.ca_cert, None);
        assert_eq!(config.command, Command::Send { x: 0.25, y: 0.75 });
    }

    #[test]
    fn parses_streaming_subcommands() {
        for (name, command) in [
            ("watch", Command::Watch),
            ("track", Command::Track),
            ("share", Command::Share),
        ] {
            let config = Cli::try_parse_from(["client", name])
                .unwrap()
                .into_config(credentials_env)
                .unwrap();
            assert_eq!(config.command, command);
        }
    }

    #[test]
    fn environment_overrides_default_url() {
        let config = Cli::try_parse_from(["client", "watch"])
            .unwrap()
            .into_config(|name| match name {
                URL_ENV => Some("https://localhost:9000".to_owned()),
                _ => credentials_env(name),
            })
            .unwrap();

        assert_eq!(config.url.to_string(), "https://localhost:9000");
    }

    #[test]
    fn option_overrides_environment() {
        // global options may come before or after the subcommand
        for args in [
            ["client", "--url", "https://localhost:7000", "watch"],
            ["client", "watch", "--url", "https://localhost:7000"],
        ] {
            let config = Cli::try_parse_from(args)
                .unwrap()
                .into_config(|name| match name {
                    URL_ENV => Some("https://localhost:9000".to_owned()),
                    _ => credentials_env(name),
                })
                .unwrap();

            assert_eq!(config.url.to_string(), "https://localhost:7000");
        }
    }

    #[test]
    fn reads_ca_cert_from_environment_and_option() {
        let from_env = Cli::try_parse_from(["client", "watch"])
            .unwrap()
            .into_config(|name| match name {
                CA_CERT_ENV => Some("env.pem".to_owned()),
                _ => credentials_env(name),
            })
            .unwrap();
        assert_eq!(from_env.ca_cert.as_deref(), Some(Path::new("env.pem")));

        let from_option = Cli::try_parse_from(["client", "--ca-cert", "cli.pem", "watch"])
            .unwrap()
            .into_config(|name| match name {
                CA_CERT_ENV => Some("env.pem".to_owned()),
                _ => credentials_env(name),
            })
            .unwrap();
        assert_eq!(from_option.ca_cert.as_deref(), Some(Path::new("cli.pem")));
    }

    #[test]
    fn reads_credentials_from_environment() {
        let config = Cli::try_parse_from(["client", "watch"])
            .unwrap()
            .into_config(credentials_env)
            .unwrap();

        assert_eq!(config.credentials.username, "env_user");
        assert_eq!(config.credentials.password, "env-password");
    }

    #[test]
    fn credential_options_override_environment() {
        let config = Cli::try_parse_from([
            "client",
            "--username",
            "cli_user",
            "--password",
            "cli-password",
            "watch",
        ])
        .unwrap()
        .into_config(credentials_env)
        .unwrap();

        assert_eq!(config.credentials.username, "cli_user");
        assert_eq!(config.credentials.password, "cli-password");
    }

    #[test]
    fn requires_credentials() {
        let missing_username = Cli::try_parse_from(["client", "--password", "secret123", "watch"])
            .unwrap()
            .into_config(|_| None);
        assert!(matches!(
            missing_username,
            Err(ConfigError::MissingUsername)
        ));

        let missing_password = Cli::try_parse_from(["client", "--username", "bob", "watch"])
            .unwrap()
            .into_config(|_| None);
        assert!(matches!(
            missing_password,
            Err(ConfigError::MissingPassword)
        ));
    }

    #[test]
    fn redacts_password_in_debug_output() {
        let config = Cli::try_parse_from(["client", "watch"])
            .unwrap()
            .into_config(credentials_env)
            .unwrap();

        let debug = format!("{config:?}");
        assert!(debug.contains("env_user"));
        assert!(!debug.contains("env-password"));
    }

    #[test]
    fn parses_coordinates() {
        assert_eq!(parse_coordinate("0"), Ok(0.0));
        assert_eq!(parse_coordinate("1"), Ok(1.0));
        assert_eq!(parse_coordinate("0.5"), Ok(0.5));
        assert!(parse_coordinate("-0.1").is_err());
        assert!(parse_coordinate("1.01").is_err());
        assert!(parse_coordinate("NaN").is_err());
        assert!(parse_coordinate("inf").is_err());
        assert!(parse_coordinate("left").is_err());
    }

    #[test]
    fn rejects_invalid_send_arguments() {
        assert!(Cli::try_parse_from(["client", "send", "0.5", "1.5"]).is_err());
        assert!(Cli::try_parse_from(["client", "send", "0.5"]).is_err());
        assert!(Cli::try_parse_from(["client", "send", "0.5", "0.5", "0.5"]).is_err());
        // a negative coordinate is out of range, not an unknown flag that hides it
        assert!(Cli::try_parse_from(["client", "send", "--", "-0.5", "0.5"]).is_err());
    }

    #[test]
    fn requires_a_subcommand() {
        assert!(Cli::try_parse_from(["client"]).is_err());
        assert!(Cli::try_parse_from(["client", "add", "alice", "7"]).is_err());
    }
}
