//! The streaming calls: `Watch`, `Track` and `Share`, each either on the
//! terminal pad or over lines of stdin and stdout.

use std::{
    future::Future,
    io::{self, Write},
    time::{Duration, Instant},
};

use crossterm::event::EventStream;
use futures_util::StreamExt;
use tokio::sync::mpsc::UnboundedReceiver;
use trevrpc::client::StreamingRpcTransport;

use crate::{
    client::{AUTHORIZATION_METADATA, ClientError},
    event::{CursorEvent, parse_position},
    pad::{Input, Pad},
    proto::cursor::v1::{
        CursorServiceClient, ShareRequest, TrackRequest, TrackResponse, WatchRequest,
    },
};

/// How often the pad drops the cursors that stopped moving.
const EXPIRE_INTERVAL: Duration = Duration::from_secs(1);

/// Lines of input, read as they arrive.
pub type Lines = UnboundedReceiver<io::Result<String>>;

fn watch_error(error: trevrpc::Error) -> ClientError {
    ClientError::Watch(error.into_status())
}

fn track_error(error: trevrpc::Error) -> ClientError {
    ClientError::Track(error.into_status())
}

fn share_error(error: trevrpc::Error) -> ClientError {
    ClientError::Share(error.into_status())
}

fn write_line(out: &mut impl Write, line: &impl std::fmt::Display) -> Result<(), ClientError> {
    writeln!(out, "{line}")
        .and_then(|()| out.flush())
        .map_err(ClientError::Output)
}

/// Returns the position of the next line of input that has one, or `None` at
/// the end of the input.
async fn next_position(lines: &mut Lines) -> Result<Option<(f64, f64)>, ClientError> {
    while let Some(line) = lines.recv().await {
        let line = line.map_err(ClientError::Stdin)?;
        if let Some(position) = parse_position(&line).map_err(ClientError::Input)? {
            return Ok(Some(position));
        }
    }
    Ok(None)
}

/// Prints each `Watch` event to `out` as a line until the server ends the
/// stream or `stop` completes.
///
/// # Errors
///
/// Returns an error if the call fails or `out` can't be written.
pub async fn watch_lines<T>(
    client: &CursorServiceClient<T>,
    authorization: &str,
    out: &mut impl Write,
    stop: impl Future<Output = ()>,
) -> Result<(), ClientError>
where
    T: StreamingRpcTransport,
{
    let mut responses = client
        .watch(&WatchRequest {})
        .metadata(AUTHORIZATION_METADATA, authorization.to_owned())
        .await
        .map_err(watch_error)?;
    tokio::pin!(stop);

    loop {
        tokio::select! {
            () = &mut stop => return Ok(()),
            response = responses.next() => match response {
                None => return Ok(()),
                Some(response) => {
                    if let Some(event) = CursorEvent::from_watch(response.map_err(watch_error)?) {
                        write_line(out, &event)?;
                    }
                }
            },
        }
    }
}

/// Sends the position of each line of `lines` with `Track` until the input
/// ends or `stop` completes, then ends the stream and returns its summary.
///
/// # Errors
///
/// Returns an error if a line is not a position or the call fails.
pub async fn track_lines<T>(
    client: &CursorServiceClient<T>,
    authorization: &str,
    lines: &mut Lines,
    stop: impl Future<Output = ()>,
) -> Result<TrackResponse, ClientError>
where
    T: StreamingRpcTransport,
{
    let call = client
        .track()
        .metadata(AUTHORIZATION_METADATA, authorization.to_owned())
        .await
        .map_err(track_error)?;
    tokio::pin!(stop);

    loop {
        tokio::select! {
            () = &mut stop => break,
            position = next_position(lines) => match position? {
                None => break,
                Some((x, y)) => call
                    .send(TrackRequest { x: Some(x), y: Some(y) })
                    .await
                    .map_err(track_error)?,
            },
        }
    }

    call.close_and_recv().await.map_err(track_error)
}

