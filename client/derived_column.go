package honeycombio

import (
	"context"
	"fmt"
)

// DerivedColumns describe all the derived columns-related methods that the
// Honeycomb API supports.
//
// API docs: https://docs.honeycomb.io/api/derived_columns/
type DerivedColumns interface {
	// List all derived columns in this dataset.
	List(ctx context.Context, dataset string) ([]DerivedColumn, error)

	// Get a derived column by its ID. Returns ErrNotFound if there is no
	// derived column with the given ID in this dataset.
	Get(ctx context.Context, dataset string, id string) (*DerivedColumn, error)

	// GetByAlias searches a derived column by its alias. Returns ErrNotFound if
	// there is no derived column with the given alias in this dataset.
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
type derivedColumns struct {
	client *Client
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
	// Expression of the derived column, this field is required and can not be
	// updated.
	// This should be an expression following the Derived Column syntax, as
	// described on https://docs.honeycomb.io/working-with-your-data/customizing-your-query/derived-columns/#derived-column-syntax
	Expression string `json:"expression"`
	// Optional.
	Description string `json:"description,omitempty"`
}

func (s *derivedColumns) List(ctx context.Context, dataset string) ([]DerivedColumn, error) {
	var c []DerivedColumn
	err := s.client.performRequest(ctx, "GET", "/1/derived_columns/"+urlEncodeDataset(dataset), nil, &c)
	return c, err
}

func (s *derivedColumns) Get(ctx context.Context, dataset string, id string) (*DerivedColumn, error) {
	var c DerivedColumn
	err := s.client.performRequest(ctx, "GET", fmt.Sprintf("/1/derived_columns/%s/%s", urlEncodeDataset(dataset), id), nil, &c)
	return &c, err
}

func (s *derivedColumns) GetByAlias(ctx context.Context, dataset string, alias string) (*DerivedColumn, error) {
	var c DerivedColumn
	err := s.client.performRequest(ctx, "GET", fmt.Sprintf("/1/derived_columns/%s?alias=%s", urlEncodeDataset(dataset), alias), nil, &c)
	return &c, err
}

func (s *derivedColumns) Create(ctx context.Context, dataset string, data *DerivedColumn) (*DerivedColumn, error) {
	var d DerivedColumn
	err := s.client.performRequest(ctx, "POST", "/1/derived_columns/"+urlEncodeDataset(dataset), data, &d)
	return &d, err
}

func (s *derivedColumns) Update(ctx context.Context, dataset string, data *DerivedColumn) (*DerivedColumn, error) {
	var d DerivedColumn
	err := s.client.performRequest(ctx, "PUT", fmt.Sprintf("/1/derived_columns/%s/%s", urlEncodeDataset(dataset), data.ID), data, &d)
	return &d, err
}

func (s *derivedColumns) Delete(ctx context.Context, dataset string, id string) error {
	return s.client.performRequest(ctx, "DELETE", fmt.Sprintf("/1/derived_columns/%s/%s", urlEncodeDataset(dataset), id), nil, nil)
}
