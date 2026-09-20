package auth

import "context"

type key struct{}

func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, key{}, claims)
}

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
