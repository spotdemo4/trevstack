package auth

import (
	"context"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestClaimsContext(t *testing.T) {
	claims := &Claims{RegisteredClaims: jwt.RegisteredClaims{Subject: "trev"}}
	ctx := WithClaims(context.Background(), claims)

	got, ok := ClaimsFromContext(ctx)
	if !ok || got != claims {
		t.Fatalf("ClaimsFromContext() = (%#v, %t), want (%#v, true)", got, ok, claims)
	}
}

func TestClaimsFromContextMissingOrNil(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{name: "missing", ctx: context.Background()},
		{name: "nil context"},
		{name: "nil claims", ctx: WithClaims(context.Background(), nil)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := ClaimsFromContext(test.ctx)
			if got != nil || ok {
				t.Fatalf("ClaimsFromContext() = (%#v, %t), want (nil, false)", got, ok)
			}
		})
	}
}
