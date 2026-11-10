package v1_test

import (
	"context"
	"math"
	"testing"
	"time"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	cursorv1 "trev.zip/template/stack/server/v3/gen/cursor/v1"
)

func send(t *testing.T, h *handlerT, username string, x, y float64) *cursorv1.Cursor {
	t.Helper()
	resp, err := h.Send(as(username), cursorv1.SendRequest_builder{X: &x, Y: &y}.Build())
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	return resp.GetCursor()
}

func storedPositions(t *testing.T, h *handlerT) int {
	t.Helper()
	return len(list(t, h, &cursorv1.ListRequest{}))
}

func TestSend(t *testing.T) {
	t.Run("records a position as a stream of its own", func(t *testing.T) {
		handler, db := newTest(t)

		first := send(t, handler, "alice", 0.25, 0.75)
		second := send(t, handler, "alice", 0.5, 0.5)

		if first.GetUsername() != "alice" || first.GetSource() != cursorv1.Source_SOURCE_SEND {
			t.Errorf("cursor = %v, want alice's from Send", first)
		}
		if first.GetX() != 0.25 || first.GetY() != 0.75 {
			t.Errorf("position = (%v, %v), want (0.25, 0.75)", first.GetX(), first.GetY())
		}
		if !first.GetTimestamp().AsTime().Equal(base.Add(time.Second)) {
			t.Errorf("timestamp = %v, want %v", first.GetTimestamp().AsTime(), base.Add(time.Second))
		}
		if first.GetStreamId() == second.GetStreamId() {
			t.Errorf("both sends share stream %d", first.GetStreamId())
		}

		var distance float64
		if err := db.QueryRowContext(t.Context(), `SELECT SUM(distance) FROM cursors`).Scan(&distance); err != nil {
			t.Fatal(err)
		}
		if distance != 0 {
			t.Errorf("distance = %v, want 0", distance)
		}
	})

	t.Run("requires a session's user", func(t *testing.T) {
		handler, _ := newTest(t)
		x, y := 0.5, 0.5
		_, err := handler.Send(context.Background(), cursorv1.SendRequest_builder{X: &x, Y: &y}.Build())
		if got := trevrpc.StatusFromError(err).Code; got != trevrpc.CodeUnauthenticated {
			t.Errorf("code = %v, want Unauthenticated", got)
		}
	})

	t.Run("validates the position", func(t *testing.T) {
		assertValid(t, cursorv1.SendRequest_builder{X: new(0.0), Y: new(1.0)}.Build())
		assertInvalid(t, cursorv1.SendRequest_builder{Y: new(0.5)}.Build())
		assertInvalid(t, cursorv1.SendRequest_builder{X: new(0.5)}.Build())
		assertInvalid(t, cursorv1.SendRequest_builder{X: new(-0.1), Y: new(0.5)}.Build())
		assertInvalid(t, cursorv1.SendRequest_builder{X: new(0.5), Y: new(1.1)}.Build())
		assertInvalid(t, cursorv1.SendRequest_builder{X: new(math.NaN()), Y: new(0.5)}.Build())
		assertInvalid(t, cursorv1.TrackRequest_builder{X: new(2.0), Y: new(0.5)}.Build())
		assertInvalid(t, cursorv1.ShareRequest_builder{X: new(0.5)}.Build())
	})
}

