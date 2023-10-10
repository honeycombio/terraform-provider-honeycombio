package client

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(t *testing.T) *Client {
	t.Helper()

	// load environment values from a .env, if available
	_ = godotenv.Load("../.env")

	apiKey, ok := os.LookupEnv("HONEYCOMB_API_KEY")
	if !ok {
		t.Fatal("expected environment variable HONEYCOMB_API_KEY")
	}
	_, debug := os.LookupEnv("HONEYCOMBIO_DEBUG")

	cfg := &Config{
		APIKey:    apiKey,
		Debug:     debug,
		UserAgent: "go-honeycombio/test",
	}
	c, err := NewClient(cfg)
	require.NoError(t, err, "failed to create client")

	return c
}

func testDataset(t *testing.T) string {
	t.Helper()

	dataset, ok := os.LookupEnv("HONEYCOMB_DATASET")
	if !ok {
		t.Fatalf("expected environment variable HONEYCOMB_DATASET")
	}

	return dataset
}

func TestClient_InvalidConfig(t *testing.T) {
	_, err := NewClient(&Config{
		APIKey: "123",
		APIUrl: "cache_object:foo/bar",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse APIUrl")
}

func TestClient_IsClassic(t *testing.T) {
	ctx := context.Background()
	c := newTestClient(t)

	assert.Equal(t, len(c.apiKey) == 32, c.IsClassic(ctx))
}
