package client

import (
	"context"
	"fmt"
	"time"
)

// markerTypes describes all the markerType-related methods that the Honeycomb API
// supports.
//
// API docs: https://docs.honeycomb.io/api/markerTypes/
type MarkerTypes interface {
	// List all markerTypes present in this dataset.
	List(ctx context.Context, dataset string) ([]MarkerType, error)

	// Create a new marker in this dataset. When creating a marker ID may not
	// be set.
	Create(ctx context.Context, dataset string, m *MarkerType) (*MarkerType, error)

	// Update an existing marker.
	Update(ctx context.Context, dataset string, m *MarkerType) (*MarkerType, error)

	// Delete a marker from the dataset.
	Delete(ctx context.Context, dataset string, id string) error
}

// markerTypes implements MarkerType.
type markerTypes struct {
	client *Client
}

// Compile-time proof of interface implementation by type markerTypes.
var _ MarkerTypes = (*markerTypes)(nil)

// MarkerType represents a Honeycomb marker.
//
// API docs: https://docs.honeycomb.io/api/markerTypes/#fields-on-a-marker
type MarkerType struct {
	ID string `json:"id,omitempty"`
	// Type is an optional marker identifier, eg 'deploy' or 'chef-run'. This
	// field is optional.
	Type string `json:"type,omitempty"`

	// Color of the marker. Colors are configured per dataset and can be set
	// per type of marker. This field is set by the API.
	Color string `json:"color,omitempty"`

	// Time the marker type was created. This field is set by the API.
	CreatedAt *time.Time `json:"created_at,omitempty"`
	// Time the marker type was last modified. This field is set by the API.
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

func (s *markerTypes) List(ctx context.Context, dataset string) ([]MarkerType, error) {
	var m []MarkerType
	err := s.client.performRequest(ctx, "GET", fmt.Sprintf("/1/markerTypes/%s", urlEncodeDataset(dataset)), nil, &m)
	return m, err
}

func (s *markerTypes) Get(ctx context.Context, dataset string, id string) (*MarkerType, error) {
	markerTypes, err := s.List(ctx, dataset)
	if err != nil {
		return nil, err
	}

	for _, m := range markerTypes {
		if m.ID == id {
			return &m, nil
		}
	}
	return nil, ErrNotFound
}

func (s *markerTypes) Create(ctx context.Context, dataset string, data *MarkerType) (*MarkerType, error) {
	var m MarkerType
	err := s.client.performRequest(ctx, "POST", fmt.Sprintf("/1/markerTypes/%s", urlEncodeDataset(dataset)), data, &m)
	return &m, err
}

func (s *markerTypes) Update(ctx context.Context, dataset string, data *MarkerType) (*MarkerType, error) {
	var m MarkerType
	err := s.client.performRequest(ctx, "PUT", fmt.Sprintf("/1/markerTypes/%s/%s", urlEncodeDataset(dataset), data.ID), data, &m)
	return &m, err
}

func (s *markerTypes) Delete(ctx context.Context, dataset string, id string) error {
	return s.client.performRequest(ctx, "DELETE", fmt.Sprintf("/1/markerTypes/%s/%s", urlEncodeDataset(dataset), id), nil, nil)
}
