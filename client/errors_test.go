package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_ParseDetailedError(t *testing.T) {
	var de DetailedError
	ctx := context.Background()
	c := newTestClient(t)

	t.Run("Post with no body should fail with 400 unparseable", func(t *testing.T) {
		err := c.performRequest(ctx, "POST", "/1/boards/", nil, nil)
		require.Error(t, err)
		assert.ErrorAs(t, err, &de)
		assert.Equal(t, de.Status, http.StatusBadRequest)
		assert.Equal(t, de.Type, "https://api.honeycomb.io/problems/unparseable")
		assert.Equal(t, de.Title, "The request body could not be parsed.")
		assert.Equal(t, de.Message, "could not parse request body")
	})

	t.Run("Get into non-existent dataset should fail with 404 'Dataset not found'", func(t *testing.T) {
		_, err := c.Markers.Get(ctx, "non-existent-dataset", "abcd1234")
		require.Error(t, err)
		assert.ErrorAs(t, err, &de)
		assert.Equal(t, de.Status, http.StatusNotFound)
		assert.Equal(t, de.Type, "https://api.honeycomb.io/problems/not-found")
		assert.Equal(t, de.Title, "The requested resource cannot be found.")
		assert.Equal(t, de.Message, "Dataset not found")
	})
}
