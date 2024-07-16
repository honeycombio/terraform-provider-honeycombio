package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/honeycombio/terraform-provider-honeycombio/client/errors"
)

// Markers describes all the marker-related methods that the Honeycomb API
// supports.
//
// API docs: https://docs.honeycomb.io/api/markers/
type Markers interface {
	// List all markers present in this dataset.
	List(ctx context.Context, dataset string) ([]Marker, error)

	// Get a marker by its ID.
	//
	// This method calls List internally since there is no API available to
	// directly get a single marker.
	Get(ctx context.Context, dataset string, id string) (*Marker, error)

	// Create a new marker in this dataset. When creating a marker ID may not
	// be set.
	Create(ctx context.Context, dataset string, m *Marker) (*Marker, error)

	// Update an existing marker.
	Update(ctx context.Context, dataset string, m *Marker) (*Marker, error)

	// Delete a marker from the dataset.
	Delete(ctx context.Context, dataset string, id string) error
}

// markers implements Markers.
type markers struct {
	client *Client
}

// Compile-time proof of interface implementation by type markers.
var _ Markers = (*markers)(nil)

// Marker represents a Honeycomb marker.
//
// API docs: https://docs.honeycomb.io/api/markers/#fields-on-a-marker
type Marker struct {
	ID string `json:"id,omitempty"`

	// The time the marker should be placed at, in Unix Time (= seconds since
	// epoch). If not set this will be set to when the request was received by
	// the API.
	StartTime int64 `json:"start_time,omitempty"`
	// The end time of the marker, in Unix Time (= seconds since epoch). This
	// can be used to indicate a time range. This field is optional.
	EndTime int64 `json:"end_time,omitempty"`
	// Message appears above the marker and can be used to desribe the marker.
	// This field is optional.
	Message string `json:"message,omitempty"`
	// Type is an optional marker identifier, eg 'deploy' or 'chef-run'. This
	// field is optional.
	Type string `json:"type,omitempty"`
	// URL is an optional url associated with the marker. This field is optional.
	URL string `json:"url,omitempty"`

	// Time the marker was created. This field is set by the API.
	CreatedAt *time.Time `json:"created_at,omitempty"`
	// Time the marker was last modified. This field is set by the API.
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	// Color of the marker. Colors are configured per dataset and can be set
	// per type of marker. This field is set by the API.
	Color string `json:"color,omitempty"`
}

func (s *markers) List(ctx context.Context, dataset string) ([]Marker, error) {
	var m []Marker
	err := s.client.Do(ctx, "GET", fmt.Sprintf("/1/markers/%s", urlEncodeDataset(dataset)), nil, &m)
	return m, err
}

func (s *markers) Get(ctx context.Context, dataset string, id string) (*Marker, error) {
	markers, err := s.List(ctx, dataset)
	if err != nil {
		return nil, err
	}

	for _, m := range markers {
		if m.ID == id {
			return &m, nil
		}
	}
	return nil, errors.DetailedError{
		Status:  http.StatusNotFound,
		Message: "Marker Not Found.",
	}
}

func (s *markers) Create(ctx context.Context, dataset string, data *Marker) (*Marker, error) {
	var m Marker
	err := s.client.Do(ctx, "POST", fmt.Sprintf("/1/markers/%s", urlEncodeDataset(dataset)), data, &m)
	return &m, err
}

func (s *markers) Update(ctx context.Context, dataset string, data *Marker) (*Marker, error) {
	var m Marker
	err := s.client.Do(ctx, "PUT", fmt.Sprintf("/1/markers/%s/%s", urlEncodeDataset(dataset), data.ID), data, &m)
	return &m, err
}

func (s *markers) Delete(ctx context.Context, dataset string, id string) error {
	return s.client.Do(ctx, "DELETE", fmt.Sprintf("/1/markers/%s/%s", urlEncodeDataset(dataset), id), nil, nil)
}
