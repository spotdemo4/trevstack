package v1

import (
	"context"
	_ "embed"

	numberv1 "github.com/spotdemo4/trevstack/server/connect/number/v1"
	"github.com/spotdemo4/trevstack/server/database"
)

//go:embed summary.sql
var summarySQL string

func (h *Handler) Summary(
	ctx context.Context,
	req *numberv1.SummaryRequest,
) (*numberv1.SummaryResponse, error) {
	db := database.FromContext(ctx)

	var startArg, endArg any
	if start := req.GetStart(); start != nil {
		startArg = start.AsTime()
	}
	if end := req.GetEnd(); end != nil {
		endArg = end.AsTime()
	}

	var totalCount int64
	var totalSum uint64
	var average float64
	var min, max, distinctNames uint32

	err := db.QueryRowContext(ctx, summarySQL,
		startArg, startArg,
		endArg, endArg,
	).Scan(
		&totalCount,
		&totalSum,
		&average,
		&min,
		&max,
		&distinctNames,
	)
	if err != nil {
		return nil, err
	}

	resp := &numberv1.SummaryResponse{}
	resp.SetTotalCount(totalCount)
	resp.SetTotalSum(totalSum)
	resp.SetAverage(average)
	resp.SetMin(min)
	resp.SetMax(max)
	resp.SetDistinctNames(distinctNames)

	return resp, nil
}
