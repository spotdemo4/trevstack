// Package v1 implements the cursor.v1 CursorService.
//
// Requests reach the handlers only after the server has checked them against
// the buf.validate rules in their protos, so the handlers rely on those rules.
// That includes each message of a client stream.
package v1

import (
	"cmp"
	"context"
	"database/sql"
	"maps"
	"slices"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	"trev.zip/template/stack/server/v2/auth"
	"trev.zip/template/stack/server/v2/broker"
	cursorv1 "trev.zip/template/stack/server/v2/gen/cursor/v1"
)

// eventBuffer is how many events a watcher may fall behind before it is
// dropped, about 17 seconds of one cursor moving at 60 Hz.
const eventBuffer = 1024

// event is a cursor event that Watch and Share stream. Exactly one field is
// set.
type event struct {
	moved *cursorv1.Cursor
	ended *cursorv1.StreamEnd
}

func (e event) streamID() int64 {
	if e.moved != nil {
		return e.moved.GetStreamId()
	}
	return e.ended.GetStreamId()
}

// Handler implements cursorv1.CursorServiceServer.
type Handler struct {
	db  *sql.DB
	now func() time.Time

	// mu orders publishing against subscribing, so a new watcher sees each
	// event either in its snapshot of live or on its subscription, never both
	// or neither.
	mu sync.Mutex
	// live holds the latest position of each open Track and Share stream.
	live   map[int64]*cursorv1.Cursor
	events *broker.Broker[event]
}

// New returns a Handler that stores cursors in db.
func New(db *sql.DB) *Handler {
	return &Handler{
		db:     db,
		now:    time.Now,
		live:   map[int64]*cursorv1.Cursor{},
		events: broker.New[event](eventBuffer),
	}
}

// Register registers the CursorService on server.
func Register(server *trevrpc.Server, db *sql.DB) {
	cursorv1.RegisterCursorServiceServer(server, New(db))
}

// subscribe returns the latest position of each open stream, oldest stream
// first, and a subscription to the events that follow.
func (h *Handler) subscribe() ([]*cursorv1.Cursor, *broker.Subscription[event]) {
	h.mu.Lock()
	defer h.mu.Unlock()

	snapshot := slices.SortedFunc(maps.Values(h.live), func(a, b *cursorv1.Cursor) int {
		return cmp.Compare(a.GetStreamId(), b.GetStreamId())
	})
	return snapshot, h.events.Subscribe()
}

// publishMove sends cursor to every watcher, and keeps it as its stream's
// latest position while the stream is open.
func (h *Handler) publishMove(cursor *cursorv1.Cursor, open bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if open {
		h.live[cursor.GetStreamId()] = cursor
	}
	h.events.Publish(event{moved: cursor})
}

// publishEnd tells every watcher that a stream ended.
func (h *Handler) publishEnd(end *cursorv1.StreamEnd) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.live, end.GetStreamId())
	h.events.Publish(event{ended: end})
}

// forward yields snapshot, then the events of sub, as responses made by
// respond, until the call's context is done, done receives, or yield returns
// false. It skips the events of the stream that skip names, if any, yields
// the error done receives, if any, and reports whether done received.
func forward[Res any](
	ctx context.Context,
	snapshot []*cursorv1.Cursor,
	sub *broker.Subscription[event],
	skip func() int64,
	done <-chan error,
	respond func(event) Res,
	yield func(Res, error) bool,
) bool {
	var zero Res
	for _, cursor := range snapshot {
		if !yield(respond(event{moved: cursor}), nil) {
			return false
		}
	}
	for {
		select {
		case <-ctx.Done():
			return false
		case err := <-done:
			if err != nil {
				yield(zero, err)
			}
			return true
		case e, ok := <-sub.C():
			if !ok {
				yield(zero, trevrpc.ResourceExhausted("fell too far behind the cursor events"))
				return false
			}
			if skip != nil && e.streamID() == skip() {
				continue
			}
			if !yield(respond(e), nil) {
				return false
			}
		}
	}
}

// username returns the user of the session the call carries.
func username(ctx context.Context) (string, error) {
	username, ok := auth.UsernameFromContext(ctx)
	if !ok {
		return "", trevrpc.Unauthenticated("a session is required")
	}
	return username, nil
}

// timeRange returns optional timestamp bounds as Unix milliseconds query
// arguments, nil when unset.
func timeRange(start, end *timestamppb.Timestamp) (startArg, endArg any) {
	if start != nil {
		startArg = start.AsTime().UnixMilli()
	}
	if end != nil {
		endArg = end.AsTime().UnixMilli()
	}
	return startArg, endArg
}

// timestamp converts Unix milliseconds to a Timestamp.
func timestamp(ms int64) *timestamppb.Timestamp {
	return timestamppb.New(time.UnixMilli(ms))
}
