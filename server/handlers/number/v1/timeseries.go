package v1

import (
	"context"
	"fmt"
	"time"

	numberv1 "github.com/spotdemo4/trevstack/server/connect/number/v1"
	"github.com/spotdemo4/trevstack/server/database"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const timeseriesSQL = `
SELECT
    %s AS bucket,
    COUNT(*),
    COALESCE(SUM(number), 0),
    COALESCE(AVG(number), 0.0)
FROM numbers
WHERE
    (? IS NULL OR timestamp >= ?)
    AND (? IS NULL OR timestamp <= ?)
GROUP BY bucket
ORDER BY bucket ASC
`

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

	expr, ok := bucketExpr(req.Interval)
	if !ok {
		return nil, fmt.Errorf("invalid interval: %v", req.Interval)
	}

	var startArg, endArg any
	if req.Start != nil {
		startArg = req.Start.AsTime()
	}
	if req.End != nil {
		endArg = req.End.AsTime()
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

		points = append(points, &numberv1.TimeSeriesPoint{
			Bucket:  timestamppb.New(bucket),
			Count:   count,
			Sum:     sum,
			Average: avg,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &numberv1.TimeSeriesResponse{Points: points}, nil
}
