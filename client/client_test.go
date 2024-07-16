package client_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/client/errors"
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

func TestClient_Errors(t *testing.T) {
	t.Parallel()

	var de errors.DetailedError
	ctx := context.Background()
	c := newTestClient(t)

	t.Run("Post with no body should fail with 400 unparseable", func(t *testing.T) {
		err := c.Do(ctx, "POST", "/1/boards/", nil, nil)
		require.Error(t, err)
		require.ErrorAs(t, err, &de)
		assert.Equal(t, http.StatusBadRequest, de.Status)
		assert.Equal(t, fmt.Sprintf("%s/problems/unparseable", c.EndpointURL()), de.Type)
		assert.Equal(t, "The request body could not be parsed.", de.Title)
		assert.Equal(t, "could not parse request body", de.Message)
	})

	t.Run("Get into non-existent dataset should fail with 404 'Dataset not found'", func(t *testing.T) {
		_, err := c.Markers.Get(ctx, "non-existent-dataset", "abcd1234")
		require.Error(t, err)
		require.ErrorAs(t, err, &de)
		assert.Equal(t, http.StatusNotFound, de.Status)
		assert.Equal(t, fmt.Sprintf("%s/problems/not-found", c.EndpointURL()), de.Type)
		assert.Equal(t, "The requested resource cannot be found.", de.Title)
		assert.Equal(t, "Dataset not found", de.Message)
	})

	t.Run("Creating a dataset without a name should return a validation error", func(t *testing.T) {
		createDatasetRequest := &client.Dataset{}
		_, err := c.Datasets.Create(ctx, createDatasetRequest)
		require.Error(t, err)
		require.ErrorAs(t, err, &de)
		assert.Equal(t, http.StatusUnprocessableEntity, de.Status)
		assert.Equal(t, fmt.Sprintf("%s/problems/validation-failed", c.EndpointURL()), de.Type)
		assert.Equal(t, "The provided input is invalid.", de.Title)
		assert.Equal(t, "The provided input is invalid.", de.Message)
		assert.Len(t, de.Details, 1)
		assert.Equal(t, "missing", de.Details[0].Code)
		assert.Equal(t, "name", de.Details[0].Field)
		assert.Equal(t, "cannot be blank", de.Details[0].Description)
		assert.Equal(t, "missing name - cannot be blank", de.Error())
	})
}
