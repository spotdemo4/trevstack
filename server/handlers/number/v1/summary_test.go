package v1_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	numberv1 "github.com/spotdemo4/trevstack/server/connect/number/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestSummary(t *testing.T) {
	t.Run("aggregates count, sum, average, min, max, distinct names", func(t *testing.T) {
		client, db := newTest(t)
		now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		seed(t, db, "a", 10, now)
		seed(t, db, "b", 20, now)
		seed(t, db, "a", 30, now)
		seed(t, db, "c", 40, now)

		resp, err := client.Summary(context.Background(), &numberv1.SummaryRequest{})
		if err != nil {
			t.Fatalf("Summary: %v", err)
		}
		if resp.TotalCount != 4 {
			t.Errorf("TotalCount = %d, want 4", resp.TotalCount)
		}
		if resp.TotalSum != 100 {
			t.Errorf("TotalSum = %d, want 100", resp.TotalSum)
		}
		if resp.Average != 25.0 {
			t.Errorf("Average = %v, want 25", resp.Average)
		}
		if resp.Min != 10 {
			t.Errorf("Min = %d, want 10", resp.Min)
		}
		if resp.Max != 40 {
			t.Errorf("Max = %d, want 40", resp.Max)
		}
		if resp.DistinctNames != 3 {
			t.Errorf("DistinctNames = %d, want 3", resp.DistinctNames)
		}
	})

	t.Run("empty table returns zeros", func(t *testing.T) {
		client, _ := newTest(t)
		resp, err := client.Summary(context.Background(), &numberv1.SummaryRequest{})
		if err != nil {
			t.Fatalf("Summary: %v", err)
		}
		if resp.TotalCount != 0 || resp.TotalSum != 0 || resp.DistinctNames != 0 {
			t.Errorf("expected zeros, got %+v", resp)
		}
	})

	t.Run("rejects end before start (CEL)", func(t *testing.T) {
		client, _ := newTest(t)
		earlier := timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
		later := timestamppb.New(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))
		_, err := client.Summary(context.Background(), &numberv1.SummaryRequest{Start: later, End: earlier})
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
		if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
			t.Errorf("code = %v, want InvalidArgument", got)
		}
	})
}
