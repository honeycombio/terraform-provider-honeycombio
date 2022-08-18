package client

import (
	"context"
	"fmt"
)

// Datasets describes all the dataset-related methods that the Honeycomb API
// supports.
//
// API docs: https://docs.honeycomb.io/api/datasets/
type DatasetDefinitions interface {
	// List all datasetDefinitions present in this dataset.
	List(ctx context.Context, dataset string) ([]DatasetDefinition, error)

	// Get a datasetDefinition by its dataset name. Returns ErrNotFound if there is no dataset
	GetAll(ctx context.Context, dataset string) (*DatasetDefinition, error)

	// Get a specific definition by its name for a specific dataset name. Returns ErrNotFound if there is no dataset
	GetByDefinition(ctx context.Context, dataset string, definition string) (*DatasetDefinition, error)

	// Create a new dataset. Only name should be set when creating a dataset,
	// all other fields are ignored.
	Create(ctx context.Context, dataset string, definition *DatasetDefinition) (*DatasetDefinition, error)

	// Update an existing dataset. Missing (optional) fields will set to their
	// respective defaults and not the currently existing values.
	Update(ctx context.Context, dataset string, definition *DatasetDefinition) (*DatasetDefinition, error)
}

type DefinitionColumn struct {
	Name *string `json:"name"`
	ID   *string `json:"id"`
}

// datasetDefinitions implements DatasetDefinitions.
type datasetDefinitions struct {
	client *Client
}

// Compile-time proof of interface implementation by type datasets.
var _ DatasetDefinitions = (*datasetDefinitions)(nil)

// DatasetDefinition represents a Honeycomb dataset metadata.
//
// API docs: https://docs.honeycomb.io/api/dataset_definitions // WIP
type DatasetDefinition struct {
	// Read only
	DurationMs     DefinitionColumn `json:"duration_ms"`
	Error          DefinitionColumn `json:"error"`
	Name           DefinitionColumn `json:"name"`
	ParentID       DefinitionColumn `json:"parent_id"`
	Route          DefinitionColumn `json:"route"`
	ServiceName    DefinitionColumn `json:"service_name"`
	SpanID         DefinitionColumn `json:"span_id"`
	SpanType       DefinitionColumn `json:"span_kind"` // Note span_kind vs span_type
	AnnotationType DefinitionColumn `json:"annotation_type"`
	LinkTraceID    DefinitionColumn `json:"link_trace_id"`
	LinkSpanID     DefinitionColumn `json:"link_span_id"`
	Status         DefinitionColumn `json:"status"`
	TraceID        DefinitionColumn `json:"trace_id"`
	User           DefinitionColumn `json:"user"`
}

func (s *datasetDefinitions) List(ctx context.Context, dataset string) ([]DatasetDefinition, error) {
	var ds []DatasetDefinition
	err := s.client.performRequest(ctx, "GET", "/1/datasets/"+urlEncodeDataset(dataset), nil, &ds)
	return ds, err
}

func (s *datasetDefinitions) GetAll(ctx context.Context, dataset string) (*DatasetDefinition, error) {
	var ds DatasetDefinition
	err := s.client.performRequest(ctx, "GET", fmt.Sprintf("/1/dataset_definitions/%s", urlEncodeDataset(dataset)), nil, &ds)
	return &ds, err
}

func (s *datasetDefinitions) GetByDefinition(ctx context.Context, dataset string, definition string) (*DatasetDefinition, error) {
	var ds DatasetDefinition
	err := s.client.performRequest(ctx, "GET", fmt.Sprintf("/1/dataset_definitions/%s/%s", urlEncodeDataset(dataset), definition), nil, &ds)
	return &ds, err
}

func (s *datasetDefinitions) Create(ctx context.Context, dataset string, data *DatasetDefinition) (*DatasetDefinition, error) {
	var ds DatasetDefinition
	err := s.client.performRequest(ctx, "POST", fmt.Sprintf("/1/datasets/%s", urlEncodeDataset(dataset)), data, &ds)
	return &ds, err
}

func (s *datasetDefinitions) Update(ctx context.Context, dataset string, data *DatasetDefinition) (*DatasetDefinition, error) {
	var ds DatasetDefinition
	err := s.client.performRequest(ctx, "PUT", fmt.Sprintf("/1/datasets/%s/%s", urlEncodeDataset(dataset), data.ID), data, &ds)
	return &ds, err
}

func (s *datasetDefinitions) Delete(ctx context.Context, dataset string, id string) error {
	return s.client.performRequest(ctx, "DELETE", fmt.Sprintf("/1/datasets/%s/%s", urlEncodeDataset(dataset), id), nil, nil)
}
