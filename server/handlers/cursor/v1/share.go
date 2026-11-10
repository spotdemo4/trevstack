package v1

import (
	"context"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	cursorv1 "trev.zip/template/stack/server/v3/gen/cursor/v1"
	"trev.zip/template/stack/server/v3/logger"
)

// Share records the positions the client streams while streaming every other
// stream's cursor events back, until the client ends its half of the stream.
func (h *Handler) Share(
	ctx context.Context,
	requests trevrpc.MessageStream[*cursorv1.ShareRequest],
) (trevrpc.MessageStream[*cursorv1.ShareResponse], error) {
	username, err := username(ctx)
	if err != nil {
		return nil, err
	}

	return trevrpc.StreamFromSeq(func(yield func(*cursorv1.ShareResponse, error) bool) {
		snapshot, sub := h.subscribe()
		defer sub.Close()

		// The client's positions are read alongside the events sent back.
		rec := h.newRecorder(username, cursorv1.Source_SOURCE_SHARE)
		recorded := make(chan error, 1)
		go func() { recorded <- recordAll(ctx, rec, requests) }()

		reading := true
		defer func() {
			if reading {
				// Unblock the reader, and wait for it, so the stream's end
				// follows its last position.
				_ = requests.Close()
				// The call is over, so the reader's error has no one to go to.
				if err := <-recorded; err != nil {
					logger.FromContext(ctx).DebugContext(ctx, "stopped recording a shared stream", "error", err)
				}
			}
			rec.end()
		}()

		reading = !forward(ctx, snapshot, sub, rec.id, recorded, shareResponse, yield)
	}), nil
}

func shareResponse(e event) *cursorv1.ShareResponse {
	resp := &cursorv1.ShareResponse{}
	if e.moved != nil {
		resp.SetMoved(e.moved)
	} else {
		resp.SetEnded(e.ended)
	}
	return resp
}
