package log

import (
	"bytes"
	"context"
	"log/slog"
	"net/netip"
	"strings"
	"testing"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
	"trev.zip/template/stack/server/v2/auth"
)

func TestInterceptorLogsCalls(t *testing.T) {
	for _, test := range []struct {
		name       string
		remoteAddr netip.AddrPort
		username   string
		code       trevrpc.Code
		wantStart  string
		wantFinish string
	}{
		{
			name:       "ok",
			remoteAddr: netip.MustParseAddrPort("203.0.113.7:4433"),
			code:       trevrpc.CodeOK,
			wantStart:  `level=DEBUG msg="request received" method=/cursor.v1.CursorService/Send client_ip=203.0.113.7`,
			wantFinish: `level=DEBUG msg="request completed" method=/cursor.v1.CursorService/Send client_ip=203.0.113.7`,
		},
		{
			name:       "error",
			remoteAddr: netip.MustParseAddrPort("[2001:db8::1]:4433"),
			code:       trevrpc.CodeInvalidArgument,
			wantStart:  `level=DEBUG msg="request received" method=/cursor.v1.CursorService/Send client_ip=2001:db8::1`,
			wantFinish: `level=ERROR msg="request error" method=/cursor.v1.CursorService/Send client_ip=2001:db8::1 code=InvalidArgument`,
		},
		{
			name:       "cancelled",
			remoteAddr: netip.MustParseAddrPort("203.0.113.7:4433"),
			code:       trevrpc.CodeCancelled,
			wantStart:  `level=DEBUG msg="request received" method=/cursor.v1.CursorService/Send client_ip=203.0.113.7`,
			wantFinish: `level=DEBUG msg="request cancelled" method=/cursor.v1.CursorService/Send client_ip=203.0.113.7`,
		},
		{
			name:       "authenticated",
			remoteAddr: netip.MustParseAddrPort("203.0.113.7:4433"),
			username:   "alice",
			code:       trevrpc.CodeOK,
			wantStart:  `level=DEBUG msg="request received" method=/cursor.v1.CursorService/Send client_ip=203.0.113.7 user=alice`,
			wantFinish: `level=DEBUG msg="request completed" method=/cursor.v1.CursorService/Send client_ip=203.0.113.7 user=alice`,
		},
		{
			name:       "unknown client",
			code:       trevrpc.CodeInvalidArgument,
			wantStart:  `level=DEBUG msg="request received" method=/cursor.v1.CursorService/Send`,
			wantFinish: `level=ERROR msg="request error" method=/cursor.v1.CursorService/Send code=InvalidArgument`,
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
			if test.username != "" {
				ctx = auth.ContextWithUsername(ctx, test.username)
			}
			next, finish, err := New(logger).Intercept(ctx, trevrpc.HandlerContext{
				FullMethod: "/cursor.v1.CursorService/Send",
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
