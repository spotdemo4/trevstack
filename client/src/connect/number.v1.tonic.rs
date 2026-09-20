// @generated
/// Generated client implementations.
pub mod number_service_client {
    #![allow(
        unused_variables,
        dead_code,
        missing_docs,
        clippy::wildcard_imports,
        clippy::let_unit_value
    )]
    use tonic::codegen::http::Uri;
    use tonic::codegen::*;
    ///
    #[derive(Debug, Clone)]
    pub struct NumberServiceClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl NumberServiceClient<tonic::transport::Channel> {
        /// Attempt to create a new client by connecting to a given endpoint.
        pub async fn connect<D>(dst: D) -> Result<Self, tonic::transport::Error>
        where
            D: TryInto<tonic::transport::Endpoint>,
            D::Error: Into<StdError>,
        {
            let conn = tonic::transport::Endpoint::new(dst)?.connect().await?;
            Ok(Self::new(conn))
        }
    }
    impl<T> NumberServiceClient<T>
    where
        T: tonic::client::GrpcService<tonic::body::Body>,
        T::Error: Into<StdError>,
        T::ResponseBody: Body<Data = Bytes> + std::marker::Send + 'static,
        <T::ResponseBody as Body>::Error: Into<StdError> + std::marker::Send,
    {
        pub fn new(inner: T) -> Self {
            let inner = tonic::client::Grpc::new(inner);
            Self { inner }
        }
        pub fn with_origin(inner: T, origin: Uri) -> Self {
            let inner = tonic::client::Grpc::with_origin(inner, origin);
            Self { inner }
        }
        pub fn with_interceptor<F>(
            inner: T,
            interceptor: F,
        ) -> NumberServiceClient<InterceptedService<T, F>>
        where
            F: tonic::service::Interceptor,
            T::ResponseBody: Default,
            T: tonic::codegen::Service<
                    http::Request<tonic::body::Body>,
                    Response = http::Response<
                        <T as tonic::client::GrpcService<tonic::body::Body>>::ResponseBody,
                    >,
                >,
            <T as tonic::codegen::Service<http::Request<tonic::body::Body>>>::Error:
                Into<StdError> + std::marker::Send + std::marker::Sync,
        {
            NumberServiceClient::new(InterceptedService::new(inner, interceptor))
        }
        /// Compress requests with the given encoding.
        ///
        /// This requires the server to support it otherwise it might respond with an
        /// error.
        #[must_use]
        pub fn send_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.inner = self.inner.send_compressed(encoding);
            self
        }
        /// Enable decompressing responses.
        #[must_use]
        pub fn accept_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.inner = self.inner.accept_compressed(encoding);
            self
        }
        /// Limits the maximum size of a decoded message.
        ///
        /// Default: `4MB`
        #[must_use]
        pub fn max_decoding_message_size(mut self, limit: usize) -> Self {
            self.inner = self.inner.max_decoding_message_size(limit);
            self
        }
        /// Limits the maximum size of an encoded message.
        ///
        /// Default: `usize::MAX`
        #[must_use]
        pub fn max_encoding_message_size(mut self, limit: usize) -> Self {
            self.inner = self.inner.max_encoding_message_size(limit);
            self
        }
        ///
        pub async fn add(
            &mut self,
            request: impl tonic::IntoRequest<super::AddRequest>,
        ) -> std::result::Result<tonic::Response<super::AddResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::unknown(format!("Service was not ready: {}", e.into()))
            })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/number.v1.NumberService/Add");
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("number.v1.NumberService", "Add"));
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn list(
            &mut self,
            request: impl tonic::IntoRequest<super::ListRequest>,
        ) -> std::result::Result<
            tonic::Response<tonic::codec::Streaming<super::ListResponse>>,
            tonic::Status,
        > {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::unknown(format!("Service was not ready: {}", e.into()))
            })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/number.v1.NumberService/List");
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("number.v1.NumberService", "List"));
            self.inner.server_streaming(req, path, codec).await
        }
        ///
        pub async fn summary(
            &mut self,
            request: impl tonic::IntoRequest<super::SummaryRequest>,
        ) -> std::result::Result<tonic::Response<super::SummaryResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::unknown(format!("Service was not ready: {}", e.into()))
            })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/number.v1.NumberService/Summary");
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("number.v1.NumberService", "Summary"));
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn time_series(
            &mut self,
            request: impl tonic::IntoRequest<super::TimeSeriesRequest>,
        ) -> std::result::Result<tonic::Response<super::TimeSeriesResponse>, tonic::Status>
        {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::unknown(format!("Service was not ready: {}", e.into()))
            })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/number.v1.NumberService/TimeSeries");
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("number.v1.NumberService", "TimeSeries"));
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn distribution(
            &mut self,
            request: impl tonic::IntoRequest<super::DistributionRequest>,
        ) -> std::result::Result<tonic::Response<super::DistributionResponse>, tonic::Status>
        {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::unknown(format!("Service was not ready: {}", e.into()))
            })?;
            let codec = tonic_prost::ProstCodec::default();
            let path =
                http::uri::PathAndQuery::from_static("/number.v1.NumberService/Distribution");
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("number.v1.NumberService", "Distribution"));
            self.inner.unary(req, path, codec).await
        }
        ///
        pub async fn top_names(
            &mut self,
            request: impl tonic::IntoRequest<super::TopNamesRequest>,
        ) -> std::result::Result<tonic::Response<super::TopNamesResponse>, tonic::Status> {
            self.inner.ready().await.map_err(|e| {
                tonic::Status::unknown(format!("Service was not ready: {}", e.into()))
            })?;
            let codec = tonic_prost::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static("/number.v1.NumberService/TopNames");
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("number.v1.NumberService", "TopNames"));
            self.inner.unary(req, path, codec).await
        }
    }
}
