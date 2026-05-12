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
	if start := req.GetStart(); start != nil {
		startArg = start.AsTime()
	}
	if end := req.GetEnd(); end != nil {
		endArg = end.AsTime()
	}

	rows, err := db.QueryContext(ctx, topnamesSQL,
		startArg, startArg,
		endArg, endArg,
		req.GetLimit(),
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
		topName := &numberv1.TopName{}
		topName.SetName(name)
		topName.SetCount(count)
		topName.SetSum(sum)
		topName.SetAverage(avg)
		names = append(names, topName)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	resp := &numberv1.TopNamesResponse{}
	resp.SetNames(names)
	return resp, nil
}
