package v1

import (
	"context"
	_ "embed"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	numberv1 "trev.zip/template/stack/server/v2/gen/number/v1"
)

//go:embed list.sql
var listSQL string

// List streams the recorded numbers.
func (h *Handler) List(
	ctx context.Context,
	req *numberv1.ListRequest,
) (trevrpc.MessageStream[*numberv1.ListResponse], error) {
	if err := validateListRequest(req); err != nil {
		return nil, err
	}

	var nameArg, minArg, maxArg, startArg, endArg any
	if req.HasName() {
		nameArg = req.GetName()
	}
	if req.HasMinimum() {
		minArg = req.GetMinimum()
	}
	if req.HasMaximum() {
		maxArg = req.GetMaximum()
	}
	if start := req.GetStart(); start != nil {
		startArg = start.AsTime()
	}
	if end := req.GetEnd(); end != nil {
		endArg = end.AsTime()
	}

	// The handler's context lasts until the stream ends, so the rows are read
	// as the client receives them.
	return trevrpc.StreamFromSeq(func(yield func(*numberv1.ListResponse, error) bool) {
		rows, err := h.db.QueryContext(ctx, listSQL,
			nameArg, nameArg,
			minArg, minArg,
			maxArg, maxArg,
			startArg, startArg,
			endArg, endArg,
		)
		if err != nil {
			yield(nil, err)
			return
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var ts time.Time
			var name string
			var number uint32
			if err := rows.Scan(&ts, &name, &number); err != nil {
				yield(nil, err)
				return
			}

			item := &numberv1.Item{}
			item.SetTimestamp(timestamppb.New(ts))
			item.SetName(name)
			item.SetNumber(number)

			resp := &numberv1.ListResponse{}
			resp.SetItem(item)
			if !yield(resp, nil) {
				return
			}
		}
		if err := rows.Err(); err != nil {
			yield(nil, err)
		}
	}), nil
}
