package v1

import (
	"context"
	_ "embed"

	cursorv1 "trev.zip/template/stack/server/v3/gen/cursor/v1"
)

//go:embed topusers.sql
var topusersSQL string

// TopUsers returns the users with the most recorded positions.
func (h *Handler) TopUsers(
	ctx context.Context,
	req *cursorv1.TopUsersRequest,
) (*cursorv1.TopUsersResponse, error) {
	startArg, endArg := timeRange(req.GetStart(), req.GetEnd())

	rows, err := h.db.QueryContext(ctx, topusersSQL,
		startArg, startArg,
		endArg, endArg,
		req.GetLimit(),
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	users := []*cursorv1.TopUser{}
	for rows.Next() {
		var username string
		var count, streams int64
		var distance float64
		if err := rows.Scan(&username, &count, &streams, &distance); err != nil {
			return nil, err
		}
		users = append(users, cursorv1.TopUser_builder{
			Username: username,
			Count:    count,
			Streams:  streams,
			Distance: distance,
		}.Build())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	resp := &cursorv1.TopUsersResponse{}
	resp.SetUsers(users)
	return resp, nil
}
