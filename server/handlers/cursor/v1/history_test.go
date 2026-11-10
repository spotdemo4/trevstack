package v1_test

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	cursorv1 "trev.zip/template/stack/server/v2/gen/cursor/v1"
)

func list(t *testing.T, h *handlerT, req *cursorv1.ListRequest) []*cursorv1.Cursor {
	t.Helper()
	stream, err := h.List(context.Background(), req)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	var cursors []*cursorv1.Cursor
	for _, resp := range collect(t, stream) {
		cursors = append(cursors, resp.GetCursor())
	}
	return cursors
}

func listStreams(t *testing.T, h *handlerT, req *cursorv1.ListStreamsRequest) []*cursorv1.StreamSummary {
	t.Helper()
	stream, err := h.ListStreams(context.Background(), req)
	if err != nil {
		t.Fatalf("ListStreams: %v", err)
	}
	var streams []*cursorv1.StreamSummary
	for _, resp := range collect(t, stream) {
		streams = append(streams, resp.GetStream())
	}
	return streams
}

func TestList(t *testing.T) {
	handler, db := newTest(t)
	alice := seedStream(t, db, "alice", cursorv1.Source_SOURCE_TRACK, base, base.Add(time.Hour), base.Add(2*time.Hour))
	bob := seedStream(t, db, "bob", cursorv1.Source_SOURCE_SEND, base.Add(3*time.Hour))

	t.Run("returns positions newest first", func(t *testing.T) {
		got := list(t, handler, &cursorv1.ListRequest{})
		if len(got) != 4 {
			t.Fatalf("len = %d, want 4", len(got))
		}
		if got[0].GetStreamId() != bob || got[0].GetUsername() != "bob" || got[0].GetSource() != cursorv1.Source_SOURCE_SEND {
			t.Errorf("got[0] = %v, want bob's Send", got[0])
		}
		if !got[1].GetTimestamp().AsTime().Equal(base.Add(2 * time.Hour)) {
			t.Errorf("got[1].Timestamp = %v, want %v", got[1].GetTimestamp().AsTime(), base.Add(2*time.Hour))
		}
	})

	tests := []struct {
		name string
		req  *cursorv1.ListRequest
		want int
	}{
		{"username substring", cursorv1.ListRequest_builder{Username: new("LIC")}.Build(), 3},
		{"source", cursorv1.ListRequest_builder{Source: cursorv1.Source_SOURCE_SEND.Enum()}.Build(), 1},
		{"stream", cursorv1.ListRequest_builder{StreamId: &alice}.Build(), 3},
		{"x bounds", cursorv1.ListRequest_builder{MinX: new(0.15), MaxX: new(0.25)}.Build(), 1},
		{"zero max x", cursorv1.ListRequest_builder{MaxX: new(0.0)}.Build(), 0},
		{"y bounds", cursorv1.ListRequest_builder{MinY: new(0.5), MaxY: new(0.5)}.Build(), 4},
		{"time bounds", cursorv1.ListRequest_builder{
			Start: timestamppb.New(base.Add(time.Hour)),
			End:   timestamppb.New(base.Add(2 * time.Hour)),
		}.Build(), 2},
	}
	for _, tt := range tests {
		t.Run("filters by "+tt.name, func(t *testing.T) {
			if got := list(t, handler, tt.req); len(got) != tt.want {
				t.Errorf("len = %d, want %d", len(got), tt.want)
			}
		})
	}

	t.Run("validates filters", func(t *testing.T) {
		earlier := timestamppb.New(base)
		later := timestamppb.New(base.Add(time.Hour))
		assertValid(t, cursorv1.ListRequest_builder{MinX: new(0.0), MaxX: new(1.0)}.Build())
		assertInvalid(t, cursorv1.ListRequest_builder{Username: new(strings.Repeat("a", 51))}.Build())
		assertInvalid(t, cursorv1.ListRequest_builder{Source: cursorv1.Source_SOURCE_UNSPECIFIED.Enum()}.Build())
		assertInvalid(t, cursorv1.ListRequest_builder{Source: cursorv1.Source(9).Enum()}.Build())
		assertInvalid(t, cursorv1.ListRequest_builder{StreamId: new(int64(0))}.Build())
		assertInvalid(t, cursorv1.ListRequest_builder{MinX: new(-0.5)}.Build())
		assertInvalid(t, cursorv1.ListRequest_builder{MinX: new(0.5), MaxX: new(0.5)}.Build())
		assertInvalid(t, cursorv1.ListRequest_builder{MinY: new(0.6), MaxY: new(0.5)}.Build())
		assertInvalid(t, cursorv1.ListRequest_builder{Start: later, End: earlier}.Build())
	})
}

