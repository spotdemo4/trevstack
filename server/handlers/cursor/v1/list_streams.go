package v1

import (
	"context"
	_ "embed"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	cursorv1 "trev.zip/template/stack/server/v2/gen/cursor/v1"
)

//go:embed list_streams.sql
var listStreamsSQL string

// ListStreams streams a summary of each recorded stream, newest first.
func (h *Handler) ListStreams(
	ctx context.Context,
	req *cursorv1.ListStreamsRequest,
) (trevrpc.MessageStream[*cursorv1.ListStreamsResponse], error) {
	var usernameArg, sourceArg any
	if req.HasUsername() {
		usernameArg = req.GetUsername()
	}
	if req.HasSource() {
		sourceArg = int32(req.GetSource())
	}
	startArg, endArg := timeRange(req.GetStart(), req.GetEnd())

	return trevrpc.StreamFromSeq(func(yield func(*cursorv1.ListStreamsResponse, error) bool) {
		rows, err := h.db.QueryContext(ctx, listStreamsSQL,
			usernameArg, usernameArg,
			sourceArg, sourceArg,
			startArg, startArg,
			endArg, endArg,
		)
		if err != nil {
			yield(nil, err)
			return
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var id, start, end, count int64
			var username string
			var source int32
			var distance float64
			if err := rows.Scan(&id, &username, &source, &start, &end, &count, &distance); err != nil {
				yield(nil, err)
				return
			}

			resp := &cursorv1.ListStreamsResponse{}
			resp.SetStream(cursorv1.StreamSummary_builder{
				Id:       id,
				Username: username,
				Source:   cursorv1.Source(source),
				Start:    timestamp(start),
				End:      timestamp(end),
				Count:    count,
				Distance: distance,
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
