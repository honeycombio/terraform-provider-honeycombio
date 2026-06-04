package limits

import (
	"context"
	"fmt"
	"net/http"
	"sync"
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
	tr := NewRateLimitingTransport(base)
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
	tr := NewRateLimitingTransport(srv)
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
