package v1

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	numberv1 "trev.zip/trev/stack/server/connect/number/v1"
	"trev.zip/trev/stack/server/database"
)

//go:embed timeseries.sql
var timeseriesSQL string

// bucketExpr returns a SQL expression that floors `timestamp` to the start of
// the chosen interval and renders it as an RFC3339 UTC string.
func bucketExpr(interval numberv1.TimeInterval) (string, bool) {
	switch interval {
	case numberv1.TimeInterval_TIME_INTERVAL_HOUR:
		return `strftime('%Y-%m-%dT%H:00:00Z', timestamp)`, true
	case numberv1.TimeInterval_TIME_INTERVAL_DAY:
		return `strftime('%Y-%m-%dT00:00:00Z', timestamp)`, true
	case numberv1.TimeInterval_TIME_INTERVAL_WEEK:
		// Floor to the most recent Sunday by subtracting strftime('%w') days.
		return `date(timestamp, '-' || strftime('%w', timestamp) || ' days') || 'T00:00:00Z'`, true
	case numberv1.TimeInterval_TIME_INTERVAL_MONTH:
		return `strftime('%Y-%m-01T00:00:00Z', timestamp)`, true
	default:
		return "", false
	}
}

func (h *Handler) TimeSeries(
	ctx context.Context,
	req *numberv1.TimeSeriesRequest,
) (*numberv1.TimeSeriesResponse, error) {
	db := database.FromContext(ctx)

	interval := req.GetInterval()
	expr, ok := bucketExpr(interval)
	if !ok {
		return nil, fmt.Errorf("invalid interval: %v", interval)
	}

	var startArg, endArg any
	if start := req.GetStart(); start != nil {
		startArg = start.AsTime()
	}
	if end := req.GetEnd(); end != nil {
		endArg = end.AsTime()
	}

	rows, err := db.QueryContext(ctx, fmt.Sprintf(timeseriesSQL, expr),
		startArg, startArg,
		endArg, endArg,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := []*numberv1.TimeSeriesPoint{}
	for rows.Next() {
		var bucketStr string
		var count int64
		var sum uint64
		var avg float64
		if err := rows.Scan(&bucketStr, &count, &sum, &avg); err != nil {
			return nil, err
		}

		bucket, err := time.Parse(time.RFC3339, bucketStr)
		if err != nil {
			return nil, err
		}

		point := &numberv1.TimeSeriesPoint{}
		point.SetBucket(timestamppb.New(bucket))
		point.SetCount(count)
		point.SetSum(sum)
		point.SetAverage(avg)
		points = append(points, point)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	resp := &numberv1.TimeSeriesResponse{}
	resp.SetPoints(points)
	return resp, nil
}
