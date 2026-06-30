package limits

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBucketKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
		want   string
	}{
		{"v1 collection", http.MethodGet, "/1/triggers/__all__", "GET 1/triggers"},
		{"v1 item", http.MethodPut, "/1/slos/prod/abc123", "PUT 1/slos"},
		{"v1 create", http.MethodPost, "/1/burn_alerts/prod", "POST 1/burn_alerts"},
		{"v2 nested", http.MethodGet, "/2/teams/foo/api-keys", "GET 2/teams"},
		{"single segment", http.MethodGet, "/1", "GET 1"},
		{"no leading slash", http.MethodGet, "1/triggers/ds", "GET 1/triggers"},
		{"empty path", http.MethodGet, "", "GET "},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, bucketKey(tc.method, tc.path))
		})
	}
}

func TestParseRateLimitPolicyWindow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		header  string
		want    time.Duration
		wantErr bool
	}{
		{"single policy", "120;w=60", time.Minute, false},
		{"shortest of several", "100;w=100, 1000;w=3600", 100 * time.Second, false},
		{"empty", "", 0, true},
		{"no window param", "120", 0, true},
		{"garbage", "not a header", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseRateLimitPolicyWindow(tc.header)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestBucketObserve(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("seeds from headers", func(t *testing.T) {
		var b bucket
		h := make(http.Header)
		h.Set(HeaderRateLimit, "limit=120, remaining=42, reset=30")
		h.Set(HeaderRateLimitPolicy, "120;w=60")

		b.observe(h, now)

		assert.True(t, b.known)
		assert.Equal(t, 120, b.limit)
		assert.Equal(t, 42, b.remaining)
		assert.Equal(t, time.Minute, b.window)
		assert.Equal(t, now.Add(30*time.Second), b.resetAt)
	})

	t.Run("defaults window when policy absent", func(t *testing.T) {
		var b bucket
		h := make(http.Header)
		h.Set(HeaderRateLimit, "limit=100, remaining=10, reset=15")

		b.observe(h, now)

		assert.Equal(t, defaultWindow, b.window)
	})

	t.Run("ignores responses without rate limit headers", func(t *testing.T) {
		var b bucket
		b.observe(make(http.Header), now)
		assert.False(t, b.known)
	})
}

// fakeFixedWindow is an http.RoundTripper that emulates Honeycomb's
// server-side fixed-window limiter: a fixed budget per window, scoped
// per-resource exactly as the real server scopes per-target, refilled in full
// at the next window boundary, advertising the standard RateLimit headers and
// returning 429 when the budget is exhausted.
type fakeFixedWindow struct {
	limit  int
	window time.Duration
	clock  func() time.Time

	mu       sync.Mutex
	windows  map[string]*serverWindow
	rejected int
}

type serverWindow struct {
	end       time.Time
	remaining int
}

func (s *fakeFixedWindow) RoundTrip(req *http.Request) (*http.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.windows == nil {
		s.windows = make(map[string]*serverWindow)
	}

	now := s.clock()
	key := bucketKey(req.Method, req.URL.Path)
	w := s.windows[key]
	if w == nil || now.After(w.end) {
		w = &serverWindow{end: now.Add(s.window), remaining: s.limit}
		s.windows[key] = w
	}

	status := http.StatusOK
	if w.remaining > 0 {
		w.remaining--
	} else {
		s.rejected++
		status = http.StatusTooManyRequests
	}

	resetSecs := int(w.end.Sub(now).Seconds())
	h := make(http.Header)
	h.Set(HeaderRateLimit, fmt.Sprintf("limit=%d, remaining=%d, reset=%d", s.limit, w.remaining, resetSecs))
	h.Set(HeaderRateLimitPolicy, fmt.Sprintf("%d;w=%d", s.limit, int(s.window.Seconds())))

	return &http.Response{
		StatusCode: status,
		Header:     h,
		Body:       http.NoBody,
		Request:    req,
	}, nil
}

// newTestTransport wraps base in a RateLimitingTransport driven by a virtual
// clock that only advances when the gate sleeps, making pacing deterministic.
func newTestTransport(base http.RoundTripper, now *time.Time) *RateLimitingTransport {
	tr := NewRateLimitingTransport(base, DefaultRateLimitRetries)
	tr.clock = func() time.Time { return *now }
	tr.sleep = func(_ context.Context, d time.Duration) error {
		*now = now.Add(d)
		return nil
	}
	return tr
}

func TestRateLimitingTransport_GatesOnBudget(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := start

	srv := &fakeFixedWindow{limit: 3, window: time.Minute, clock: func() time.Time { return now }}
	tr := newTestTransport(srv, &now)

	const n = 10
	for i := 0; i < n; i++ {
		req, err := http.NewRequest(http.MethodGet, "https://api.honeycomb.io/1/triggers/__all__", nil)
		require.NoError(t, err)

		resp, err := tr.RoundTrip(req)
		require.NoError(t, err)
		require.Equalf(t, http.StatusOK, resp.StatusCode, "request %d was rate limited", i)
	}

	// With a budget of 3/window, 10 requests span 4 windows, so the gate must
	// have paused across 3 window boundaries — and never drawn a 429.
	assert.Zero(t, srv.rejected, "gate should have prevented every 429")
	assert.GreaterOrEqual(t, now.Sub(start), 3*srv.window)
}

func TestRateLimitingTransport_PassesThroughWithinBudget(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := start

	// Budget comfortably exceeds the number of requests: nothing should pause.
	srv := &fakeFixedWindow{limit: 120, window: time.Minute, clock: func() time.Time { return now }}
	tr := newTestTransport(srv, &now)

	for i := 0; i < 20; i++ {
		req, err := http.NewRequest(http.MethodGet, "https://api.honeycomb.io/1/triggers/__all__", nil)
		require.NoError(t, err)

		resp, err := tr.RoundTrip(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}

	assert.Equal(t, start, now, "no pacing should occur while within budget")
	assert.Zero(t, srv.rejected)
}

func TestRateLimitingTransport_SeparateBucketsPerResource(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := start

	srv := &fakeFixedWindow{limit: 2, window: time.Minute, clock: func() time.Time { return now }}
	tr := newTestTransport(srv, &now)

	// Alternating between two resources should let four requests through
	// without pausing, since each resource has its own budget of two.
	paths := []string{
		"https://api.honeycomb.io/1/triggers/__all__",
		"https://api.honeycomb.io/1/slos/__all__",
		"https://api.honeycomb.io/1/triggers/__all__",
		"https://api.honeycomb.io/1/slos/__all__",
	}
	for _, p := range paths {
		req, err := http.NewRequest(http.MethodGet, p, nil)
		require.NoError(t, err)

		resp, err := tr.RoundTrip(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	}

	assert.Equal(t, start, now, "independent buckets should not pace each other")
}

func TestRateLimitingTransport_ContextCancelledWhileWaiting(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := start

	srv := &fakeFixedWindow{limit: 1, window: time.Minute, clock: func() time.Time { return now }}
	tr := NewRateLimitingTransport(srv, DefaultRateLimitRetries)
	tr.clock = func() time.Time { return now }

	ctx, cancel := context.WithCancel(context.Background())
	tr.sleep = func(_ context.Context, _ time.Duration) error {
		// Simulate the caller cancelling the context while the gate is waiting.
		cancel()
		return ctx.Err()
	}

	// First request consumes the only token in the window.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.honeycomb.io/1/triggers/__all__", nil)
	require.NoError(t, err)
	_, err = tr.RoundTrip(req)
	require.NoError(t, err)

	// Second request must wait for the window to reset; cancellation aborts it.
	req2, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.honeycomb.io/1/triggers/__all__", nil)
	require.NoError(t, err)
	_, err = tr.RoundTrip(req2)
	assert.ErrorIs(t, err, context.Canceled)
}

// roundTripFunc adapts a function to an http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestRateLimitingTransport_Absorbs429(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := start

	// First call is rate limited; the retry succeeds.
	calls := 0
	base := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		h := make(http.Header)
		h.Set(HeaderRateLimitPolicy, "120;w=60")
		if calls == 1 {
			h.Set(HeaderRateLimit, "limit=120, remaining=0, reset=30")
			return &http.Response{StatusCode: http.StatusTooManyRequests, Header: h, Body: http.NoBody, Request: r}, nil
		}
		h.Set(HeaderRateLimit, "limit=120, remaining=119, reset=60")
		return &http.Response{StatusCode: http.StatusOK, Header: h, Body: http.NoBody, Request: r}, nil
	})

	tr := newTestTransport(base, &now)

	req, err := http.NewRequest(http.MethodGet, "https://api.honeycomb.io/1/triggers/x", nil)
	require.NoError(t, err)
	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode, "the 429 should be absorbed and the request retried")
	assert.Equal(t, 2, calls)
	assert.GreaterOrEqual(t, now.Sub(start), 30*time.Second, "should have waited out the window before retrying")
}

