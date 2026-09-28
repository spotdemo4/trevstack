use std::{str::FromStr, time::Duration};

use http::Uri;
use thiserror::Error;
use tonic::{
    codegen::{Body, Bytes, StdError},
    metadata::MetadataValue,
    transport::Endpoint,
};
use url::Url;

use crate::proto::{
    auth::v1::{LoginRequest, auth_service_client::AuthServiceClient},
    number::v1::{AddRequest, number_service_client::NumberServiceClient},
};

const CONNECT_TIMEOUT: Duration = Duration::from_secs(5);
const RPC_TIMEOUT: Duration = Duration::from_secs(10);
const SESSION_COOKIE: &str = "stack_session";

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
    endpoint: String,
    origin: Uri,
}

impl ServiceUrl {
    pub async fn add(
        &self,
        credentials: Credentials,
        name: String,
        number: u32,
    ) -> Result<u64, ClientError> {
        let channel = Endpoint::from_shared(self.endpoint.clone())?
            .connect_timeout(CONNECT_TIMEOUT)
            .timeout(RPC_TIMEOUT)
            .connect()
            .await?;

        authenticated_add(channel, self.origin.clone(), credentials, name, number).await
    }
}

/// Logs in to obtain a session token, then sends it as the session cookie on `Add`.
async fn authenticated_add<T>(
    service: T,
    origin: Uri,
    credentials: Credentials,
    name: String,
    number: u32,
) -> Result<u64, ClientError>
where
    T: tonic::client::GrpcService<tonic::body::Body> + Clone,
    T::Error: Into<StdError>,
    T::ResponseBody: Body<Data = Bytes> + Send + 'static,
    <T::ResponseBody as Body>::Error: Into<StdError> + Send,
{
    let session = AuthServiceClient::with_origin(service.clone(), origin.clone())
        .login(LoginRequest {
            username: Some(credentials.username),
            password: Some(credentials.password),
        })
        .await
        .map_err(ClientError::Login)?
        .into_inner();
    if session.jwt.is_empty() {
        return Err(ClientError::InvalidSession);
    }
    let cookie = MetadataValue::try_from(format!("{SESSION_COOKIE}={}", session.jwt))
        .map_err(|_| ClientError::InvalidSession)?;

    let mut request = tonic::Request::new(AddRequest {
        name: Some(name),
        number: Some(number),
    });
    request.metadata_mut().insert("cookie", cookie);
    let response = NumberServiceClient::with_origin(service, origin)
        .add(request)
        .await
        .map_err(ClientError::Add)?;

    Ok(response.into_inner().sum)
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
    #[error("Login failed with {code:?}: {message}", code = .0.code(), message = .0.message())]
    Login(tonic::Status),
    #[error("Login returned an invalid session token")]
    InvalidSession,
    #[error("Add failed with {code:?}: {message}", code = .0.code(), message = .0.message())]
    Add(tonic::Status),
}

#[cfg(test)]
mod tests {
    use std::{
        convert::Infallible,
        sync::{Arc, Mutex},
    };

    use http::{HeaderMap, Request, Response, Uri};
    use http_body_util::{BodyExt, Full};
    use prost::Message;
    use tonic::{Code, body::Body, codegen::Bytes};
    use tower::service_fn;

    use super::{ClientError, Credentials, ServiceUrl, authenticated_add};
    use crate::proto::{
        auth::v1::LoginResponse,
        number::v1::{AddRequest, AddResponse, number_service_client::NumberServiceClient},
    };

    type SeenRequests = Arc<Mutex<Vec<(String, Option<String>)>>>;

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

    #[tokio::test]
    async fn logs_in_before_adding_with_session_cookie() {
        let (seen, service) = mock_service(|path| match path {
            "/grpc/auth.v1.AuthService/Login" => grpc_ok(LoginResponse {
                jwt: "token".to_owned(),
                sub: "bob".to_owned(),
                exp: 1_893_456_000,
            }),
            "/grpc/number.v1.NumberService/Add" => grpc_ok(AddResponse { sum: 42 }),
            _ => grpc_error(Code::Unimplemented),
        });

        let sum = authenticated_add(
            service,
            "http://localhost:8080/grpc".parse().unwrap(),
            credentials(),
            "alice".to_owned(),
            7,
        )
        .await
        .unwrap();

        assert_eq!(sum, 42);
        assert_eq!(
            *seen.lock().unwrap(),
            [
                ("/grpc/auth.v1.AuthService/Login".to_owned(), None),
                (
                    "/grpc/number.v1.NumberService/Add".to_owned(),
                    Some("stack_session=token".to_owned())
                ),
            ]
        );
    }

    #[tokio::test]
    async fn does_not_add_when_login_fails() {
        let (seen, service) = mock_service(|_| grpc_error(Code::Unauthenticated));

        let error = authenticated_add(
            service,
            "http://localhost:8080/grpc".parse().unwrap(),
            credentials(),
            "alice".to_owned(),
            7,
        )
        .await
        .unwrap_err();

        assert!(
            matches!(error, ClientError::Login(status) if status.code() == Code::Unauthenticated)
        );
        assert_eq!(seen.lock().unwrap().len(), 1);
    }

    fn credentials() -> Credentials {
        Credentials {
            username: "bob".to_owned(),
            password: "password123".to_owned(),
        }
    }

    fn mock_service(
        respond: fn(&str) -> Response<Body>,
    ) -> (
        SeenRequests,
        impl tonic::client::GrpcService<Body, ResponseBody = Body, Error = Infallible> + Clone,
    ) {
        let seen = SeenRequests::default();
        let captured = Arc::clone(&seen);
        let service = service_fn(move |request: Request<Body>| {
            let path = request.uri().path().to_owned();
            let cookie = request
                .headers()
                .get("cookie")
                .map(|value| value.to_str().unwrap().to_owned());
            let response = respond(&path);
            captured.lock().unwrap().push((path, cookie));
            async { Ok::<_, Infallible>(response) }
        });
        (seen, service)
    }

    fn grpc_ok(message: impl Message) -> Response<Body> {
        let encoded = message.encode_to_vec();
        let mut frame = vec![0];
        frame.extend_from_slice(&u32::try_from(encoded.len()).unwrap().to_be_bytes());
        frame.extend_from_slice(&encoded);

        let mut trailers = HeaderMap::new();
        trailers.insert("grpc-status", "0".parse().unwrap());
        let body = Full::new(Bytes::from(frame))
            .map_err(|never| match never {})
            .with_trailers(async { Some(Ok(trailers)) });

        Response::builder()
            .status(200)
            .header("content-type", "application/grpc")
            .body(Body::new(body))
            .unwrap()
    }

    fn grpc_error(code: Code) -> Response<Body> {
        Response::builder()
            .status(200)
            .header("content-type", "application/grpc")
            .header("grpc-status", (code as i32).to_string())
            .body(Body::empty())
            .unwrap()
    }
}
