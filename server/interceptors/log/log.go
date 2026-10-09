package log

import (
	"context"
	"log/slog"

	"connectrpc.com/connect/v2"
)

type LogInterceptor struct {
	log *slog.Logger
}

func NewLogInterceptor(log *slog.Logger) *LogInterceptor {
	return &LogInterceptor{
		log: log,
	}
}

func (i *LogInterceptor) WrapServer(next connect.ServerFunc) connect.ServerFunc {
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