func TestTrack(t *testing.T) {
	t.Run("records the stream and summarizes it", func(t *testing.T) {
		handler, _ := newTest(t)

		resp, err := handler.Track(as("alice"), trevrpc.StreamOf(
			cursorv1.TrackRequest_builder{X: new(0.0), Y: new(0.0)}.Build(),
			cursorv1.TrackRequest_builder{X: new(0.3), Y: new(0.4)}.Build(),
			cursorv1.TrackRequest_builder{X: new(0.3), Y: new(0.0)}.Build(),
		))
		if err != nil {
			t.Fatalf("Track: %v", err)
		}

		stream := resp.GetStream()
		if stream.GetCount() != 3 {
			t.Errorf("count = %d, want 3", stream.GetCount())
		}
		if math.Abs(stream.GetDistance()-0.9) > 1e-9 {
			t.Errorf("distance = %v, want 0.9", stream.GetDistance())
		}
		if got := stream.GetEnd().AsTime().Sub(stream.GetStart().AsTime()); got != 2*time.Second {
			t.Errorf("duration = %v, want 2s", got)
		}
		if stream.GetUsername() != "alice" || stream.GetSource() != cursorv1.Source_SOURCE_TRACK {
			t.Errorf("stream = %v, want alice's from Track", stream)
		}
		if got := storedPositions(t, handler); got != 3 {
			t.Errorf("stored %d positions, want 3", got)
		}
	})

	t.Run("an empty stream records nothing", func(t *testing.T) {
		handler, db := newTest(t)

		resp, err := handler.Track(as("alice"), trevrpc.StreamOf[*cursorv1.TrackRequest]())
		if err != nil {
			t.Fatalf("Track: %v", err)
		}
		if resp.HasStream() {
			t.Errorf("stream = %v, want unset", resp.GetStream())
		}
		var streams int
		if err := db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM cursor_streams`).Scan(&streams); err != nil {
			t.Fatal(err)
		}
		if streams != 0 {
			t.Errorf("stored %d streams, want 0", streams)
		}
	})
}

func TestWatch(t *testing.T) {
	t.Run("streams moves and ends as they happen", func(t *testing.T) {
		handler, _ := newTest(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		watch, err := handler.Watch(ctx, &cursorv1.WatchRequest{})
		if err != nil {
			t.Fatalf("Watch: %v", err)
		}
		defer func() { _ = watch.Close() }()

		requests := newPipe[*cursorv1.TrackRequest]()
		tracked := make(chan error, 1)
		go func() {
			_, err := handler.Track(as("alice"), requests)
			tracked <- err
		}()

		// The watcher subscribes on its first Recv, so wait for it before
		// the first position.
		first := recvAsync(watch)
		for handler.Watchers() == 0 {
			time.Sleep(time.Millisecond)
		}

		requests.Send(cursorv1.TrackRequest_builder{X: new(0.1), Y: new(0.2)}.Build())
		moved := (<-first).GetMoved()
		if moved.GetUsername() != "alice" || moved.GetX() != 0.1 || moved.GetY() != 0.2 {
			t.Errorf("moved = %v, want alice at (0.1, 0.2)", moved)
		}

		sent := send(t, handler, "bob", 0.9, 0.9)
		if got := recv(t, watch).GetMoved(); got.GetStreamId() != sent.GetStreamId() {
			t.Errorf("moved stream = %d, want bob's %d", got.GetStreamId(), sent.GetStreamId())
		}

		requests.CloseSend()
		if err := <-tracked; err != nil {
			t.Fatalf("Track: %v", err)
		}
		ended := recv(t, watch).GetEnded()
		if ended.GetStreamId() != moved.GetStreamId() || ended.GetUsername() != "alice" {
			t.Errorf("ended = %v, want alice's stream %d", ended, moved.GetStreamId())
		}

		cancel()
		if _, err := watch.Recv(); err == nil {
			t.Error("Recv after cancel succeeded")
		}
	})

	t.Run("starts with the latest position of each open stream", func(t *testing.T) {
		handler, _ := newTest(t)

		requests := newPipe[*cursorv1.TrackRequest]()
		tracked := make(chan error, 1)
		go func() {
			_, err := handler.Track(as("alice"), requests)
			tracked <- err
		}()
		requests.Send(cursorv1.TrackRequest_builder{X: new(0.1), Y: new(0.1)}.Build())
		requests.Send(cursorv1.TrackRequest_builder{X: new(0.2), Y: new(0.2)}.Build())
		// Sends are not open streams, so they are not replayed.
		send(t, handler, "bob", 0.9, 0.9)
		// Wait until the second position was recorded.
		requests.Send(cursorv1.TrackRequest_builder{X: new(0.3), Y: new(0.3)}.Build())

		watch, err := handler.Watch(context.Background(), &cursorv1.WatchRequest{})
		if err != nil {
			t.Fatalf("Watch: %v", err)
		}
		defer func() { _ = watch.Close() }()

		moved := recv(t, watch).GetMoved()
		if moved.GetUsername() != "alice" || moved.GetX() < 0.2 {
			t.Errorf("first event = %v, want alice's latest position", moved)
		}

		requests.CloseSend()
		if err := <-tracked; err != nil {
			t.Fatalf("Track: %v", err)
		}
	})

	t.Run("fails a watcher that falls too far behind", func(t *testing.T) {
		handler, _ := newTest(t)

		requests := newPipe[*cursorv1.TrackRequest]()
		go func() { _, _ = handler.Track(as("alice"), requests) }()
		defer requests.CloseSend()
		requests.Send(cursorv1.TrackRequest_builder{X: new(0.1), Y: new(0.1)}.Build())
		requests.Send(cursorv1.TrackRequest_builder{X: new(0.2), Y: new(0.2)}.Build())

		watch, err := handler.Watch(context.Background(), &cursorv1.WatchRequest{})
		if err != nil {
			t.Fatalf("Watch: %v", err)
		}
		defer func() { _ = watch.Close() }()
		// The open stream's position arrives once the watcher subscribed.
		recv(t, watch)

		for range 2000 {
			send(t, handler, "bob", 0.5, 0.5)
		}

		for {
			_, err := watch.Recv()
			if err == nil {
				continue
			}
			if got := trevrpc.StatusFromError(err).Code; got != trevrpc.CodeResourceExhausted {
				t.Errorf("code = %v, want ResourceExhausted", got)
			}
			break
		}
	})
}

func TestShare(t *testing.T) {
	handler, _ := newTest(t)

	aliceRequests := newPipe[*cursorv1.ShareRequest]()
	alice, err := handler.Share(as("alice"), aliceRequests)
	if err != nil {
		t.Fatalf("Share: %v", err)
	}
	defer func() { _ = alice.Close() }()

	bobRequests := newPipe[*cursorv1.ShareRequest]()
	bob, err := handler.Share(as("bob"), bobRequests)
	if err != nil {
		t.Fatalf("Share: %v", err)
	}
	defer func() { _ = bob.Close() }()

	// Both subscribe once they are read.
	aliceEvents := make(chan *cursorv1.ShareResponse, 16)
	go func() {
		for {
			resp, err := alice.Recv()
			if err != nil {
				close(aliceEvents)
				return
			}
			aliceEvents <- resp
		}
	}()
	bobEvent := recvAsync(bob)
	for handler.Watchers() < 2 {
		time.Sleep(time.Millisecond)
	}

	// Alice never hears her own position, so the first she hears is Bob's.
	aliceRequests.Send(cursorv1.ShareRequest_builder{X: new(0.1), Y: new(0.1)}.Build())
	bobRequests.Send(cursorv1.ShareRequest_builder{X: new(0.9), Y: new(0.9)}.Build())

	if got := (<-bobEvent).GetMoved(); got.GetUsername() != "alice" {
		t.Errorf("Bob heard %v, want Alice's position", got)
	}
	if got := (<-aliceEvents).GetMoved(); got.GetUsername() != "bob" {
		t.Errorf("Alice heard %v, want Bob's position", got)
	}

	// Ending Alice's half ends her call, and Bob hears that her stream ended.
	aliceRequests.CloseSend()
	if _, ok := <-aliceEvents; ok {
		t.Error("Alice's call continued after she ended her half")
	}
	if got := recv(t, bob).GetEnded(); got.GetUsername() != "alice" {
		t.Errorf("Bob heard %v, want the end of Alice's stream", got)
	}
	if got := storedPositions(t, handler); got != 2 {
		t.Errorf("stored %d positions, want 2", got)
	}
}