func TestListStreams(t *testing.T) {
	handler, db := newTest(t)
	alice := seedStream(t, db, "alice", cursorv1.Source_SOURCE_TRACK, base, base.Add(time.Minute), base.Add(2*time.Minute))
	bob := seedStream(t, db, "bob", cursorv1.Source_SOURCE_SHARE, base.Add(time.Hour))

	t.Run("summarizes streams newest first", func(t *testing.T) {
		got := listStreams(t, handler, &cursorv1.ListStreamsRequest{})
		if len(got) != 2 {
			t.Fatalf("len = %d, want 2", len(got))
		}
		if got[0].GetId() != bob || got[1].GetId() != alice {
			t.Fatalf("ids = %d, %d, want %d, %d", got[0].GetId(), got[1].GetId(), bob, alice)
		}
		a := got[1]
		if a.GetUsername() != "alice" || a.GetSource() != cursorv1.Source_SOURCE_TRACK || a.GetCount() != 3 {
			t.Errorf("alice's stream = %v", a)
		}
		if math.Abs(a.GetDistance()-0.2) > 1e-9 {
			t.Errorf("distance = %v, want 0.2", a.GetDistance())
		}
		if !a.GetStart().AsTime().Equal(base) || !a.GetEnd().AsTime().Equal(base.Add(2*time.Minute)) {
			t.Errorf("span = %v to %v", a.GetStart().AsTime(), a.GetEnd().AsTime())
		}
	})

	t.Run("filters", func(t *testing.T) {
		if got := listStreams(t, handler, cursorv1.ListStreamsRequest_builder{Username: new("bo")}.Build()); len(got) != 1 || got[0].GetId() != bob {
			t.Errorf("by username = %v, want bob's", got)
		}
		if got := listStreams(t, handler, cursorv1.ListStreamsRequest_builder{Source: cursorv1.Source_SOURCE_TRACK.Enum()}.Build()); len(got) != 1 || got[0].GetId() != alice {
			t.Errorf("by source = %v, want alice's", got)
		}
		// Only the first position counts, so Alice's stream is outside.
		from := timestamppb.New(base.Add(time.Minute))
		if got := listStreams(t, handler, cursorv1.ListStreamsRequest_builder{Start: from}.Build()); len(got) != 1 || got[0].GetId() != bob {
			t.Errorf("by start = %v, want bob's", got)
		}
		if got := listStreams(t, handler, cursorv1.ListStreamsRequest_builder{End: from}.Build()); len(got) != 1 || got[0].GetId() != alice {
			t.Errorf("by end = %v, want alice's", got)
		}
	})
}

