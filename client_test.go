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

	dataset, ok := os.LookupEnv("HONEYCOMBIO_DATASET")
	if !ok {
		t.Fatal("expected environment variable HONEYCOMBIO_DATASET")
	}

	cfg := &Config{
		APIKey:  apiKey,
		Dataset: dataset,
	}
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}

	return c
}

func TestConfig_merge(t *testing.T) {
	config1 := &Config{
		APIKey:    "123",
		Dataset:   "",
		APIUrl:    "http://localhost",
		UserAgent: "",
	}
	config2 := &Config{
		APIKey:    "",
		Dataset:   "dataset",
		APIUrl:    "",
		UserAgent: "go-honeycombio",
	}

	expected := &Config{
		APIKey:    "123",
		Dataset:   "dataset",
		APIUrl:    "http://localhost",
		UserAgent: "go-honeycombio",
	}

	config1.merge(config2)

	assert.Equal(t, expected, config1)
}

func TestClient_invalidConfig(t *testing.T) {
	cfg := &Config{
		APIKey:  "",
		Dataset: "dataset",
	}
	_, err := NewClient(cfg)
	assert.Error(t, err, "APIKey must be configured")

	cfg = &Config{
		APIKey:  "123",
		Dataset: "",
	}
	_, err = NewClient(cfg)
	assert.Error(t, err, "Dataset must be configured")

	cfg = &Config{
		APIKey:  "123",
		Dataset: "dataset",
		APIUrl:  "cache_object:foo/bar",
	}
	_, err = NewClient(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not parse APIUrl")
}
