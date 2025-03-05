package limits

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/dunglas/httpsfv"
	"github.com/hashicorp/go-retryablehttp"
)

const (
	// HeaderRateLimit is the (draft07) recommended header from the IETF
	// on rate limiting.
	//
	// The value of the header is expected to be a HTTP Structured Field Value (SFV)
	// dictionary with the keys "limit", "remaining", and "reset".
	//
	// Where "limit" is the maximum number of requests allowed in the window,
	// "remaining" is the number of requests remaining in the window,
	// and "reset" is the number of seconds until the limit resets.
	HeaderRateLimit = "Ratelimit"

	// HeaderRetryAfter is the RFC7231 header used to indicate when a client should
	// retry requests in UTC time.
	HeaderRetryAfter = "Retry-After"
)

// RetryHTTPBackoff is a retryablehttp.Backoff function that will
// use a linear backoff for all status codes except 429, which will
// attempt to use the rate limit headers to determine the backoff time
func RetryHTTPBackoff(
	mini, maxi time.Duration,
	attemptNum int,
	r *http.Response,
) time.Duration {
	if r != nil && r.StatusCode == http.StatusTooManyRequests {
		return rateLimitBackoff(mini, maxi, r)
	}

	// if we've not been rate limited, use a linear backoff
	// but increase the minimum and maximum backoff times
	// and hand it off to retryablehttp.LinearJitterBackoff
	mini = 500 * time.Millisecond
	maxi = 950 * time.Millisecond
	return retryablehttp.LinearJitterBackoff(mini, maxi, attemptNum, r)
}

// rateLimitBackoff calculates the backoff time for a rate limited request
// based on the possible response headers.
// The function will first try to get the reset time from the rate limit header.
//
// If the rate limit header is not present, or the reset time is in the past,
// the function will return a random backoff time between mini and maxi.
func rateLimitBackoff(mini, maxi time.Duration, r *http.Response) time.Duration {
	// calculate some jitter for a little extra fuzziness to avoid thundering herds
	jitter := time.Duration(rand.Float64() * float64(maxi-mini))

	var reset time.Duration
	if v := r.Header.Get(HeaderRateLimit); v != "" {
		// we currently only care about the reset time
		_, _, resetSeconds, err := parseRateLimitHeader(v)
		if err == nil {
			reset = time.Duration(resetSeconds) * time.Second
		}
	}
	// if we didn't get a reset value from the ratelimit header
	// try the retry-after header
	if reset == 0 {
		if v := r.Header.Get(HeaderRetryAfter); v != "" {
			retryTime, err := time.Parse(time.RFC3339, v)
			if err == nil {
				reset = time.Until(retryTime)
			}
		}
	}

	// only update min if the time to wait is longer
	if reset > mini {
		mini = reset
	}
	return mini + jitter
}

// retryHTTPCheck is a retryablehttp.CheckRetry function that will
// retry on a 429 or any 5xx status code
func RetryHTTPCheck(
	ctx context.Context,
	r *http.Response,
	err error,
) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if err != nil {
		return true, err
	}
	if r != nil {
		if r.StatusCode == http.StatusTooManyRequests || r.StatusCode >= 500 {
			return true, nil
		}
	}
	return false, nil
}

// parseRateLimitHeader parses the rate limit header into its constituent parts.
//
// The header is expected to be in the format "limit=X, remaining=Y, reset=Z".
// Where:
//   - X is the maximum number of requests allowed in the window
//   - Y is the number of requests remaining in the window
//   - Z is the number of seconds until the limit resets
func parseRateLimitHeader(h string) (limit, remaining, reset int64, err error) {
	vals, err := httpsfv.UnmarshalDictionary([]string{h})
	if err != nil {
		err = errors.New("invalid ratelimit header")
		return
	}

	limit, err = valueFromSFVDictionary[int64](vals, "limit")
	if err != nil {
		err = fmt.Errorf("could not get \"limit\" from header: %w", err)
		return
	}
	remaining, err = valueFromSFVDictionary[int64](vals, "remaining")
	if err != nil {
		err = fmt.Errorf("could not get \"remaining\" from header: %w", err)
		return
	}
	reset, err = valueFromSFVDictionary[int64](vals, "reset")
	if err != nil {
		err = fmt.Errorf("could not get \"reset\" from header: %w", err)
		return
	}
	return
}

func valueFromSFVDictionary[T any](d *httpsfv.Dictionary, key string) (T, error) {
	var zero T
	k, ok := d.Get(key)
	if !ok {
		return zero, errors.New("key not found")
	}
	v, ok := k.(httpsfv.Item).Value.(T)
	if !ok {
		return zero, fmt.Errorf("value is not a %T", zero)
	}
	return v, nil
}
