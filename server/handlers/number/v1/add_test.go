package v1_test

import (
	"context"
	"strings"
	"testing"

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
			assertInvalid(t, tc.req)
		})
	}

	t.Run("accepts the bounds", func(t *testing.T) {
		// Names are measured in characters, not bytes.
		assertValid(t, numberv1.AddRequest_builder{Name: new(strings.Repeat("é", 50)), Number: new(uint32(1))}.Build())
		assertValid(t, numberv1.AddRequest_builder{Name: new("a"), Number: new(uint32(1_000_000))}.Build())
	})
}