func TestSummary(t *testing.T) {
	t.Run("aggregates positions, streams, users and distance", func(t *testing.T) {
		handler, db := newTest(t)
		seedStream(t, db, "alice", cursorv1.Source_SOURCE_TRACK, base, base, base)
		seedStream(t, db, "alice", cursorv1.Source_SOURCE_SEND, base)
		seedStream(t, db, "bob", cursorv1.Source_SOURCE_SHARE, base, base, base.Add(time.Hour))

		resp, err := handler.Summary(context.Background(), &cursorv1.SummaryRequest{})
		if err != nil {
			t.Fatalf("Summary: %v", err)
		}
		if resp.GetCount() != 7 || resp.GetStreams() != 3 || resp.GetUsers() != 2 {
			t.Errorf("counts = %d, %d, %d, want 7, 3, 2", resp.GetCount(), resp.GetStreams(), resp.GetUsers())
		}
		if math.Abs(resp.GetDistance()-0.4) > 1e-9 {
			t.Errorf("distance = %v, want 0.4", resp.GetDistance())
		}
		if math.Abs(resp.GetAverageCount()-7.0/3) > 1e-9 || math.Abs(resp.GetAverageDistance()-0.4/3) > 1e-9 {
			t.Errorf("averages = %v, %v", resp.GetAverageCount(), resp.GetAverageDistance())
		}

		resp, err = handler.Summary(context.Background(), cursorv1.SummaryRequest_builder{End: timestamppb.New(base)}.Build())
		if err != nil {
			t.Fatalf("Summary: %v", err)
		}
		if resp.GetCount() != 6 {
			t.Errorf("count before the hour = %d, want 6", resp.GetCount())
		}
	})

	t.Run("no positions returns zeros", func(t *testing.T) {
		handler, _ := newTest(t)
		resp, err := handler.Summary(context.Background(), &cursorv1.SummaryRequest{})
		if err != nil {
			t.Fatalf("Summary: %v", err)
		}
		if resp.GetCount() != 0 || resp.GetStreams() != 0 || resp.GetAverageCount() != 0 {
			t.Errorf("expected zeros, got %v", resp)
		}
	})

	t.Run("rejects end before start", func(t *testing.T) {
		assertInvalid(t, cursorv1.SummaryRequest_builder{Start: timestamppb.New(base.Add(time.Hour)), End: timestamppb.New(base)}.Build())
	})
}

func TestTimeSeries(t *testing.T) {
	handler, db := newTest(t)
	// 2026-01-01 is a Thursday.
	seedStream(t, db, "alice", cursorv1.Source_SOURCE_TRACK, base, base.Add(30*time.Second), base.Add(90*time.Second))
	seedStream(t, db, "bob", cursorv1.Source_SOURCE_SEND, base.Add(40*24*time.Hour))

	tests := []struct {
		interval cursorv1.TimeInterval
		want     []time.Time
	}{
		{cursorv1.TimeInterval_TIME_INTERVAL_MINUTE, []time.Time{base, base.Add(time.Minute), base.Add(40 * 24 * time.Hour)}},
		{cursorv1.TimeInterval_TIME_INTERVAL_DAY, []time.Time{base, base.Add(40 * 24 * time.Hour)}},
		{cursorv1.TimeInterval_TIME_INTERVAL_WEEK, []time.Time{
			time.Date(2025, 12, 28, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 2, 8, 0, 0, 0, 0, time.UTC),
		}},
		{cursorv1.TimeInterval_TIME_INTERVAL_MONTH, []time.Time{base, time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)}},
	}
	for _, tt := range tests {
		t.Run(tt.interval.String(), func(t *testing.T) {
			resp, err := handler.TimeSeries(context.Background(), cursorv1.TimeSeriesRequest_builder{Interval: tt.interval.Enum()}.Build())
			if err != nil {
				t.Fatalf("TimeSeries: %v", err)
			}
			points := resp.GetPoints()
			if len(points) != len(tt.want) {
				t.Fatalf("len = %d, want %d", len(points), len(tt.want))
			}
			for i, want := range tt.want {
				if got := points[i].GetBucket().AsTime(); !got.Equal(want) {
					t.Errorf("bucket %d = %v, want %v", i, got, want)
				}
			}
		})
	}

	t.Run("counts positions, streams and users", func(t *testing.T) {
		resp, err := handler.TimeSeries(context.Background(), cursorv1.TimeSeriesRequest_builder{Interval: cursorv1.TimeInterval_TIME_INTERVAL_MONTH.Enum()}.Build())
		if err != nil {
			t.Fatalf("TimeSeries: %v", err)
		}
		first := resp.GetPoints()[0]
		if first.GetCount() != 3 || first.GetStreams() != 1 || first.GetUsers() != 1 {
			t.Errorf("first = %v, want 3 positions of 1 stream of 1 user", first)
		}
	})

	t.Run("validates the interval", func(t *testing.T) {
		assertInvalid(t, &cursorv1.TimeSeriesRequest{})
		assertInvalid(t, cursorv1.TimeSeriesRequest_builder{Interval: cursorv1.TimeInterval_TIME_INTERVAL_UNSPECIFIED.Enum()}.Build())
	})
}

