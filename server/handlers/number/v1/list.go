package v1

import (
	"context"
	_ "embed"
	"time"

	numberv1 "github.com/spotdemo4/trevstack/server/connect/number/v1"
	"github.com/spotdemo4/trevstack/server/database"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//go:embed list.sql
var listSQL string

//go:embed count.sql
var countSQL string

const listPageSize = 50

func (h *Handler) List(
	ctx context.Context,
	req *numberv1.ListRequest,
) (*numberv1.ListResponse, error) {
	db := database.FromContext(ctx)

	var nameArg, minArg, maxArg, startArg, endArg, cursorArg any
	if req.Name != nil {
		nameArg = *req.Name
	}
	if req.Min != nil {
		minArg = *req.Min
	}
	if req.Max != nil {
		maxArg = *req.Max
	}
	if req.Start != nil {
		startArg = req.Start.AsTime()
	}
	if req.End != nil {
		endArg = req.End.AsTime()
	}
	if req.Cursor != nil {
		cursorArg = *req.Cursor
	}

	rows, err := db.QueryContext(ctx, listSQL,
		nameArg, nameArg,
		minArg, minArg,
		maxArg, maxArg,
		startArg, startArg,
		endArg, endArg,
		cursorArg, cursorArg,
		listPageSize,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*numberv1.Item, 0, listPageSize)
	rowIDs := make([]int64, 0, listPageSize)
	for rows.Next() {
		var rowid int64
		var ts time.Time
		var name string
		var number uint32
		if err := rows.Scan(&rowid, &ts, &name, &number); err != nil {
			return nil, err
		}

		items = append(items, &numberv1.Item{
			Timestamp: timestamppb.New(ts),
			Name:      name,
			Number:    number,
		})
		rowIDs = append(rowIDs, rowid)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalCount := int64(0)
	if err := db.QueryRowContext(ctx, countSQL,
		nameArg, nameArg,
		minArg, minArg,
		maxArg, maxArg,
		startArg, startArg,
		endArg, endArg,
	).Scan(&totalCount); err != nil {
		return nil, err
	}

	nextCursor := int64(0)
	if len(items) > 0 {
		nextCursor = rowIDs[len(items)-1]
	} else if req.Cursor != nil {
		nextCursor = *req.Cursor
	}

	resp := &numberv1.ListResponse{
		Items:      items,
		TotalCount: totalCount,
		NextCursor: nextCursor,
	}

	return resp, nil
}
