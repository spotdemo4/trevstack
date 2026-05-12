package v1

import (
	"context"
	_ "embed"
	"time"

	"connectrpc.com/connect"
	numberv1 "github.com/spotdemo4/trevstack/server/connect/number/v1"
	"github.com/spotdemo4/trevstack/server/database"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//go:embed list.sql
var listSQL string

func (h *Handler) List(
	ctx context.Context,
	req *numberv1.ListRequest,
	stream *connect.ServerStream[numberv1.ListResponse],
) error {
	db := database.FromContext(ctx)

	var nameArg, minArg, maxArg, startArg, endArg any
	if req.HasName() {
		nameArg = req.GetName()
	}
	if req.HasMin() {
		minArg = req.GetMin()
	}
	if req.HasMax() {
		maxArg = req.GetMax()
	}
	if start := req.GetStart(); start != nil {
		startArg = start.AsTime()
	}
	if end := req.GetEnd(); end != nil {
		endArg = end.AsTime()
	}

	rows, err := db.QueryContext(ctx, listSQL,
		nameArg, nameArg,
		minArg, minArg,
		maxArg, maxArg,
		startArg, startArg,
		endArg, endArg,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var rowid int64
		var ts time.Time
		var name string
		var number uint32
		if err := rows.Scan(&rowid, &ts, &name, &number); err != nil {
			return err
		}

		item := &numberv1.Item{}
		item.SetTimestamp(timestamppb.New(ts))
		item.SetName(name)
		item.SetNumber(number)

		resp := &numberv1.ListResponse{}
		resp.SetItem(item)
		if err := stream.Send(resp); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}