/// Sends the position of each line of `lines` with `Share`, printing each
/// event the server streams back to `out` as a line. Ends the request stream
/// when the input ends or `stop` completes, and returns once the server ends
/// the call.
///
/// # Errors
///
/// Returns an error if a line is not a position, the call fails, or `out`
/// can't be written.
pub async fn share_lines<T>(
    client: &CursorServiceClient<T>,
    authorization: &str,
    lines: &mut Lines,
    out: &mut impl Write,
    stop: impl Future<Output = ()>,
) -> Result<(), ClientError>
where
    T: StreamingRpcTransport,
{
    let (sender, mut responses) = client
        .share()
        .metadata(AUTHORIZATION_METADATA, authorization.to_owned())
        .await
        .map_err(share_error)?
        .split();
    // None once the request stream is finished
    let mut sender = Some(sender);
    tokio::pin!(stop);

    loop {
        tokio::select! {
            () = &mut stop, if sender.is_some() => {
                if let Some(sender) = sender.take() {
                    sender.finish().map_err(share_error)?;
                }
            }
            position = next_position(lines), if sender.is_some() => match position? {
                None => {
                    if let Some(sender) = sender.take() {
                        sender.finish().map_err(share_error)?;
                    }
                }
                Some((x, y)) => {
                    if let Some(sender) = &sender {
                        sender
                            .send(ShareRequest { x: Some(x), y: Some(y) })
                            .await
                            .map_err(share_error)?;
                    }
                }
            },
            response = responses.next() => match response {
                None => return Ok(()),
                Some(response) => {
                    if let Some(event) = CursorEvent::from_share(response.map_err(share_error)?) {
                        write_line(out, &event)?;
                    }
                }
            },
        }
    }
}

/// Shows every `Watch` event on the terminal pad until the user quits or the
/// server ends the stream.
///
/// # Errors
///
/// Returns an error if the call fails or the terminal can't be used.
pub async fn watch_pad<T>(
    client: &CursorServiceClient<T>,
    authorization: &str,
    username: &str,
) -> Result<(), ClientError>
where
    T: StreamingRpcTransport,
{
    let mut responses = client
        .watch(&WatchRequest {})
        .metadata(AUTHORIZATION_METADATA, authorization.to_owned())
        .await
        .map_err(watch_error)?;
    let mut pad = Pad::open().map_err(ClientError::Terminal)?;
    let mut terminal = EventStream::new();
    let mut expire = tokio::time::interval(EXPIRE_INTERVAL);

    loop {
        pad.set_status(format!(
            "watching as {username} · {} cursors · q to quit",
            pad.cursor_count()
        ));
        pad.draw().map_err(ClientError::Terminal)?;

        tokio::select! {
            response = responses.next() => match response {
                None => return Ok(()),
                Some(response) => {
                    if let Some(event) = CursorEvent::from_watch(response.map_err(watch_error)?) {
                        pad.apply(event, Instant::now());
                    }
                }
            },
            event = terminal.next() => match event {
                None => return Ok(()),
                Some(event) => {
                    if pad.input(&event.map_err(ClientError::Terminal)?) == Input::Quit {
                        return Ok(());
                    }
                }
            },
            _ = expire.tick() => pad.expire(Instant::now()),
        }
    }
}

