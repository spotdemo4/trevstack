// Package ratelimit limits the rate of Connect calls per client.
package ratelimit

import (
	"container/list"
	"context"
	"math"
	"net"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect/v2"
)

const (
	maxForwardedForBytes        = 4 * 1024
	maxForwardedForHops         = 32
	maxRateLimitBuckets         = 10_000
	maxRateLimitCleanupInterval = time.Minute
)

// Policy allows Requests calls per Window.
type Policy struct {
	Requests int
	Window   time.Duration
}

type rateLimitKey struct {
	procedure string
	clientIP  string
}

type rateLimitBucket struct {
	tokens     float64
	lastRefill time.Time
	element    *list.Element
}

// Interceptor limits calls per procedure and client IP with token buckets.
type Interceptor struct {
	mu              sync.Mutex
	policies        map[string]Policy
	buckets         map[rateLimitKey]*rateLimitBucket
	bucketOrder     *list.List
	maxBuckets      int
	trustedProxies  []netip.Prefix
	now             func() time.Time
	lastCleanup     time.Time
	cleanupInterval time.Duration
}

// New returns an Interceptor that applies policies by procedure. Clients in
// trustedProxies may set X-Forwarded-For. It panics on an invalid policy or
// prefix.
func New(
	policies map[string]Policy,
	trustedProxies []netip.Prefix,
) *Interceptor {
	copiedPolicies := make(map[string]Policy, len(policies))
	cleanupInterval := maxRateLimitCleanupInterval
	for procedure, policy := range policies {
		if procedure == "" {
			panic("rate limit procedure must not be empty")
		}
		if policy.Requests <= 0 {
			panic("rate limit requests must be positive")
		}
		if policy.Window <= 0 {
			panic("rate limit window must be positive")
		}
		copiedPolicies[procedure] = policy
		if policy.Window < cleanupInterval {
			cleanupInterval = policy.Window
		}
	}

	copiedProxies := make([]netip.Prefix, len(trustedProxies))
	for index, prefix := range trustedProxies {
		if !prefix.IsValid() {
			panic("trusted proxy prefix must be valid")
		}
		copiedProxies[index] = prefix.Masked()
	}

	return &Interceptor{
		policies:        copiedPolicies,
		buckets:         make(map[rateLimitKey]*rateLimitBucket),
		bucketOrder:     list.New(),
		maxBuckets:      maxRateLimitBuckets,
		trustedProxies:  copiedProxies,
		now:             time.Now,
		cleanupInterval: cleanupInterval,
	}
}

// WrapServer rejects calls over their procedure's limit with
// CodeResourceExhausted and a Retry-After header.
func (i *Interceptor) WrapServer(next connect.ServerFunc) connect.ServerFunc {
	return func(
		ctx context.Context,
		spec connect.Spec,
		stream connect.ServerStream,
	) error {
		info, ok := connect.CallInfoForServerContext(ctx)
		if !ok {
			return connect.NewError(connect.CodeInternal, "connect call info unavailable")
		}
		if retryAfter, allowed := i.allow(
			spec.Procedure,
			info.PeerAddr,
			info.RequestHeader(),
		); !allowed {
			seconds := max(int64(math.Ceil(retryAfter.Seconds())), 1)
			info.ResponseHeader().Set("Retry-After", strconv.FormatInt(seconds, 10))
			return connect.NewError(connect.CodeResourceExhausted, "rate limit exceeded")
		}
		return next(ctx, spec, stream)
	}
}

func (i *Interceptor) allow(
	procedure string,
	peerAddr string,
	header *connect.Header,
) (time.Duration, bool) {
	policy, ok := i.policies[procedure]
	if !ok {
		return 0, true
	}

	now := i.now()
	key := rateLimitKey{
		procedure: procedure,
		clientIP:  i.clientIP(peerAddr, header),
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
		if now.Sub(bucket.lastRefill) >= i.policies[key.procedure].Window {
			delete(i.buckets, key)
			i.bucketOrder.Remove(bucket.element)
		}
	}
	i.lastCleanup = now
}

func (i *Interceptor) clientIP(peerAddr string, header *connect.Header) string {
	direct, ok := parsePeerIP(peerAddr)
	if !ok {
		return "unknown"
	}
	if !i.isTrustedProxy(direct) {
		return direct.String()
	}

	values := header.Values("X-Forwarded-For")
	if len(values) == 0 {
		return direct.String()
	}

	remainingBytes := maxForwardedForBytes
	hops := 0
	var candidate netip.Addr
	for _, value := range slices.Backward(values) {
		for value != "" {
			if hops >= maxForwardedForHops || remainingBytes <= 0 {
				return candidateOrDirect(candidate, direct)
			}

			start := max(len(value)-remainingBytes, 0)
			comma := strings.LastIndexByte(value[start:], ',')
			if comma < 0 && start > 0 {
				return candidateOrDirect(candidate, direct)
			}
			comma += start
			part := value[comma+1:]
			remainingBytes -= len(part) + 1
			hops++

			addr, err := netip.ParseAddr(strings.TrimSpace(part))
			if err != nil {
				return candidateOrDirect(candidate, direct)
			}
			candidate = addr.Unmap()
			if !i.isTrustedProxy(candidate) {
				return candidate.String()
			}
			if comma < 0 {
				break
			}
			value = value[:comma]
		}
	}
	return candidateOrDirect(candidate, direct)
}

func candidateOrDirect(candidate, direct netip.Addr) string {
	if candidate.IsValid() {
		return candidate.String()
	}
	return direct.String()
}

func (i *Interceptor) isTrustedProxy(addr netip.Addr) bool {
	for _, prefix := range i.trustedProxies {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

func parsePeerIP(peerAddr string) (netip.Addr, bool) {
	host, _, err := net.SplitHostPort(peerAddr)
	if err != nil {
		return netip.Addr{}, false
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}, false
	}
	return addr.Unmap(), true
}
