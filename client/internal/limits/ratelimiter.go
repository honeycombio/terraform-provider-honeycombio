package limits

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dunglas/httpsfv"
)

// errInvalidPolicy is returned when the RateLimit-Policy header is missing or
// cannot be parsed into a usable window.
var errInvalidPolicy = errors.New("invalid ratelimit policy header")

const (
	// HeaderRateLimitPolicy is the (draft07) IETF header describing the active
	// rate limit policy. The value is a Structured Field list of Items, each
	// whose value is the request limit and whose "w" parameter is the window
	// length in seconds, e.g. "120;w=60".
	HeaderRateLimitPolicy = "RateLimit-Policy"

	// defaultWindow is assumed when the server does not advertise a window via
	// the RateLimit-Policy header. Honeycomb's APIs use a one-minute window.
	defaultWindow = time.Minute

	// resetPadding is added to the time we wait for a window to reset. The
	// server reports the reset as a whole number of seconds (truncated), so the
	// true boundary can be up to a second later than reported; padding avoids
	// waking early and immediately drawing a 429.
	resetPadding = time.Second
)

// RateLimitingTransport is an http.RoundTripper that proactively keeps the
// client within Honeycomb's advertised rate limits.
//
// Honeycomb enforces a fixed-window limit server-side, scoped per-team,
// per-resource, and per-action: each window grants a fixed number of requests
// (e.g. 120 reads/minute), refilled in full at the next window boundary rather
// than trickled back continuously. Every response — including a 429 — reports
// the active limit, the requests remaining in the window, and the seconds until
// it resets, via the RateLimit and RateLimit-Policy headers.
//
// This transport mirrors that model. It keeps a per-(method, resource) bucket
// seeded from those headers and, while a bucket still has budget, passes
// requests straight through — so workloads that fit within the limit behave
// exactly as they would without it. Only once a bucket's budget is exhausted
// does it pause new requests until the window resets, then resume. This turns a
// large plan, apply, or refresh from a burst of HTTP 429s and retry-backoff
// into smooth, predictable throughput.
//
// It complements rather than replaces the reactive RetryHTTPBackoff: the gate
// prevents the 429s it can foresee, and the backoff remains the safety net for
// the first request to a bucket (before any limit is known) and for budget
// shared with other clients.
type RateLimitingTransport struct {
	base  http.RoundTripper
	clock func() time.Time
	sleep func(context.Context, time.Duration) error

	mu      sync.Mutex
	buckets map[string]*bucket
}

// bucket tracks the server's reported rate limit state for a single
// (method, resource) pair.
type bucket struct {
	mu sync.Mutex

	known     bool          // whether the limit headers have been observed yet
	limit     int           // requests permitted per window
	remaining int           // requests remaining in the current window
	window    time.Duration // length of the window
	resetAt   time.Time     // when the current window resets
}

// NewRateLimitingTransport wraps base with proactive client-side rate limiting.
// If base is nil, http.DefaultTransport is used.
func NewRateLimitingTransport(base http.RoundTripper) *RateLimitingTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &RateLimitingTransport{
		base:    base,
		clock:   time.Now,
		sleep:   sleepFor,
		buckets: make(map[string]*bucket),
	}
}

// RoundTrip waits until the bucket for this request has budget, performs the
// request, and updates the bucket from the response's rate limit headers.
func (t *RateLimitingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	b := t.bucketFor(bucketKey(req.Method, req.URL.Path))

	if err := t.acquire(req.Context(), b); err != nil {
		return nil, err
	}

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	b.observe(resp.Header, t.clock())
	return resp, nil
}

func (t *RateLimitingTransport) bucketFor(key string) *bucket {
	t.mu.Lock()
	defer t.mu.Unlock()

	if b, ok := t.buckets[key]; ok {
		return b
	}
	b := &bucket{}
	t.buckets[key] = b
	return b
}

// acquire blocks until the bucket can admit another request. Until the first
// response has been observed the limit is unknown and requests pass through
// unimpeded, leaving the reactive backoff to handle any miss.
func (t *RateLimitingTransport) acquire(ctx context.Context, b *bucket) error {
	for {
		b.mu.Lock()
		now := t.clock()

		if !b.known {
			b.mu.Unlock()
			return nil
		}

		// Roll the window over once we've reached its reset, optimistically
		// assuming a full refill; the next observed response corrects it.
		if !now.Before(b.resetAt) {
			b.remaining = b.limit
			b.resetAt = now.Add(b.window)
		}

		if b.remaining > 0 {
			b.remaining--
			b.mu.Unlock()
			return nil
		}

		wait := b.resetAt.Sub(now) + resetPadding
		b.mu.Unlock()

		if err := t.sleep(ctx, wait); err != nil {
			return err
		}
		// re-evaluate now that the window should have reset
	}
}

// observe updates the bucket from a response's rate limit headers. Responses
// without usable headers leave the bucket unchanged.
//
// The server's remaining count is authoritative and overwrites our local
// estimate. Requests still in flight when this response was produced are not
// yet reflected in it, so the estimate can run briefly optimistic under high
// concurrency; the reactive backoff absorbs the occasional resulting 429.
func (b *bucket) observe(h http.Header, now time.Time) {
	v := h.Get(HeaderRateLimit)
	if v == "" {
		return
	}
	limit, remaining, reset, err := parseRateLimitHeader(v)
	if err != nil {
		return
	}

	window := defaultWindow
	if w, err := parseRateLimitPolicyWindow(h.Get(HeaderRateLimitPolicy)); err == nil && w > 0 {
		window = w
	}

	b.mu.Lock()
	b.known = true
	b.limit = int(limit)
	b.remaining = int(remaining)
	b.window = window
	b.resetAt = now.Add(time.Duration(reset) * time.Second)
	b.mu.Unlock()
}

// bucketKey derives a rate limit bucket from the request method and path.
//
// Honeycomb scopes its limits per-resource, so the API version and resource
// collection — the first two path segments — form the bucket. All
// "/1/triggers/<dataset>" requests therefore share a bucket, mirroring the
// server's per-target accounting, while e.g. triggers and SLOs are paced
// independently.
func bucketKey(method, path string) string {
	segs := strings.SplitN(strings.TrimPrefix(path, "/"), "/", 3)

	var resource string
	switch {
	case len(segs) >= 2:
		resource = segs[0] + "/" + segs[1]
	case len(segs) == 1:
		resource = segs[0]
	}

	return method + " " + resource
}

// parseRateLimitPolicyWindow returns the window length from a RateLimit-Policy
// header value such as "120;w=60". When several policies are listed the
// shortest window is used, as that is the binding constraint for pacing.
func parseRateLimitPolicyWindow(h string) (time.Duration, error) {
	if h == "" {
		return 0, errInvalidPolicy
	}

	list, err := httpsfv.UnmarshalList([]string{h})
	if err != nil {
		return 0, errInvalidPolicy
	}

	var shortest time.Duration
	for _, member := range list {
		item, isItem := member.(httpsfv.Item)
		if !isItem {
			continue
		}
		w, ok := item.Params.Get("w")
		if !ok {
			continue
		}
		seconds, ok := w.(int64)
		if !ok || seconds <= 0 {
			continue
		}
		window := time.Duration(seconds) * time.Second
		if shortest == 0 || window < shortest {
			shortest = window
		}
	}

	if shortest == 0 {
		return 0, errInvalidPolicy
	}
	return shortest, nil
}

// sleepFor blocks for d or until ctx is cancelled, whichever comes first.
func sleepFor(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