func TestRateLimitingTransport_AbsorbRespectsMax(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := start

	// Always rate limited: the transport should give up after maxRetries and
	// surface the 429 rather than looping forever.
	calls := 0
	base := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		h := make(http.Header)
		h.Set(HeaderRateLimitPolicy, "120;w=60")
		h.Set(HeaderRateLimit, "limit=120, remaining=0, reset=1")
		return &http.Response{StatusCode: http.StatusTooManyRequests, Header: h, Body: http.NoBody, Request: r}, nil
	})

	tr := newTestTransport(base, &now)
	tr.maxRetries = 2

	req, err := http.NewRequest(http.MethodGet, "https://api.honeycomb.io/1/triggers/x", nil)
	require.NoError(t, err)
	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode, "after maxRetries the 429 should surface")
	assert.Equal(t, 3, calls, "1 initial attempt + 2 retries")
}

// reorderingTransport delays each response by a small random amount *after* the
// inner transport has produced it. The fake server stamps each response's
// "remaining" count at processing time, so delaying delivery shuffles the order
// in which the gate observes those counts relative to the order they were
// produced. That reproduces the real-world condition — stale, out-of-order
// rate limit observations — that stresses the gate's bookkeeping under
// concurrency, rather than relying on incidental goroutine scheduling.
type reorderingTransport struct {
	inner http.RoundTripper
	max   time.Duration
}

