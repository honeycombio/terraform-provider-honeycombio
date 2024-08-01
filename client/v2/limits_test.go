package v2

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
	// init min and max to zero to remove jitter
	min, max := time.Duration(0), time.Duration(0)

	tests := []struct {
		name          string
		headerName    string
		headerValue   string
		expectedValue time.Duration
	}{
		{
			name:          "no header",
			expectedValue: min,
		},
		{
			name:          "invalid ratelimit header",
			headerName:    HeaderRateLimit,
			headerValue:   "foobar",
			expectedValue: min,
		},
		{
			name:          "invalid retry-after header",
			headerName:    HeaderRetryAfter,
			headerValue:   "three hours from now",
			expectedValue: min,
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
			headerValue:   now.Add(2 * time.Minute).UTC().Format(time.RFC3339),
			expectedValue: 2 * time.Minute,
		},
		{
			name:          "negative retry-after header",
			headerName:    HeaderRateLimit,
			headerValue:   "limit=100, remaining=-1, reset=-10",
			expectedValue: min,
		},
		{
			name:          "retry-after in the past",
			headerName:    HeaderRetryAfter,
			headerValue:   now.Add(-2 * time.Minute).UTC().Format(time.RFC3339),
			expectedValue: min,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			w.Header().Add(tc.headerName, tc.headerValue)
			w.WriteHeader(http.StatusTooManyRequests)

			r := rateLimitBackoff(min, max, w.Result())
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
		w.Header().Add(HeaderRetryAfter, now.Add(2*time.Minute).UTC().Format(time.RFC3339))
		w.WriteHeader(http.StatusTooManyRequests)

		r := rateLimitBackoff(min, max, w.Result())
		assert.WithinDuration(t,
			now.Add(60*time.Second),
			now.Add(r),
			time.Second,
		)
	})
}

func TestClient_parseRateLimitHeader(t *testing.T) {
	t.Parallel()

	type expect struct {
		limit     int
		remaining int
		reset     int
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
			name:      "valid but missing reset",
			header:    "limit=100, remaining=50",
			expectErr: true,
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
