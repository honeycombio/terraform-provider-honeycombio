package v2

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

const (
	// HeaderRateLimit is the (draft07) recommended header from the IETF
	// on rate limiting.
	HeaderRateLimit = "RateLimit"
	// The value of the header is formatted "limit=X, remaining=Y, reset=Z".
	// Where:
	//   - X is the maximum number of requests allowed in the window
	//   - Y is the number of requests remaining in the window
	//   - Z is the number of seconds until the limit resets
	HeaderRateLimitFmt = "limit=%d, remaining=%d, reset=%d"

	// HeaderRetryAfter is the RFC7231 header used to indicate when a client should
	// retry requests in UTC time.
	HeaderRetryAfter = "Retry-After"
)

var rng *rand.Rand

func init() {
	// initialize the random number generator
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// rateLimitBackoff calculates the backoff time for a rate limited request
// based on the possible response headers.
// The function will first try to get the reset time from the rate limit header.
//
// If the rate limit header is not present, or the reset time is in the past,
// the function will return a random backoff time between min and max.
func rateLimitBackoff(min, max time.Duration, r *http.Response) time.Duration {
	// calculate some jitter for a little extra fuzziness to avoid thundering herds
	jitter := time.Duration(rng.Float64() * float64(max-min))

	var reset time.Duration
	if v := r.Header.Get(HeaderRateLimit); v != "" {
		// we currently only care about the reset time
		_, _, resetSeconds, err := parseRateLimitHeader(v)
		if err == nil {
			reset = time.Duration(resetSeconds) * time.Second
		}
	} else if v := r.Header.Get(HeaderRetryAfter); v != "" {
		// if we can't get the ratelimit header, try the retry-after header
		retryTime, err := time.Parse(time.RFC3339, v)
		if err == nil {
			reset = time.Until(retryTime)
		}
	}

	// only update min if the time to wait is longer
	if reset > min {
		min = reset
	}
	return min + jitter
}

// parseRateLimitHeader parses the rate limit header into its constituent parts.
//
// The header is expected to be in the format "limit=X, remaining=Y, reset=Z".
// Where:
//   - X is the maximum number of requests allowed in the window
//   - Y is the number of requests remaining in the window
//   - Z is the number of seconds until the limit resets
func parseRateLimitHeader(v string) (limit, remaining, reset int, err error) {
	_, err = fmt.Sscanf(v, HeaderRateLimitFmt, &limit, &remaining, &reset)
	return
}
