package limits

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_rateLimitBackoff(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	// init mini and maxi to zero to remove jitter
	mini, maxi := time.Duration(0), time.Duration(0)

	tests := []struct {
		name          string
		headerName    string
		headerValue   string
		expectedValue time.Duration
	}{
		{
			name:          "no header",
			expectedValue: mini,
		},
		{
			name:          "invalid ratelimit header",
			headerName:    HeaderRateLimit,
			headerValue:   "foobar",
			expectedValue: mini,
		},
		{
			name:          "invalid retry-after header",
			headerName:    HeaderRetryAfter,
			headerValue:   "three hours from now",
			expectedValue: mini,
		},
		{
			name:          "valid ratelimit header",
			headerName:    HeaderRateLimit,
			headerValue:   "limit=100, remaining=50, reset=60",
			expectedValue: 60 * time.Second,
		},
		{
			name:          "valid retry-after header",
			headerName:    HeaderRetryAfter,
			headerValue:   now.Add(2 * time.Minute).UTC().Format(http.TimeFormat),
			expectedValue: 2 * time.Minute,
		},
		{
			name:          "negative reset value in ratelimit header",
			headerName:    HeaderRateLimit,
			headerValue:   "limit=100, remaining=-1, reset=-10",
			expectedValue: mini,
		},
		{
			name:          "retry-after in the past",
			headerName:    HeaderRetryAfter,
			headerValue:   now.Add(-2 * time.Minute).UTC().Format(time.RFC3339),
			expectedValue: mini,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			w.Header().Add(tc.headerName, tc.headerValue)
			w.WriteHeader(http.StatusTooManyRequests)

			r := rateLimitBackoff(mini, maxi, w.Result())
			assert.WithinDuration(t,
				now.Add(tc.expectedValue),
				now.Add(r),
				time.Second,
			)
		})
	}

	t.Run("ratelimit header takes precedence", func(t *testing.T) {
		w := httptest.NewRecorder()
		w.Header().Add(HeaderRateLimit, "limit=100, remaining=50, reset=60")
		w.Header().Add(HeaderRetryAfter, now.Add(2*time.Minute).UTC().Format(http.TimeFormat))
		w.WriteHeader(http.StatusTooManyRequests)

		r := rateLimitBackoff(mini, maxi, w.Result())
		assert.WithinDuration(t,
			now.Add(60*time.Second),
			now.Add(r),
			time.Second,
		)
	})

	t.Run("reset value is fuzzed with jitter", func(t *testing.T) {
		w := httptest.NewRecorder()
		w.Header().Add(HeaderRateLimit, "limit=100, remaining=50, reset=60")
		w.WriteHeader(http.StatusTooManyRequests)

		mini = 100 * time.Millisecond
		maxi = 500 * time.Millisecond
		assert.Greater(t,
			rateLimitBackoff(mini, maxi, w.Result()),
			60*time.Second,
			"expected backoff to be 60sec+jitter",
		)
	})

	t.Run("without supported rate limit header jitter is between min and max", func(t *testing.T) {
		w := httptest.NewRecorder()
		w.WriteHeader(http.StatusTooManyRequests)

		mini = 200 * time.Millisecond
		maxi = 900 * time.Millisecond

		now := time.Now().UTC()
		assert.WithinRange(t,
			now.Add(rateLimitBackoff(mini, maxi, w.Result())),
			now.Add(mini),
			now.Add(maxi),
			"expected backoff to be between min and max",
		)
	})
}

func TestClient_parseRateLimitHeader(t *testing.T) {
	t.Parallel()

	type expect struct {
		limit     int64
		remaining int64
		reset     int64
	}
	tests := []struct {
		name      string
		header    string
		expect    expect
		expectErr bool
	}{
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
			name:   "valid",
			header: "limit=100, remaining=50, reset=60",
			expect: expect{
				limit:     100,
				remaining: 50,
				reset:     60,
			},
		},
		{
			name:   "valid, no spacing",
			header: "limit=250,remaining=199,reset=120",
			expect: expect{
				limit:     250,
				remaining: 199,
				reset:     120,
			},
		},
		{
			name:   "mixed up member order",
			header: "remaining=50, limit=100, reset=60",
			expect: expect{
				limit:     100,
				remaining: 50,
				reset:     60,
			},
		},
		{
			name:   "additional key, otherwise valid",
			header: "limit=100, remaining=50, reset=120, foo=bar",
			expect: expect{
				limit:     100,
				remaining: 50,
				reset:     120,
			},
		},
		{
			name:      "missing member",
			header:    "limit=100, remaining=50",
			expectErr: true,
		},
		{
			name:      "wrong type value of member",
			header:    "limit=100, remaining=50, reset=now",
			expectErr: true,
		},
		{
			name:   "additional key, otherwise valid",
			header: "limit=100, remaining=50, reset=120, foo=bar",
			expect: expect{
				limit:     100,
				remaining: 50,
				reset:     120,
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			limit, remaining, reset, err := parseRateLimitHeader(tc.header)
			if tc.expectErr {
				require.Error(t, err, "expected an error")
				return
			}
			require.NoError(t, err, "expected no error")
			assert.Equal(t, tc.expect.limit, limit, "limit doesn't match")
			assert.Equal(t, tc.expect.remaining, remaining, "remaining doesn't match")
			assert.Equal(t, tc.expect.reset, reset, "reset doesn't match")
		})
	}
}
