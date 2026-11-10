package v1

import (
	"context"
	_ "embed"

	cursorv1 "trev.zip/template/stack/server/v3/gen/cursor/v1"
)

//go:embed summary.sql
var summarySQL string

// Summary returns aggregate statistics of the recorded positions.
func (h *Handler) Summary(
	ctx context.Context,
	req *cursorv1.SummaryRequest,
) (*cursorv1.SummaryResponse, error) {
	startArg, endArg := timeRange(req.GetStart(), req.GetEnd())

	var count, streams, users int64
	var distance float64
	err := h.db.QueryRowContext(ctx, summarySQL,
		startArg, startArg,
		endArg, endArg,
	).Scan(&count, &streams, &users, &distance)
	if err != nil {
		return nil, err
	}

	resp := &cursorv1.SummaryResponse{}
	resp.SetCount(count)
	resp.SetStreams(streams)
	resp.SetUsers(users)
	resp.SetDistance(distance)
	if streams > 0 {
		resp.SetAverageCount(float64(count) / float64(streams))
		resp.SetAverageDistance(distance / float64(streams))
	}
	return resp, nil
}
