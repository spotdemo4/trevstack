package v1

import (
	"context"
	_ "embed"

	cursorv1 "trev.zip/template/stack/server/v2/gen/cursor/v1"
)

//go:embed heatmap.sql
var heatmapSQL string

// Heatmap counts the recorded positions in a grid of equal cells.
func (h *Handler) Heatmap(
	ctx context.Context,
	req *cursorv1.HeatmapRequest,
) (*cursorv1.HeatmapResponse, error) {
	columns, rows := req.GetColumns(), req.GetRows()
	startArg, endArg := timeRange(req.GetStart(), req.GetEnd())

	result, err := h.db.QueryContext(ctx, heatmapSQL,
		columns, columns,
		rows, rows,
		startArg, startArg,
		endArg, endArg,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = result.Close() }()

	// Every cell is sent, so empty ones are zero.
	counts := make([]int64, columns*rows)
	for result.Next() {
		var column, row uint32
		var count int64
		if err := result.Scan(&column, &row, &count); err != nil {
			return nil, err
		}
		counts[row*columns+column] = count
	}
	if err := result.Err(); err != nil {
		return nil, err
	}

	resp := &cursorv1.HeatmapResponse{}
	resp.SetColumns(columns)
	resp.SetRows(rows)
	resp.SetCounts(counts)
	return resp, nil
}
