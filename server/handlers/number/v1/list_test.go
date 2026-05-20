package v1_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	numberv1 "github.com/spotdemo4/trevstack/server/connect/number/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func collectListItems(
	t *testing.T,
	stream *connect.ServerStreamForClient[numberv1.ListResponse],
) []*numberv1.Item {
	t.Helper()

	var items []*numberv1.Item
	for stream.Receive() {
		item := stream.Msg().GetItem()
		if item == nil {
			t.Fatal("List yielded response without item")
		}
		items = append(items, item)
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("List stream: %v", err)
	}
	return items
}

func TestList(t *testing.T) {
	t.Run("returns rows newest first", func(t *testing.T) {
		client, db := newTest(t)
		base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		seed(t, db, "alice", 10, base)
		seed(t, db, "bob", 20, base.Add(time.Hour))
		seed(t, db, "carol", 30, base.Add(2*time.Hour))

		stream, err := client.List(context.Background(), &numberv1.ListRequest{})
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		items := collectListItems(t, stream)
		if len(items) != 3 {
			t.Fatalf("len(items) = %d, want 3", len(items))
		}
		gotNames := []string{items[0].GetName(), items[1].GetName(), items[2].GetName()}
		want := []string{"carol", "bob", "alice"}
		for i, n := range want {
			if gotNames[i] != n {
				t.Errorf("items[%d].Name = %q, want %q", i, gotNames[i], n)
			}
		}
	})

	t.Run("filters by name substring", func(t *testing.T) {
		client, db := newTest(t)
		now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		seed(t, db, "alice", 1, now)
		seed(t, db, "alex", 2, now)
		seed(t, db, "bob", 3, now)

		needle := "al"
		stream, err := client.List(context.Background(), numberv1.ListRequest_builder{Name: &needle}.Build())
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		items := collectListItems(t, stream)
		if len(items) != 2 {
			t.Fatalf("len(items) = %d, want 2", len(items))
		}
	})

	t.Run("filters by min/max range", func(t *testing.T) {
		client, db := newTest(t)
		now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		seed(t, db, "a", 5, now)
		seed(t, db, "b", 50, now)
		seed(t, db, "c", 500, now)

		min, max := uint32(10), uint32(100)
		stream, err := client.List(context.Background(), numberv1.ListRequest_builder{Minimum: &min, Maximum: &max}.Build())
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		items := collectListItems(t, stream)
		if len(items) != 1 {
			t.Fatalf("len(items) = %d, want 1", len(items))
		}
		if items[0].GetNumber() != 50 {
			t.Errorf("items[0].Number = %d, want 50", items[0].GetNumber())
		}
	})

	tooLong := strings.Repeat("a", 51)
	lo, hi := uint32(50), uint32(10)
	huge := uint32(1_000_001)
	earlier := timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	later := timestamppb.New(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))

	validationCases := []struct {
		name string
		req  *numberv1.ListRequest
	}{
		{"name too long", numberv1.ListRequest_builder{Name: &tooLong}.Build()},
		{"minimum above max value", numberv1.ListRequest_builder{Minimum: &huge}.Build()},
		{"maximum above max value", numberv1.ListRequest_builder{Maximum: &huge}.Build()},
		{"minimum greater than maximum (CEL)", numberv1.ListRequest_builder{Minimum: &lo, Maximum: &hi}.Build()},
		{"end before start (CEL)", numberv1.ListRequest_builder{Start: later, End: earlier}.Build()},
	}
	for _, tc := range validationCases {
		t.Run("rejects "+tc.name, func(t *testing.T) {
			client, _ := newTest(t)
			stream, err := client.List(context.Background(), tc.req)
			if err != nil {
				if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
					t.Errorf("code = %v, want InvalidArgument", got)
				}
				return
			}
			if stream.Receive() {
				t.Fatal("expected validation error, got stream item")
			}
			err = stream.Err()
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
				t.Errorf("code = %v, want InvalidArgument", got)
			}
		})
	}
}
