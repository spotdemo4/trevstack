package v1

import (
	"context"
	_ "embed"

	numberv1 "github.com/spotdemo4/trevstack/server/connect/number/v1"
	"github.com/spotdemo4/trevstack/server/database"
)

//go:embed topnames.sql
var topnamesSQL string

func (h *Handler) TopNames(
	ctx context.Context,
	req *numberv1.TopNamesRequest,
) (*numberv1.TopNamesResponse, error) {
	db := database.FromContext(ctx)

	var startArg, endArg any
	if req.Start != nil {
		startArg = req.Start.AsTime()
	}
	if req.End != nil {
		endArg = req.End.AsTime()
	}

	rows, err := db.QueryContext(ctx, topnamesSQL,
		startArg, startArg,
		endArg, endArg,
		req.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := []*numberv1.TopName{}
	for rows.Next() {
		var name string
		var count int64
		var sum uint64
		var avg float64
		if err := rows.Scan(&name, &count, &sum, &avg); err != nil {
			return nil, err
		}
		names = append(names, &numberv1.TopName{
			Name:    name,
			Count:   count,
			Sum:     sum,
			Average: avg,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &numberv1.TopNamesResponse{Names: names}, nil
}
