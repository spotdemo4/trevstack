//! The terminal pad: the terminal as a surface to move the mouse over, which
//! shows other streams' cursors.
//!
//! Positions are fractions of the pad's width and height, as on the web app's
//! pad, so the same position shows in the same place on every client whatever
//! its size.

use std::{
    collections::HashMap,
    io::{self, Stdout, Write},
    sync::Once,
    time::{Duration, Instant},
};

use crossterm::{
    cursor::{Hide, MoveTo, Show},
    event::{
        DisableMouseCapture, EnableMouseCapture, Event, KeyCode, KeyEventKind, KeyModifiers,
        MouseEventKind,
    },
    execute, queue,
    style::{Attribute, Color, Print, ResetColor, SetAttribute, SetForegroundColor},
    terminal::{
        self, Clear, ClearType, EnterAlternateScreen, LeaveAlternateScreen, disable_raw_mode,
        enable_raw_mode,
    },
};

use crate::event::CursorEvent;

/// How long a cursor stays on the pad without moving.
pub const IDLE_TIMEOUT: Duration = Duration::from_secs(10);

const MARKER: char = '●';
const COLORS: [Color; 6] = [
    Color::Red,
    Color::Green,
    Color::Yellow,
    Color::Blue,
    Color::Magenta,
    Color::Cyan,
];

/// What a terminal event asks the pad's owner to do.
#[derive(Clone, Copy, Debug, PartialEq)]
pub enum Input {
    /// The user asked to quit, with q, Esc or Ctrl-C.
    Quit,
    /// The mouse moved over the pad, to the position `(x, y)`.
    Move(f64, f64),
    /// Nothing for the owner to do.
    Ignore,
}

#[derive(Clone, Debug)]
struct Remote {
    username: String,
    x: f64,
    y: f64,
    moved: Instant,
}

/// The terminal in raw mode on its alternate screen, with mouse capture, until
/// the pad is dropped.
pub struct Pad {
    out: Stdout,
    columns: u16,
    rows: u16,
    cursors: HashMap<i64, Remote>,
    status: String,
}

impl Pad {
    /// Takes over the terminal. It is restored when the pad is dropped, and
    /// before a panic message is printed.
    ///
    /// # Errors
    ///
    /// Returns an error if the terminal can't be set up.
    pub fn open() -> io::Result<Self> {
        let (columns, rows) = terminal::size()?;
        install_panic_hook();
        enable_raw_mode()?;
        // created before the rest of the setup, so dropping it restores what succeeded
        let mut pad = Self {
            out: io::stdout(),
            columns,
            rows,
            cursors: HashMap::new(),
            status: String::new(),
        };
        execute!(
            pad.out,
            EnterAlternateScreen,
            EnableMouseCapture,
            Hide,
            Clear(ClearType::All)
        )?;
        Ok(pad)
    }

    /// Returns what a terminal event asks for, and follows resizes.
    pub fn input(&mut self, event: &Event) -> Input {
        match event {
            Event::Key(key) if key.kind != KeyEventKind::Release => match key.code {
                KeyCode::Char('q') | KeyCode::Esc => Input::Quit,
                KeyCode::Char('c') if key.modifiers.contains(KeyModifiers::CONTROL) => Input::Quit,
                _ => Input::Ignore,
            },
            Event::Mouse(mouse)
                if matches!(
                    mouse.kind,
                    MouseEventKind::Moved | MouseEventKind::Drag(_) | MouseEventKind::Down(_)
                ) =>
            {
                position(mouse.column, mouse.row, self.columns, self.rows)
                    .map_or(Input::Ignore, |(x, y)| Input::Move(x, y))
            }
            Event::Resize(columns, rows) => {
                self.columns = *columns;
                self.rows = *rows;
                Input::Ignore
            }
            _ => Input::Ignore,
        }
    }

    /// Sets the text of the status bar.
    pub fn set_status(&mut self, status: impl Into<String>) {
        self.status = status.into();
    }

    /// Returns how many cursors the pad shows.
    #[must_use]
    pub fn cursor_count(&self) -> usize {
        self.cursors.len()
    }

    /// Moves or removes the cursor of an event's stream.
    pub fn apply(&mut self, event: CursorEvent, now: Instant) {
        match event {
            CursorEvent::Moved {
                stream_id,
                username,
                x,
                y,
                ..
            } => {
                self.cursors.insert(
                    stream_id,
                    Remote {
                        username,
                        x,
                        y,
                        moved: now,
                    },
                );
            }
            CursorEvent::Ended { stream_id, .. } => {
                self.cursors.remove(&stream_id);
            }
        }
    }

    /// Removes the cursors that have not moved for [`IDLE_TIMEOUT`].
    pub fn expire(&mut self, now: Instant) {
        self.cursors
            .retain(|_, cursor| now.saturating_duration_since(cursor.moved) < IDLE_TIMEOUT);
    }

