package ratelimit

import (
	"context"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect"
)

const (
	testLoginProcedure  = "/auth.v1.AuthService/Login"
	testSignupProcedure = "/auth.v1.AuthService/Signup"
)

type rateLimitStreamingConn struct {
	spec connect.Spec
	peer connect.Peer
}

func (c *rateLimitStreamingConn) Spec() connect.Spec           { return c.spec }
func (c *rateLimitStreamingConn) Peer() connect.Peer           { return c.peer }
func (c *rateLimitStreamingConn) Receive(any) error            { return nil }
func (c *rateLimitStreamingConn) RequestHeader() http.Header   { return make(http.Header) }
func (c *rateLimitStreamingConn) Send(any) error               { return nil }
func (c *rateLimitStreamingConn) ResponseHeader() http.Header  { return make(http.Header) }
func (c *rateLimitStreamingConn) ResponseTrailer() http.Header { return make(http.Header) }

func TestRateLimitInterceptorBurstAndRefill(t *testing.T) {
	now := time.Date(2026, time.September, 16, 12, 0, 0, 0, time.UTC)
	interceptor := NewRateLimitInterceptor(map[string]RateLimitPolicy{
		testLoginProcedure: {Requests: 2, Window: 10 * time.Second},
	}, nil)
	interceptor.now = func() time.Time { return now }

	for request := 1; request <= 2; request++ {
		if retryAfter, allowed := interceptor.allow(testLoginProcedure, "192.0.2.1:1000", nil); !allowed {
			t.Fatalf("request %d allowed = false, retry after %v", request, retryAfter)
		}
	}
	if retryAfter, allowed := interceptor.allow(testLoginProcedure, "192.0.2.1:2000", nil); allowed || retryAfter != 5*time.Second {
		t.Fatalf("exhausted request = (%v, %t), want (5s, false)", retryAfter, allowed)
	}

	now = now.Add(5 * time.Second)
	if retryAfter, allowed := interceptor.allow(testLoginProcedure, "192.0.2.1:3000", nil); !allowed {
		t.Fatalf("refilled request allowed = false, retry after %v", retryAfter)
	}
}

func TestRateLimitInterceptorIsolatesProceduresAndPeers(t *testing.T) {
	interceptor := NewRateLimitInterceptor(map[string]RateLimitPolicy{
		testLoginProcedure:  {Requests: 1, Window: time.Minute},
		testSignupProcedure: {Requests: 1, Window: time.Minute},
	}, nil)

	if _, allowed := interceptor.allow(testLoginProcedure, "192.0.2.1:1000", nil); !allowed {
		t.Fatal("initial login request was rejected")
	}
	if _, allowed := interceptor.allow(testLoginProcedure, "192.0.2.1:2000", nil); allowed {
		t.Fatal("login request with a different source port was allowed")
	}
	if _, allowed := interceptor.allow(testSignupProcedure, "192.0.2.1:3000", nil); !allowed {
		t.Fatal("signup request shared the login bucket")
	}
	if _, allowed := interceptor.allow(testLoginProcedure, "192.0.2.2:1000", nil); !allowed {
		t.Fatal("different peer IP shared the login bucket")
	}
}

func TestRateLimitInterceptorBypassesUnconfiguredProcedures(t *testing.T) {
	interceptor := NewRateLimitInterceptor(map[string]RateLimitPolicy{
		testLoginProcedure: {Requests: 1, Window: time.Minute},
	}, nil)

	for range 10 {
		if _, allowed := interceptor.allow("/auth.v1.AuthService/Logout", "not-a-peer", nil); !allowed {
			t.Fatal("unconfigured procedure was rejected")
		}
	}
	if len(interceptor.buckets) != 0 {
		t.Fatalf("bucket count = %d, want 0", len(interceptor.buckets))
	}
}

