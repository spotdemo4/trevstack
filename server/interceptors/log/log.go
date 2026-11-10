// Package log logs TrevRPC calls.
package log

import (
	"context"
	"log/slog"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	"trev.zip/template/stack/server/v2/auth"
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

// Intercept logs each call when it starts, and when it ends with its status,
// at error unless it succeeded or the client cancelled it, along with the client's IP when the transport knows it and the user of an
// authenticated call.
func (i *Interceptor) Intercept(
	ctx context.Context,
	info trevrpc.HandlerContext,
) (context.Context, func(trevrpc.RPCFinished), error) {
	attrs := []any{"method", info.FullMethod}
	if info.RemoteAddr.IsValid() {
		attrs = append(attrs, "client_ip", info.RemoteAddr.Addr().String())
	}
	if username, ok := auth.UsernameFromContext(ctx); ok {
		attrs = append(attrs, "user", username)
	}
	i.log.DebugContext(ctx, "request received", attrs...)

	return ctx, func(finished trevrpc.RPCFinished) {
		switch finished.Code {
		case trevrpc.CodeOK:
			i.log.DebugContext(ctx, "request completed", attrs...)
		case trevrpc.CodeCancelled:
			// The client gave up on the call, such as a page that stopped
			// watching a stream when it closed, which is not a server error.
			i.log.DebugContext(ctx, "request cancelled", attrs...)
		default:
			i.log.ErrorContext(ctx, "request error", append(attrs, "code", finished.Code.String())...)
		}
	}, nil
}
