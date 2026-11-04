use std::{path::Path, str::FromStr, time::Duration};

use quinn::rustls::pki_types::{CertificateDer, pem::PemObject};
use thiserror::Error;
use trevrpc::{
    Status,
    client::{Channel, RpcTransport},
};
use url::Url;

use crate::proto::{
    auth::v1::{AuthServiceClient, LoginRequest},
    number::v1::{AddRequest, NumberServiceClient},
};

const CONNECT_TIMEOUT: Duration = Duration::from_secs(5);
const RPC_TIMEOUT: Duration = Duration::from_secs(10);
const DEFAULT_PORT: u16 = 443;
const AUTHORIZATION_METADATA: &str = "authorization";

#[derive(Clone, Eq, PartialEq)]
pub struct Credentials {
    pub username: String,
    pub password: String,
}

impl std::fmt::Debug for Credentials {
    fn fmt(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        formatter
            .debug_struct("Credentials")
            .field("username", &self.username)
            .field("password", &"<redacted>")
            .finish()
    }
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct ServiceUrl {
    host: String,
    port: u16,
}

impl ServiceUrl {
    pub async fn add(
        &self,
        ca_cert: Option<&Path>,
        credentials: Credentials,
        name: String,
        number: u32,
    ) -> Result<u64, ClientError> {
        // The channel verifies the server against the bundled webpki roots
        // and tries QUIC first, then QMux over TLS when UDP is blocked.
        let mut builder = Channel::builder(&self.to_string())
            .map_err(ClientError::Connect)?
            .connect_timeout(CONNECT_TIMEOUT);
        if let Some(path) = ca_cert {
            builder = builder.add_root_certificates(read_certificates(path)?);
        }
        let channel = builder.connect().await.map_err(ClientError::Connect)?;

        let result = authenticated_add(&channel, credentials, name, number).await;
        channel.close();
        channel.wait_idle().await;
        result
    }
}

/// Reads the PEM certificates in `path`, to trust in addition to the bundled
/// roots.
fn read_certificates(path: &Path) -> Result<Vec<CertificateDer<'static>>, ClientError> {
    let certs = CertificateDer::pem_file_iter(path)
        .and_then(Iterator::collect::<Result<Vec<_>, _>>)
        .map_err(|error| ClientError::CaCert(error.to_string()))?;
    if certs.is_empty() {
        return Err(ClientError::CaCert("no certificates found".to_owned()));
    }
    Ok(certs)
}

/// Logs in to obtain a session token, then sends it as a bearer token on `Add`.
async fn authenticated_add<T>(
    transport: &T,
    credentials: Credentials,
    name: String,
    number: u32,
) -> Result<u64, ClientError>
where
    T: RpcTransport + Clone,
{
    let session = AuthServiceClient::new(transport.clone())
        .login(&LoginRequest {
            username: Some(credentials.username),
            password: Some(credentials.password),
        })
        .timeout(RPC_TIMEOUT)
        .await
        .map_err(|error| ClientError::Login(error.into_status()))?;
    if session.jwt.is_empty() {
        return Err(ClientError::InvalidSession);
    }

    let response = NumberServiceClient::new(transport.clone())
        .add(&AddRequest {
            name: Some(name),
            number: Some(number),
        })
        .timeout(RPC_TIMEOUT)
        .metadata(AUTHORIZATION_METADATA, format!("Bearer {}", session.jwt))
        .await
        .map_err(|error| ClientError::Add(error.into_status()))?;

    Ok(response.sum)
}

impl FromStr for ServiceUrl {
    type Err = ServiceUrlError;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let url = Url::parse(input).map_err(ServiceUrlError::Invalid)?;
        if url.scheme() != "https" {
            return Err(ServiceUrlError::UnsupportedScheme(url.scheme().to_owned()));
        }
        let host = match url.host() {
            Some(url::Host::Domain(domain)) => domain.to_owned(),
            Some(url::Host::Ipv4(address)) => address.to_string(),
            Some(url::Host::Ipv6(address)) => address.to_string(),
            None => return Err(ServiceUrlError::MissingHost),
        };
        if !url.username().is_empty() || url.password().is_some() {
            return Err(ServiceUrlError::Credentials);
        }
        if url.path() != "/" {
            return Err(ServiceUrlError::Path);
        }
        if url.query().is_some() {
            return Err(ServiceUrlError::Query);
        }
        if url.fragment().is_some() {
            return Err(ServiceUrlError::Fragment);
        }

        Ok(Self {
            host,
            port: url.port().unwrap_or(DEFAULT_PORT),
        })
    }
}

impl std::fmt::Display for ServiceUrl {
    fn fmt(&self, formatter: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        if self.host.contains(':') {
            write!(formatter, "https://[{}]:{}", self.host, self.port)
        } else {
            write!(formatter, "https://{}:{}", self.host, self.port)
        }
    }
}

#[derive(Debug, Error)]
pub enum ServiceUrlError {
    #[error("invalid service URL: {0}")]
    Invalid(url::ParseError),
    #[error("service URL must use https, not {0}")]
    UnsupportedScheme(String),
    #[error("service URL must include a host")]
    MissingHost,
    #[error("service URL must not contain credentials")]
    Credentials,
    #[error("service URL must not contain a path")]
    Path,
    #[error("service URL must not contain a query string")]
    Query,
    #[error("service URL must not contain a fragment")]
    Fragment,
}

