package client

import (
	"context"
	"fmt"
	"time"
)

// markerSettings describes all the markerType-related methods that the Honeycomb API
// supports.
//
// API docs: https://docs.honeycomb.io/api/marker-settings
type MarkerSettings interface {
	// List all marker settings present in this dataset.
	List(ctx context.Context, dataset string) ([]MarkerSetting, error)

	// Get a marker type by its ID. Returns ErrNotFound if there is no marker with
	// the given ID in this dataset.
	//
	// This method calls List internally since there is no API available to
	// directly get a single marker.
	Get(ctx context.Context, dataset string, id string) (*MarkerSetting, error)

	// Create a new marker setting in this dataset.
	Create(ctx context.Context, dataset string, m *MarkerSetting) (*MarkerSetting, error)

	// Update an existing marker setting.
	Update(ctx context.Context, dataset string, m *MarkerSetting) (*MarkerSetting, error)

	// Delete a marker setting from the dataset.
	Delete(ctx context.Context, dataset string, id string) error
}

// markerSettings implements MarkerSettings.
type markerSettings struct {
	client *Client
}

// Compile-time proof of interface implementation by type markerSettings.
var _ MarkerSettings = (*markerSettings)(nil)

// MarkerSettings represents settings on a Honeycomb marker.
//
// API docs: https://docs.honeycomb.io/api/marker-settings/
type MarkerSetting struct {
	// Unique identifier of a marker type setting. This field is set by the API.
	ID string `json:"id,omitempty"`

	// Type is a required marker type setting identifier, eg 'deploy'.
	Type string `json:"type,omitempty"`

	// Color of the marker type setting. Colors are configured per dataset and can be set
	// per type of marker.
	Color string `json:"color,omitempty"`

	// Time the marker type was created. This field is set by the API.
	CreatedAt *time.Time `json:"created_at,omitempty"`
	// Time the marker type was last modified. This field is set by the API.
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func (s *markerSettings) List(ctx context.Context, dataset string) ([]MarkerSetting, error) {
	var m []MarkerSetting
	err := s.client.performRequest(ctx, "GET", fmt.Sprintf("/1/marker_settings/%s", urlEncodeDataset(dataset)), nil, &m)
	return m, err
}

func (s *markerSettings) Get(ctx context.Context, dataset string, id string) (*MarkerSetting, error) {
	markerSettings, err := s.List(ctx, dataset)
	if err != nil {
		return nil, err
	}

	for _, m := range markerSettings {
		if m.ID == id {
			return &m, nil
		}
	}
	return nil, ErrNotFound
}

func (s *markerSettings) Create(ctx context.Context, dataset string, data *MarkerSetting) (*MarkerSetting, error) {
	var m MarkerSetting
	err := s.client.performRequest(ctx, "POST", fmt.Sprintf("/1/marker_settings/%s", urlEncodeDataset(dataset)), data, &m)
	return &m, err
}

func (s *markerSettings) Update(ctx context.Context, dataset string, data *MarkerSetting) (*MarkerSetting, error) {
	var m MarkerSetting
	err := s.client.performRequest(ctx, "PUT", fmt.Sprintf("/1/marker_settings/%s/%s", urlEncodeDataset(dataset), data), data, &m)
	return &m, err
}

func (s *markerSettings) Delete(ctx context.Context, dataset string, id string) error {
	return s.client.performRequest(ctx, "DELETE", fmt.Sprintf("/1/marker_settings/%s/%s", urlEncodeDataset(dataset), id), nil, nil)
}