func (r *reorderingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := r.inner.RoundTrip(req)
	if err == nil && r.max > 0 {
		//nolint:gosec // non-cryptographic jitter is fine for a test
		time.Sleep(time.Duration(rand.Int63n(int64(r.max))))
	}
	return resp, err
}

// scaledClock reports a simulated time that advances at a fixed multiple of
// real time. It lets the test use a realistic minute-long window — so the
// integer-second RateLimit/reset headers stay meaningful, exactly as in
// production — while the whole test still completes in milliseconds. It is safe
// for concurrent use: Now derives purely from the (immutable) real start time,
// with no shared mutable state.
type scaledClock struct {
	realStart time.Time
	simStart  time.Time
	scale     int64 // simulated nanoseconds per real nanosecond
}

func newScaledClock(scale int64) *scaledClock {
	return &scaledClock{
		realStart: time.Now(),
		simStart:  time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		scale:     scale,
	}
}

func (c *scaledClock) Now() time.Time {
	return c.simStart.Add(time.Since(c.realStart) * time.Duration(c.scale))
}

// sleep blocks for the real-time equivalent of a simulated duration.
func (c *scaledClock) sleep(ctx context.Context, sim time.Duration) error {
	return sleepFor(ctx, sim/time.Duration(c.scale))
}

