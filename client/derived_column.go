package client

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/honeycombio/terraform-provider-honeycombio/client/internal/cache"
)

// derivedColumnCacheTTL is how long a dataset's cached derived column
// list is served before being refetched. Kept short: the cache exists to
// collapse the burst of per-resource reads within a single Terraform or
// Pulumi operation, not to avoid refetching across operations.
const derivedColumnCacheTTL = time.Minute

// DerivedColumns describe all the derived columns-related methods that the
// Honeycomb API supports.
//
// API docs: https://docs.honeycomb.io/api/derived_columns/
type DerivedColumns interface {
	// List all derived columns in this dataset.
	List(ctx context.Context, dataset string) ([]DerivedColumn, error)

	// Get a derived column by its ID.
	Get(ctx context.Context, dataset string, id string) (*DerivedColumn, error)

	// GetByAlias searches a derived column by its alias.
	GetByAlias(ctx context.Context, dataset string, alias string) (*DerivedColumn, error)

	// Create a new derived column in this dataset. When creating a new derived
	// column ID may not be set. The Alias must be unique for this dataset.
	Create(ctx context.Context, dataset string, d *DerivedColumn) (*DerivedColumn, error)

	// Update an existing derived column.
	Update(ctx context.Context, dataset string, d *DerivedColumn) (*DerivedColumn, error)

	// Delete a derived column.
	Delete(ctx context.Context, dataset string, id string) error
}

// derivedColumns implements DerivedColumns.
//
// Reads are served from a short-lived per-dataset cache of the full
// column list so that large plans, refreshes, and applies don't issue
// one API request per managed derived column. Writes invalidate the
// dataset's cached list.
type derivedColumns struct {
	client *Client
	cache  *cache.ListCache[DerivedColumn]
}

// Compile-time proof of interface implementation by type derivedColumns.
var _ DerivedColumns = (*derivedColumns)(nil)

// Column represents a Honeycomb derived column in a dataset.
//
// API docs: https://docs.honeycomb.io/api/derived_columns/#fields-on-a-derivedcolumn
type DerivedColumn struct {
	ID string `json:"id,omitempty"`
	// Alias of the derived column, this field is required and can not be
	// updated.
	Alias string `json:"alias"`
	// Expression of the derived column, this field is required.
	// This should be an expression following the Derived Column syntax, as
	// described on https://docs.honeycomb.io/working-with-your-data/customizing-your-query/derived-columns/#derived-column-syntax
	Expression string `json:"expression"`
	// Optional.
	Description string `json:"description,omitempty"`
}

func (s *derivedColumns) List(ctx context.Context, dataset string) ([]DerivedColumn, error) {
	return s.cache.Get(ctx, urlEncodeDataset(dataset), func(ctx context.Context) ([]DerivedColumn, error) {
		var c []DerivedColumn
		err := s.client.Do(ctx, "GET", fmt.Sprintf("/1/derived_columns/%s", urlEncodeDataset(dataset)), nil, &c)
		return c, err
	})
}

func (s *derivedColumns) Get(ctx context.Context, dataset string, id string) (*DerivedColumn, error) {
	var c DerivedColumn
	err := s.client.Do(ctx, "GET", fmt.Sprintf("/1/derived_columns/%s/%s", urlEncodeDataset(dataset), id), nil, &c)
	return &c, err
}

func (s *derivedColumns) GetByAlias(ctx context.Context, dataset string, alias string) (*DerivedColumn, error) {
	columns, err := s.List(ctx, dataset)
	if err == nil {
		for i := range columns {
			if columns[i].Alias == alias {
				return &columns[i], nil
			}
		}
	}

	// fall back to a direct lookup: the alias may be resolvable
	// server-side even when absent from the dataset's cached list, and
	// this preserves the API's error responses (e.g. a 404 for a column
	// which truly doesn't exist) exactly as they were
	var c DerivedColumn
	err = s.client.Do(ctx, "GET", fmt.Sprintf("/1/derived_columns/%s?alias=%s", urlEncodeDataset(dataset), url.QueryEscape(alias)), nil, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *derivedColumns) Create(ctx context.Context, dataset string, data *DerivedColumn) (*DerivedColumn, error) {
	var d DerivedColumn
	err := s.client.Do(ctx, "POST", fmt.Sprintf("/1/derived_columns/%s", urlEncodeDataset(dataset)), data, &d)
	s.cache.Invalidate(urlEncodeDataset(dataset))
	return &d, err
}

func (s *derivedColumns) Update(ctx context.Context, dataset string, data *DerivedColumn) (*DerivedColumn, error) {
	var d DerivedColumn
	err := s.client.Do(ctx, "PUT", fmt.Sprintf("/1/derived_columns/%s/%s", urlEncodeDataset(dataset), data.ID), data, &d)
	s.cache.Invalidate(urlEncodeDataset(dataset))
	return &d, err
}

func (s *derivedColumns) Delete(ctx context.Context, dataset string, id string) error {
	err := s.client.Do(ctx, "DELETE", fmt.Sprintf("/1/derived_columns/%s/%s", urlEncodeDataset(dataset), id), nil, nil)
	s.cache.Invalidate(urlEncodeDataset(dataset))
	return err
}
