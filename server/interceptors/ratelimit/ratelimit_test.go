package ratelimit

import (
	"context"
	"net/netip"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
)

const (
	testLoginMethod  = "/auth.v1.AuthService/Login"
	testSignupMethod = "/auth.v1.AuthService/Signup"
)

func addr(value string) netip.AddrPort {
	return netip.MustParseAddrPort(value)
}

func TestRateLimitBurstAndRefill(t *testing.T) {
	now := time.Date(2026, time.September, 16, 12, 0, 0, 0, time.UTC)
	interceptor := New(map[string]Policy{
		testLoginMethod: {Requests: 2, Window: 10 * time.Second},
	})
	interceptor.now = func() time.Time { return now }

	for request := 1; request <= 2; request++ {
		if retryAfter, allowed := interceptor.allow(testLoginMethod, addr("192.0.2.1:1000")); !allowed {
			t.Fatalf("request %d allowed = false, retry after %v", request, retryAfter)
		}
	}
	if retryAfter, allowed := interceptor.allow(testLoginMethod, addr("192.0.2.1:2000")); allowed || retryAfter != 5*time.Second {
		t.Fatalf("exhausted request = (%v, %t), want (5s, false)", retryAfter, allowed)
	}

	now = now.Add(5 * time.Second)
	if retryAfter, allowed := interceptor.allow(testLoginMethod, addr("192.0.2.1:3000")); !allowed {
		t.Fatalf("refilled request allowed = false, retry after %v", retryAfter)
	}
}

func TestRateLimitIsolatesMethodsAndClients(t *testing.T) {
	interceptor := New(map[string]Policy{
		testLoginMethod:  {Requests: 1, Window: time.Minute},
		testSignupMethod: {Requests: 1, Window: time.Minute},
	})

	if _, allowed := interceptor.allow(testLoginMethod, addr("192.0.2.1:1000")); !allowed {
		t.Fatal("initial login request was rejected")
	}
	if _, allowed := interceptor.allow(testLoginMethod, addr("192.0.2.1:2000")); allowed {
		t.Fatal("login request with a different source port was allowed")
	}
	if _, allowed := interceptor.allow(testSignupMethod, addr("192.0.2.1:3000")); !allowed {
		t.Fatal("signup request shared the login bucket")
	}
	if _, allowed := interceptor.allow(testLoginMethod, addr("192.0.2.2:1000")); !allowed {
		t.Fatal("different client IP shared the login bucket")
	}
}

func TestRateLimitBypassesUnconfiguredMethods(t *testing.T) {
	interceptor := New(map[string]Policy{
		testLoginMethod: {Requests: 1, Window: time.Minute},
	})

	for range 10 {
		if _, allowed := interceptor.allow("/cursor.v1.CursorService/Send", netip.AddrPort{}); !allowed {
			t.Fatal("unconfigured method was rejected")
		}
	}
	if len(interceptor.buckets) != 0 {
		t.Fatalf("bucket count = %d, want 0", len(interceptor.buckets))
	}
}

func TestRateLimitClientIP(t *testing.T) {
	for _, test := range []struct {
		remoteAddr netip.AddrPort
		want       string
	}{
		{remoteAddr: addr("192.0.2.1:1000"), want: "192.0.2.1"},
		{remoteAddr: addr("[2001:db8::1]:1000"), want: "2001:db8::1"},
		{remoteAddr: netip.AddrPort{}, want: "unknown"},
	} {
		if got := clientIP(test.remoteAddr); got != test.want {
			t.Errorf("clientIP(%v) = %q, want %q", test.remoteAddr, got, test.want)
		}
	}
}

func TestRateLimitEvictsStaleBuckets(t *testing.T) {
	now := time.Date(2026, time.September, 16, 12, 0, 0, 0, time.UTC)
	interceptor := New(map[string]Policy{
		testLoginMethod: {Requests: 1, Window: 10 * time.Second},
	})
	interceptor.now = func() time.Time { return now }

	interceptor.allow(testLoginMethod, addr("192.0.2.1:1000"))
	now = now.Add(10 * time.Second)
	interceptor.allow(testLoginMethod, addr("192.0.2.2:1000"))

	if len(interceptor.buckets) != 1 {
		t.Fatalf("bucket count = %d, want 1", len(interceptor.buckets))
	}
	for key := range interceptor.buckets {
		if key.clientIP != "192.0.2.2" {
			t.Errorf("remaining client IP = %q, want 192.0.2.2", key.clientIP)
		}
	}
}

func TestRateLimitBoundsBuckets(t *testing.T) {
	interceptor := New(map[string]Policy{
		testLoginMethod: {Requests: 1, Window: time.Hour},
	})
	interceptor.maxBuckets = 2

	interceptor.allow(testLoginMethod, addr("192.0.2.1:1000"))
	interceptor.allow(testLoginMethod, addr("192.0.2.2:1000"))
	interceptor.allow(testLoginMethod, addr("192.0.2.3:1000"))

	if len(interceptor.buckets) != interceptor.maxBuckets {
		t.Fatalf("bucket count = %d, want %d", len(interceptor.buckets), interceptor.maxBuckets)
	}
	if _, ok := interceptor.buckets[rateLimitKey{method: testLoginMethod, clientIP: "192.0.2.1"}]; ok {
		t.Error("oldest bucket was not evicted")
	}
	if interceptor.bucketOrder.Len() != interceptor.maxBuckets {
		t.Errorf("bucket order length = %d, want %d", interceptor.bucketOrder.Len(), interceptor.maxBuckets)
	}
}

func TestRateLimitConcurrentRequests(t *testing.T) {
	const burst = 10
	interceptor := New(map[string]Policy{
		testLoginMethod: {Requests: burst, Window: time.Minute},
	})

	var allowed atomic.Int64
	var wait sync.WaitGroup
	for range 100 {
		wait.Go(func() {
			if _, ok := interceptor.allow(testLoginMethod, addr("192.0.2.1:1000")); ok {
				allowed.Add(1)
			}
		})
	}
	wait.Wait()

	if got := allowed.Load(); got != burst {
		t.Errorf("allowed requests = %d, want %d", got, burst)
	}
}

func TestRateLimitIntercept(t *testing.T) {
	now := time.Date(2026, time.September, 16, 12, 0, 0, 0, time.UTC)
	interceptor := New(map[string]Policy{
		testLoginMethod: {Requests: 1, Window: 1500 * time.Millisecond},
	})
	interceptor.now = func() time.Time { return now }
	intercept := func(method string) error {
		ctx := context.Background()
		next, finish, err := interceptor.Intercept(ctx, trevrpc.HandlerContext{
			FullMethod: method,
			RemoteAddr: addr("192.0.2.1:1000"),
		})
		if next != ctx || finish != nil {
			t.Errorf("Intercept(%s) = (%v, finish %t), want the same context and no finish", method, next, finish != nil)
		}
		return err
	}

	if err := intercept(testLoginMethod); err != nil {
		t.Fatalf("first call error = %v", err)
	}
	status := trevrpc.StatusFromError(intercept(testLoginMethod))
	if status.Code != trevrpc.CodeResourceExhausted {
		t.Fatalf("second call code = %v, want ResourceExhausted", status.Code)
	}
	if got := string(status.Metadata[RetryAfterKey]); got != "2" {
		t.Errorf("%s = %q, want 2", RetryAfterKey, got)
	}

	if err := intercept("/cursor.v1.CursorService/Send"); err != nil {
		t.Fatalf("unconfigured call error = %v", err)
	}
}
