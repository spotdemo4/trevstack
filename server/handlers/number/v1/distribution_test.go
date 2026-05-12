package v1_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	numberv1 "github.com/spotdemo4/trevstack/server/connect/number/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestDistribution(t *testing.T) {
	t.Run("empty table returns no buckets", func(t *testing.T) {
		client, _ := newTest(t)
		resp, err := client.Distribution(context.Background(), numberv1.DistributionRequest_builder{
			BucketCount: ptr(uint32(5)),
		}.Build())
		if err != nil {
			t.Fatalf("Distribution: %v", err)
		}
		if len(resp.GetBuckets()) != 0 {
			t.Errorf("len(Buckets) = %d, want 0", len(resp.GetBuckets()))
		}
	})

	t.Run("all identical values collapse to one bucket", func(t *testing.T) {
		client, db := newTest(t)
		now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		seed(t, db, "a", 7, now)
		seed(t, db, "b", 7, now)
		seed(t, db, "c", 7, now)

		resp, err := client.Distribution(context.Background(), numberv1.DistributionRequest_builder{
			BucketCount: ptr(uint32(5)),
		}.Build())
		if err != nil {
			t.Fatalf("Distribution: %v", err)
		}
		if len(resp.GetBuckets()) != 1 {
			t.Fatalf("len(Buckets) = %d, want 1", len(resp.GetBuckets()))
		}
		bucket := resp.GetBuckets()[0]
		if bucket.GetLower() != 7 || bucket.GetUpper() != 7 || bucket.GetCount() != 3 {
			t.Errorf("Buckets[0] = %+v, want {Lower:7 Upper:7 Count:3}", bucket)
		}
	})

	t.Run("splits range into requested bucket count", func(t *testing.T) {
		client, db := newTest(t)
		now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		for _, n := range []uint32{1, 2, 3, 8, 9, 10} {
			seed(t, db, "x", n, now)
		}

		resp, err := client.Distribution(context.Background(), numberv1.DistributionRequest_builder{
			BucketCount: ptr(uint32(2)),
		}.Build())
		if err != nil {
			t.Fatalf("Distribution: %v", err)
		}
		if len(resp.GetBuckets()) != 2 {
			t.Fatalf("len(Buckets) = %d, want 2", len(resp.GetBuckets()))
		}
		// Range is [1, 10], split into 2 buckets of equal span. Lower half holds 1/2/3, upper half holds 8/9/10.
		total := resp.GetBuckets()[0].GetCount() + resp.GetBuckets()[1].GetCount()
		if total != 6 {
			t.Errorf("total count across buckets = %d, want 6", total)
		}
		if resp.GetBuckets()[0].GetCount() != 3 || resp.GetBuckets()[1].GetCount() != 3 {
			t.Errorf("Buckets = %+v, want each bucket count = 3", resp.GetBuckets())
		}
	})

	earlier := timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	later := timestamppb.New(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))

	validationCases := []struct {
		name string
		req  *numberv1.DistributionRequest
	}{
		{"bucket count zero", numberv1.DistributionRequest_builder{BucketCount: ptr(uint32(0))}.Build()},
		{"bucket count above max", numberv1.DistributionRequest_builder{BucketCount: ptr(uint32(101))}.Build()},
		{"end before start (CEL)", numberv1.DistributionRequest_builder{BucketCount: ptr(uint32(5)), Start: later, End: earlier}.Build()},
	}
	for _, tc := range validationCases {
		t.Run("rejects "+tc.name, func(t *testing.T) {
			client, _ := newTest(t)
			_, err := client.Distribution(context.Background(), tc.req)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
				t.Errorf("code = %v, want InvalidArgument", got)
			}
		})
	}
}