func TestRateLimitInterceptorClientIP(t *testing.T) {
	tests := []struct {
		name             string
		trustedProxies   []netip.Prefix
		peer             string
		xForwardedFor    string
		want             string
	}{
		{
			name:          "direct peer by default",
			peer:          "192.0.2.1:1000",
			xForwardedFor: "198.51.100.1",
			want:          "192.0.2.1",
		},
		{
			name:           "trusted proxy",
			trustedProxies: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
			peer:           "10.0.0.2:1000",
			xForwardedFor:  "198.51.100.1",
			want:           "198.51.100.1",
		},
		{
			name: "trusted proxy chain",
			trustedProxies: []netip.Prefix{
				netip.MustParsePrefix("10.0.0.0/8"),
				netip.MustParsePrefix("192.0.2.0/24"),
			},
			peer:          "10.0.0.2:1000",
			xForwardedFor: "203.0.113.1, 192.0.2.3",
			want:          "203.0.113.1",
		},
		{
			name:           "malformed forwarded address",
			trustedProxies: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
			peer:           "10.0.0.2:1000",
			xForwardedFor:  "198.51.100.1, invalid",
			want:           "10.0.0.2",
		},
		{
			name:           "malformed spoofed prefix",
			trustedProxies: []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
			peer:           "10.0.0.2:1000",
			xForwardedFor:  "invalid, 198.51.100.1",
			want:           "198.51.100.1",
		},
		{
			name: "too many forwarded hops",
			trustedProxies: []netip.Prefix{
				netip.MustParsePrefix("10.0.0.0/8"),
				netip.MustParsePrefix("198.51.100.0/24"),
			},
			peer:           "10.0.0.2:1000",
			xForwardedFor:  strings.Repeat("198.51.100.1,", maxForwardedForHops) + "198.51.100.1",
			want:           "198.51.100.1",
		},
		{
			name: "all forwarded hops trusted",
			trustedProxies: []netip.Prefix{
				netip.MustParsePrefix("10.0.0.0/8"),
				netip.MustParsePrefix("192.0.2.0/24"),
			},
			peer:          "10.0.0.2:1000",
			xForwardedFor: "192.0.2.3",
			want:          "192.0.2.3",
		},
		{
			name: "IPv4-mapped peer",
			peer: "[::ffff:192.0.2.1]:1000",
			want: "192.0.2.1",
		},
		{
			name: "malformed direct peer",
			peer: "invalid",
			want: "unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			interceptor := NewRateLimitInterceptor(nil, test.trustedProxies)
			header := make(http.Header)
			if test.xForwardedFor != "" {
				header.Set("X-Forwarded-For", test.xForwardedFor)
			}
			if got := interceptor.clientIP(test.peer, header); got != test.want {
				t.Errorf("clientIP() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestRateLimitInterceptorEvictsStaleBuckets(t *testing.T) {
	now := time.Date(2026, time.September, 16, 12, 0, 0, 0, time.UTC)
	interceptor := NewRateLimitInterceptor(map[string]RateLimitPolicy{
		testLoginProcedure: {Requests: 1, Window: 10 * time.Second},
	}, nil)
	interceptor.now = func() time.Time { return now }

	interceptor.allow(testLoginProcedure, "192.0.2.1:1000", nil)
	now = now.Add(10 * time.Second)
	interceptor.allow(testLoginProcedure, "192.0.2.2:1000", nil)

	if len(interceptor.buckets) != 1 {
		t.Fatalf("bucket count = %d, want 1", len(interceptor.buckets))
	}
	for key := range interceptor.buckets {
		if key.clientIP != "192.0.2.2" {
			t.Errorf("remaining client IP = %q, want 192.0.2.2", key.clientIP)
		}
	}
}

func TestRateLimitInterceptorBoundsBuckets(t *testing.T) {
	interceptor := NewRateLimitInterceptor(map[string]RateLimitPolicy{
		testLoginProcedure: {Requests: 1, Window: time.Hour},
	}, nil)
	interceptor.maxBuckets = 2

	interceptor.allow(testLoginProcedure, "192.0.2.1:1000", nil)
	interceptor.allow(testLoginProcedure, "192.0.2.2:1000", nil)
	interceptor.allow(testLoginProcedure, "192.0.2.3:1000", nil)

	if len(interceptor.buckets) != interceptor.maxBuckets {
		t.Fatalf("bucket count = %d, want %d", len(interceptor.buckets), interceptor.maxBuckets)
	}
	if _, ok := interceptor.buckets[rateLimitKey{procedure: testLoginProcedure, clientIP: "192.0.2.1"}]; ok {
		t.Error("oldest bucket was not evicted")
	}
	if interceptor.bucketOrder.Len() != interceptor.maxBuckets {
		t.Errorf("bucket order length = %d, want %d", interceptor.bucketOrder.Len(), interceptor.maxBuckets)
	}
}

func TestRateLimitInterceptorConcurrentRequests(t *testing.T) {
	const burst = 10
	interceptor := NewRateLimitInterceptor(map[string]RateLimitPolicy{
		testLoginProcedure: {Requests: burst, Window: time.Minute},
	}, nil)

	var allowed atomic.Int64
	var wait sync.WaitGroup
	for range 100 {
		wait.Go(func() {
			if _, ok := interceptor.allow(testLoginProcedure, "192.0.2.1:1000", nil); ok {
				allowed.Add(1)
			}
		})
	}
	wait.Wait()

	if got := allowed.Load(); got != burst {
		t.Errorf("allowed requests = %d, want %d", got, burst)
	}
}

func TestRateLimitInterceptorStreamingHandler(t *testing.T) {
	interceptor := NewRateLimitInterceptor(map[string]RateLimitPolicy{
		testLoginProcedure: {Requests: 1, Window: time.Minute},
	}, nil)
	conn := &rateLimitStreamingConn{
		spec: connect.Spec{Procedure: testLoginProcedure},
		peer: connect.Peer{Addr: "192.0.2.1:1000"},
	}
	called := 0
	wrapped := interceptor.WrapStreamingHandler(func(context.Context, connect.StreamingHandlerConn) error {
		called++
		return nil
	})

	if err := wrapped(context.Background(), conn); err != nil {
		t.Fatalf("first streaming call error = %v", err)
	}
	err := wrapped(context.Background(), conn)
	if got := connect.CodeOf(err); got != connect.CodeResourceExhausted {
		t.Fatalf("second streaming call code = %v, want ResourceExhausted", got)
	}
	if called != 1 {
		t.Errorf("next called = %d times, want 1", called)
	}

	conn.spec.Procedure = "/auth.v1.AuthService/Logout"
	if err := wrapped(context.Background(), conn); err != nil {
		t.Fatalf("unconfigured streaming call error = %v", err)
	}
	if called != 2 {
		t.Errorf("next called = %d times after bypass, want 2", called)
	}
}

func TestRateLimitError(t *testing.T) {
	err := rateLimitError(1500 * time.Millisecond)
	if got := connect.CodeOf(err); got != connect.CodeResourceExhausted {
		t.Errorf("code = %v, want ResourceExhausted", got)
	}
	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("error type = %T, want *connect.Error", err)
	}
	if got := connectErr.Meta().Get("Retry-After"); got != "2" {
		t.Errorf("Retry-After = %q, want 2", got)
	}
}