#[derive(Debug, Error)]
pub enum ClientError {
    #[error("could not read the CA certificate: {0}")]
    CaCert(String),
    #[error("could not connect to the number service: {0}")]
    Connect(trevrpc::Error),
    #[error("Login failed with {0}")]
    Login(Status),
    #[error("Login returned an invalid session token")]
    InvalidSession,
    #[error("Add failed with {0}")]
    Add(Status),
}

#[cfg(test)]
mod tests {
    use std::sync::{Arc, Mutex};

    use bytes::Bytes;
    use prost::Message;
    use trevrpc::{Code, ResponseEnvelope, Status, client::RpcTransport, wire::RpcRequest};

    use super::{ClientError, Credentials, ServiceUrl, authenticated_add};
    use crate::proto::{auth::v1::LoginResponse, number::v1::AddResponse};

    type SeenRequests = Arc<Mutex<Vec<(String, Option<String>)>>>;

    #[derive(Clone)]
    struct MockTransport {
        seen: SeenRequests,
        respond: fn(&str) -> Result<Vec<u8>, Status>,
    }

    #[trevrpc::async_trait]
    impl RpcTransport for MockTransport {
        async fn call(
            &self,
            request: RpcRequest,
            _body: Vec<u8>,
        ) -> trevrpc::Result<ResponseEnvelope<Bytes>> {
            let authorization = request
                .metadata
                .get("authorization")
                .map(|value| String::from_utf8(value.clone()).unwrap());
            let response = (self.respond)(&request.method);
            self.seen
                .lock()
                .unwrap()
                .push((request.method, authorization));
            Ok(ResponseEnvelope::new(Bytes::from(response?)))
        }
    }

    fn mock_transport(respond: fn(&str) -> Result<Vec<u8>, Status>) -> MockTransport {
        MockTransport {
            seen: SeenRequests::default(),
            respond,
        }
    }

    #[test]
    fn parses_service_urls() {
        let url: ServiceUrl = "https://localhost:8080".parse().unwrap();
        assert_eq!(url.to_string(), "https://localhost:8080");

        let default_port: ServiceUrl = "https://example.com/".parse().unwrap();
        assert_eq!(default_port.to_string(), "https://example.com:443");

        let ipv6: ServiceUrl = "https://[::1]:8080".parse().unwrap();
        assert_eq!(ipv6.to_string(), "https://[::1]:8080");
    }

    #[test]
    fn rejects_unsupported_url_parts() {
        assert!("http://localhost:8080".parse::<ServiceUrl>().is_err());
        assert!("ftp://localhost".parse::<ServiceUrl>().is_err());
        assert!("https://user@localhost".parse::<ServiceUrl>().is_err());
        assert!("https://localhost/grpc".parse::<ServiceUrl>().is_err());
        assert!("https://localhost/?x=1".parse::<ServiceUrl>().is_err());
        assert!("https://localhost/#section".parse::<ServiceUrl>().is_err());
        assert!("localhost:8080".parse::<ServiceUrl>().is_err());
    }

    #[tokio::test]
    async fn logs_in_before_adding_with_bearer_token() {
        let transport = mock_transport(|method| match method {
            "/auth.v1.AuthService/Login" => Ok(LoginResponse {
                jwt: "token".to_owned(),
                sub: "bob".to_owned(),
                exp: 1_893_456_000,
            }
            .encode_to_vec()),
            "/number.v1.NumberService/Add" => Ok(AddResponse { sum: 42 }.encode_to_vec()),
            _ => Err(Status::unimplemented("unknown method")),
        });

        let sum = authenticated_add(&transport, credentials(), "alice".to_owned(), 7)
            .await
            .unwrap();

        assert_eq!(sum, 42);
        assert_eq!(
            *transport.seen.lock().unwrap(),
            [
                ("/auth.v1.AuthService/Login".to_owned(), None),
                (
                    "/number.v1.NumberService/Add".to_owned(),
                    Some("Bearer token".to_owned())
                ),
            ]
        );
    }

    #[tokio::test]
    async fn does_not_add_when_login_fails() {
        let transport = mock_transport(|_| Err(Status::unauthenticated("invalid credentials")));

        let error = authenticated_add(&transport, credentials(), "alice".to_owned(), 7)
            .await
            .unwrap_err();

        assert!(
            matches!(error, ClientError::Login(status) if status.code() == Code::Unauthenticated)
        );
        assert_eq!(transport.seen.lock().unwrap().len(), 1);
    }

    #[tokio::test]
    async fn rejects_empty_session_token() {
        let transport = mock_transport(|_| Ok(LoginResponse::default().encode_to_vec()));

        let error = authenticated_add(&transport, credentials(), "alice".to_owned(), 7)
            .await
            .unwrap_err();

        assert!(matches!(error, ClientError::InvalidSession));
        assert_eq!(transport.seen.lock().unwrap().len(), 1);
    }

    fn credentials() -> Credentials {
        Credentials {
            username: "bob".to_owned(),
            password: "password123".to_owned(),
        }
    }
}
