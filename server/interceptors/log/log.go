// Package log logs Connect calls.
package log

import (
	"context"
	"log/slog"

	"connectrpc.com/connect/v2"
)

// Interceptor logs Connect calls and their errors.
type Interceptor struct {
	log *slog.Logger
}

// New returns an Interceptor that logs to log.
func New(log *slog.Logger) *Interceptor {
	return &Interceptor{
		log: log,
	}
}

// WrapServer logs each call and any error it returns.
func (i *Interceptor) WrapServer(next connect.ServerFunc) connect.ServerFunc {
	return func(
		ctx context.Context,
		spec connect.Spec,
		stream connect.ServerStream,
	) error {
		i.log.DebugContext(ctx, "request received", "method", spec.Procedure)

		err := next(ctx, spec, stream)
		if err != nil {
			i.log.ErrorContext(ctx, "request error", "error", err)
		} else {
			i.log.DebugContext(ctx, "request completed", "method", spec.Procedure)
		}

		return err
	}
}
