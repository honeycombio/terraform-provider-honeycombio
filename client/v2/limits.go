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

var rnd *rand.Rand

// init initializes the random number generator
func init() {
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// rateLimitBackoff calculates the backoff time for a rate limited request
// based on the possible response headers.
//
// If the rate limit header is not present, the function will return a random
// backoff time between min and max.
func rateLimitBackoff(min, max time.Duration, r *http.Response) time.Duration {
	// calculate some jitter for a little extra fuzziness to avoid thundering herds
	jitter := time.Duration(rnd.Float64() * float64(max-min))

	// try to get the next reset from the response headers
	if v := r.Header.Get(HeaderRateLimit); v != "" {
		// we currently only care about the reset time
		_, _, reset, err := parseRateLimitHeader(v)
		if err == nil {
			min = time.Duration(reset) * time.Second
		}
	} else if v := r.Header.Get(HeaderRetryAfter); v != "" {
		// if we can't get the ratelimit header, try the retry-after header
		retryTime, err := time.Parse(time.RFC3339, v)
		if err == nil {
			min = time.Until(retryTime)
		}
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
