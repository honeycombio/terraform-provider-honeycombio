package client_test

import (
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestMarkers(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	var m *client.Marker
	var err error

	c := newTestClient(t)
	dataset := testDataset(t)

	t.Run("Create", func(t *testing.T) {
		data := &client.Marker{
			Message:   fmt.Sprintf("Test run at %v", time.Now()),
			Type:      test.RandomStringWithPrefix("test.", 8),
			URL:       test.RandomURL(),
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
		// this has proven to be a bit racey after the create above, so we'll retry a few times
		assert.EventuallyWithT(t, func(col *assert.CollectT) {
			ml, err := c.Markers.List(ctx, dataset)
			require.NoError(col, err)

			// not doing an Equal here because the timestamps may be different
			// and confirming the ID is in the listing is sufficient
			assert.Condition(col, func() bool {
				return slices.ContainsFunc(ml, func(cm client.Marker) bool {
					return cm.ID == m.ID
				})
			})
		}, time.Second, 200*time.Millisecond, "could not find newly created Marker in List")
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
