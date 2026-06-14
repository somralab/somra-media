package api

import "net/http"

// RateLimiter is the contract every rate-limit strategy implements. The API
// gateway depends only on this interface so a real implementation (token
// bucket, sliding window, Redis-backed cluster limiter) can land without
// touching the router.
//
// Allow returns true if the request should be served. When false, the caller
// is expected to surface a 429 response. The interface intentionally does
// not return retry-after data; that lives in the response handler shaped by
// the limiter implementation.
type RateLimiter interface {
	Allow(r *http.Request) bool
}

// NoopRateLimiter accepts every request. Used as the safe default in
// Sprint 01 where no real limiter is wired yet; replacing this is a single
// constructor argument when a real implementation exists.
type NoopRateLimiter struct{}

// Allow always returns true.
func (NoopRateLimiter) Allow(*http.Request) bool { return true }

// RateLimitMiddleware blocks requests rejected by limiter with a 429
// JSON envelope. limiter must not be nil; the constructor wires
// NoopRateLimiter when callers leave the option unset.
func RateLimitMiddleware(limiter RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limiter == nil || limiter.Allow(r) {
				next.ServeHTTP(w, r)
				return
			}
			writeError(w, r, errTooManyRequests)
		})
	}
}
