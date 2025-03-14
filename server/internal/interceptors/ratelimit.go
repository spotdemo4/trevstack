package interceptors

import (
	"context"
	"log"
	"sync"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type ratelimitInterceptor struct {
	key      string
	visitors map[string]*visitor
	mu       sync.Mutex
}

func NewRateLimitInterceptor(key string) *ratelimitInterceptor {
	rl := &ratelimitInterceptor{
		key:      key,
		visitors: make(map[string]*visitor),
		mu:       sync.Mutex{},
	}

	go rl.cleanupVisitors()

	return rl
}

func (i *ratelimitInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	// Same as previous UnaryInterceptorFunc.
	return connect.UnaryFunc(func(
		ctx context.Context,
		req connect.AnyRequest,
	) (connect.AnyResponse, error) {
		// Check if the request is from a client
		if req.Spec().IsClient {
			return next(ctx, req)
		}

		// Get ip
		log.Println(req.Peer().Addr)

		return next(ctx, req)
	})
}

func (*ratelimitInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return connect.StreamingClientFunc(func(
		ctx context.Context,
		spec connect.Spec,
	) connect.StreamingClientConn {
		return next(ctx, spec)
	})
}

func (i *ratelimitInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return connect.StreamingHandlerFunc(func(
		ctx context.Context,
		conn connect.StreamingHandlerConn,
	) error {
		// Get ip
		log.Println(conn.Peer().Query)

		return next(ctx, conn)
	})
}

func (i *ratelimitInterceptor) getVisitor(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	v, exists := i.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(1, 3)
		// Include the current time when creating a new visitor.
		i.visitors[ip] = &visitor{limiter, time.Now()}
		return limiter
	}

	// Update the last seen time for the visitor.
	v.lastSeen = time.Now()
	return v.limiter
}

// Every minute check the map for visitors that haven't been seen for
// more than 3 minutes and delete the entries.
func (i *ratelimitInterceptor) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		i.mu.Lock()
		for ip, v := range i.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(i.visitors, ip)
			}
		}
		i.mu.Unlock()
	}
}
