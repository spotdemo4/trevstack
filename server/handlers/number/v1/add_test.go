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

		resp, err := client.Add(ctx, &numberv1.AddRequest{Name: "foo", Number: 10})
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
		if resp.Sum != 10 {
			t.Errorf("Sum = %d, want 10", resp.Sum)
		}

		resp, err = client.Add(ctx, &numberv1.AddRequest{Name: "bar", Number: 5})
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
		if resp.Sum != 15 {
			t.Errorf("Sum = %d, want 15", resp.Sum)
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
		{"empty name", &numberv1.AddRequest{Name: "", Number: 10}},
		{"name too long", &numberv1.AddRequest{Name: strings.Repeat("a", 51), Number: 10}},
		{"number below min", &numberv1.AddRequest{Name: "foo", Number: 0}},
		{"number above max", &numberv1.AddRequest{Name: "foo", Number: 1_000_001}},
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
