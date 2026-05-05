package v1

import (
	"context"
	_ "embed"

	numberv1 "github.com/spotdemo4/trevstack/server/connect/number/v1"
	"github.com/spotdemo4/trevstack/server/database"
)

//go:embed distribution_bounds.sql
var distributionBoundsSQL string

//go:embed distribution_buckets.sql
var distributionBucketsSQL string

func (h *Handler) Distribution(
	ctx context.Context,
	req *numberv1.DistributionRequest,
) (*numberv1.DistributionResponse, error) {
	db := database.FromContext(ctx)

	var startArg, endArg any
	if req.Start != nil {
		startArg = req.Start.AsTime()
	}
	if req.End != nil {
		endArg = req.End.AsTime()
	}

	var lo, hi uint32
	var total int64
	err := db.QueryRowContext(ctx, distributionBoundsSQL,
		startArg, startArg,
		endArg, endArg,
	).Scan(&lo, &hi, &total)
	if err != nil {
		return nil, err
	}

	// No data in range — return empty buckets.
	if total == 0 {
		return &numberv1.DistributionResponse{}, nil
	}

	n := req.BucketCount
	// All values identical — collapse to a single bucket.
	if hi == lo {
		return &numberv1.DistributionResponse{
			Buckets: []*numberv1.DistributionBucket{{
				Lower: lo,
				Upper: hi,
				Count: total,
			}},
		}, nil
	}

	span := uint64(hi-lo) + 1

	rows, err := db.QueryContext(ctx, distributionBucketsSQL,
		n-1, lo, n, span,
		startArg, startArg,
		endArg, endArg,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[uint32]int64, n)
	for rows.Next() {
		var idx uint32
		var count int64
		if err := rows.Scan(&idx, &count); err != nil {
			return nil, err
		}
		counts[idx] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Emit every bucket so the chart has a continuous x-axis.
	buckets := make([]*numberv1.DistributionBucket, 0, n)
	for i := uint32(0); i < n; i++ {
		lower := lo + uint32(uint64(i)*span/uint64(n))
		var upper uint32
		if i == n-1 {
			upper = hi
		} else {
			upper = lo + uint32(uint64(i+1)*span/uint64(n)) - 1
		}
		buckets = append(buckets, &numberv1.DistributionBucket{
			Lower: lower,
			Upper: upper,
			Count: counts[i],
		})
	}

	return &numberv1.DistributionResponse{Buckets: buckets}, nil
}
