package client

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// Columns describe all the columns-related methods that the Honeycomb API
// supports.
//
// API docs: https://docs.honeycomb.io/api/columns/
type Columns interface {
	// List all columns in this dataset.
	List(ctx context.Context, dataset string) ([]Column, error)

	// Get a column by its ID.
	Get(ctx context.Context, dataset string, id string) (*Column, error)

	// GetByKeyName searches a column by its key name.
	GetByKeyName(ctx context.Context, dataset string, keyName string) (*Column, error)

	// Create a new column in this dataset. When creating a new column ID may
	// not be set. The KeyName must be unique for this dataset.
	Create(ctx context.Context, dataset string, c *Column) (*Column, error)

	// Update an existing column.
	Update(ctx context.Context, dataset string, c *Column) (*Column, error)

	// Delete a column.
	Delete(ctx context.Context, dataset string, id string) error
}

// columns implements Columns.
type columns struct {
	client *Client
}

// Compile-time proof of interface implementation by type columns.
var _ Columns = (*columns)(nil)

// Column represents a Honeycomb column in a dataset.
//
// API docs: https://docs.honeycomb.io/api/columns/#fields-on-a-column
type Column struct {
	ID string `json:"id,omitempty"`

	// Name of the column, this field is required.
	KeyName string `json:"key_name"`
	// Deprecated, optional.
	Alias string `json:"alias,omitempty"`
	// Optional, defaults to false.
	Hidden *bool `json:"hidden,omitempty"`
	// Optional.
	Description string `json:"description,omitempty"`
	// Optional, defaults to string.
	Type *ColumnType `json:"type,omitempty"`

	// Read only
	LastWrittenAt time.Time `json:"last_written,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

// ColumnType determines the type of column.
type ColumnType string

// Declaration of column types.
const (
	ColumnTypeString  ColumnType = "string"
	ColumnTypeFloat   ColumnType = "float"
	ColumnTypeInteger ColumnType = "integer"
	ColumnTypeBoolean ColumnType = "boolean"
)

// ColumnTypes returns an exhaustive list of column types.
func ColumnTypes() []ColumnType {
	return []ColumnType{ColumnTypeString, ColumnTypeFloat, ColumnTypeInteger, ColumnTypeBoolean}
}

func (s *columns) List(ctx context.Context, dataset string) ([]Column, error) {
	var c []Column
	err := s.client.performRequest(ctx, "GET", "/1/columns/"+urlEncodeDataset(dataset), nil, &c)
	return c, err
}

func (s *columns) Get(ctx context.Context, dataset string, id string) (*Column, error) {
	var c Column
	err := s.client.performRequest(ctx, "GET", fmt.Sprintf("/1/columns/%s/%s", urlEncodeDataset(dataset), id), nil, &c)
	return &c, err
}

func (s *columns) GetByKeyName(ctx context.Context, dataset string, keyName string) (*Column, error) {
	var c Column
	err := s.client.performRequest(ctx, "GET", fmt.Sprintf("/1/columns/%s?key_name=%s", urlEncodeDataset(dataset), url.QueryEscape(keyName)), nil, &c)
	return &c, err
}

func (s *columns) Create(ctx context.Context, dataset string, data *Column) (*Column, error) {
	var c Column
	err := s.client.performRequest(ctx, "POST", "/1/columns/"+urlEncodeDataset(dataset), data, &c)
	return &c, err
}

func (s *columns) Update(ctx context.Context, dataset string, data *Column) (*Column, error) {
	var c Column
	err := s.client.performRequest(ctx, "PUT", fmt.Sprintf("/1/columns/%s/%s", urlEncodeDataset(dataset), data.ID), data, &c)
	return &c, err
}

func (s *columns) Delete(ctx context.Context, dataset string, id string) error {
	return s.client.performRequest(ctx, "DELETE", fmt.Sprintf("/1/columns/%s/%s", urlEncodeDataset(dataset), id), nil, nil)
}
