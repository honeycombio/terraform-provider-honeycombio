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

	// Get a specific definition by its name for a specific dataset name. Returns ErrNotFound if there is no dataset
	Get(ctx context.Context, dataset string, definitionName string) (*DatasetDefinition, error)

	// Create a new datasetd definition from the data passed in.
	Create(ctx context.Context, dataset string, definitionName string, data *DatasetDefinition) (*DatasetDefinition, error)

	// Delete specific definition for an existing dataset.
	Delete(ctx context.Context, dataset string, definitionName string) error
}

type DefinitionColumn struct {
	Name *string `json:"name"`
	ID   *string `json:"id"`
}

// Compile-time proof of interface implementation by type datasets definiitions.
var _ DatasetDefinitions = (*datasetDefinitions)(nil)

// datasetDefinitions implements DatasetDefinitions.
type datasetDefinitions struct {
	client *Client
}

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
	err := s.client.performRequest(ctx, "GET", "/1/dataset_definitions/"+urlEncodeDataset(dataset), nil, &ds)
	return ds, err
}

func (s *datasetDefinitions) Get(ctx context.Context, dataset string, definitionName string) (*DatasetDefinition, error) {
	var ds DatasetDefinition
	err := s.client.performRequest(ctx, "GET", fmt.Sprintf("/1/dataset_definitions/%s/%s", urlEncodeDataset(dataset), definitionName), nil, &ds)
	return &ds, err
}

func (s *datasetDefinitions) Create(ctx context.Context, dataset string, definitionName string, data *DatasetDefinition) (*DatasetDefinition, error) {
	var ds DatasetDefinition
	err := s.client.performRequest(ctx, "POST", fmt.Sprintf("/1/dataset_definitions/%s/%s", urlEncodeDataset(dataset), definitionName), data, &ds)
	return &ds, err
}

func (s *datasetDefinitions) Delete(ctx context.Context, dataset string, definitionName string) error {
	return s.client.performRequest(ctx, "DELETE", fmt.Sprintf("/1/dataset_definitions/%s/%s", urlEncodeDataset(dataset), definitionName), nil, nil)
}
