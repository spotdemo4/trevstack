package v1

import (
	"context"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	cursorv1 "trev.zip/template/stack/server/v2/gen/cursor/v1"
)

// Track records the positions the client streams, and summarizes them once
// the client ends the stream.
func (h *Handler) Track(
	ctx context.Context,
	requests trevrpc.MessageStream[*cursorv1.TrackRequest],
) (*cursorv1.TrackResponse, error) {
	username, err := username(ctx)
	if err != nil {
		return nil, err
	}

	rec := h.newRecorder(username, cursorv1.Source_SOURCE_TRACK)
	defer rec.end()

	if err := recordAll(ctx, rec, requests); err != nil {
		return nil, err
	}

	resp := &cursorv1.TrackResponse{}
	resp.SetStream(rec.summary())
	return resp, nil
}
