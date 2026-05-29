package v1_test

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
	numberv1 "trev.zip/llc/stack/server/connect/number/v1"
)

func TestTopNames(t *testing.T) {
	t.Run("returns names ranked by count then sum", func(t *testing.T) {
		client, db := newTest(t)
		now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		seed(t, db, "a", 1, now)
		seed(t, db, "a", 1, now)
		seed(t, db, "a", 1, now)
		seed(t, db, "b", 5, now)
		seed(t, db, "b", 5, now)
		seed(t, db, "c", 100, now)

		resp, err := client.TopNames(context.Background(), numberv1.TopNamesRequest_builder{
			Limit: ptr(uint32(2)),
		}.Build())
		if err != nil {
			t.Fatalf("TopNames: %v", err)
		}
		if len(resp.GetNames()) != 2 {
			t.Fatalf("len(Names) = %d, want 2", len(resp.GetNames()))
		}
		if resp.GetNames()[0].GetName() != "a" || resp.GetNames()[0].GetCount() != 3 || resp.GetNames()[0].GetSum() != 3 {
			t.Errorf("Names[0] = %+v, want {a 3 3}", resp.GetNames()[0])
		}
		if resp.GetNames()[1].GetName() != "b" || resp.GetNames()[1].GetCount() != 2 || resp.GetNames()[1].GetSum() != 10 {
			t.Errorf("Names[1] = %+v, want {b 2 10}", resp.GetNames()[1])
		}
	})

	earlier := timestamppb.New(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	later := timestamppb.New(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))

	validationCases := []struct {
		name string
		req  *numberv1.TopNamesRequest
	}{
		{"limit zero", numberv1.TopNamesRequest_builder{Limit: ptr(uint32(0))}.Build()},
		{"limit above max", numberv1.TopNamesRequest_builder{Limit: ptr(uint32(101))}.Build()},
		{"end before start (CEL)", numberv1.TopNamesRequest_builder{Limit: ptr(uint32(5)), Start: later, End: earlier}.Build()},
	}
	for _, tc := range validationCases {
		t.Run("rejects "+tc.name, func(t *testing.T) {
			client, _ := newTest(t)
			_, err := client.TopNames(context.Background(), tc.req)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
				t.Errorf("code = %v, want InvalidArgument", got)
			}
		})
	}
}
