package client_test

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

const testUserAgent = "go-honeycombio/test"

func newTestClient(t *testing.T) *client.Client {
	t.Helper()

	// load environment values from a .env, if available
	_ = godotenv.Load("../.env")

	apiKey, ok := os.LookupEnv(client.DefaultAPIKeyEnv)
	if !ok {
		t.Fatal("expected environment variable " + client.DefaultAPIKeyEnv)
	}
	_, debug := os.LookupEnv("HONEYCOMBIO_DEBUG")

	c, err := client.NewClientWithConfig(&client.Config{
		APIKey:    apiKey,
		Debug:     debug,
		UserAgent: testUserAgent,
	})
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
	t.Parallel()

	_, err := client.NewClientWithConfig(&client.Config{
		APIKey: "123",
		APIUrl: "cache_object:foo/bar",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse APIUrl")
}

func TestClient_IsClassic(t *testing.T) {
	t.Parallel()

	// load environment values from a .env, if available
	_ = godotenv.Load("../.env")

	ctx := context.Background()
	apiKey, ok := os.LookupEnv(client.DefaultAPIKeyEnv)
	if !ok {
		t.Fatal("expected environment variable " + client.DefaultAPIKeyEnv)
	}
	c, err := client.NewClientWithConfig(&client.Config{
		APIKey:    apiKey,
		UserAgent: testUserAgent,
	})
	require.NoError(t, err, "failed to create client")

	assert.Equal(t, len(apiKey) == 32, c.IsClassic(ctx))
}

func TestClient_EndpointURL(t *testing.T) {
	t.Parallel()

	endpointUrl := "https://api.example.com"
	c, err := client.NewClientWithConfig(&client.Config{
		APIUrl: endpointUrl,
		APIKey: "abcd123",
	})
	require.NoError(t, err, "failed to create client")

	assert.Equal(t, endpointUrl, c.EndpointURL().String())
}
