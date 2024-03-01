package client_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestMarkers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var m *client.Marker
	var err error

	c := newTestClient(t)
	dataset := testDataset(t)

	t.Run("Create", func(t *testing.T) {
		data := &client.Marker{
			Message:   fmt.Sprintf("Test run at %v", time.Now()),
			Type:      test.RandomStringWithPrefix("test.", 8),
			URL:       "http://example.com",
			StartTime: time.Now().Unix(),
		}
		m, err = c.Markers.Create(ctx, dataset, data)

		require.NoError(t, err)
		assert.NotNil(t, m.ID)
		assert.Equal(t, data.Message, m.Message)
		assert.Equal(t, data.Type, m.Type)
		assert.Equal(t, data.URL, m.URL)
		assert.WithinDuration(t, time.UnixMilli(data.StartTime), time.UnixMilli(m.StartTime), 5*time.Second)
		assert.WithinDuration(t, time.UnixMilli(data.EndTime), time.UnixMilli(m.EndTime), 5*time.Second)
	})

	t.Run("List", func(t *testing.T) {
		markers, err := c.Markers.List(ctx, dataset)

		require.NoError(t, err)
		assert.Contains(t, markers, *m, "could not find newly created marker with List")
	})

	t.Run("Get", func(t *testing.T) {
		result, err := c.Markers.Get(ctx, dataset, m.ID)

		require.NoError(t, err)
		assert.Equal(t, m.ID, result.ID)
		assert.Equal(t, m.Message, result.Message)
		assert.Equal(t, m.Type, result.Type)
		assert.Equal(t, m.URL, result.URL)
		assert.WithinDuration(t, time.UnixMilli(m.StartTime), time.UnixMilli(result.StartTime), 5*time.Second)
		assert.WithinDuration(t, time.UnixMilli(m.EndTime), time.UnixMilli(result.EndTime), 5*time.Second)
	})

	t.Run("Update", func(t *testing.T) {
		m.EndTime = time.Now().Unix()

		result, err := c.Markers.Update(ctx, dataset, m)

		require.NoError(t, err)
		assert.Equal(t, m.ID, result.ID)
		assert.Equal(t, m.Message, result.Message)
		assert.Equal(t, m.Type, result.Type)
		assert.Equal(t, m.URL, result.URL)
		assert.WithinDuration(t, time.UnixMilli(m.StartTime), time.UnixMilli(result.StartTime), 5*time.Second)
		assert.WithinDuration(t, time.UnixMilli(m.EndTime), time.UnixMilli(result.EndTime), 5*time.Second)
	})

	t.Run("Delete", func(t *testing.T) {
		err := c.Markers.Delete(ctx, dataset, m.ID)

		require.NoError(t, err)
	})

	t.Run("Fail to Get deleted Marker", func(t *testing.T) {
		_, err := c.Markers.Get(ctx, dataset, m.ID)

		var de client.DetailedError
		require.Error(t, err)
		require.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})
}
