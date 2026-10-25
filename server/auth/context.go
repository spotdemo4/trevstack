package auth

import "context"

type key struct{}

// WithClaims returns a copy of ctx that carries claims.
func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, key{}, claims)
}

// ClaimsFromContext returns the claims stored in ctx by WithClaims.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	if ctx == nil {
		return nil, false
	}

	claims, ok := ctx.Value(key{}).(*Claims)
	if !ok || claims == nil {
		return nil, false
	}

	return claims, true
}
