package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client/internal/limits"
)

func TestClient_ProactiveThrottlingOptIn(t *testing.T) {
	t.Parallel()

	t.Run("disabled by default", func(t *testing.T) {
		c, err := NewClientWithConfig(&Config{
			APIKeyID:     "abcd1234",
			APIKeySecret: "abcd1234",
		})
		require.NoError(t, err)

		_, throttled := c.http.HTTPClient.Transport.(*limits.ThrottledTransport)
		assert.False(t, throttled, "transport should not be throttled unless enabled")
	})

	t.Run("enabled when configured", func(t *testing.T) {
		c, err := NewClientWithConfig(&Config{
			APIKeyID:            "abcd1234",
			APIKeySecret:        "abcd1234",
			ProactiveThrottling: true,
		})
		require.NoError(t, err)

		_, throttled := c.http.HTTPClient.Transport.(*limits.ThrottledTransport)
		assert.True(t, throttled, "transport should be throttled when enabled")
	})
}
