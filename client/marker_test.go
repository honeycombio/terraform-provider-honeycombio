package client

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMarkers(t *testing.T) {
	ctx := context.Background()

	var m *Marker
	var err error

	c := newTestClient(t)
	dataset := testDataset(t)

	t.Run("Create", func(t *testing.T) {
		data := &Marker{
			Message:   fmt.Sprintf("Test run at %v", time.Now()),
			Type:      "deploys",
			URL:       "http://example.com",
			StartTime: time.Now().Unix(),
		}
		m, err = c.Markers.Create(ctx, dataset, data)

		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(t, m.ID)
		assert.Equal(t, data.Message, m.Message)
		assert.Equal(t, data.Type, m.Type)
		assert.Equal(t, data.URL, m.URL)
		assert.WithinDuration(t, time.UnixMilli(data.StartTime), time.UnixMilli(m.StartTime), 5*time.Second)
		assert.WithinDuration(t, time.UnixMilli(data.EndTime), time.UnixMilli(m.EndTime), 5*time.Second)
	})

	t.Run("List", func(t *testing.T) {
		markers, err := c.Markers.List(ctx, dataset)

		assert.NoError(t, err)
		assert.Contains(t, markers, *m, "could not find newly created marker with List")
	})

	t.Run("Get", func(t *testing.T) {
		result, err := c.Markers.Get(ctx, dataset, m.ID)

		assert.NoError(t, err)
		assert.Equal(t, *m, *result)
	})

	t.Run("Update", func(t *testing.T) {
		m.EndTime = time.Now().Unix()

		result, err := c.Markers.Update(ctx, dataset, m)

		assert.NoError(t, err)
		assert.Equal(t, m.ID, result.ID)
		assert.Equal(t, m.Message, result.Message)
		assert.Equal(t, m.Type, result.Type)
		assert.Equal(t, m.URL, result.URL)
		assert.WithinDuration(t, time.UnixMilli(m.StartTime), time.UnixMilli(result.StartTime), 5*time.Second)
		assert.WithinDuration(t, time.UnixMilli(m.EndTime), time.UnixMilli(result.EndTime), 5*time.Second)
	})

	t.Run("Delete", func(t *testing.T) {
		err := c.Markers.Delete(ctx, dataset, m.ID)

		assert.NoError(t, err)
	})

	t.Run("Get_deletedMarker", func(t *testing.T) {
		_, err := c.Markers.Get(ctx, dataset, m.ID)

		assert.Equal(t, ErrNotFound, err)
	})
}
