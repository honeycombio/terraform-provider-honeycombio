package limits

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dunglas/httpsfv"
)

const (
	// HeaderRateLimitPolicy is the (draft07) recommended header from the
	// IETF describing the rate limit policy in effect.
	//
	// The value of the header is expected to be a HTTP Structured Field Value (SFV)
	// list of policies, each an Item whose value is the request quota with a
	// "w" parameter carrying the window in seconds. e.g. "240;w=60"
	HeaderRateLimitPolicy = "Ratelimit-Policy"

	// defaultPaceInterval is the pacing interval used when the response
	// headers don't give us enough information to derive one.
	defaultPaceInterval = 500 * time.Millisecond

	// pacingThresholdDivisor controls when pacing kicks in: once the
	// remaining budget drops to 1/Nth of the limit the remaining requests
	// are spread across the rest of the window instead of sent as fast
	// as possible.
	pacingThresholdDivisor = 5
)

// ThrottledTransport is an http.RoundTripper which proactively throttles
// requests based on the rate limit headers returned by the Honeycomb API.
//
// Rate limits are applied per-endpoint, so requests are grouped by method
// and the first two path segments (e.g. "GET /1/derived_columns").
// While an endpoint's budget is healthy requests pass through untouched.
// When the reported remaining budget runs low the sender is paced to spread
// the remaining requests across the rest of the window, and when the budget
// is exhausted (or a 429 is received) all of the endpoint's requests are
// held until the window resets.
//
// Without send-side coordination each concurrent caller independently
// slams into the limit and retries, which starves other consumers of the
// same rate limit budget with a storm of doomed requests.
type ThrottledTransport struct {
	base http.RoundTripper

	mu        sync.Mutex
	endpoints map[string]*endpointState

	// now is time.Now, replaceable for testing
	now func() time.Time
}

type endpointState struct {
	// nextAt is the earliest time the next request may be sent.
	nextAt time.Time
	// interval is the minimum time between requests. Zero means unpaced.
	interval time.Duration
}

// NewThrottledTransport wraps base with a rate-limit-aware throttle.
// If base is nil, http.DefaultTransport is used. If base is already a
// ThrottledTransport it is returned as-is.
func NewThrottledTransport(base http.RoundTripper) http.RoundTripper {
	if t, ok := base.(*ThrottledTransport); ok {
		return t
	}
	if base == nil {
		base = http.DefaultTransport
	}
	return &ThrottledTransport{
		base:      base,
		endpoints: make(map[string]*endpointState),
		now:       time.Now,
	}
}

func (t *ThrottledTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	key := throttleKey(req)
	if err := t.wait(req.Context(), key); err != nil {
		return nil, err
	}
	resp, err := t.base.RoundTrip(req)
	if err == nil && resp != nil {
		t.observe(key, resp)
	}
	return resp, err
}

// wait blocks until the endpoint's throttle allows another request to be
// sent, or the request's context is cancelled.
func (t *ThrottledTransport) wait(ctx context.Context, key string) error {
	for {
		t.mu.Lock()
		st, ok := t.endpoints[key]
		if !ok {
			t.mu.Unlock()
			return nil
		}
		now := t.now()
		if !now.Before(st.nextAt) {
			if st.interval > 0 {
				st.nextAt = now.Add(st.interval)
			}
			t.mu.Unlock()
			return nil
		}
		d := st.nextAt.Sub(now)
		t.mu.Unlock()

		timer := time.NewTimer(d)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

// observe updates the endpoint's throttle state from a response's
// rate limit headers.
func (t *ThrottledTransport) observe(key string, resp *http.Response) {
	limit, remaining, reset, rlErr := parseRateLimitHeader(resp.Header.Get(HeaderRateLimit))
	rateLimited := resp.StatusCode == http.StatusTooManyRequests
	if rlErr != nil && !rateLimited {
		// no rate limit information and not limited: nothing to learn
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	st, ok := t.endpoints[key]
	if !ok {
		st = &endpointState{}
		t.endpoints[key] = st
	}
	now := t.now()

	// the pacing interval once the budget is exhausted: spread the next
	// window's budget evenly rather than bursting at reset
	pace := defaultPaceInterval
	if pLimit, pWindow, err := parseRateLimitPolicyHeader(resp.Header.Get(HeaderRateLimitPolicy)); err == nil {
		pace = pWindow / time.Duration(pLimit)
	} else if rlErr == nil && limit > 0 && reset > 0 {
		pace = time.Duration(reset) * time.Second / time.Duration(limit)
	}

	switch {
	case rateLimited:
		// hold this endpoint's requests until the budget resets, then pace.
		// rateLimitBackoff parses the reset from the response and fuzzes
		// it with a little jitter.
		st.deferUntil(now.Add(rateLimitBackoff(0, defaultPaceInterval, resp)))
		st.interval = pace
	case remaining <= 0:
		st.deferUntil(now.Add(time.Duration(reset) * time.Second))
		st.interval = pace
	case limit > 0 && remaining <= limit/pacingThresholdDivisor:
		// running low: spread what's left of the budget across the rest
		// of the window
		if reset > 0 {
			st.interval = time.Duration(reset) * time.Second / time.Duration(remaining)
		} else {
			st.interval = pace
		}
	default:
		st.interval = 0
	}
}

func (s *endpointState) deferUntil(t time.Time) {
	if t.After(s.nextAt) {
		s.nextAt = t
	}
}

// throttleKey groups requests by the rate limit budget they consume:
// the method and the first two path segments. e.g. "GET /1/derived_columns"
func throttleKey(req *http.Request) string {
	segments := strings.SplitN(strings.TrimPrefix(req.URL.EscapedPath(), "/"), "/", 3)
	return req.Method + " /" + strings.Join(segments[:min(len(segments), 2)], "/")
}

// parseRateLimitPolicyHeader parses the rate limit policy header into its
// quota and window.
//
// The header is expected to be in the format "<limit>;w=<window>". e.g. "240;w=60"
// If multiple policies are present only the first is considered.
func parseRateLimitPolicyHeader(h string) (limit int64, window time.Duration, err error) {
	list, err := httpsfv.UnmarshalList([]string{h})
	if err != nil {
		err = errors.New("invalid ratelimit-policy header")
		return
	}
	if len(list) == 0 {
		err = errors.New("empty ratelimit-policy header")
		return
	}
	item, ok := list[0].(httpsfv.Item)
	if !ok {
		err = errors.New("invalid ratelimit-policy header")
		return
	}
	limit, ok = item.Value.(int64)
	if !ok || limit <= 0 {
		err = errors.New("invalid limit in ratelimit-policy header")
		return
	}
	w, ok := item.Params.Get("w")
	if !ok {
		err = errors.New("no window in ratelimit-policy header")
		return
	}
	windowSeconds, ok := w.(int64)
	if !ok || windowSeconds <= 0 {
		err = fmt.Errorf("invalid window %q in ratelimit-policy header", w)
		return
	}
	window = time.Duration(windowSeconds) * time.Second
	return
}