func TestHeatmap(t *testing.T) {
	handler, db := newTest(t)
	// x = 0.1, 0.2, 0.3 at y = 0.5.
	seedStream(t, db, "alice", cursorv1.Source_SOURCE_TRACK, base, base, base)
	if _, err := db.ExecContext(t.Context(), `INSERT INTO cursors (stream_id, x, y, distance, timestamp) VALUES (1, 1, 1, 0, 0), (1, 0, 0, 0, 0)`); err != nil {
		t.Fatal(err)
	}

	resp, err := handler.Heatmap(context.Background(), cursorv1.HeatmapRequest_builder{Columns: new(uint32(4)), Rows: new(uint32(2))}.Build())
	if err != nil {
		t.Fatalf("Heatmap: %v", err)
	}
	if resp.GetColumns() != 4 || resp.GetRows() != 2 {
		t.Errorf("size = %dx%d, want 4x2", resp.GetColumns(), resp.GetRows())
	}
	want := []int64{
		1, 0, 0, 0,
		2, 1, 0, 1,
	}
	got := resp.GetCounts()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("counts = %v, want %v", got, want)
			break
		}
	}

	assertInvalid(t, cursorv1.HeatmapRequest_builder{Columns: new(uint32(0)), Rows: new(uint32(1))}.Build())
	assertInvalid(t, cursorv1.HeatmapRequest_builder{Columns: new(uint32(1)), Rows: new(uint32(101))}.Build())
	assertInvalid(t, cursorv1.HeatmapRequest_builder{Columns: new(uint32(1))}.Build())
}

func TestTopUsers(t *testing.T) {
	handler, db := newTest(t)
	seedStream(t, db, "alice", cursorv1.Source_SOURCE_TRACK, base, base)
	seedStream(t, db, "bob", cursorv1.Source_SOURCE_TRACK, base, base, base)
	seedStream(t, db, "bob", cursorv1.Source_SOURCE_SEND, base)

	resp, err := handler.TopUsers(context.Background(), cursorv1.TopUsersRequest_builder{Limit: new(uint32(10))}.Build())
	if err != nil {
		t.Fatalf("TopUsers: %v", err)
	}
	users := resp.GetUsers()
	if len(users) != 2 || users[0].GetUsername() != "bob" || users[1].GetUsername() != "alice" {
		t.Fatalf("users = %v, want bob then alice", users)
	}
	if users[0].GetCount() != 4 || users[0].GetStreams() != 2 || math.Abs(users[0].GetDistance()-0.2) > 1e-9 {
		t.Errorf("bob = %v", users[0])
	}

	resp, err = handler.TopUsers(context.Background(), cursorv1.TopUsersRequest_builder{Limit: new(uint32(1))}.Build())
	if err != nil {
		t.Fatalf("TopUsers: %v", err)
	}
	if len(resp.GetUsers()) != 1 {
		t.Errorf("len = %d, want 1", len(resp.GetUsers()))
	}

	assertInvalid(t, cursorv1.TopUsersRequest_builder{Limit: new(uint32(0))}.Build())
	assertInvalid(t, &cursorv1.TopUsersRequest{})
}