// TestRateLimitingTransport_ConcurrentLoad drives the gate the way Terraform
// does: many goroutines hammering a single bucket at once (mirroring
// -parallelism) with out-of-order rate limit observations.
//
// Under concurrency the gate's view is approximate, so the server will reject
// some requests — that is expected and acceptable. What this test guarantees is
// that those 429s are absorbed and replayed, so no job fails. (That the gate
// *prevents* 429s in the first place is shown by
// TestRateLimitingTransport_GateAvoidsBurstRejections.)
func TestRateLimitingTransport_ConcurrentLoad(t *testing.T) {
	t.Parallel()

	const (
		limit   = 30          // per-window server budget
		window  = time.Minute // realistic window; integer-second headers stay meaningful
		workers = 10          // < limit, mirrors Terraform parallelism below the budget
		jobs    = 150         // > limit, forces pacing across several windows
	)

	clk := newScaledClock(1000) // 1ms real == 1s simulated: a 60s window elapses in 60ms
	srv := &fakeFixedWindow{limit: limit, window: window, clock: clk.Now}
	base := &reorderingTransport{inner: srv, max: 200 * time.Microsecond}
	tr := NewRateLimitingTransport(base, DefaultRateLimitRetries)
	tr.clock = clk.Now
	tr.sleep = clk.sleep

	work := make(chan int, jobs)
	for i := range jobs {
		work <- i
	}
	close(work)

	var wg sync.WaitGroup
	var got200, got429 int64
	for range workers {
		wg.Go(func() {
			for range work {
				req, err := http.NewRequest(http.MethodGet, "https://api.honeycomb.io/1/triggers/__all__", nil)
				if err != nil {
					t.Errorf("building request: %v", err)
					return
				}
				resp, err := tr.RoundTrip(req)
				if err != nil {
					t.Errorf("RoundTrip: %v", err)
					return
				}
				switch resp.StatusCode {
				case http.StatusOK:
					atomic.AddInt64(&got200, 1)
				case http.StatusTooManyRequests:
					atomic.AddInt64(&got429, 1)
				}
			}
		})
	}
	wg.Wait()

	require.Equal(t, int64(jobs), got200+got429, "every job should produce a response")
	assert.Zerof(t, got429,
		"client saw %d unrecovered HTTP 429(s): a job failed under concurrency instead of being retried", got429)

	// Server-side rejections are expected here and are not a failure: they were
	// absorbed and replayed. Logged for visibility into how much the gate
	// over-issued under this load.
	srv.mu.Lock()
	rejected := srv.rejected
	srv.mu.Unlock()
	t.Logf("server rejected %d/%d request(s) under concurrency; all were absorbed", rejected, jobs)
}

// TestRateLimitingTransport_GateAvoidsBurstRejections shows the gate doing its
// job. Given a burst that oversubscribes a window's budget, the gate paces the
// requests across windows so the server rejects none — whereas the same burst
// sent without the gate piles into a single window and most are rejected. This
// is the proactive value the gate adds on top of the (reactive) 429 absorb.
func TestRateLimitingTransport_GateAvoidsBurstRejections(t *testing.T) {
	t.Parallel()

	const (
		limit  = 20
		window = time.Minute
		jobs   = 100
	)

	sendAll := func(rt http.RoundTripper) {
		for range jobs {
			req, err := http.NewRequest(http.MethodGet, "https://api.honeycomb.io/1/triggers/__all__", nil)
			require.NoError(t, err)
			resp, err := rt.RoundTrip(req)
			require.NoError(t, err)
			_ = resp.Body.Close()
		}
	}

	// Without the gate: a fast burst with no pacing. Time is frozen, modelling
	// every request landing within a single window — so only one window's budget
	// succeeds and the rest are rejected.
	frozen := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	naive := &fakeFixedWindow{limit: limit, window: window, clock: func() time.Time { return frozen }}
	sendAll(naive)

	// With the gate: the same load, paced. A virtual clock advances only when the
	// gate sleeps, so the burst is spread across windows and nothing is rejected.
	now := frozen
	gated := &fakeFixedWindow{limit: limit, window: window, clock: func() time.Time { return now }}
	sendAll(newTestTransport(gated, &now))

	assert.Equal(t, jobs-limit, naive.rejected,
		"without the gate, every request beyond the first window's budget is rejected")
	assert.Zero(t, gated.rejected,
		"with the gate, the burst is paced across windows and the server rejects nothing")
}
