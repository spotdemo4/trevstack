CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL COLLATE nocase UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    token_hash BLOB PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    expires_at INTEGER NOT NULL
) WITHOUT ROWID;
CREATE INDEX IF NOT EXISTS sessions_user_id ON sessions (user_id);
CREATE INDEX IF NOT EXISTS sessions_expires_at ON sessions (expires_at);

-- One row per call that sent positions: a Send, or a Track or Share stream.
CREATE TABLE IF NOT EXISTS cursor_streams (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    -- cursor.v1.Source
    source INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS cursor_streams_user_id ON cursor_streams (user_id);

CREATE TABLE IF NOT EXISTS cursors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    stream_id INTEGER NOT NULL REFERENCES cursor_streams (id) ON DELETE CASCADE,
    -- fractions of the pad's width and height, from its top-left corner
    x REAL NOT NULL,
    y REAL NOT NULL,
    -- distance from the stream's previous position, 0 for its first
    distance REAL NOT NULL,
    -- Unix milliseconds
    timestamp INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS cursors_stream_id ON cursors (stream_id);
CREATE INDEX IF NOT EXISTS cursors_timestamp ON cursors (timestamp);