    /// Redraws the cursors and the status bar.
    ///
    /// # Errors
    ///
    /// Returns an error if writing to the terminal fails.
    pub fn draw(&mut self) -> io::Result<()> {
        let pad_rows = pad_rows(self.rows);
        queue!(self.out, Clear(ClearType::All))?;

        let mut cursors: Vec<_> = self.cursors.iter().collect();
        cursors.sort_unstable_by_key(|(stream_id, _)| **stream_id);
        for (_, cursor) in cursors {
            let column = cell(cursor.x, self.columns);
            let row = cell(cursor.y, pad_rows);
            // the label goes after the marker and a space, as far as it fits
            let room = usize::from(self.columns.saturating_sub(column).saturating_sub(2));
            let label: String = cursor.username.chars().take(room).collect();
            queue!(
                self.out,
                MoveTo(column, row),
                SetForegroundColor(color(&cursor.username)),
                Print(MARKER),
                Print(' '),
                Print(label),
                ResetColor,
            )?;
        }

        let width = usize::from(self.columns);
        let status: String = self.status.chars().take(width).collect();
        queue!(
            self.out,
            MoveTo(0, self.rows.saturating_sub(1)),
            SetAttribute(Attribute::Reverse),
            Print(format!("{status:<width$}")),
            SetAttribute(Attribute::Reset),
        )?;
        self.out.flush()
    }
}

impl Drop for Pad {
    fn drop(&mut self) {
        restore();
    }
}

/// Puts the terminal back as it was before [`Pad::open`]; harmless when it
/// already is.
fn restore() {
    let _ = execute!(
        io::stdout(),
        DisableMouseCapture,
        Show,
        LeaveAlternateScreen
    );
    let _ = disable_raw_mode();
}

/// Restores the terminal before the panic message is printed, which would
/// otherwise be printed on the alternate screen and lost.
fn install_panic_hook() {
    static INSTALL: Once = Once::new();
    INSTALL.call_once(|| {
        let previous = std::panic::take_hook();
        std::panic::set_hook(Box::new(move |info| {
            restore();
            previous(info);
        }));
    });
}

/// Returns the number of rows of the pad: every row except the status bar's.
fn pad_rows(rows: u16) -> u16 {
    rows.saturating_sub(1).max(1)
}

/// Returns the position of a cell of the pad as fractions of its width and
/// height, or `None` for a cell of the status bar.
#[must_use]
pub fn position(column: u16, row: u16, columns: u16, rows: u16) -> Option<(f64, f64)> {
    let pad_rows = pad_rows(rows);
    if row >= pad_rows {
        return None;
    }
    Some((fraction(column, columns), fraction(row, pad_rows)))
}

/// Returns cell `index` of `cells` as a fraction from 0 (the first) to 1 (the
/// last).
fn fraction(index: u16, cells: u16) -> f64 {
    let last = cells.saturating_sub(1);
    if last == 0 {
        return 0.0;
    }
    (f64::from(index) / f64::from(last)).clamp(0.0, 1.0)
}

/// Returns the cell of `cells` nearest to `fraction` of the way from the first
/// to the last.
// the value is clamped to 0..=last first, so it neither truncates nor loses its sign
#[allow(
    clippy::as_conversions,
    clippy::cast_possible_truncation,
    clippy::cast_sign_loss
)]
fn cell(fraction: f64, cells: u16) -> u16 {
    let last = f64::from(cells.saturating_sub(1));
    (fraction.clamp(0.0, 1.0) * last).round().clamp(0.0, last) as u16
}

/// Returns the colour of a username's cursor, the same on every run.
fn color(username: &str) -> Color {
    // FNV-1a, which unlike the standard library's hasher is stable across releases
    let hash = username
        .bytes()
        .fold(0xcbf2_9ce4_8422_2325_u64, |hash, byte| {
            (hash ^ u64::from(byte)).wrapping_mul(0x0100_0000_01b3)
        });
    let count = u64::try_from(COLORS.len()).unwrap_or(1);
    usize::try_from(hash.checked_rem(count).unwrap_or(0))
        .ok()
        .and_then(|index| COLORS.get(index).copied())
        .unwrap_or(Color::White)
}

#[cfg(test)]
mod tests {
    use super::{cell, color, fraction, position};

    #[test]
    fn maps_cells_to_fractions() {
        // an 81 by 21 terminal has a pad of 81 by 20, above the status bar
        assert_eq!(position(0, 0, 81, 21), Some((0.0, 0.0)));
        assert_eq!(position(80, 19, 81, 21), Some((1.0, 1.0)));
        assert_eq!(position(20, 0, 81, 21), Some((0.25, 0.0)));
        assert_eq!(position(0, 20, 81, 21), None);
    }

    #[test]
    fn handles_tiny_terminals() {
        assert_eq!(fraction(0, 1), 0.0);
        assert_eq!(fraction(0, 0), 0.0);
        assert_eq!(position(0, 0, 1, 1), Some((0.0, 0.0)));
        assert_eq!(cell(0.5, 0), 0);
        assert_eq!(cell(0.5, 1), 0);
    }

    #[test]
    fn maps_fractions_to_cells() {
        assert_eq!(cell(0.0, 81), 0);
        assert_eq!(cell(1.0, 81), 80);
        assert_eq!(cell(0.25, 81), 20);
        assert_eq!(cell(-1.0, 81), 0);
        assert_eq!(cell(2.0, 81), 80);
        assert_eq!(cell(f64::NAN, 81), 0);
    }

    #[test]
    fn round_trips_cells() {
        for column in 0..80 {
            let (x, _) = position(column, 0, 80, 24).unwrap();
            assert_eq!(cell(x, 80), column);
        }
    }

    #[test]
    fn colors_usernames_consistently() {
        assert_eq!(color("alice"), color("alice"));
        let distinct: std::collections::HashSet<_> = ["alice", "bob", "carol", "dave", "erin"]
            .iter()
            .map(|name| format!("{:?}", color(name)))
            .collect();
        assert!(distinct.len() > 1);
    }
}
