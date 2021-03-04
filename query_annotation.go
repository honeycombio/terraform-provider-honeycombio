package honeycombio

import (
	"context"
	"fmt"
	"time"
)

// QueryAnnotations describes all the query annotation-related methods that the
// Honeycomb API supports.
//
// API docs: https://docs.honeycomb.io/api/query-annotations/
type QueryAnnotations interface {
	// List all query annotations.
	List(ctx context.Context, dataset string) ([]QueryAnnotation, error)

	// Get a query annotation by its ID. Returns ErrNotFound if there is no
	// query annotation with the given ID.
	Get(ctx context.Context, dataset string, id string) (*QueryAnnotation, error)

	// Create a new query annotation. When creating a new query annotation ID
	// may not be set.
	Create(ctx context.Context, dataset string, b *QueryAnnotation) (*QueryAnnotation, error)

	// Update an existing query annotation.
	Update(ctx context.Context, dataset string, b *QueryAnnotation) (*QueryAnnotation, error)

	// Delete a query annotation.
	Delete(ctx context.Context, dataset string, id string) error
}

// queryAnnotations implements QueryAnnotations.
type queryAnnotations struct {
	client *Client
}

// Compile-time proof of interface implementation by type queryAnnotations.
var _ QueryAnnotations = (*queryAnnotations)(nil)

// QueryAnnotation represents a Honeycomb query annotation.
//
// API docs: https://docs.honeycomb.io/api/query-annotations/#fields-on-a-query-annotation
type QueryAnnotation struct {
	ID string `json:"id,omitempty"`

	Name        string `json:"name"`
	Description string `json:"description"`
	QueryID     string `json:"query_id"`

	CreatedAt *time.Time `json:"created-at,omitempty"`
	UpdatedAt *time.Time `json:"updated-at,omitempty"`
}

func (s *queryAnnotations) List(ctx context.Context, dataset string) ([]QueryAnnotation, error) {
	var q []QueryAnnotation
	err := s.client.performRequest(ctx, "GET", fmt.Sprintf("/1/query_annotations/%s", urlEncodeDataset(dataset)), nil, &q)
	return q, err
}

func (s *queryAnnotations) Get(ctx context.Context, dataset string, ID string) (*QueryAnnotation, error) {
	var q QueryAnnotation
	err := s.client.performRequest(ctx, "GET", fmt.Sprintf("/1/query_annotations/%s/%s", urlEncodeDataset(dataset), ID), nil, &q)
	return &q, err
}

func (s *queryAnnotations) Create(ctx context.Context, dataset string, data *QueryAnnotation) (*QueryAnnotation, error) {
	var q QueryAnnotation
	err := s.client.performRequest(ctx, "POST", fmt.Sprintf("/1/query_annotations/%s", urlEncodeDataset(dataset)), data, &q)
	return &q, err
}

func (s *queryAnnotations) Update(ctx context.Context, dataset string, data *QueryAnnotation) (*QueryAnnotation, error) {
	var q QueryAnnotation
	err := s.client.performRequest(ctx, "PUT", fmt.Sprintf("/1/query_annotations/%s/%s", urlEncodeDataset(dataset), data.ID), data, &q)
	return &q, err
}

func (s *queryAnnotations) Delete(ctx context.Context, dataset string, id string) error {
	return s.client.performRequest(ctx, "DELETE", fmt.Sprintf("/1/query_annotations/%s/%s", urlEncodeDataset(dataset), id), nil, nil)
}