/// Sends the mouse position over the terminal pad with `Track` until the user
/// quits, then ends the stream and returns its summary.
///
/// # Errors
///
/// Returns an error if the call fails or the terminal can't be used.
pub async fn track_pad<T>(
    client: &CursorServiceClient<T>,
    authorization: &str,
    username: &str,
) -> Result<TrackResponse, ClientError>
where
    T: StreamingRpcTransport,
{
    let call = client
        .track()
        .metadata(AUTHORIZATION_METADATA, authorization.to_owned())
        .await
        .map_err(track_error)?;
    let mut pad = Pad::open().map_err(ClientError::Terminal)?;
    let mut terminal = EventStream::new();
    let mut count: u64 = 0;

    loop {
        pad.set_status(format!(
            "tracking as {username} · {count} positions · q to stop"
        ));
        pad.draw().map_err(ClientError::Terminal)?;

        let Some(event) = terminal.next().await else {
            break;
        };
        match pad.input(&event.map_err(ClientError::Terminal)?) {
            Input::Quit => break,
            Input::Move(x, y) => {
                call.send(TrackRequest {
                    x: Some(x),
                    y: Some(y),
                })
                .await
                .map_err(track_error)?;
                count = count.saturating_add(1);
            }
            Input::Ignore => {}
        }
    }

    drop(pad);
    call.close_and_recv().await.map_err(track_error)
}

