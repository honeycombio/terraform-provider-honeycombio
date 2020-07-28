package honeycombiosdk

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMarker(t *testing.T) {
	var m Marker
	var err error

	c := newTestClient(t)

	t.Run("CreateMarker", func(t *testing.T) {
		data := CreateMarkerData{
			Message: fmt.Sprintf("Test run at %v", time.Now()),
			Type:    "deploys",
			URL:     "http://example.com",
		}
		m, err = c.CreateMarker(data)
		if err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, m.ID)
		assert.Equal(t, data.Message, m.Message)
		assert.Equal(t, data.Type, m.Type)
		assert.Equal(t, data.URL, m.URL)
	})

	t.Run("ListMarkers", func(t *testing.T) {
		markers, err := c.ListMarkers()
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
}
