use std::{future::Future, path::Path, str::FromStr, time::Duration};

use quinn::rustls::pki_types::{CertificateDer, pem::PemObject};
use thiserror::Error;
use trevrpc::{
    Status,
    client::{Channel, RpcTransport},
};
use url::Url;

use crate::proto::{
    auth::v1::{AuthServiceClient, LoginRequest, LogoutRequest},
    cursor::v1::{CursorServiceClient, SendRequest},
};

const CONNECT_TIMEOUT: Duration = Duration::from_secs(5);
const RPC_TIMEOUT: Duration = Duration::from_secs(10);
const DEFAULT_PORT: u16 = 443;
pub(crate) const AUTHORIZATION_METADATA: &str = "authorization";

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
    /// Connects to the server, trusting the certificates in `ca_cert` in
    /// addition to the bundled roots.
    ///
    /// # Errors
    ///
    /// Returns an error if the CA certificates can't be read or the connection
    /// fails.
    pub async fn connect(&self, ca_cert: Option<&Path>) -> Result<Channel, ClientError> {
        // The channel verifies the server against the bundled webpki roots
        // and tries QUIC first, then QMux over TLS when UDP is blocked.
        let mut builder = Channel::builder(&self.to_string())
            .map_err(ClientError::Connect)?
            .connect_timeout(CONNECT_TIMEOUT);
        if let Some(path) = ca_cert {
            builder = builder.add_root_certificates(read_certificates(path)?);
        }
        builder.connect().await.map_err(ClientError::Connect)
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

/// Logs in to obtain a session token, runs `run` with the `Bearer <token>`
/// authorization to send as metadata, then ends the session, since the token
/// is not kept. The session ends whether `run` succeeds or fails.
///
/// # Errors
///
/// Returns an error if logging in fails, or the error of `run`.
pub async fn with_session<T, F, Fut, R>(
    transport: &T,
    credentials: Credentials,
    run: F,
) -> Result<R, ClientError>
where
    T: RpcTransport + Clone,
    F: FnOnce(String) -> Fut,
    Fut: Future<Output = Result<R, ClientError>>,
{
    let session = AuthServiceClient::new(transport.clone())
        .login(&LoginRequest {
            username: Some(credentials.username),
            password: Some(credentials.password),
        })
        .timeout(RPC_TIMEOUT)
        .await
        .map_err(|error| ClientError::Login(error.into_status()))?;
    if session.token.is_empty() {
        return Err(ClientError::InvalidSession);
    }
    let authorization = format!("Bearer {}", session.token);

    let result = run(authorization.clone()).await;

    // Without this the session would stay valid until it expires. A failure
    // only leaves it to expire, so it does not change the result.
    let _ = AuthServiceClient::new(transport.clone())
        .logout(&LogoutRequest {})
        .timeout(RPC_TIMEOUT)
        .metadata(AUTHORIZATION_METADATA, authorization)
        .await;

    result
}

/// Records the position `(x, y)` with `Send`, returning its stream ID.
///
/// # Errors
///
/// Returns an error if the server rejects the request.
pub async fn send<T>(
    client: &CursorServiceClient<T>,
    authorization: &str,
    x: f64,
    y: f64,
) -> Result<i64, ClientError>
where
    T: RpcTransport,
{
    client
        .send(&SendRequest {
            x: Some(x),
            y: Some(y),
        })
        .timeout(RPC_TIMEOUT)
        .metadata(AUTHORIZATION_METADATA, authorization.to_owned())
        .await
        .map(|response| response.cursor.map_or(0, |cursor| cursor.stream_id))
        .map_err(|error| ClientError::Send(error.into_status()))
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
    #[error("could not connect to the cursor service: {0}")]
    Connect(trevrpc::Error),
    #[error("Login failed with {0}")]
    Login(Status),
    #[error("Login returned an invalid session token")]
    InvalidSession,
    #[error("Send failed with {0}")]
    Send(Status),
    #[error("Watch failed with {0}")]
    Watch(Status),
    #[error("Track failed with {0}")]
    Track(Status),
    #[error("Share failed with {0}")]
    Share(Status),
    #[error("invalid input: {0}")]
    Input(String),
    #[error("could not read stdin: {0}")]
    Stdin(std::io::Error),
    #[error("could not write stdout: {0}")]
    Output(std::io::Error),
    #[error("terminal error: {0}")]
    Terminal(std::io::Error),
}

#[cfg(test)]
pub(crate) mod tests {
    use std::sync::{Arc, Mutex};

    use bytes::Bytes;
    use prost::Message;
    use trevrpc::{Code, ResponseEnvelope, Status, client::RpcTransport, wire::RpcRequest};

    use super::{ClientError, Credentials, ServiceUrl, send, with_session};
    use crate::proto::{
        auth::v1::{LoginResponse, LogoutResponse},
        cursor::v1::{Cursor, CursorServiceClient, SendRequest, SendResponse},
    };

    /// The method, authorization metadata and body of each request a mock
    /// transport saw.
    pub type SeenRequests = Arc<Mutex<Vec<(String, Option<String>, Vec<u8>)>>>;

    #[derive(Clone)]
    pub struct MockTransport {
        pub seen: SeenRequests,
        pub respond: fn(&str) -> Result<Vec<u8>, Status>,
    }

    pub fn authorization(request: &RpcRequest) -> Option<String> {
        request
            .metadata
            .get("authorization")
            .map(|value| String::from_utf8(value.clone()).unwrap())
    }

    #[trevrpc::async_trait]
    impl RpcTransport for MockTransport {
        async fn call(
            &self,
            request: RpcRequest,
            body: Vec<u8>,
        ) -> trevrpc::Result<ResponseEnvelope<Bytes>> {
            let response = (self.respond)(&request.method);
            self.seen
                .lock()
                .unwrap()
                .push((request.method.clone(), authorization(&request), body));
            Ok(ResponseEnvelope::new(Bytes::from(response?)))
        }
    }

    pub fn mock_transport(respond: fn(&str) -> Result<Vec<u8>, Status>) -> MockTransport {
        MockTransport {
            seen: SeenRequests::default(),
            respond,
        }
    }

    impl MockTransport {
        /// Returns the method and authorization of each request.
        pub fn calls(&self) -> Vec<(String, Option<String>)> {
            self.seen
                .lock()
                .unwrap()
                .iter()
                .map(|(method, authorization, _)| (method.clone(), authorization.clone()))
                .collect()
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

    pub fn login() -> Vec<u8> {
        LoginResponse {
            token: "token".to_owned(),
            sub: "bob".to_owned(),
            exp: 1_893_456_000,
        }
        .encode_to_vec()
    }

    fn authenticated_calls() -> [(String, Option<String>); 3] {
        [
            ("/auth.v1.AuthService/Login".to_owned(), None),
            (
                "/cursor.v1.CursorService/Send".to_owned(),
                Some("Bearer token".to_owned()),
            ),
            (
                "/auth.v1.AuthService/Logout".to_owned(),
                Some("Bearer token".to_owned()),
            ),
        ]
    }

    fn sent() -> Vec<u8> {
        SendResponse {
            cursor: Some(Cursor {
                stream_id: 42,
                ..Cursor::default()
            }),
        }
        .encode_to_vec()
    }

    async fn send_in_session(transport: &MockTransport) -> Result<i64, ClientError> {
        let client = CursorServiceClient::new(transport.clone());
        with_session(transport, credentials(), |authorization| async move {
            send(&client, &authorization, 0.25, 0.75).await
        })
        .await
    }

    #[tokio::test]
    async fn sends_with_bearer_token_between_login_and_logout() {
        let transport = mock_transport(|method| match method {
            "/auth.v1.AuthService/Login" => Ok(login()),
            "/cursor.v1.CursorService/Send" => Ok(sent()),
            "/auth.v1.AuthService/Logout" => Ok(LogoutResponse {}.encode_to_vec()),
            _ => Err(Status::unimplemented("unknown method")),
        });

        let stream_id = send_in_session(&transport).await.unwrap();

        assert_eq!(stream_id, 42);
        assert_eq!(transport.calls(), authenticated_calls());
        let seen = transport.seen.lock().unwrap();
        let request = SendRequest::decode(seen[1].2.as_slice()).unwrap();
        assert_eq!(
            request,
            SendRequest {
                x: Some(0.25),
                y: Some(0.75)
            }
        );
    }

    #[tokio::test]
    async fn logs_out_when_send_fails() {
        let transport = mock_transport(|method| match method {
            "/auth.v1.AuthService/Login" => Ok(login()),
            "/auth.v1.AuthService/Logout" => Ok(LogoutResponse {}.encode_to_vec()),
            _ => Err(Status::invalid_argument("x out of range")),
        });

        let error = send_in_session(&transport).await.unwrap_err();

        assert!(
            matches!(error, ClientError::Send(status) if status.code() == Code::InvalidArgument)
        );
        assert_eq!(transport.calls(), authenticated_calls());
    }

    #[tokio::test]
    async fn ignores_logout_failures() {
        let transport = mock_transport(|method| match method {
            "/auth.v1.AuthService/Login" => Ok(login()),
            "/cursor.v1.CursorService/Send" => Ok(sent()),
            _ => Err(Status::unavailable("server went away")),
        });

        let stream_id = send_in_session(&transport).await.unwrap();

        assert_eq!(stream_id, 42);
        assert_eq!(transport.calls(), authenticated_calls());
    }

    #[tokio::test]
    async fn does_not_send_when_login_fails() {
        let transport = mock_transport(|_| Err(Status::unauthenticated("invalid credentials")));

        let error = send_in_session(&transport).await.unwrap_err();

        assert!(
            matches!(error, ClientError::Login(status) if status.code() == Code::Unauthenticated)
        );
        assert_eq!(transport.calls().len(), 1);
    }

    #[tokio::test]
    async fn rejects_empty_session_token() {
        let transport = mock_transport(|_| Ok(LoginResponse::default().encode_to_vec()));

        let error = send_in_session(&transport).await.unwrap_err();

        assert!(matches!(error, ClientError::InvalidSession));
        assert_eq!(transport.calls().len(), 1);
    }

    fn credentials() -> Credentials {
        Credentials {
            username: "bob".to_owned(),
            password: "password123".to_owned(),
        }
    }
}
