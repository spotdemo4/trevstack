package v1_test

import (
	"context"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	numberv1 "trev.zip/template/stack/server/v2/gen/number/v1"
)

func TestTimeSeries(t *testing.T) {
	t.Run("buckets by day", func(t *testing.T) {
		handler, db := newTest(t)
		day1 := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
		day2 := time.Date(2026, 1, 2, 5, 0, 0, 0, time.UTC)
		seed(t, db, "a", 10, day1)
		seed(t, db, "b", 20, day1.Add(time.Hour))
		seed(t, db, "c", 5, day2)

		resp, err := handler.TimeSeries(context.Background(), numberv1.TimeSeriesRequest_builder{
			Interval: new(numberv1.TimeInterval_TIME_INTERVAL_DAY),
		}.Build())
		if err != nil {
			t.Fatalf("TimeSeries: %v", err)
		}
		if len(resp.GetPoints()) != 2 {
			t.Fatalf("len(Points) = %d, want 2", len(resp.GetPoints()))
		}
		if resp.GetPoints()[0].GetCount() != 2 || resp.GetPoints()[0].GetSum() != 30 {
			t.Errorf("Points[0] = %+v, want count=2 sum=30", resp.GetPoints()[0])
		}
		if resp.GetPoints()[1].GetCount() != 1 || resp.GetPoints()[1].GetSum() != 5 {
			t.Errorf("Points[1] = %+v, want count=1 sum=5", resp.GetPoints()[1])
		}
	})

	t.Run("rejects missing interval", func(t *testing.T) {
		assertInvalid(t, &numberv1.TimeSeriesRequest{})
	})

	t.Run("rejects unspecified interval", func(t *testing.T) {
		assertInvalid(t, numberv1.TimeSeriesRequest_builder{
			Interval: new(numberv1.TimeInterval_TIME_INTERVAL_UNSPECIFIED),
		}.Build())
	})

	t.Run("rejects undefined interval", func(t *testing.T) {
		assertInvalid(t, numberv1.TimeSeriesRequest_builder{
			Interval: new(numberv1.TimeInterval(5)),
		}.Build())
	})

	t.Run("rejects end before start (CEL)", func(t *testing.T) {
		earlier := timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
		later := timestamppb.New(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))
		assertInvalid(t, numberv1.TimeSeriesRequest_builder{
			Interval: new(numberv1.TimeInterval_TIME_INTERVAL_DAY),
			Start:    later,
			End:      earlier,
		}.Build())
	})
}
