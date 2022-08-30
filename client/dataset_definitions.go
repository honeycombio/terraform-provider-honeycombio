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
	Create(ctx context.Context, dataset string, data *DatasetDefinition) (*DatasetDefinition, error)

	// Update specific dataset definition value
	Update(ctx context.Context, dataset string, data *DatasetDefinition) (*DatasetDefinition, error)

	// Delete specific definition for an existing dataset.
	Delete(ctx context.Context, dataset string, definitionName string) error
}

// Compile-time proof of interface implementation by type datasets definiitions.
var _ DatasetDefinitions = (*datasetDefinitions)(nil)

// datasetDefinitions implements DatasetDefinitions.
type datasetDefinitions struct {
	client *Client
}

// DatasetDefinition represents a Honeycomb dataset metadata.
//
type DefinitionColumn struct {
	ID         *string `json:"id"`
	Name       *string `json:"name"`
	ColumnType *string `json:"column_type"`
}

// DatasetDefinition represents a Honeycomb dataset metadata.
// API docs: https://docs.honeycomb.io/api/dataset-definitions/
type DatasetDefinition struct {
	// Read only
	SpanID         DefinitionColumn `json:"span_id"`
	TraceID        DefinitionColumn `json:"trace_id"`
	ParentID       DefinitionColumn `json:"parent_id"`
	Name           DefinitionColumn `json:"name"`
	ServiceName    DefinitionColumn `json:"service_name"`
	DurationMs     DefinitionColumn `json:"duration_ms"`
	SpanKind       DefinitionColumn `json:"span_kind"` // Note span_kind vs span_type
	AnnotationType DefinitionColumn `json:"annotation_type"`
	LinkSpanID     DefinitionColumn `json:"link_span_id"`
	LinkTraceID    DefinitionColumn `json:"link_trace_id"`
	Error          DefinitionColumn `json:"error"`
	Status         DefinitionColumn `json:"status"`
	Route          DefinitionColumn `json:"route"`
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

func (s *datasetDefinitions) Create(ctx context.Context, dataset string, data *DatasetDefinition) (*DatasetDefinition, error) {
	var ds DatasetDefinition
	dsName := data.Name
	err := s.client.performRequest(ctx, "POST", fmt.Sprintf("/1/dataset_definitions/%s/%v", urlEncodeDataset(dataset), dsName), data, &ds)
	return &ds, err
}

func (s *datasetDefinitions) Update(ctx context.Context, dataset string, data *DatasetDefinition) (*DatasetDefinition, error) {
	var ds DatasetDefinition
	dsName := data.Name
	err := s.client.performRequest(ctx, "PATCH", fmt.Sprintf("/1/dataset_definitions/%s/%v", urlEncodeDataset(dataset), dsName), data, &ds)
	return &ds, err
}

func (s *datasetDefinitions) Delete(ctx context.Context, dataset string, definitionName string) error {
	return s.client.performRequest(ctx, "DELETE", fmt.Sprintf("/1/dataset_definitions/%s/%s", urlEncodeDataset(dataset), definitionName), nil, nil)
}
