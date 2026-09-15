package client

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/honeycombio/terraform-provider-honeycombio/client/internal/cache"
)

// columnCacheTTL is how long a dataset's cached column list is served
// before being refetched. Kept short: the cache exists to collapse the
// burst of per-resource reads within a single Terraform or Pulumi
// operation, not to avoid refetching across operations.
const columnCacheTTL = time.Minute

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
//
// When read caching is enabled (Config.ReadCaching) reads are served
// from a short-lived per-dataset cache of the full column list so that
// large plans, refreshes, and applies don't issue one API request per
// managed column. Writes invalidate the dataset's cached list.
type columns struct {
	client *Client
	// cache is nil unless read caching is enabled.
	cache *cache.Cache[Column]
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
	// ColumnTypeHistogram applies only to metrics datasets.
	ColumnTypeHistogram ColumnType = "histogram"
)

// ColumnTypes returns an exhaustive list of column types.
func ColumnTypes() []ColumnType {
	return []ColumnType{
		ColumnTypeString,
		ColumnTypeFloat,
		ColumnTypeInteger,
		ColumnTypeBoolean,
		ColumnTypeHistogram,
	}
}

func (s *columns) List(ctx context.Context, dataset string) ([]Column, error) {
	fetch := func(ctx context.Context) ([]Column, error) {
		var c []Column
		err := s.client.Do(ctx, "GET", "/1/columns/"+urlEncodeDataset(dataset), nil, &c)
		return c, err
	}
	if s.cache == nil {
		return fetch(ctx)
	}
	return s.cache.Get(ctx, urlEncodeDataset(dataset), fetch)
}

func (s *columns) Get(ctx context.Context, dataset string, id string) (*Column, error) {
	var c Column
	err := s.client.Do(ctx, "GET", fmt.Sprintf("/1/columns/%s/%s", urlEncodeDataset(dataset), id), nil, &c)
	return &c, err
}

func (s *columns) GetByKeyName(ctx context.Context, dataset string, keyName string) (*Column, error) {
	if s.cache != nil {
		columns, err := s.List(ctx, dataset)
		if err == nil {
			for i := range columns {
				if columns[i].KeyName == keyName {
					return &columns[i], nil
				}
			}
		}
	}

	// direct lookup: the only path when read caching is disabled, and
	// the fallback for a cache miss — this preserves the API's error
	// responses (e.g. a 404 for a column which truly doesn't exist)
	// exactly as they were
	var c Column
	err := s.client.Do(ctx, "GET", fmt.Sprintf("/1/columns/%s?key_name=%s", urlEncodeDataset(dataset), url.QueryEscape(keyName)), nil, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *columns) Create(ctx context.Context, dataset string, data *Column) (*Column, error) {
	var c Column
	err := s.client.Do(ctx, "POST", "/1/columns/"+urlEncodeDataset(dataset), data, &c)
	s.invalidateCache(dataset)
	return &c, err
}

func (s *columns) Update(ctx context.Context, dataset string, data *Column) (*Column, error) {
	var c Column
	err := s.client.Do(ctx, "PUT", fmt.Sprintf("/1/columns/%s/%s", urlEncodeDataset(dataset), data.ID), data, &c)
	s.invalidateCache(dataset)
	return &c, err
}

func (s *columns) Delete(ctx context.Context, dataset string, id string) error {
	err := s.client.Do(ctx, "DELETE", fmt.Sprintf("/1/columns/%s/%s", urlEncodeDataset(dataset), id), nil, nil)
	s.invalidateCache(dataset)
	return err
}

// invalidateCache drops the cached column list a write may have made
// stale; the next read fetches a fresh list.
//
// Columns are always dataset-scoped -- there is no environment-wide
// column -- so a write can only affect its own dataset's list.
func (s *columns) invalidateCache(dataset string) {
	if s.cache == nil {
		return
	}
	s.cache.Invalidate(urlEncodeDataset(dataset))
}
