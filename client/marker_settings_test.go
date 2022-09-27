package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkerSettings(t *testing.T) {
	ctx := context.Background()

	c := newTestClient(t)
	dataset := testDataset(t)

	var m *MarkerSetting
	var err error

	currentMarkerSetting := &MarkerSetting{
		Type:  "deploy",
		Color: "#b71c1c",
	}

	t.Run("Create", func(t *testing.T) {
		m, err = c.MarkerSettings.Create(ctx, dataset, currentMarkerSetting)

		assert.NoError(t, err)
		assert.NotNil(t, m.UpdatedAt, "updated at is empty")
		assert.NotNil(t, m.CreatedAt, "created at is empty")

		if m.Type == currentMarkerSetting.Type {
			currentMarkerSetting.UpdatedAt = m.UpdatedAt
			currentMarkerSetting.CreatedAt = m.CreatedAt
		}
		// assert.Equal(t, *currentMarkerSetting, *m)

		assert.Equal(t, currentMarkerSetting.Type, m.Type)
		assert.Equal(t, currentMarkerSetting.Color, m.Color)
	})

	// t.Run("List", func(t *testing.T) {
	// 	markerSettings, err := c.MarkerSettings.List(ctx, dataset)

	// 	assert.NoError(t, err)
	// 	assert.Contains(t, markerSettings, *m, "could not find newly created marker settings with List")
	// })

	// t.Run("Get", func(t *testing.T) {
	// 	result, err := c.MarkerSettings.Get(ctx, dataset, m.ID)

	// 	assert.NoError(t, err)
	// 	assert.Equal(t, *m, *result)
	// })

	// t.Run("Update", func(t *testing.T) {

	// 	result, err := c.MarkerSettings.Update(ctx, dataset, m)

	// 	assert.NoError(t, err)
	// 	assert.Equal(t, m, result)
	// })

	// t.Run("Delete", func(t *testing.T) {
	// 	err := c.MarkerSettings.Delete(ctx, dataset, m.ID)

	// 	assert.NoError(t, err)
	// })

	// t.Run("Get_deletedMarker", func(t *testing.T) {
	// 	_, err := c.MarkerSettings.Get(ctx, dataset, m.ID)

	// 	assert.Equal(t, ErrNotFound, err)
	// })
}