/// Sends the mouse position over the terminal pad with `Share`, showing every
/// other stream's cursor, until the user quits or the server ends the call.
///
/// # Errors
///
/// Returns an error if the call fails or the terminal can't be used.
pub async fn share_pad<T>(
    client: &CursorServiceClient<T>,
    authorization: &str,
    username: &str,
) -> Result<(), ClientError>
where
    T: StreamingRpcTransport,
{
    let (sender, mut responses) = client
        .share()
        .metadata(AUTHORIZATION_METADATA, authorization.to_owned())
        .await
        .map_err(share_error)?
        .split();
    let mut pad = Pad::open().map_err(ClientError::Terminal)?;
    let mut terminal = EventStream::new();
    let mut expire = tokio::time::interval(EXPIRE_INTERVAL);
    let mut count: u64 = 0;

    loop {
        pad.set_status(format!(
            "sharing as {username} · {count} positions · {} cursors · q to quit",
            pad.cursor_count()
        ));
        pad.draw().map_err(ClientError::Terminal)?;

        tokio::select! {
            response = responses.next() => match response {
                None => return Ok(()),
                Some(response) => {
                    if let Some(event) = CursorEvent::from_share(response.map_err(share_error)?) {
                        pad.apply(event, Instant::now());
                    }
                }
            },
            event = terminal.next() => {
                let Some(event) = event else { break };
                match pad.input(&event.map_err(ClientError::Terminal)?) {
                    Input::Quit => break,
                    Input::Move(x, y) => {
                        sender
                            .send(ShareRequest { x: Some(x), y: Some(y) })
                            .await
                            .map_err(share_error)?;
                        count = count.saturating_add(1);
                    }
                    Input::Ignore => {}
                }
            },
            _ = expire.tick() => pad.expire(Instant::now()),
        }
    }

    // The server ends the call once the request stream ends; read on until it
    // does, so a failure is still reported.
    drop(pad);
    sender.finish().map_err(share_error)?;
    while let Some(response) = responses.next().await {
        response.map_err(share_error)?;
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use std::{
        future::pending,
        io,
        sync::{Arc, Mutex},
    };

    use bytes::Bytes;
    use futures_util::{StreamExt, stream};
    use prost::Message;
    use tokio::sync::mpsc;
    use trevrpc::{
        BoxStream, Code, ResponseEnvelope, ResponseFrame, Status,
        client::{RpcTransport, StreamingRpcTransport},
        wire::RpcRequest,
    };

    use super::{Lines, share_lines, track_lines, watch_lines};
    use crate::{
        client::{ClientError, tests::authorization},
        proto::cursor::v1::{
            Cursor, CursorServiceClient, ShareRequest, ShareResponse, Source, StreamEnd,
            StreamSummary, TrackRequest, TrackResponse, WatchResponse, share_response,
            watch_response,
        },
    };

    /// The method, authorization metadata and request messages of each call.
    type SeenCalls = Arc<Mutex<Vec<(String, Option<String>, Vec<Vec<u8>>)>>>;

    /// Answers a streaming call to a method with response messages, given its
    /// request messages.
    type Respond = fn(&str, &[Vec<u8>]) -> Result<Vec<Vec<u8>>, Status>;

    /// A transport whose streaming calls read every request message, then
    /// answer with the messages `respond` returns for them.
    #[derive(Clone)]
    struct StreamingMock {
        seen: SeenCalls,
        respond: Respond,
    }

    fn streaming_mock(respond: Respond) -> StreamingMock {
        StreamingMock {
            seen: SeenCalls::default(),
            respond,
        }
    }

    #[trevrpc::async_trait]
    impl RpcTransport for StreamingMock {
        async fn call(
            &self,
            _request: RpcRequest,
            _body: Vec<u8>,
        ) -> trevrpc::Result<ResponseEnvelope<Bytes>> {
            Err(Status::unimplemented("unary calls are not mocked").into())
        }
    }

    #[trevrpc::async_trait]
    impl StreamingRpcTransport for StreamingMock {
        async fn streaming_call(
            &self,
            request: RpcRequest,
            request_body: BoxStream<Vec<u8>>,
        ) -> trevrpc::Result<BoxStream<ResponseFrame>> {
            let seen = Arc::clone(&self.seen);
            let respond = self.respond;
            // the responses come once the request stream ends, which a
            // client-streaming call only ends after it starts
            let frames = stream::once(async move {
                let bodies: Vec<Vec<u8>> = request_body
                    .filter_map(|body| async move { body.ok() })
                    .collect()
                    .await;
                let response = respond(&request.method, &bodies);
                seen.lock().unwrap().push((
                    request.method.clone(),
                    authorization(&request),
                    bodies,
                ));
                let frames: Vec<trevrpc::Result<ResponseFrame>> = match response {
                    Ok(messages) => messages
                        .into_iter()
                        .map(|message| Ok(ResponseFrame::message(message)))
                        .chain([Ok(ResponseFrame::Status(Status::ok()))])
                        .collect(),
                    Err(status) => vec![Ok(ResponseFrame::Status(status))],
                };
                stream::iter(frames)
            })
            .flatten();
            Ok(Box::pin(frames))
        }
    }

    /// Returns input of `input`'s lines, which ends after them.
    fn lines(input: &[&str]) -> Lines {
        let (sender, receiver) = mpsc::unbounded_channel();
        for line in input {
            sender.send(Ok((*line).to_owned())).unwrap();
        }
        receiver
    }

    fn cursor(stream_id: i64) -> Cursor {
        Cursor {
            stream_id,
            username: "alice".to_owned(),
            source: Source::Share.into(),
            x: 0.5,
            y: 0.25,
            timestamp: None,
        }
    }

    #[tokio::test]
    async fn prints_watch_events() {
        let transport = streaming_mock(|_, _| {
            Ok(vec![
                WatchResponse {
                    event: Some(watch_response::Event::Moved(cursor(4))),
                }
                .encode_to_vec(),
                WatchResponse {
                    event: Some(watch_response::Event::Ended(StreamEnd {
                        stream_id: 4,
                        username: "alice".to_owned(),
                    })),
                }
                .encode_to_vec(),
            ])
        });
        let client = CursorServiceClient::new(transport.clone());
        let mut out = Vec::new();

        watch_lines(&client, "Bearer token", &mut out, pending())
            .await
            .unwrap();

        assert_eq!(
            String::from_utf8(out).unwrap(),
            "moved 4 alice share 0.5000 0.2500\nended 4 alice\n"
        );
        let seen = transport.seen.lock().unwrap();
        assert_eq!(seen[0].0, "/cursor.v1.CursorService/Watch");
        assert_eq!(seen[0].1.as_deref(), Some("Bearer token"));
    }

    #[tokio::test]
    async fn stops_watching_when_asked() {
        let transport = streaming_mock(|_, _| Ok(vec![]));
        let client = CursorServiceClient::new(transport);
        let mut out = Vec::new();

        watch_lines(&client, "Bearer token", &mut out, async {})
            .await
            .unwrap();
    }

    #[tokio::test]
    async fn tracks_each_line_then_returns_the_summary() {
        let transport = streaming_mock(|_, bodies| {
            Ok(vec![
                TrackResponse {
                    stream: Some(StreamSummary {
                        id: 9,
                        count: i64::try_from(bodies.len()).unwrap(),
                        ..StreamSummary::default()
                    }),
                }
                .encode_to_vec(),
            ])
        });
        let client = CursorServiceClient::new(transport.clone());
        let mut input = lines(&["0.1 0.2", "", "0.3 0.4"]);

        let response = track_lines(&client, "Bearer token", &mut input, pending())
            .await
            .unwrap();

        assert_eq!(response.stream.unwrap().count, 2);
        let seen = transport.seen.lock().unwrap();
        assert_eq!(seen[0].0, "/cursor.v1.CursorService/Track");
        let requests: Vec<_> = seen[0]
            .2
            .iter()
            .map(|body| TrackRequest::decode(body.as_slice()).unwrap())
            .collect();
        assert_eq!(
            requests,
            [
                TrackRequest {
                    x: Some(0.1),
                    y: Some(0.2)
                },
                TrackRequest {
                    x: Some(0.3),
                    y: Some(0.4)
                },
            ]
        );
    }

    #[tokio::test]
    async fn rejects_malformed_lines() {
        let transport = streaming_mock(|_, _| Ok(vec![TrackResponse::default().encode_to_vec()]));
        let client = CursorServiceClient::new(transport);
        let mut input = lines(&["0.1 0.2", "0.3"]);

        let error = track_lines(&client, "Bearer token", &mut input, pending())
            .await
            .unwrap_err();

        assert!(matches!(error, ClientError::Input(_)));
    }

    #[tokio::test]
    async fn reports_stdin_errors() {
        let transport = streaming_mock(|_, _| Ok(vec![TrackResponse::default().encode_to_vec()]));
        let client = CursorServiceClient::new(transport);
        let (sender, mut input) = mpsc::unbounded_channel();
        sender.send(Err(io::Error::other("broken"))).unwrap();

        let error = track_lines(&client, "Bearer token", &mut input, pending())
            .await
            .unwrap_err();

        assert!(matches!(error, ClientError::Stdin(_)));
    }

    #[tokio::test]
    async fn reports_track_failures() {
        let transport = streaming_mock(|_, _| Err(Status::invalid_argument("x out of range")));
        let client = CursorServiceClient::new(transport);
        let mut input = lines(&["0.1 0.2"]);

        let error = track_lines(&client, "Bearer token", &mut input, pending())
            .await
            .unwrap_err();

        assert!(
            matches!(error, ClientError::Track(status) if status.code() == Code::InvalidArgument)
        );
    }

    #[tokio::test]
    async fn shares_each_line_and_prints_the_events() {
        let transport = streaming_mock(|_, bodies| {
            Ok(bodies
                .iter()
                .enumerate()
                .map(|(index, body)| {
                    let request = ShareRequest::decode(body.as_slice()).unwrap();
                    ShareResponse {
                        event: Some(share_response::Event::Moved(Cursor {
                            x: request.x.unwrap(),
                            y: request.y.unwrap(),
                            ..cursor(i64::try_from(index).unwrap())
                        })),
                    }
                    .encode_to_vec()
                })
                .collect())
        });
        let client = CursorServiceClient::new(transport.clone());
        let mut input = lines(&["0.1 0.2", "0.3 0.4"]);
        let mut out = Vec::new();

        share_lines(&client, "Bearer token", &mut input, &mut out, pending())
            .await
            .unwrap();

        assert_eq!(
            String::from_utf8(out).unwrap(),
            "moved 0 alice share 0.1000 0.2000\nmoved 1 alice share 0.3000 0.4000\n"
        );
        assert_eq!(
            transport.seen.lock().unwrap()[0].0,
            "/cursor.v1.CursorService/Share"
        );
    }
}
