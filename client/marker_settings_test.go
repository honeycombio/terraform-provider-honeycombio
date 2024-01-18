package client_test

import (
	"context"
	"testing"
	"time"

	"golang.org/x/exp/slices"

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

	t.Run("Create", func(t *testing.T) {
		markerSetting := &client.MarkerSetting{
			Type:  test.RandomStringWithPrefix("test.", 8),
			Color: "#b71c1c",
		}
		m, err = c.MarkerSettings.Create(ctx, dataset, markerSetting)
		require.NoError(t, err)

		assert.NotEmpty(t, m.ID)
		assert.Equal(t, markerSetting.Type, m.Type)
		assert.Equal(t, markerSetting.Color, m.Color)
		assert.WithinDuration(t, *m.CreatedAt, time.Now(), 5*time.Second)
		assert.WithinDuration(t, *m.UpdatedAt, time.Now(), 5*time.Second)
	})

	t.Run("List", func(t *testing.T) {
		// this has proven to be a bit racey after the create above, so we'll retry a few times
		assert.EventuallyWithT(t, func(col *assert.CollectT) {
			ml, err := c.MarkerSettings.List(ctx, dataset)
			assert.NoError(col, err)

			// not doing an Equal here because the timestamps may be different
			// and confirming the ID is in the listing is sufficient
			assert.Condition(col, func() bool {
				return slices.ContainsFunc(ml, func(ms client.MarkerSetting) bool {
					return ms.ID == m.ID
				})
			})
		}, time.Second, 200*time.Millisecond, "could not find created Marker Setting in List")
	})

	t.Run("Get", func(t *testing.T) {
		result, err := c.MarkerSettings.Get(ctx, dataset, m.ID)
		assert.NoError(t, err)

		assert.Equal(t, m.ID, result.ID)
		assert.Equal(t, m.Type, result.Type)
		assert.Equal(t, m.Color, result.Color)
		assert.WithinDuration(t, *m.CreatedAt, *result.CreatedAt, 5*time.Second)
		assert.WithinDuration(t, *m.UpdatedAt, *result.UpdatedAt, 5*time.Second)
	})

	t.Run("Update", func(t *testing.T) {
		m.Color = "#000000"
		result, err := c.MarkerSettings.Update(ctx, dataset, m)
		assert.NoError(t, err)

		assert.Equal(t, m.ID, result.ID)
		assert.Equal(t, m.Type, result.Type)
		assert.Equal(t, m.Color, result.Color)
		assert.WithinDuration(t, *m.CreatedAt, *result.CreatedAt, 5*time.Second)
		assert.WithinDuration(t, *m.UpdatedAt, *result.UpdatedAt, 5*time.Second)
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
