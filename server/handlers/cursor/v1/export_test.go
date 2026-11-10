package v1

import "time"

// SetNow replaces the clock that timestamps recorded positions.
func (h *Handler) SetNow(now func() time.Time) {
	h.now = now
}

// Watchers returns the number of open Watch and Share streams.
func (h *Handler) Watchers() int {
	return h.events.Len()
}
