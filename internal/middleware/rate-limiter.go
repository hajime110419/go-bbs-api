package middleware

import (
	"net/http"

	"github.com/juju/ratelimit"
)

// RateLimiterMiddleware returns an HTTP middleware that applies a rate limit
// using the juju/ratelimit token bucket.
func RateLimiterMiddleware(bucket *ratelimit.Bucket) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Check if a token is available. TakeAvailable(1) attempts to consume 1 token
			// immediately and returns 0 if none are avaliable.
			if bucket.TakeAvailable(1) == 0 {
				// If rate limit is exceeded, return 429 Too Many Requests.
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				http.Error(w, `{"error": "Too many requests. Please try again later."}`, http.StatusTooManyRequests)
				return
			}
			// If a token is consumed, proceed to the next handler.
			next(w, r)
		}
	}
}
