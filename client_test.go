package honeycombio

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestClient(t *testing.T) *Client {
	apiKey, ok := os.LookupEnv("HONEYCOMBIO_APIKEY")
	if !ok {
		t.Fatal("expected environment variable HONEYCOMBIO_APIKEY")
	}

	cfg := &Config{
		APIKey: apiKey,
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
		APIKey:    "123",
		APIUrl:    "",
		UserAgent: "",
	}
	config2 := &Config{
		APIKey:    "",
		APIUrl:    "http://localhost",
		UserAgent: "",
	}

	expected := &Config{
		APIKey:    "123",
		APIUrl:    "http://localhost",
		UserAgent: "",
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
