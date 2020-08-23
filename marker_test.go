package honeycombio

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
		assert.Equal(t, data.StartTime, m.StartTime)
		assert.Equal(t, data.EndTime, m.EndTime)
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
		assert.Equal(t, m, result)
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
