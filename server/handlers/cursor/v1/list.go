package v1

import (
	"context"
	_ "embed"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	cursorv1 "trev.zip/template/stack/server/v3/gen/cursor/v1"
)

//go:embed list.sql
var listSQL string

// List streams the recorded positions, newest first.
func (h *Handler) List(
	ctx context.Context,
	req *cursorv1.ListRequest,
) (trevrpc.MessageStream[*cursorv1.ListResponse], error) {
	var usernameArg, sourceArg, streamArg, minXArg, maxXArg, minYArg, maxYArg any
	if req.HasUsername() {
		usernameArg = req.GetUsername()
	}
	if req.HasSource() {
		sourceArg = int32(req.GetSource())
	}
	if req.HasStreamId() {
		streamArg = req.GetStreamId()
	}
	if req.HasMinX() {
		minXArg = req.GetMinX()
	}
	if req.HasMaxX() {
		maxXArg = req.GetMaxX()
	}
	if req.HasMinY() {
		minYArg = req.GetMinY()
	}
	if req.HasMaxY() {
		maxYArg = req.GetMaxY()
	}
	startArg, endArg := timeRange(req.GetStart(), req.GetEnd())

	// The handler's context lasts until the stream ends, so the rows are read
	// as the client receives them.
	return trevrpc.StreamFromSeq(func(yield func(*cursorv1.ListResponse, error) bool) {
		rows, err := h.db.QueryContext(ctx, listSQL,
			usernameArg, usernameArg,
			sourceArg, sourceArg,
			streamArg, streamArg,
			minXArg, minXArg,
			maxXArg, maxXArg,
			minYArg, minYArg,
			maxYArg, maxYArg,
			startArg, startArg,
			endArg, endArg,
		)
		if err != nil {
			yield(nil, err)
			return
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var streamID, ms int64
			var username string
			var source int32
			var x, y float64
			if err := rows.Scan(&streamID, &username, &source, &x, &y, &ms); err != nil {
				yield(nil, err)
				return
			}

			resp := &cursorv1.ListResponse{}
			resp.SetCursor(cursorv1.Cursor_builder{
				StreamId:  streamID,
				Username:  username,
				Source:    cursorv1.Source(source),
				X:         x,
				Y:         y,
				Timestamp: timestamp(ms),
			}.Build())
			if !yield(resp, nil) {
				return
			}
		}
		if err := rows.Err(); err != nil {
			yield(nil, err)
		}
	}), nil
}
