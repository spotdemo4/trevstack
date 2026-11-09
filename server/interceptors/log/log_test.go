package log

import (
	"bytes"
	"context"
	"log/slog"
	"net/netip"
	"strings"
	"testing"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
)

func TestInterceptorLogsCalls(t *testing.T) {
	for _, test := range []struct {
		name       string
		remoteAddr netip.AddrPort
		code       trevrpc.Code
		wantStart  string
		wantFinish string
	}{
		{
			name:       "ok",
			remoteAddr: netip.MustParseAddrPort("203.0.113.7:4433"),
			code:       trevrpc.CodeOK,
			wantStart:  `level=DEBUG msg="request received" method=/number.v1.NumberService/Add client_ip=203.0.113.7`,
			wantFinish: `level=DEBUG msg="request completed" method=/number.v1.NumberService/Add client_ip=203.0.113.7`,
		},
		{
			name:       "error",
			remoteAddr: netip.MustParseAddrPort("[2001:db8::1]:4433"),
			code:       trevrpc.CodeInvalidArgument,
			wantStart:  `level=DEBUG msg="request received" method=/number.v1.NumberService/Add client_ip=2001:db8::1`,
			wantFinish: `level=ERROR msg="request error" method=/number.v1.NumberService/Add client_ip=2001:db8::1 code=InvalidArgument`,
		},
		{
			name:       "unknown client",
			code:       trevrpc.CodeInvalidArgument,
			wantStart:  `level=DEBUG msg="request received" method=/number.v1.NumberService/Add`,
			wantFinish: `level=ERROR msg="request error" method=/number.v1.NumberService/Add code=InvalidArgument`,
		},
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
			next, finish, err := New(logger).Intercept(ctx, trevrpc.HandlerContext{
				FullMethod: "/number.v1.NumberService/Add",
				RemoteAddr: test.remoteAddr,
			})
			if err != nil || next != ctx || finish == nil {
				t.Fatalf("Intercept() = (%v, finish %t, %v), want the same context, a finish and no error", next, finish != nil, err)
			}
			finish(trevrpc.RPCFinished{Code: test.code})

			lines := strings.Split(strings.TrimSpace(output.String()), "\n")
			if len(lines) != 2 {
				t.Fatalf("logged %d lines, want 2:\n%s", len(lines), output.String())
			}
			if lines[0] != test.wantStart {
				t.Errorf("start line = %q, want %q", lines[0], test.wantStart)
			}
			if lines[1] != test.wantFinish {
				t.Errorf("finish line = %q, want %q", lines[1], test.wantFinish)
			}
		})
	}
}
