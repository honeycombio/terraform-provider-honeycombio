package limits

import (
	"bytes"
	"context"
	"errors"
	"io"
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

	// DefaultRateLimitRetries is the default number of times a rate-limited
	// (HTTP 429) request is replayed — waiting out the window each time —
	// before the 429 is surfaced. It matches the prior reactive retry budget so
	// the default behaviour is unchanged; raise it (via the provider's
	// rate_limit_retries) to ride out more contention without failing.
	DefaultRateLimitRetries = 10
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
// The gate prevents the 429s it can foresee. For those it cannot — the first
// request to a bucket before any limit is known, or budget shared with other
// concurrent clients — the transport treats a 429 not as a failure but as the
// server saying "come back when the window resets": it waits out the window and
// replays the request, up to maxRetries times. Because a rate-limited request
// is retried rather than failed, contention slows a run down instead of
// breaking it. maxRetries bounds this so a pathologically starved request still
// eventually surfaces the 429 (real, non-rate-limit errors and 5xx remain the
// job of the reactive RetryHTTPBackoff).
type RateLimitingTransport struct {
	base       http.RoundTripper
	clock      func() time.Time
	sleep      func(context.Context, time.Duration) error
	maxRetries int

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
// If base is nil, http.DefaultTransport is used. maxRetries bounds how many
// times a 429'd request is replayed (waiting out the window each time) before
// the 429 is surfaced; values <= 0 fall back to DefaultRateLimitRetries.
func NewRateLimitingTransport(base http.RoundTripper, maxRetries int) *RateLimitingTransport {
	if base == nil {
		base = http.DefaultTransport
	}
	if maxRetries <= 0 {
		maxRetries = DefaultRateLimitRetries
	}
	return &RateLimitingTransport{
		base:       base,
		clock:      time.Now,
		sleep:      sleepFor,
		maxRetries: maxRetries,
		buckets:    make(map[string]*bucket),
	}
}

// RoundTrip paces the request against its bucket, performs it, and updates the
// bucket from the response headers. A 429 is absorbed — the window is waited
// out and the request replayed — up to maxRetries times before being surfaced.
func (t *RateLimitingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	b := t.bucketFor(bucketKey(req.Method, req.URL.Path))

	// Buffer the body so the request can be replayed across rate-limit waits.
	// Request bodies here are small JSON payloads.
	var body []byte
	if req.Body != nil && req.Body != http.NoBody {
		var err error
		if body, err = io.ReadAll(req.Body); err != nil {
			_ = req.Body.Close()
			return nil, err
		}
		_ = req.Body.Close()
	}
	rewind := func() {
		if body != nil {
			req.Body = io.NopCloser(bytes.NewReader(body))
		}
	}
	rewind()

	for attempt := 0; ; attempt++ {
		if err := t.acquire(req.Context(), b); err != nil {
			return nil, err
		}

		resp, err := t.base.RoundTrip(req)
		if err != nil {
			return resp, err
		}
		b.observe(resp.Header, t.clock())

		if resp.StatusCode != http.StatusTooManyRequests || attempt >= t.maxRetries {
			return resp, nil
		}

		// Absorb the 429: make sure the bucket will block until the window
		// resets (observe has normally recorded it; ensureBlocked is a guard for
		// a 429 that arrives without usable headers), then replay the request.
		b.ensureBlocked(resp.Header, t.clock())
		drainClose(resp.Body)
		rewind()
	}
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

// ensureBlocked guarantees the bucket will make the next acquire wait, so a 429
// is never replayed in a hot loop. Normally observe has already recorded a
// future reset from the RateLimit header; this only acts when a 429 arrived
// without usable headers, falling back to the Retry-After value or one window.
func (b *bucket) ensureBlocked(h http.Header, now time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.known && b.remaining <= 0 && b.resetAt.After(now) {
		return
	}

	wait := defaultWindow
	if ra := h.Get(HeaderRetryAfter); ra != "" {
		if when, err := time.Parse(http.TimeFormat, ra); err == nil {
			if d := when.Sub(now); d > 0 {
				wait = d
			}
		}
	}

	b.known = true
	b.remaining = 0
	if b.window == 0 {
		b.window = defaultWindow
	}
	b.resetAt = now.Add(wait)
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

// drainClose drains and closes a response body so its connection can be reused
// before the request is replayed.
func drainClose(rc io.ReadCloser) {
	if rc != nil {
		_, _ = io.Copy(io.Discard, rc)
		_ = rc.Close()
	}
}
