// Package ratelimit limits the rate of calls per client.
package ratelimit

import (
	"container/list"
	"context"
	"math"
	"net/netip"
	"strconv"
	"sync"
	"time"

	trevrpc "trev.zip/llc/trevrpc/trevrpc-go"
)

// RetryAfterKey is the metadata of a rejected call that holds the number of
// seconds to wait before retrying.
const RetryAfterKey = "retry-after"

const (
	maxRateLimitBuckets         = 10_000
	maxRateLimitCleanupInterval = time.Minute
)

// Policy allows Requests calls per Window.
type Policy struct {
	Requests int
	Window   time.Duration
}

type rateLimitKey struct {
	method   string
	clientIP string
}

type rateLimitBucket struct {
	tokens     float64
	lastRefill time.Time
	element    *list.Element
}

// Interceptor limits calls per method and client IP with token buckets. It
// runs once per handler invocation, so a call TrevRPC replays for optimistic
// stream compression spends one token.
type Interceptor struct {
	mu              sync.Mutex
	policies        map[string]Policy
	buckets         map[rateLimitKey]*rateLimitBucket
	bucketOrder     *list.List
	maxBuckets      int
	now             func() time.Time
	lastCleanup     time.Time
	cleanupInterval time.Duration
}

// New returns an Interceptor that applies policies by full method name, such
// as "/auth.v1.AuthService/Login". It panics on an invalid policy.
func New(policies map[string]Policy) *Interceptor {
	copiedPolicies := make(map[string]Policy, len(policies))
	cleanupInterval := maxRateLimitCleanupInterval
	for method, policy := range policies {
		if method == "" {
			panic("rate limit method must not be empty")
		}
		if policy.Requests <= 0 {
			panic("rate limit requests must be positive")
		}
		if policy.Window <= 0 {
			panic("rate limit window must be positive")
		}
		copiedPolicies[method] = policy
		if policy.Window < cleanupInterval {
			cleanupInterval = policy.Window
		}
	}

	return &Interceptor{
		policies:        copiedPolicies,
		buckets:         make(map[rateLimitKey]*rateLimitBucket),
		bucketOrder:     list.New(),
		maxBuckets:      maxRateLimitBuckets,
		now:             time.Now,
		cleanupInterval: cleanupInterval,
	}
}

// Intercept rejects calls over their method's limit with ResourceExhausted
// and RetryAfterKey metadata.
func (i *Interceptor) Intercept(
	ctx context.Context,
	info trevrpc.HandlerContext,
) (context.Context, func(trevrpc.RPCFinished), error) {
	retryAfter, allowed := i.allow(info.FullMethod, info.RemoteAddr)
	if allowed {
		return ctx, nil, nil
	}
	seconds := max(int64(math.Ceil(retryAfter.Seconds())), 1)
	status := trevrpc.ResourceExhausted("rate limit exceeded")
	status.Metadata = trevrpc.Metadata{RetryAfterKey: []byte(strconv.FormatInt(seconds, 10))}
	return ctx, nil, status
}

func (i *Interceptor) allow(method string, remoteAddr netip.AddrPort) (time.Duration, bool) {
	policy, ok := i.policies[method]
	if !ok {
		return 0, true
	}

	now := i.now()
	key := rateLimitKey{
		method:   method,
		clientIP: clientIP(remoteAddr),
	}

	i.mu.Lock()
	defer i.mu.Unlock()

	i.cleanup(now)
	bucket, ok := i.buckets[key]
	if !ok {
		if len(i.buckets) >= i.maxBuckets {
			oldest := i.bucketOrder.Front()
			delete(i.buckets, oldest.Value.(rateLimitKey))
			i.bucketOrder.Remove(oldest)
		}
		bucket = &rateLimitBucket{
			tokens:     float64(policy.Requests),
			lastRefill: now,
			element:    i.bucketOrder.PushBack(key),
		}
		i.buckets[key] = bucket
	} else {
		i.bucketOrder.MoveToBack(bucket.element)
	}

	if elapsed := now.Sub(bucket.lastRefill); elapsed > 0 {
		refill := float64(elapsed) * float64(policy.Requests) / float64(policy.Window)
		bucket.tokens = min(float64(policy.Requests), bucket.tokens+refill)
		bucket.lastRefill = now
	}

	if bucket.tokens >= 1 {
		bucket.tokens--
		return 0, true
	}

	retryAfter := time.Duration(math.Ceil(
		(1 - bucket.tokens) * float64(policy.Window) / float64(policy.Requests),
	))
	return max(retryAfter, time.Nanosecond), false
}

func (i *Interceptor) cleanup(now time.Time) {
	if !i.lastCleanup.IsZero() && now.Sub(i.lastCleanup) < i.cleanupInterval {
		return
	}
	for key, bucket := range i.buckets {
		if now.Sub(bucket.lastRefill) >= i.policies[key.method].Window {
			delete(i.buckets, key)
			i.bucketOrder.Remove(bucket.element)
		}
	}
	i.lastCleanup = now
}

// clientIP returns the address that identifies the client. TrevRPC reports
// IPv4 clients with IPv4 addresses on every transport.
func clientIP(remoteAddr netip.AddrPort) string {
	if !remoteAddr.IsValid() {
		return "unknown"
	}
	return remoteAddr.Addr().String()
}
