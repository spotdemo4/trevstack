//! Cursor events, and the lines the client reads and prints when it is not
//! attached to a terminal.

use std::fmt;

use crate::proto::cursor::v1::{
    Cursor, ShareResponse, Source, StreamEnd, TrackResponse, WatchResponse, share_response,
    watch_response,
};

/// A cursor event from `Watch` or `Share`.
#[derive(Clone, Debug, PartialEq)]
pub enum CursorEvent {
    /// A stream's cursor moved.
    Moved {
        stream_id: i64,
        username: String,
        source: Source,
        x: f64,
        y: f64,
    },
    /// A stream ended, so its cursor can be removed.
    Ended { stream_id: i64, username: String },
}

impl CursorEvent {
    fn moved(cursor: Cursor) -> Self {
        Self::Moved {
            stream_id: cursor.stream_id,
            username: cursor.username,
            source: Source::try_from(cursor.source).unwrap_or(Source::Unspecified),
            x: cursor.x,
            y: cursor.y,
        }
    }

    fn ended(end: StreamEnd) -> Self {
        Self::Ended {
            stream_id: end.stream_id,
            username: end.username,
        }
    }

    /// Returns the event of a `Watch` response, or `None` for a response
    /// without one, such as an event this client does not know.
    #[must_use]
    pub fn from_watch(response: WatchResponse) -> Option<Self> {
        match response.event? {
            watch_response::Event::Moved(cursor) => Some(Self::moved(cursor)),
            watch_response::Event::Ended(end) => Some(Self::ended(end)),
        }
    }

    /// Returns the event of a `Share` response, or `None` for a response
    /// without one, such as an event this client does not know.
    #[must_use]
    pub fn from_share(response: ShareResponse) -> Option<Self> {
        match response.event? {
            share_response::Event::Moved(cursor) => Some(Self::moved(cursor)),
            share_response::Event::Ended(end) => Some(Self::ended(end)),
        }
    }
}

/// Formats an event as one output line, without the newline:
/// `moved <stream_id> <username> <source> <x> <y>` or
/// `ended <stream_id> <username>`.
impl fmt::Display for CursorEvent {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::Moved {
                stream_id,
                username,
                source,
                x,
                y,
            } => write!(
                formatter,
                "moved {stream_id} {username} {} {x:.4} {y:.4}",
                source_name(*source)
            ),
            Self::Ended {
                stream_id,
                username,
            } => write!(formatter, "ended {stream_id} {username}"),
        }
    }
}

/// Returns the lowercase name of the RPC that recorded a stream.
#[must_use]
pub const fn source_name(source: Source) -> &'static str {
    match source {
        Source::Unspecified => "unspecified",
        Source::Send => "send",
        Source::Track => "track",
        Source::Share => "share",
    }
}

/// Formats the summary of a `Track` stream as one output line, without the
/// newline: `<stream_id> <count> <distance>`, all zero when the stream sent no
/// positions.
#[must_use]
pub fn format_summary(response: &TrackResponse) -> String {
    let stream = response.stream.clone().unwrap_or_default();
    format!("{} {} {:.4}", stream.id, stream.count, stream.distance)
}

/// Parses a line of input as a whitespace-separated `x y` position, each from
/// 0 to 1. Returns `Ok(None)` for a blank line.
///
/// # Errors
///
/// Returns a description of the problem if the line is not blank and not a
/// valid position.
pub fn parse_position(line: &str) -> Result<Option<(f64, f64)>, String> {
    let mut parts = line.split_whitespace();
    let Some(x) = parts.next() else {
        return Ok(None);
    };
    let (Some(y), None) = (parts.next(), parts.next()) else {
        return Err(format!("expected \"x y\", got {line:?}"));
    };
    let x = crate::cli::parse_coordinate(x)?;
    let y = crate::cli::parse_coordinate(y)?;
    Ok(Some((x, y)))
}

#[cfg(test)]
mod tests {
    use super::{CursorEvent, format_summary, parse_position};
    use crate::proto::cursor::v1::{
        Cursor, ShareResponse, Source, StreamEnd, StreamSummary, TrackResponse, WatchResponse,
        share_response, watch_response,
    };

    fn cursor() -> Cursor {
        Cursor {
            stream_id: 12,
            username: "alice".to_owned(),
            source: Source::Track.into(),
            x: 0.25,
            y: 0.123_456,
            timestamp: None,
        }
    }

    #[test]
    fn formats_moved_events() {
        let event = CursorEvent::from_watch(WatchResponse {
            event: Some(watch_response::Event::Moved(cursor())),
        })
        .unwrap();

        assert_eq!(event.to_string(), "moved 12 alice track 0.2500 0.1235");
    }

    #[test]
    fn formats_ended_events() {
        let event = CursorEvent::from_share(ShareResponse {
            event: Some(share_response::Event::Ended(StreamEnd {
                stream_id: 3,
                username: "bob".to_owned(),
            })),
        })
        .unwrap();

        assert_eq!(event.to_string(), "ended 3 bob");
    }

    #[test]
    fn skips_responses_without_an_event() {
        assert_eq!(CursorEvent::from_watch(WatchResponse { event: None }), None);
        assert_eq!(CursorEvent::from_share(ShareResponse { event: None }), None);
    }

    #[test]
    fn names_unknown_sources_unspecified() {
        let event = CursorEvent::from_watch(WatchResponse {
            event: Some(watch_response::Event::Moved(Cursor {
                source: 99,
                ..cursor()
            })),
        })
        .unwrap();

        assert_eq!(
            event.to_string(),
            "moved 12 alice unspecified 0.2500 0.1235"
        );
    }

    #[test]
    fn formats_summaries() {
        let summary = format_summary(&TrackResponse {
            stream: Some(StreamSummary {
                id: 7,
                count: 3,
                distance: 1.5,
                ..StreamSummary::default()
            }),
        });
        assert_eq!(summary, "7 3 1.5000");

        assert_eq!(
            format_summary(&TrackResponse { stream: None }),
            "0 0 0.0000"
        );
    }

    #[test]
    fn parses_positions() {
        assert_eq!(parse_position("0.25 0.75"), Ok(Some((0.25, 0.75))));
        assert_eq!(parse_position("  1\t0  "), Ok(Some((1.0, 0.0))));
        assert_eq!(parse_position(""), Ok(None));
        assert_eq!(parse_position("   "), Ok(None));
        assert!(parse_position("0.5").is_err());
        assert!(parse_position("0.5 0.5 0.5").is_err());
        assert!(parse_position("0.5 2").is_err());
        assert!(parse_position("x y").is_err());
    }
}
