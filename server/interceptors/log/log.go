// Package log logs TrevRPC calls.
package log

import (
	"context"
	"log/slog"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
)

// Interceptor logs the calls that reach a handler and their errors.
type Interceptor struct {
	log *slog.Logger
}

// New returns an Interceptor that logs to log.
func New(log *slog.Logger) *Interceptor {
	return &Interceptor{
		log: log,
	}
}

// Intercept logs each call when it starts, and when it ends with its status.
func (i *Interceptor) Intercept(
	ctx context.Context,
	info trevrpc.HandlerContext,
) (context.Context, func(trevrpc.RPCFinished), error) {
	i.log.DebugContext(ctx, "request received", "method", info.FullMethod)

	return ctx, func(finished trevrpc.RPCFinished) {
		if finished.Code != trevrpc.CodeOK {
			i.log.ErrorContext(ctx, "request error", "method", info.FullMethod, "code", finished.Code.String())
		} else {
			i.log.DebugContext(ctx, "request completed", "method", info.FullMethod)
		}
	}, nil
}
