package v1

import (
	"context"

	cursorv1 "trev.zip/template/stack/server/v2/gen/cursor/v1"
)

// Send records a single position as a stream of its own.
func (h *Handler) Send(ctx context.Context, req *cursorv1.SendRequest) (*cursorv1.SendResponse, error) {
	username, err := username(ctx)
	if err != nil {
		return nil, err
	}

	cursor, err := h.newRecorder(username, cursorv1.Source_SOURCE_SEND).record(ctx, req.GetX(), req.GetY())
	if err != nil {
		return nil, err
	}

	resp := &cursorv1.SendResponse{}
	resp.SetCursor(cursor)
	return resp, nil
}
