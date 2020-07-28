package honeycombiosdk

import (
	"os"
	"testing"
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

	return NewClient(apiKey, dataset)
}
