package honeycombiosdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// The marker type, as described by https://honeycomb.io/docs/reference/api/#markers
// This struct is shamelessly copied from https://github.com/honeycombio/honeymarker/blob/master/marker.go
type Marker struct {
	ID string `json:"id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// StartTime unix timestamp truncates to seconds
	StartTime int64 `json:"start_time,omitempty"`
	// EndTime unix timestamp truncates to seconds
	EndTime int64 `json:"end_time,omitempty"`
	// Message is optional free-form text associated with the message
	Message string `json:"message,omitempty"`
	// Type is an optional marker identifier, eg 'deploy' or 'chef-run'
	Type string `json:"type,omitempty"`
	// URL is an optional url associated with the marker
	URL string `json:"url,omitempty"`
	// Color is not stored in the marker table but populated by a join
	Color string `json:"color,omitempty"`
}

func (c *Client) ListMarkers() (m []Marker, err error) {
	url := buildMarkersURL(c.dataset)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	c.populateHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if !is2xx(resp.StatusCode) {
		err = fmt.Errorf("Request failed with status code %d", resp.StatusCode)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&m)
	return
}

type CreateMarkerData struct {
	StartTime int64  `json:"start_time,omitempty"`
	EndTime   int64  `json:"end_time,omitempty"`
	Message   string `json:"message,omitempty"`
	Type      string `json:"type,omitempty"`
	URL       string `json:"url,omitempty"`
}

func (c *Client) CreateMarker(d CreateMarkerData) (m Marker, err error) {
	url := buildMarkersURL(c.dataset)

	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(d)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return
	}
	c.populateHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if !is2xx(resp.StatusCode) {
		err = fmt.Errorf("Request failed with status code %d", resp.StatusCode)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&m)
	return
}

func buildMarkersURL(dataset string) string {
	return fmt.Sprintf("%s/1/markers/%s", apiURL, dataset)
}

func is2xx(status int) bool {
	return status >= 200 && status < 300
}
