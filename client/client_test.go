package client

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func init() {
	// load environment values from a .env, if available
	_ = godotenv.Load("../.env")
}

func newTestClient(t *testing.T) *Client {
	apiKey, ok := os.LookupEnv("HONEYCOMB_API_KEY")
	if !ok {
		t.Fatal("expected environment variable HONEYCOMB_API_KEY")
	}
	_, debug := os.LookupEnv("DEBUG")

	cfg := &Config{
		APIKey: apiKey,
		Debug:  debug,
	}
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}

	return c
}

func testDataset(t *testing.T) string {
	dataset, ok := os.LookupEnv("HONEYCOMBIO_DATASET")
	if !ok {
		t.Fatalf("expected environment variable HONEYCOMBIO_DATASET")
	}

	return dataset
}

func TestConfig_merge(t *testing.T) {
	config1 := &Config{
		APIKey:    "",
		APIUrl:    "",
		UserAgent: "go-honeycombio",
	}
	config2 := &Config{
		APIKey:    "",
		APIUrl:    "http://localhost",
		UserAgent: "",
	}

	expected := &Config{
		APIKey:    "",
		APIUrl:    "http://localhost",
		UserAgent: "go-honeycombio",
	}

	config1.merge(config2)

	assert.Equal(t, expected, config1)
}

func TestClient_invalidConfig(t *testing.T) {
	cfg := &Config{
		APIKey: "",
	}
	_, err := NewClient(cfg)
	assert.Error(t, err, "APIKey must be configured")

	cfg = &Config{
		APIKey: "123",
		APIUrl: "cache_object:foo/bar",
	}
	_, err = NewClient(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse APIUrl")
}

func TestClient_parseHoneycombioError(t *testing.T) {
	ctx := context.Background()
	c := newTestClient(t)

	err := c.performRequest(ctx, "POST", "/1/boards/", nil, nil)

	assert.Error(t, err, "400 Bad Request: request body should not be empty")
}
