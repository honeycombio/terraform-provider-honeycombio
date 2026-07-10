package limits

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThrottle_throttleKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		method   string
		url      string
		expected string
	}{
		{http.MethodGet, "https://api.honeycomb.io/1/derived_columns/my-dataset?alias=foo", "GET /1/derived_columns"},
		{http.MethodGet, "https://api.honeycomb.io/1/derived_columns/__all__", "GET /1/derived_columns"},
		{http.MethodPut, "https://api.honeycomb.io/1/triggers/my-dataset/abcd1234", "PUT /1/triggers"},
		{http.MethodGet, "https://api.honeycomb.io/2/teams/my-team/environments", "GET /2/teams"},
		{http.MethodGet, "https://api.honeycomb.io/1/auth", "GET /1/auth"},
		{http.MethodGet, "https://api.honeycomb.io/", "GET /"},
	}
	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			u, err := url.Parse(tc.url)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, throttleKey(&http.Request{Method: tc.method, URL: u}))
		})
	}
}

func TestThrottle_parseRateLimitPolicyHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		header    string
		limit     int64
		window    time.Duration
		expectErr bool
	}{
		{
			name:   "valid",
			header: "240;w=60",
			limit:  240,
			window: time.Minute,
		},
		{
			name:   "multiple policies takes first",
			header: "10;w=1, 240;w=60",
			limit:  10,
			window: time.Second,
		},
		{
			name:      "empty",
			expectErr: true,
		},
		{
			name:      "invalid",
			header:    "foobar",
			expectErr: true,
		},
		{
			name:      "missing window",
			header:    "240",
			expectErr: true,
		},
		{
			name:      "non-numeric window",
			header:    "240;w=soon",
			expectErr: true,
		},
		{
			name:      "negative limit",
			header:    "-1;w=60",
			expectErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			limit, window, err := parseRateLimitPolicyHeader(tc.header)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.limit, limit)
			assert.Equal(t, tc.window, window)
		})
	}
}

func TestThrottle_observe(t *testing.T) {
	t.Parallel()

	const key = "GET /1/derived_columns"
	now := time.Now()

	newTransport := func() *ThrottledTransport {
		tt, ok := NewThrottledTransport(nil).(*ThrottledTransport)
		require.True(t, ok)
		return tt
	}
	response := func(status int, headers map[string]string) *http.Response {
		w := httptest.NewRecorder()
		for k, v := range headers {
			w.Header().Add(k, v)
		}
		w.WriteHeader(status)
		return w.Result()
	}

	t.Run("healthy budget does not throttle", func(t *testing.T) {
		tt := newTransport()
		tt.observe(key, response(http.StatusOK, map[string]string{
			HeaderRateLimit:       "limit=240, remaining=200, reset=30",
			HeaderRateLimitPolicy: "240;w=60",
		}))

		st := tt.endpoints[key]
		require.NotNil(t, st)
		assert.Zero(t, st.interval)
		assert.True(t, st.nextAt.IsZero())
	})

	t.Run("no rate limit headers does not throttle", func(t *testing.T) {
		tt := newTransport()
		tt.observe(key, response(http.StatusOK, nil))
		assert.Empty(t, tt.endpoints)
	})

	t.Run("429 holds until reset and paces", func(t *testing.T) {
		tt := newTransport()
		tt.observe(key, response(http.StatusTooManyRequests, map[string]string{
			HeaderRateLimit:       "limit=240, remaining=0, reset=30",
			HeaderRateLimitPolicy: "240;w=60",
		}))

		st := tt.endpoints[key]
		require.NotNil(t, st)
		// held until the reported reset (+ up to defaultPaceInterval of jitter)
		assert.WithinRange(t, st.nextAt,
			now.Add(30*time.Second),
			now.Add(31*time.Second+defaultPaceInterval),
		)
		// paced to the policy's rate: 60s / 240 = 250ms
		assert.Equal(t, 250*time.Millisecond, st.interval)
	})

	t.Run("429 without headers still holds", func(t *testing.T) {
		tt := newTransport()
		tt.observe(key, response(http.StatusTooManyRequests, nil))

		st := tt.endpoints[key]
		require.NotNil(t, st)
		assert.Equal(t, defaultPaceInterval, st.interval)
	})

	t.Run("exhausted budget holds until reset", func(t *testing.T) {
		tt := newTransport()
		tt.observe(key, response(http.StatusOK, map[string]string{
			HeaderRateLimit: "limit=240, remaining=0, reset=10",
		}))

		st := tt.endpoints[key]
		require.NotNil(t, st)
		assert.WithinRange(t, st.nextAt,
			now.Add(9*time.Second),
			now.Add(11*time.Second),
		)
		// no policy header: pace derived from limit and reset: 10s / 240
		assert.Equal(t, 10*time.Second/240, st.interval)
	})

	t.Run("low budget paces the remaining window", func(t *testing.T) {
		tt := newTransport()
		tt.observe(key, response(http.StatusOK, map[string]string{
			HeaderRateLimit: "limit=240, remaining=40, reset=20",
		}))

		st := tt.endpoints[key]
		require.NotNil(t, st)
		assert.True(t, st.nextAt.IsZero(), "low budget should pace, not hold")
		// spread the remaining budget across the rest of the window: 20s / 40
		assert.Equal(t, 500*time.Millisecond, st.interval)
	})

	t.Run("recovered budget clears pacing", func(t *testing.T) {
		tt := newTransport()
		tt.observe(key, response(http.StatusOK, map[string]string{
			HeaderRateLimit: "limit=240, remaining=10, reset=5",
		}))
		require.NotZero(t, tt.endpoints[key].interval)

		tt.observe(key, response(http.StatusOK, map[string]string{
			HeaderRateLimit: "limit=240, remaining=239, reset=60",
		}))
		assert.Zero(t, tt.endpoints[key].interval)
	})

	t.Run("throttling is per-endpoint", func(t *testing.T) {
		tt := newTransport()
		tt.observe(key, response(http.StatusTooManyRequests, map[string]string{
			HeaderRateLimit: "limit=240, remaining=0, reset=30",
		}))

		assert.Nil(t, tt.endpoints["GET /1/triggers"])
	})
}

