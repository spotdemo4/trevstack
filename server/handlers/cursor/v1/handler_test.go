package v1_test

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/proto"
	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	"trev.zip/llc/trevrpc/trevrpc-go/protovalidatetrevrpc"
	"trev.zip/template/stack/server/v2/auth"
	"trev.zip/template/stack/server/v2/database"
	cursorv1 "trev.zip/template/stack/server/v2/gen/cursor/v1"
	cursorv1handler "trev.zip/template/stack/server/v2/handlers/cursor/v1"
)

type handlerT = cursorv1handler.Handler

// base is the time the test clock starts at.
var base = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// newTest opens an in-memory SQLite DB with the users alice and bob, runs
// migrations, and returns a handler that uses it, whose clock advances a
// second each time it is read. The returned *sql.DB is the same instance the
// handler uses, so callers can seed rows or assert state directly.
func newTest(t *testing.T) (*cursorv1handler.Handler, *sql.DB) {
	t.Helper()

	db, err := sql.Open("sqlite3", "file::memory:?_foreign_keys=true")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	// Each :memory: connection is its own database, so pin the pool to one
	// connection so every query in the test sees the same DB.
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })

	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	for _, username := range []string{"alice", "bob"} {
		if _, err := db.ExecContext(t.Context(), `INSERT INTO users (username, password_hash) VALUES (?, '')`, username); err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}

	handler := cursorv1handler.New(db)
	var mu sync.Mutex
	now := base
	handler.SetNow(func() time.Time {
		mu.Lock()
		defer mu.Unlock()
		now = now.Add(time.Second)
		return now
	})
	return handler, db
}

// as returns a context of a call authenticated as username.
func as(username string) context.Context {
	return auth.ContextWithUsername(context.Background(), username)
}

// seedStream inserts a stream of username's with a position at each time,
// moving right by 0.1 each time, starting from (0.1, 0.5), and returns its
// ID.
func seedStream(t *testing.T, db *sql.DB, username string, source cursorv1.Source, times ...time.Time) int64 {
	t.Helper()
	var id int64
	err := db.QueryRowContext(t.Context(),
		`INSERT INTO cursor_streams (user_id, source) SELECT id, ? FROM users WHERE username = ? RETURNING id`,
		int32(source), username,
	).Scan(&id)
	if err != nil {
		t.Fatalf("seed stream: %v", err)
	}
	for i, ts := range times {
		distance := 0.1
		if i == 0 {
			distance = 0
		}
		_, err := db.ExecContext(t.Context(),
			`INSERT INTO cursors (stream_id, x, y, distance, timestamp) VALUES (?, ?, 0.5, ?, ?)`,
			id, 0.1*float64(i+1), distance, ts.UnixMilli(),
		)
		if err != nil {
			t.Fatalf("seed cursor: %v", err)
		}
	}
	return id
}

// pipe is a request stream that the test sends messages on, as a client
// would.
type pipe[T any] struct {
	messages chan T
	closed   chan struct{}
	once     sync.Once
}

func newPipe[T any]() *pipe[T] {
	return &pipe[T]{messages: make(chan T), closed: make(chan struct{})}
}

// Send hands message to the handler, and waits until it has received it.
func (p *pipe[T]) Send(message T) {
	p.messages <- message
}

// CloseSend ends the stream, as a client ending its half of the call.
func (p *pipe[T]) CloseSend() {
	close(p.messages)
}

func (p *pipe[T]) Recv() (T, error) {
	var zero T
	select {
	case message, ok := <-p.messages:
		if !ok {
			return zero, io.EOF
		}
		return message, nil
	case <-p.closed:
		return zero, io.ErrClosedPipe
	}
}

func (p *pipe[T]) Close() error {
	p.once.Do(func() { close(p.closed) })
	return nil
}

// recv returns the next message of stream, failing the test on an error.
func recv[T any](t *testing.T, stream trevrpc.MessageStream[T]) T {
	t.Helper()
	message, err := stream.Recv()
	if err != nil {
		t.Fatalf("Recv: %v", err)
	}
	return message
}

// recvAsync receives the next message of stream in the background.
func recvAsync[T any](stream trevrpc.MessageStream[T]) <-chan T {
	ch := make(chan T, 1)
	go func() {
		message, _ := stream.Recv()
		ch <- message
	}()
	return ch
}

// collect returns the remaining messages of stream.
func collect[T any](t *testing.T, stream trevrpc.MessageStream[T]) []T {
	t.Helper()
	var messages []T
	for {
		message, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return messages
		}
		if err != nil {
			t.Fatalf("Recv: %v", err)
		}
		messages = append(messages, message)
	}
}

// validate checks req against the buf.validate rules of its message, as the
// server does before it calls a handler.
func validate(t *testing.T, req proto.Message) error {
	t.Helper()
	desc := req.ProtoReflect().Descriptor()
	check, err := protovalidatetrevrpc.New(nil).ValidateFunc(desc)
	if err != nil {
		t.Fatalf("compile rules of %s: %v", desc.FullName(), err)
	}
	if check == nil {
		t.Fatalf("%s has no rules", desc.FullName())
	}
	return check(req)
}

// assertInvalid checks that the server rejects req with InvalidArgument
// before it reaches the handler.
func assertInvalid(t *testing.T, req proto.Message) {
	t.Helper()
	if got := trevrpc.StatusFromError(validate(t, req)).Code; got != trevrpc.CodeInvalidArgument {
		t.Errorf("validate(%v) code = %v, want InvalidArgument", req, got)
	}
}

// assertValid checks that the server passes req on to the handler.
func assertValid(t *testing.T, req proto.Message) {
	t.Helper()
	if err := validate(t, req); err != nil {
		t.Errorf("validate(%v) = %v, want nil", req, err)
	}
}
