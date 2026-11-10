package v1

import (
	"context"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	cursorv1 "trev.zip/template/stack/server/v3/gen/cursor/v1"
)

// Watch streams every cursor event until the client cancels the call.
func (h *Handler) Watch(
	ctx context.Context,
	_ *cursorv1.WatchRequest,
) (trevrpc.MessageStream[*cursorv1.WatchResponse], error) {
	// The subscription is made once the client reads the stream, so a stream
	// that is closed unread leaves none behind.
	return trevrpc.StreamFromSeq(func(yield func(*cursorv1.WatchResponse, error) bool) {
		snapshot, sub := h.subscribe()
		defer sub.Close()

		forward(ctx, snapshot, sub, nil, nil, watchResponse, yield)
	}), nil
}

func watchResponse(e event) *cursorv1.WatchResponse {
	resp := &cursorv1.WatchResponse{}
	if e.moved != nil {
		resp.SetMoved(e.moved)
	} else {
		resp.SetEnded(e.ended)
	}
	return resp
}
