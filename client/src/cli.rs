use clap::Parser;
use thiserror::Error;

use crate::client::{Credentials, ServiceUrl, ServiceUrlError};

const DEFAULT_URL: &str = "http://127.0.0.1:8080/grpc";
const URL_ENV: &str = "NUMBER_SERVICE_URL";
const USERNAME_ENV: &str = "NUMBER_SERVICE_USERNAME";
const PASSWORD_ENV: &str = "NUMBER_SERVICE_PASSWORD";

#[derive(Debug, Parser)]
#[command(version, about = "Send a name and number to the Number service")]
pub struct Cli {
    /// Number service base URL [env: NUMBER_SERVICE_URL]
    #[arg(long)]
    url: Option<String>,

    /// Account username [env: NUMBER_SERVICE_USERNAME]
    #[arg(long)]
    username: Option<String>,

    /// Account password; prefer the environment variable to keep it out of the
    /// process list [env: NUMBER_SERVICE_PASSWORD]
    #[arg(long)]
    password: Option<String>,

    /// Name associated with the number
    name: String,

    /// Number to add
    number: u32,
}

#[derive(Debug)]
pub struct Config {
    pub url: ServiceUrl,
    pub credentials: Credentials,
    pub name: String,
    pub number: u32,
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
    pub fn into_config(
        self,
        get_env: impl Fn(&str) -> Option<String>,
    ) -> Result<Config, ConfigError> {
        let url = self
            .url
            .or_else(|| get_env(URL_ENV))
            .unwrap_or_else(|| DEFAULT_URL.to_owned());
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
            credentials: Credentials { username, password },
            name: self.name,
            number: self.number,
        })
    }
}

#[cfg(test)]
mod tests {
    use clap::Parser;

    use super::{Cli, ConfigError, DEFAULT_URL, PASSWORD_ENV, URL_ENV, USERNAME_ENV};

    fn credentials_env(name: &str) -> Option<String> {
        match name {
            USERNAME_ENV => Some("env_user".to_owned()),
            PASSWORD_ENV => Some("env-password".to_owned()),
            _ => None,
        }
    }

    #[test]
    fn uses_default_url() {
        let config = Cli::try_parse_from(["client", "alice", "7"])
            .unwrap()
            .into_config(credentials_env)
            .unwrap();

        assert_eq!(config.url.to_string(), DEFAULT_URL);
        assert_eq!(config.name, "alice");
        assert_eq!(config.number, 7);
    }

    #[test]
    fn environment_overrides_default_url() {
        let config = Cli::try_parse_from(["client", "alice", "7"])
            .unwrap()
            .into_config(|name| match name {
                URL_ENV => Some("http://localhost:9000/rpc".to_owned()),
                _ => credentials_env(name),
            })
            .unwrap();

        assert_eq!(config.url.to_string(), "http://localhost:9000/rpc");
    }

    #[test]
    fn option_overrides_environment() {
        let config = Cli::try_parse_from([
            "client",
            "--url",
            "http://localhost:7000/grpc",
            "alice",
            "7",
        ])
        .unwrap()
        .into_config(|name| match name {
            URL_ENV => Some("http://localhost:9000/rpc".to_owned()),
            _ => credentials_env(name),
        })
        .unwrap();

        assert_eq!(config.url.to_string(), "http://localhost:7000/grpc");
    }

    #[test]
    fn reads_credentials_from_environment() {
        let config = Cli::try_parse_from(["client", "alice", "7"])
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
            "alice",
            "7",
        ])
        .unwrap()
        .into_config(credentials_env)
        .unwrap();

        assert_eq!(config.credentials.username, "cli_user");
        assert_eq!(config.credentials.password, "cli-password");
    }

    #[test]
    fn requires_credentials() {
        let missing_username = Cli::try_parse_from(["client", "--password", "secret123", "a", "7"])
            .unwrap()
            .into_config(|_| None);
        assert!(matches!(
            missing_username,
            Err(ConfigError::MissingUsername)
        ));

        let missing_password = Cli::try_parse_from(["client", "--username", "bob", "a", "7"])
            .unwrap()
            .into_config(|_| None);
        assert!(matches!(
            missing_password,
            Err(ConfigError::MissingPassword)
        ));
    }

    #[test]
    fn redacts_password_in_debug_output() {
        let config = Cli::try_parse_from(["client", "alice", "7"])
            .unwrap()
            .into_config(credentials_env)
            .unwrap();

        let debug = format!("{config:?}");
        assert!(debug.contains("env_user"));
        assert!(!debug.contains("env-password"));
    }

    #[test]
    fn rejects_invalid_number() {
        assert!(Cli::try_parse_from(["client", "alice", "not-a-number"]).is_err());
        assert!(Cli::try_parse_from(["client", "alice", "4294967296"]).is_err());
    }

    #[test]
    fn requires_exactly_one_pair() {
        assert!(Cli::try_parse_from(["client", "alice"]).is_err());
        assert!(Cli::try_parse_from(["client", "alice", "7", "extra"]).is_err());
    }
}
