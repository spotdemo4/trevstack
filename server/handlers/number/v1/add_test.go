package v1_test

import (
	"context"
	"strings"
	"testing"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	numberv1 "trev.zip/template/stack/server/v2/gen/number/v1"
)

func TestAdd(t *testing.T) {
	t.Run("inserts row and returns running sum", func(t *testing.T) {
		handler, db := newTest(t)
		ctx := context.Background()

		resp, err := handler.Add(ctx, numberv1.AddRequest_builder{
			Name:   new("foo"),
			Number: new(uint32(10)),
		}.Build())
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
		if resp.GetSum() != 10 {
			t.Errorf("Sum = %d, want 10", resp.GetSum())
		}

		resp, err = handler.Add(ctx, numberv1.AddRequest_builder{
			Name:   new("bar"),
			Number: new(uint32(5)),
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
		{"missing name", numberv1.AddRequest_builder{Number: new(uint32(10))}.Build()},
		{"missing number", numberv1.AddRequest_builder{Name: new("foo")}.Build()},
		{"empty name", numberv1.AddRequest_builder{Name: new(""), Number: new(uint32(10))}.Build()},
		{"name too long", numberv1.AddRequest_builder{Name: new(strings.Repeat("a", 51)), Number: new(uint32(10))}.Build()},
		{"number below min", numberv1.AddRequest_builder{Name: new("foo"), Number: new(uint32(0))}.Build()},
		{"number above max", numberv1.AddRequest_builder{Name: new("foo"), Number: new(uint32(1_000_001))}.Build()},
	}
	for _, tc := range validationCases {
		t.Run("rejects "+tc.name, func(t *testing.T) {
			handler, _ := newTest(t)

			_, err := handler.Add(context.Background(), tc.req)
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if got := trevrpc.StatusFromError(err).Code; got != trevrpc.CodeInvalidArgument {
				t.Errorf("code = %v, want InvalidArgument", got)
			}
		})
	}
}
