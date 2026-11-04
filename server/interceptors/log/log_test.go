package log

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
)

func TestInterceptorLogsCalls(t *testing.T) {
	for _, test := range []struct {
		name string
		code trevrpc.Code
		want string
	}{
		{name: "ok", code: trevrpc.CodeOK, want: `level=DEBUG msg="request completed" method=/number.v1.NumberService/Add`},
		{name: "error", code: trevrpc.CodeInvalidArgument, want: `level=ERROR msg="request error" method=/number.v1.NumberService/Add code=InvalidArgument`},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&output, &slog.HandlerOptions{
				Level: slog.LevelDebug,
				ReplaceAttr: func(_ []string, attr slog.Attr) slog.Attr {
					if attr.Key == slog.TimeKey {
						return slog.Attr{}
					}
					return attr
				},
			}))

			ctx := context.Background()
			next, finish, err := New(logger).Intercept(ctx, trevrpc.HandlerContext{FullMethod: "/number.v1.NumberService/Add"})
			if err != nil || next != ctx || finish == nil {
				t.Fatalf("Intercept() = (%v, finish %t, %v), want the same context, a finish and no error", next, finish != nil, err)
			}
			finish(trevrpc.RPCFinished{Code: test.code})

			lines := strings.Split(strings.TrimSpace(output.String()), "\n")
			if len(lines) != 2 {
				t.Fatalf("logged %d lines, want 2:\n%s", len(lines), output.String())
			}
			if want := `level=DEBUG msg="request received" method=/number.v1.NumberService/Add`; lines[0] != want {
				t.Errorf("start line = %q, want %q", lines[0], want)
			}
			if lines[1] != test.want {
				t.Errorf("finish line = %q, want %q", lines[1], test.want)
			}
		})
	}
}
