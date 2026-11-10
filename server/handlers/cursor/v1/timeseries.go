package v1

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	cursorv1 "trev.zip/template/stack/server/v3/gen/cursor/v1"
)

//go:embed timeseries.sql
var timeseriesSQL string

// seconds is the timestamp column, Unix milliseconds, as Unix seconds for
// SQLite's date functions.
const seconds = `c.timestamp / 1000, 'unixepoch'`

// bucketExpr returns a SQL expression that floors the timestamp to the start
// of the chosen interval and renders it as an RFC3339 UTC string.
func bucketExpr(interval cursorv1.TimeInterval) (string, bool) {
	switch interval {
	case cursorv1.TimeInterval_TIME_INTERVAL_MINUTE:
		return `strftime('%Y-%m-%dT%H:%M:00Z', ` + seconds + `)`, true
	case cursorv1.TimeInterval_TIME_INTERVAL_HOUR:
		return `strftime('%Y-%m-%dT%H:00:00Z', ` + seconds + `)`, true
	case cursorv1.TimeInterval_TIME_INTERVAL_DAY:
		return `strftime('%Y-%m-%dT00:00:00Z', ` + seconds + `)`, true
	case cursorv1.TimeInterval_TIME_INTERVAL_WEEK:
		// Floor to the most recent Sunday by subtracting strftime('%w') days.
		return `date(` + seconds + `, '-' || strftime('%w', ` + seconds + `) || ' days') || 'T00:00:00Z'`, true
	case cursorv1.TimeInterval_TIME_INTERVAL_MONTH:
		return `strftime('%Y-%m-01T00:00:00Z', ` + seconds + `)`, true
	default:
		return "", false
	}
}

// TimeSeries returns the recorded positions grouped into time buckets.
func (h *Handler) TimeSeries(
	ctx context.Context,
	req *cursorv1.TimeSeriesRequest,
) (*cursorv1.TimeSeriesResponse, error) {
	interval := req.GetInterval()
	expr, ok := bucketExpr(interval)
	if !ok {
		return nil, fmt.Errorf("invalid interval: %v", interval)
	}
	startArg, endArg := timeRange(req.GetStart(), req.GetEnd())

	rows, err := h.db.QueryContext(ctx, fmt.Sprintf(timeseriesSQL, expr),
		startArg, startArg,
		endArg, endArg,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	points := []*cursorv1.TimeSeriesPoint{}
	for rows.Next() {
		var bucketStr string
		var count, streams, users int64
		if err := rows.Scan(&bucketStr, &count, &streams, &users); err != nil {
			return nil, err
		}

		bucket, err := time.Parse(time.RFC3339, bucketStr)
		if err != nil {
			return nil, err
		}

		points = append(points, cursorv1.TimeSeriesPoint_builder{
			Bucket:  timestamppb.New(bucket),
			Count:   count,
			Streams: streams,
			Users:   users,
		}.Build())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	resp := &cursorv1.TimeSeriesResponse{}
	resp.SetPoints(points)
	return resp, nil
}
