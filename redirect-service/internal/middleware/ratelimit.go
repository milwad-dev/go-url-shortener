package middleware

import (
	"context"
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	// idleClientTTL is how long a client's limiter is kept after its last request.
	idleClientTTL = 3 * time.Minute
	// cleanupInterval is how often idle limiters are evicted.
	cleanupInterval = time.Minute
)

// skipPaths are never rate limited: orchestrator probes must keep working under load.
var skipPaths = map[string]bool{
	"/healthz": true,
	"/readyz":  true,
}

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter is a per-client-IP token-bucket rate limiter.
type RateLimiter struct {
	rps   rate.Limit
	burst int

	mu      sync.Mutex
	clients map[string]*client
}

// NewRateLimiter allows each client IP rps requests per second, with bursts up to burst.
func NewRateLimiter(ctx context.Context, rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		rps:     rate.Limit(rps),
		burst:   burst,
		clients: make(map[string]*client),
	}

	go rl.cleanup(ctx)

	return rl
}

// Middleware rejects requests over the client's limit with 429 and a Retry-After header.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if skipPaths[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		reservation := rl.limiterFor(clientIP(r)).Reserve()
		if delay := reservation.Delay(); delay > 0 {
			reservation.Cancel()
			w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(delay.Seconds()))))
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) limiterFor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	c, ok := rl.clients[ip]
	if !ok {
		c = &client{limiter: rate.NewLimiter(rl.rps, rl.burst)}
		rl.clients[ip] = c
	}
	c.lastSeen = time.Now()

	return c.limiter
}

func (rl *RateLimiter) cleanup(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rl.mu.Lock()
			for ip, c := range rl.clients {
				if time.Since(c.lastSeen) > idleClientTTL {
					delete(rl.clients, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// clientIP returns the caller's IP from the TCP connection.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
