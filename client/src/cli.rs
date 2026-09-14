use clap::Parser;

use crate::client::{ServiceUrl, ServiceUrlError};

const DEFAULT_URL: &str = "http://127.0.0.1:8080/grpc";
const URL_ENV: &str = "NUMBER_SERVICE_URL";

#[derive(Debug, Parser)]
#[command(version, about = "Send a name and number to the Number service")]
pub struct Cli {
    /// Number service base URL
    #[arg(long)]
    url: Option<String>,

    /// Name associated with the number
    name: String,

    /// Number to add
    number: u32,
}

#[derive(Debug)]
pub struct Config {
    pub url: ServiceUrl,
    pub name: String,
    pub number: u32,
}

impl Cli {
    pub fn into_config(
        self,
        get_env: impl FnOnce(&str) -> Option<String>,
    ) -> Result<Config, ServiceUrlError> {
        let url = self
            .url
            .or_else(|| get_env(URL_ENV))
            .unwrap_or_else(|| DEFAULT_URL.to_owned());

        Ok(Config {
            url: url.parse()?,
            name: self.name,
            number: self.number,
        })
    }
}

#[cfg(test)]
mod tests {
    use clap::Parser;

    use super::{Cli, DEFAULT_URL, URL_ENV};

    #[test]
    fn uses_default_url() {
        let config = Cli::try_parse_from(["client", "alice", "7"])
            .unwrap()
            .into_config(|_| None)
            .unwrap();

        assert_eq!(config.url.to_string(), DEFAULT_URL);
        assert_eq!(config.name, "alice");
        assert_eq!(config.number, 7);
    }

    #[test]
    fn environment_overrides_default_url() {
        let config = Cli::try_parse_from(["client", "alice", "7"])
            .unwrap()
            .into_config(|name| {
                assert_eq!(name, URL_ENV);
                Some("http://localhost:9000/rpc".to_owned())
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
        .into_config(|_| Some("http://localhost:9000/rpc".to_owned()))
        .unwrap();

        assert_eq!(config.url.to_string(), "http://localhost:7000/grpc");
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
