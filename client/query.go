package client

import (
	"context"
	"fmt"
)

// Queries describe all the query-related methods that the Honeycomb API
// supports.
//
// API docs: https://docs.honeycomb.io/api/queries/
type Queries interface {
	// Get a query by its ID.
	Get(ctx context.Context, dataset string, id string) (*QuerySpec, error)

	// Create a new query in this dataset. When creating a new query ID may
	// not be set.
	Create(ctx context.Context, dataset string, c *QuerySpec) (*QuerySpec, error)
}

// queries implements Queries.
type queries struct {
	client *Client
}

// Compile-time proof of interface implementation by type queries.
var _ Queries = (*queries)(nil)

func (s *queries) Get(ctx context.Context, dataset string, id string) (*QuerySpec, error) {
	var q QuerySpec
	err := s.client.Do(ctx, "GET", fmt.Sprintf("/1/queries/%s/%s", urlEncodeDataset(dataset), id), nil, &q)
	return &q, err
}

func (s *queries) Create(ctx context.Context, dataset string, data *QuerySpec) (*QuerySpec, error) {
	var q QuerySpec
	err := s.client.Do(ctx, "POST", "/1/queries/"+urlEncodeDataset(dataset), data, &q)
	return &q, err
}
