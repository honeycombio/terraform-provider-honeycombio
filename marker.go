package honeycombiosdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Compile-time proof of interface implementation.
var _ Markers = (*markers)(nil)

// Markers describes all the markers related methods that Honeycomb supports.
type Markers interface {
	// List all markers present in this dataset.
	List() ([]Marker, error)

	// Get a marker by its ID.
	//
	// This method calls List internally since there is no API available to
	// directly get a single marker.
	Get(id string) (*Marker, error)

	// Create a new marker in this dataset.
	Create(data CreateData) (Marker, error)
}

// markers implements Markers.
type markers struct {
	client *Client
}

// Marker represents a Honeycomb marker, as described by https://docs.honeycomb.io/api/markers/#fields-on-a-marker
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

func (s *markers) List() (m []Marker, err error) {
	url := buildMarkersURL(s.client.dataset)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	s.client.populateHeaders(req)

	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if !is2xx(resp.StatusCode) {
		err = fmt.Errorf("request failed with status code %d", resp.StatusCode)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&m)
	return
}

func (s *markers) Get(ID string) (*Marker, error) {
	markers, err := s.List()
	if err != nil {
		return nil, err
	}

	for _, m := range markers {
		if m.ID == ID {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("marker with ID = %s was not found", ID)
}

// CreateData holds the data to create a new marker.
type CreateData struct {
	StartTime int64  `json:"start_time,omitempty"`
	EndTime   int64  `json:"end_time,omitempty"`
	Message   string `json:"message,omitempty"`
	Type      string `json:"type,omitempty"`
	URL       string `json:"url,omitempty"`
}

func (s *markers) Create(d CreateData) (m Marker, err error) {
	url := buildMarkersURL(s.client.dataset)

	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(d)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", url, buf)
	if err != nil {
		return
	}
	s.client.populateHeaders(req)

	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if !is2xx(resp.StatusCode) {
		err = fmt.Errorf("request failed with status code %d", resp.StatusCode)
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
