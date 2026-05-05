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
	if req.Start != nil {
		startArg = req.Start.AsTime()
	}
	if req.End != nil {
		endArg = req.End.AsTime()
	}

	resp := &numberv1.SummaryResponse{}
	err := db.QueryRowContext(ctx, summarySQL,
		startArg, startArg,
		endArg, endArg,
	).Scan(
		&resp.TotalCount,
		&resp.TotalSum,
		&resp.Average,
		&resp.Min,
		&resp.Max,
		&resp.DistinctNames,
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
