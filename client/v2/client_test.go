package v2

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/jsonapi"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hnyclient "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
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
			APIKeyID:     "123",
			APIKeySecret: "456",
			BaseURL:      "https://api.example.com",
			UserAgent:    testUserAgent,
		})
		require.NoError(t, err)
		assert.Equal(t, "Bearer 123:456", c.Headers.Get("Authorization"))
		assert.Equal(t, "https://api.example.com", c.BaseURL.String())
		assert.Equal(t, testUserAgent, c.UserAgent)
	})

	t.Run("fails to construct a client with an invalid API URL", func(t *testing.T) {
		_, err := NewClientWithConfig(&Config{
			APIKeyID:     "123",
			APIKeySecret: "456",
			BaseURL:      "cache_object:foo/bar",
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
		c, err := NewClientWithConfig(&Config{
			APIKeyID:     "foo",
			APIKeySecret: "bar",
			UserAgent:    testUserAgent,
		})
		require.NoError(t, err)

		_, err = c.AuthInfo(ctx)

		var de hnyclient.DetailedError
		require.ErrorAs(t, err, &de)
		assert.Equal(t, http.StatusUnauthorized, de.Status)
		assert.Empty(t, de.Message)
		assert.Equal(t, "invalid API key", de.Title)
		assert.Equal(t, "unauthenticated/invalid-key", de.Type)
		if assert.Len(t, de.Details, 1) {
			assert.Equal(t, "Authorization header", de.Details[0].Field)
			assert.Equal(t, "invalid API key", de.Details[0].Description)
		}
		assert.Equal(t, "Authorization header - invalid API key", de.Error())
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

// newTestEnvironment creates a new Environment with a random name and description
// for testing purposes.
// The Environment is automatically deleted when the test completes.
func newTestEnvironment(ctx context.Context, t *testing.T, c *Client) *Environment {
	t.Helper()

	env, err := c.Environments.Create(ctx, &Environment{
		Name:        test.RandomStringWithPrefix("test.", 20),
		Description: helper.ToPtr(test.RandomString(50)),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		// disable deletion protection and delete the Environment
		c.Environments.Update(context.Background(), &Environment{
			ID: env.ID,
			Settings: &EnvironmentSettings{
				DeleteProtected: helper.ToPtr(false),
			},
		})
		c.Environments.Delete(ctx, env.ID)
	})

	return env
}

// fakeAPI serves the routes a test registers, plus /2/auth for the team slug,
// and records every other request it gets.
//
// Responses are hand-written JSON rather than output of the jsonapi package
// the client uses, so a bug in the client's encoding can't cancel itself out.
type fakeAPI struct {
	mu       sync.Mutex
	requests []recordedRequest
}

type recordedRequest struct {
	Method string
	Path   string
	Query  string
	Body   string
}

func newFakeAPIClient(t *testing.T, routes map[string]http.HandlerFunc) (*Client, *fakeAPI) {
	t.Helper()

	f := &fakeAPI{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /2/auth", func(w http.ResponseWriter, _ *http.Request) {
		writeJSONAPI(w, http.StatusOK, `{
			"data": {"id": "key1", "type": "api-keys",
				"attributes": {"name": "k", "key_type": "management", "disabled": false, "scopes": []},
				"relationships": {"team": {"data": {"type": "teams", "id": "team1"}}}},
			"included": [{"id": "team1", "type": "teams", "attributes": {"name": "Test", "slug": "`+testTeamSlug+`"}}]
		}`)
	})
	for pattern, h := range routes {
		mux.HandleFunc(pattern, h)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(body))
		if r.URL.Path != "/2/auth" {
			f.mu.Lock()
			f.requests = append(f.requests, recordedRequest{r.Method, r.URL.Path, r.URL.RawQuery, string(body)})
			f.mu.Unlock()
		}
		mux.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)

	c, err := NewClientWithConfig(&Config{
		APIKeyID:     "id",
		APIKeySecret: "secret",
		BaseURL:      srv.URL,
		UserAgent:    testUserAgent,
	})
	require.NoError(t, err)
	return c, f
}

func (f *fakeAPI) recorded() []recordedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]recordedRequest(nil), f.requests...)
}

func writeJSONAPI(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", jsonapi.MediaType)
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}

func noContent(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }
