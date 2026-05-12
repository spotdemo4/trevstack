package v1_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	numberv1 "github.com/spotdemo4/trevstack/server/connect/number/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTimeSeries(t *testing.T) {
	t.Run("buckets by day", func(t *testing.T) {
		client, db := newTest(t)
		day1 := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
		day2 := time.Date(2026, 1, 2, 5, 0, 0, 0, time.UTC)
		seed(t, db, "a", 10, day1)
		seed(t, db, "b", 20, day1.Add(time.Hour))
		seed(t, db, "c", 5, day2)

		resp, err := client.TimeSeries(context.Background(), &numberv1.TimeSeriesRequest{
			Interval: numberv1.TimeInterval_TIME_INTERVAL_DAY,
		})
		if err != nil {
			t.Fatalf("TimeSeries: %v", err)
		}
		if len(resp.Points) != 2 {
			t.Fatalf("len(Points) = %d, want 2", len(resp.Points))
		}
		if resp.Points[0].Count != 2 || resp.Points[0].Sum != 30 {
			t.Errorf("Points[0] = %+v, want count=2 sum=30", resp.Points[0])
		}
		if resp.Points[1].Count != 1 || resp.Points[1].Sum != 5 {
			t.Errorf("Points[1] = %+v, want count=1 sum=5", resp.Points[1])
		}
	})

	t.Run("rejects unspecified interval", func(t *testing.T) {
		client, _ := newTest(t)
		_, err := client.TimeSeries(context.Background(), &numberv1.TimeSeriesRequest{
			Interval: numberv1.TimeInterval_TIME_INTERVAL_UNSPECIFIED,
		})
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
		if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
			t.Errorf("code = %v, want InvalidArgument", got)
		}
	})

	t.Run("rejects end before start (CEL)", func(t *testing.T) {
		client, _ := newTest(t)
		earlier := timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
		later := timestamppb.New(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))
		_, err := client.TimeSeries(context.Background(), &numberv1.TimeSeriesRequest{
			Interval: numberv1.TimeInterval_TIME_INTERVAL_DAY,
			Start:    later,
			End:      earlier,
		})
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}
		if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
			t.Errorf("code = %v, want InvalidArgument", got)
		}
	})
}
