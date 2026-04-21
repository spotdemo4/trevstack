package interceptors

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"
)

type LogInterceptor struct {
	log *slog.Logger
}

func NewLogInterceptor(log *slog.Logger) *LogInterceptor {
	return &LogInterceptor{
		log: log,
	}
}

func (i *LogInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		i.log.DebugContext(ctx, "request received", "method", req.Spec().Procedure)

		resp, err := next(ctx, req)
		if err != nil {
			i.log.ErrorContext(ctx, "request error", "error", err)
		} else {
			i.log.DebugContext(ctx, "request completed", "method", req.Spec().Procedure)
		}

		return resp, err
	})
}

func (i *LogInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(
		ctx context.Context,
		spec connect.Spec,
	) connect.StreamingClientConn {
		i.log.DebugContext(ctx, "streaming client started", "method", spec.Procedure)

		return next(ctx, spec)
	})
}

func (i *LogInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(
		ctx context.Context,
		conn connect.StreamingHandlerConn,
	) error {
		i.log.DebugContext(ctx, "streaming handler started", "method", conn.Spec().Procedure)

		return next(ctx, conn)
	})
}
