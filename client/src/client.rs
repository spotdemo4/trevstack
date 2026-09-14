use std::{str::FromStr, time::Duration};

use http::Uri;
use thiserror::Error;
use tonic::transport::Endpoint;
use url::Url;

use crate::proto::number::v1::{AddRequest, number_service_client::NumberServiceClient};

const CONNECT_TIMEOUT: Duration = Duration::from_secs(5);
const RPC_TIMEOUT: Duration = Duration::from_secs(10);

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct ServiceUrl {
    endpoint: String,
    origin: Uri,
}

impl ServiceUrl {
    pub async fn add(&self, name: String, number: u32) -> Result<u64, ClientError> {
        let channel = Endpoint::from_shared(self.endpoint.clone())?
            .connect_timeout(CONNECT_TIMEOUT)
            .timeout(RPC_TIMEOUT)
            .connect()
            .await?;
        let mut client = NumberServiceClient::with_origin(channel, self.origin.clone());
        let response = client
            .add(AddRequest {
                name: Some(name),
                number: Some(number),
            })
            .await?;

        Ok(response.into_inner().sum)
    }
}

impl FromStr for ServiceUrl {
    type Err = ServiceUrlError;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut url = Url::parse(input).map_err(ServiceUrlError::Invalid)?;
        if !matches!(url.scheme(), "http" | "https") {
            return Err(ServiceUrlError::UnsupportedScheme(url.scheme().to_owned()));
        }
        if url.host().is_none() {
            return Err(ServiceUrlError::MissingHost);
        }
        if !url.username().is_empty() || url.password().is_some() {
            return Err(ServiceUrlError::Credentials);
        }
        if url.query().is_some() {
            return Err(ServiceUrlError::Query);
        }
        if url.fragment().is_some() {
            return Err(ServiceUrlError::Fragment);
        }

        let path = url.path().trim_end_matches('/').to_owned();
        url.set_path(if path.is_empty() { "/" } else { &path });

        let endpoint = url.origin().ascii_serialization();
        let origin = url.as_str().parse().map_err(ServiceUrlError::InvalidUri)?;

        Ok(Self { endpoint, origin })
    }
}

impl std::fmt::Display for ServiceUrl {
    fn fmt(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        self.origin.fmt(formatter)
    }
}

#[derive(Debug, Error)]
pub enum ServiceUrlError {
    #[error("invalid service URL: {0}")]
    Invalid(url::ParseError),
    #[error("service URL must use http or https, not {0}")]
    UnsupportedScheme(String),
    #[error("service URL must include a host")]
    MissingHost,
    #[error("service URL must not contain credentials")]
    Credentials,
    #[error("service URL must not contain a query string")]
    Query,
    #[error("service URL must not contain a fragment")]
    Fragment,
    #[error("service URL cannot be represented as an HTTP URI: {0}")]
    InvalidUri(http::uri::InvalidUri),
}

#[derive(Debug, Error)]
pub enum ClientError {
    #[error("could not connect to the number service: {0}")]
    Transport(#[from] tonic::transport::Error),
    #[error("Add failed with {code:?}: {message}", code = .0.code(), message = .0.message())]
    Status(#[from] tonic::Status),
}

#[cfg(test)]
mod tests {
    use std::{
        convert::Infallible,
        sync::{Arc, Mutex},
    };

    use http::{Request, Response, Uri};
    use tonic::body::Body;
    use tower::service_fn;

    use super::ServiceUrl;
    use crate::proto::number::v1::{AddRequest, number_service_client::NumberServiceClient};

    #[test]
    fn normalizes_trailing_slashes() {
        let url: ServiceUrl = "http://localhost:8080/grpc///".parse().unwrap();
        assert_eq!(url.to_string(), "http://localhost:8080/grpc");

        let root: ServiceUrl = "http://localhost:8080/".parse().unwrap();
        assert_eq!(root.to_string(), "http://localhost:8080/");
    }

    #[test]
    fn rejects_unsupported_url_parts() {
        assert!("ftp://localhost/grpc".parse::<ServiceUrl>().is_err());
        assert!("http://user@localhost/grpc".parse::<ServiceUrl>().is_err());
        assert!("http://localhost/grpc?x=1".parse::<ServiceUrl>().is_err());
        assert!(
            "http://localhost/grpc#section"
                .parse::<ServiceUrl>()
                .is_err()
        );
        assert!("localhost:8080/grpc".parse::<ServiceUrl>().is_err());
    }

    #[tokio::test]
    async fn prefixes_generated_method_path() {
        assert_method_path(
            "http://localhost:8080/grpc",
            "/grpc/number.v1.NumberService/Add",
        )
        .await;
        assert_method_path("http://localhost:8080/", "/number.v1.NumberService/Add").await;
    }

    async fn assert_method_path(origin: &str, expected: &str) {
        let seen = Arc::new(Mutex::new(None));
        let captured = Arc::clone(&seen);
        let service = service_fn(move |request: Request<Body>| {
            *captured.lock().unwrap() = Some(request.uri().clone());
            async {
                Ok::<_, Infallible>(
                    Response::builder()
                        .status(200)
                        .header("content-type", "application/grpc")
                        .header("grpc-status", "12")
                        .body(Body::empty())
                        .unwrap(),
                )
            }
        });
        let origin: Uri = origin.parse().unwrap();
        let mut client = NumberServiceClient::with_origin(service, origin);

        let _ = client
            .add(AddRequest {
                name: Some("alice".to_owned()),
                number: Some(7),
            })
            .await;

        assert_eq!(seen.lock().unwrap().as_ref().unwrap().path(), expected);
    }
}
