package v1_test

import (
	"context"
	"strings"
	"testing"

	"connectrpc.com/connect"
	numberv1 "github.com/spotdemo4/trevstack/server/connect/number/v1"
)

func TestAdd(t *testing.T) {
	t.Run("inserts row and returns running sum", func(t *testing.T) {
		client, db := newTest(t)
		ctx := context.Background()

		resp, err := client.Add(ctx, numberv1.AddRequest_builder{
			Name:   ptr("foo"),
			Number: ptr(uint32(10)),
		}.Build())
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
		if resp.GetSum() != 10 {
			t.Errorf("Sum = %d, want 10", resp.GetSum())
		}

		resp, err = client.Add(ctx, numberv1.AddRequest_builder{
			Name:   ptr("bar"),
			Number: ptr(uint32(5)),
		}.Build())
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
		if resp.GetSum() != 15 {
			t.Errorf("Sum = %d, want 15", resp.GetSum())
		}

		var count int
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM numbers").Scan(&count); err != nil {
			t.Fatalf("count: %v", err)
		}
		if count != 2 {
			t.Errorf("row count = %d, want 2", count)
		}
	})

	validationCases := []struct {
		name string
		req  *numberv1.AddRequest
	}{
		{"empty name", numberv1.AddRequest_builder{Name: ptr(""), Number: ptr(uint32(10))}.Build()},
		{"name too long", numberv1.AddRequest_builder{Name: ptr(strings.Repeat("a", 51)), Number: ptr(uint32(10))}.Build()},
		{"number below min", numberv1.AddRequest_builder{Name: ptr("foo"), Number: ptr(uint32(0))}.Build()},
		{"number above max", numberv1.AddRequest_builder{Name: ptr("foo"), Number: ptr(uint32(1_000_001))}.Build()},
	}
	for _, tc := range validationCases {
		t.Run("rejects "+tc.name, func(t *testing.T) {
			client, _ := newTest(t)

			_, err := client.Add(context.Background(), tc.req)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if got := connect.CodeOf(err); got != connect.CodeInvalidArgument {
				t.Errorf("code = %v, want InvalidArgument", got)
			}
		})
	}
}