func TestThrottle_RoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("healthy responses pass through unthrottled", func(t *testing.T) {
		var requests int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			requests++
			w.Header().Set(HeaderRateLimit, "limit=240, remaining=200, reset=30")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := &http.Client{Transport: NewThrottledTransport(nil)}
		start := time.Now()
		for range 5 {
			resp, err := client.Get(server.URL + "/1/derived_columns/test")
			require.NoError(t, err)
			resp.Body.Close()
		}
		assert.Equal(t, 5, requests)
		assert.Less(t, time.Since(start), time.Second)
	})

	t.Run("exhausted budget delays the next request until reset", func(t *testing.T) {
		var timestamps []time.Time
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			timestamps = append(timestamps, time.Now())
			w.Header().Set(HeaderRateLimit, "limit=240, remaining=0, reset=1")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := &http.Client{Transport: NewThrottledTransport(nil)}
		for range 2 {
			resp, err := client.Get(server.URL + "/1/derived_columns/test")
			require.NoError(t, err)
			resp.Body.Close()
		}

		require.Len(t, timestamps, 2)
		assert.GreaterOrEqual(t,
			timestamps[1].Sub(timestamps[0]),
			900*time.Millisecond,
			"second request should have been held until the budget reset",
		)
	})

	t.Run("concurrent requests are held after a 429", func(t *testing.T) {
		var (
			mu         sync.Mutex
			timestamps []time.Time
		)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			mu.Lock()
			timestamps = append(timestamps, time.Now())
			first := len(timestamps) == 1
			mu.Unlock()

			w.Header().Set(HeaderRateLimit, "limit=240, remaining=0, reset=1")
			if first {
				w.WriteHeader(http.StatusTooManyRequests)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer server.Close()

		transport := NewThrottledTransport(nil)
		start := time.Now()

		// the first request trips the 429 and the rest queue behind
		// the reset instead of racing to their own 429s
		client := &http.Client{Transport: transport}
		resp, err := client.Get(server.URL + "/1/derived_columns/test")
		require.NoError(t, err)
		resp.Body.Close()

		var wg sync.WaitGroup
		for range 3 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				resp, err := client.Get(server.URL + "/1/derived_columns/test")
				assert.NoError(t, err)
				if resp != nil {
					resp.Body.Close()
				}
			}()
		}
		wg.Wait()

		require.Len(t, timestamps, 4)
		for _, ts := range timestamps[1:] {
			assert.GreaterOrEqual(t, ts.Sub(start), 900*time.Millisecond,
				"queued requests should not be sent before the budget reset")
		}
	})

	t.Run("wait respects context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set(HeaderRateLimit, "limit=240, remaining=0, reset=30")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := &http.Client{Transport: NewThrottledTransport(nil)}
		resp, err := client.Get(server.URL + "/1/derived_columns/test")
		require.NoError(t, err)
		resp.Body.Close()

		// the endpoint is now held for 30s: a context-limited request
		// should give up promptly
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/1/derived_columns/test", nil)
		require.NoError(t, err)

		start := time.Now()
		_, err = client.Do(req) //nolint:bodyclose // the request never happens
		require.Error(t, err)
		assert.Less(t, time.Since(start), time.Second)
	})
}
