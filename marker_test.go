package honeycombiosdk

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMarkers(t *testing.T) {
	var m Marker
	var err error

	c := newTestClient(t)

	t.Run("Create", func(t *testing.T) {
		data := MarkerCreateData{
			Message: fmt.Sprintf("Test run at %v", time.Now()),
			Type:    "deploys",
			URL:     "http://example.com",
		}
		m, err = c.Markers.Create(data)
		if err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, m.ID)
		assert.Equal(t, data.Message, m.Message)
		assert.Equal(t, data.Type, m.Type)
		assert.Equal(t, data.URL, m.URL)
	})

	t.Run("List", func(t *testing.T) {
		markers, err := c.Markers.List()
		if err != nil {
			t.Fatal(err)
		}

		var createdMarker *Marker

		for _, marker := range markers {
			if marker.ID == m.ID {
				createdMarker = &marker
			}
		}
		if createdMarker == nil {
			t.Fatalf("could not find newly created marker with ID = %s", m.ID)
		}

		assert.Equal(t, m, *createdMarker)
	})

	t.Run("Get", func(t *testing.T) {
		marker, err := c.Markers.Get(m.ID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, m, *marker)
	})

	t.Run("Get_unexistingID", func(t *testing.T) {
		_, err := c.Markers.Get("0")
		assert.Equal(t, ErrNotFound, err)
	})
}
