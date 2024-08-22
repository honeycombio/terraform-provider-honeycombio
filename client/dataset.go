package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Datasets describes all the dataset-related methods that the Honeycomb API
// supports.
//
// API docs: https://docs.honeycomb.io/api/datasets/
type Datasets interface {
	// List all datasets.
	List(ctx context.Context) ([]Dataset, error)

	// Get a dataset by its slug.
	Get(ctx context.Context, slug string) (*Dataset, error)

	// Creates a new dataset.
	//
	// Will return DatasetExistsErr if a dataset with that name already exists
	// in the Environment.
	Create(ctx context.Context, dataset *Dataset) (*Dataset, error)

	// Update an existing dataset. Missing (optional) fields will set to their
	// respective defaults and not the currently existing values.
	Update(ctx context.Context, dataset *Dataset) (*Dataset, error)

	// Delete a dataset by its slug.
	Delete(ctx context.Context, slug string) error
}

// datasets implements Datasets.
type datasets struct {
	client *Client
}

// DatasetExistsErr is returned by Create when the dataset already exists.
var DatasetExistsErr = fmt.Errorf("dataset already exists")

// Compile-time proof of interface implementation by type datasets.
var _ Datasets = (*datasets)(nil)

// Dataset represents a Honeycomb dataset.
//
// API docs: https://docs.honeycomb.io/api/datasets/
type Dataset struct {
	Name            string          `json:"name"`
	Description     string          `json:"description,omitempty"`
	Slug            string          `json:"slug,omitempty"`
	ExpandJSONDepth int             `json:"expand_json_depth,omitempty"`
	Settings        DatasetSettings `json:"settings,omitempty"`
	// Read only
	LastWrittenAt time.Time `json:"last_written_at,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
}

type DatasetSettings struct {
	// Optional, defaults to true. Cannot be set on creation.
	DeleteProtected *bool `json:"delete_protected,omitempty"`
}

func (s datasets) List(ctx context.Context) ([]Dataset, error) {
	var datasets []Dataset
	err := s.client.Do(ctx, "GET", "/1/datasets", nil, &datasets)
	return datasets, err
}

func (s datasets) Get(ctx context.Context, slug string) (*Dataset, error) {
	var dataset Dataset
	err := s.client.Do(ctx, "GET", fmt.Sprintf("/1/datasets/%s", urlEncodeDataset(slug)), nil, &dataset)
	return &dataset, err
}

func (s datasets) Create(ctx context.Context, d *Dataset) (*Dataset, error) {
	req, err := s.client.newRequest(ctx, "POST", "/1/datasets", d)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK: // the API doesn't consider this an error, but we do
		return nil, DatasetExistsErr
	case http.StatusCreated:
		var dataset Dataset
		err = json.NewDecoder(resp.Body).Decode(&dataset)
		if err != nil {
			return nil, err
		}
		return &dataset, err
	default:
		return nil, ErrorFromResponse(resp)
	}
}

func (s datasets) Update(ctx context.Context, d *Dataset) (*Dataset, error) {
	var dataset Dataset
	err := s.client.Do(ctx, "PUT", fmt.Sprintf("/1/datasets/%s", urlEncodeDataset(d.Slug)), d, &dataset)
	return &dataset, err
}

func (s datasets) Delete(ctx context.Context, slug string) error {
	return s.client.Do(ctx, "DELETE", fmt.Sprintf("/1/datasets/%s", urlEncodeDataset(slug)), nil, nil)
}
