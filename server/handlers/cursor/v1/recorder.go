package v1

import (
	"context"
	_ "embed"
	"math"
	"sync/atomic"
	"time"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	cursorv1 "trev.zip/template/stack/server/v2/gen/cursor/v1"
)

//go:embed insert_stream.sql
var insertStreamSQL string

//go:embed insert_cursor.sql
var insertCursorSQL string

// position is a request that carries a position.
type position interface {
	GetX() float64
	GetY() float64
}

// recorder records the positions of one stream. Its stream is created with
// its first position, so a stream that sends none leaves no trace.
//
// Only one goroutine may call record and end, but any may call id.
type recorder struct {
	h        *Handler
	username string
	source   cursorv1.Source

	streamID atomic.Int64
	count    int64
	distance float64
	start    time.Time
	last     time.Time
	x, y     float64
}

func (h *Handler) newRecorder(username string, source cursorv1.Source) *recorder {
	return &recorder{h: h, username: username, source: source}
}

// id returns the stream's ID, or 0 before its first position.
func (r *recorder) id() int64 {
	return r.streamID.Load()
}

// open reports whether the stream's cursor stays on watchers' pads until the
// stream ends, which a single Send never does.
func (r *recorder) open() bool {
	return r.source != cursorv1.Source_SOURCE_SEND
}

// record sends the position to watchers, then stores it.
func (r *recorder) record(ctx context.Context, x, y float64) (*cursorv1.Cursor, error) {
	now := time.UnixMilli(r.h.now().UnixMilli())

	if r.count == 0 {
		var id int64
		err := r.h.db.QueryRowContext(ctx, insertStreamSQL, r.source, r.username).Scan(&id)
		if err != nil {
			return nil, err
		}
		r.streamID.Store(id)
		r.start = now
	}

	var step float64
	if r.count > 0 {
		step = math.Hypot(x-r.x, y-r.y)
	}

	cursor := cursorv1.Cursor_builder{
		StreamId:  r.id(),
		Username:  r.username,
		Source:    r.source,
		X:         x,
		Y:         y,
		Timestamp: timestamp(now.UnixMilli()),
	}.Build()

	// Watchers see the position without waiting for the database.
	r.h.publishMove(cursor, r.open())

	_, err := r.h.db.ExecContext(ctx, insertCursorSQL, r.id(), x, y, step, now.UnixMilli())
	if err != nil {
		return nil, err
	}

	r.count++
	r.distance += step
	r.last = now
	r.x, r.y = x, y
	return cursor, nil
}

// recordAll records every position of requests until the client ends the
// stream.
func recordAll[Req position](ctx context.Context, r *recorder, requests trevrpc.MessageStream[Req]) error {
	for req, err := range trevrpc.Messages(requests) {
		if err != nil {
			return err
		}
		if _, err := r.record(ctx, req.GetX(), req.GetY()); err != nil {
			return err
		}
	}
	return nil
}

// end tells watchers that an open stream with positions ended.
func (r *recorder) end() {
	if r.count == 0 || !r.open() {
		return
	}
	r.h.publishEnd(cursorv1.StreamEnd_builder{
		StreamId: r.id(),
		Username: r.username,
	}.Build())
}

// summary returns the stream's positions in aggregate, or nil before its
// first position.
func (r *recorder) summary() *cursorv1.StreamSummary {
	if r.count == 0 {
		return nil
	}
	return cursorv1.StreamSummary_builder{
		Id:       r.id(),
		Username: r.username,
		Source:   r.source,
		Start:    timestamp(r.start.UnixMilli()),
		End:      timestamp(r.last.UnixMilli()),
		Count:    r.count,
		Distance: r.distance,
	}.Build()
}
