package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestMarkerSettings(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var m *client.MarkerSetting
	var err error

	c := newTestClient(t)
	dataset := testDataset(t)

	currentMarkerSetting := &client.MarkerSetting{
		Type:  test.RandomStringWithPrefix("test.", 8),
		Color: "#b71c1c",
	}

	t.Run("Create", func(t *testing.T) {

		m, err = c.MarkerSettings.Create(ctx, dataset, currentMarkerSetting)
		require.NoError(t, err)

		assert.NotNil(t, m.ID)
		assert.Equal(t, currentMarkerSetting.Type, m.Type)
		assert.Equal(t, currentMarkerSetting.Color, m.Color)
	})

	t.Run("List", func(t *testing.T) {
		markerSettings, err := c.MarkerSettings.List(ctx, dataset)

		assert.NoError(t, err)

		for _, m := range markerSettings {
			assert.NotNil(t, m.UpdatedAt, "updated at is empty")
			assert.NotNil(t, m.CreatedAt, "created at is empty")
			if m.Type == currentMarkerSetting.Type {
				currentMarkerSetting.UpdatedAt = m.UpdatedAt
				currentMarkerSetting.CreatedAt = m.CreatedAt
			}
		}
		assert.Contains(t, markerSettings, *m, "could not find marker settings with List")
	})

	t.Run("Get", func(t *testing.T) {
		result, err := c.MarkerSettings.Get(ctx, dataset, m.ID)

		assert.NoError(t, err)
		assert.Equal(t, *m, *result)
	})

	t.Run("Update", func(t *testing.T) {
		result, err := c.MarkerSettings.Update(ctx, dataset, m)

		assert.NoError(t, err)
		assert.Equal(t, m, result)
	})

	t.Run("Delete", func(t *testing.T) {
		err := c.MarkerSettings.Delete(ctx, dataset, m.ID)

		assert.NoError(t, err)
	})

	t.Run("Fail to Get deleted Marker Setting", func(t *testing.T) {
		_, err := c.MarkerSettings.Get(ctx, dataset, m.ID)

		var de client.DetailedError
		assert.Error(t, err)
		assert.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})
}
