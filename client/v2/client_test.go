package v2

import (
	"context"
	"net/http"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hnyclient "github.com/honeycombio/terraform-provider-honeycombio/client"
)

const (
	testEnvFilePath = "../../.env"
	testUserAgent   = "go-honeycombio/test"
)

func TestClient_Config(t *testing.T) {
	// load environment values from a .env, if available
	_ = godotenv.Load(testEnvFilePath)

	t.Run("constructs a default client", func(t *testing.T) {
		c, err := NewClient()
		require.NoError(t, err)
		assert.Equal(t, defaultUserAgent, c.UserAgent)
		assert.NotEmpty(t, c.BaseURL.String())
	})

	t.Run("constructs a client with config overrides", func(t *testing.T) {
		c, err := NewClientWithConfig(&Config{
			APIKeyID:           "123",
			APIKeySecret:       "456",
			BaseURL:            "https://api.example.com",
			UserAgent:          testUserAgent,
			skipInitialization: true,
		})
		require.NoError(t, err)
		assert.Equal(t, "Bearer 123:456", c.Headers.Get("Authorization"))
		assert.Equal(t, "https://api.example.com", c.BaseURL.String())
		assert.Equal(t, testUserAgent, c.UserAgent)
	})

	t.Run("fails to construct a client with an invalid API URL", func(t *testing.T) {
		_, err := NewClientWithConfig(&Config{
			BaseURL:            "cache_object:foo/bar",
			skipInitialization: true,
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid BaseURL")
	})

	t.Run("fails to construct a client without key id and secret pair", func(t *testing.T) {
		t.Run("with both missing", func(t *testing.T) {
			// load environment values from a .env, if available
			_ = godotenv.Load(testEnvFilePath)
			t.Setenv(DefaultAPIKeyIDEnv, "")
			t.Setenv(DefaultAPIKeySecretEnv, "")

			_, err := NewClient()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "missing API Key ID and Secret pair")
		})

		t.Run("with key ID missing", func(t *testing.T) {
			// load environment values from a .env, if available
			_ = godotenv.Load(testEnvFilePath)
			t.Setenv(DefaultAPIKeyIDEnv, "")

			_, err := NewClient()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "missing API Key ID and Secret pair")
		})

		t.Run("with key secret missing", func(t *testing.T) {
			// load environment values from a .env, if available
			_ = godotenv.Load(testEnvFilePath)
			t.Setenv(DefaultAPIKeyIDEnv, "")

			_, err := NewClient()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "missing API Key ID and Secret pair")
		})
	})
}

func TestClient_AuthInfo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("happy path", func(t *testing.T) {
		c := newTestClient(t)
		metadata, err := c.AuthInfo(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, metadata.ID)
		assert.NotEmpty(t, metadata.Name)
		assert.Equal(t, "management", metadata.KeyType)
		assert.False(t, metadata.Disabled)
		assert.NotEmpty(t, metadata.Scopes)
		if assert.NotNil(t, metadata.Timestamps) {
			assert.NotEmpty(t, metadata.Timestamps.CreatedAt)
			assert.NotEmpty(t, metadata.Timestamps.UpdatedAt)
		}
		if assert.NotNil(t, metadata.Team) {
			assert.NotEmpty(t, metadata.Team.ID)
			assert.NotEmpty(t, metadata.Team.Name)
			assert.NotEmpty(t, metadata.Team.Slug)
		}
	})

	t.Run("handles unauthorized gracefully", func(t *testing.T) {
		_, err := NewClientWithConfig(&Config{
			APIKeyID:     "foo",
			APIKeySecret: "bar",
			UserAgent:    testUserAgent,
		})

		var de hnyclient.DetailedError
		require.ErrorAs(t, err, &de)
		assert.Equal(t, http.StatusUnauthorized, de.Status)
	})
}

func newTestClient(t *testing.T) *Client {
	t.Helper()

	// load environment values from a .env, if available
	_ = godotenv.Load(testEnvFilePath)

	c, err := NewClientWithConfig(&Config{
		UserAgent: testUserAgent,
	})
	require.NoError(t, err, "failed to create test client")

	return c
}
