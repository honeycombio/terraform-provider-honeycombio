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
	// Get All Dataset Definitions for a Dataset
	List(ctx context.Context, dataset string) ([]DatasetDefinition, error)

	// Get All Dataset Definitions
	Update(ctx context.Context, dataset string, data *DatasetDefinition) (*DatasetDefinition, error)
}

// Compile-time proof of interface implementation by type datasets definiitions.
var _ DatasetDefinitions = (*datasetDefinitions)(nil)

// datasetDefinitions implements DatasetDefinitions.
type datasetDefinitions struct {
	client *Client
}

// DatasetDefinition represents a Honeycomb dataset metadata.
type DefinitionColumn struct {
	Name       string `json:"name"`
	ColumnType string `json:"column_type,omitempty"`
}

// DatasetDefinition represents a Honeycomb dataset metadata.
// API docs: https://docs.honeycomb.io/api/dataset-definitions/
type DatasetDefinition struct {
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

// Required by Terraform Provider Client
func (s *datasetDefinitions) List(ctx context.Context, dataset string) ([]DatasetDefinition, error) {
	var definitions []DatasetDefinition
	err := s.client.performRequest(ctx, "GET", "/1/dataset_definitions/"+urlEncodeDataset(dataset), nil, &definitions)
	return definitions, err
}

func (s *datasetDefinitions) Update(ctx context.Context, dataset string, data *DatasetDefinition) (*DatasetDefinition, error) {
	var definition DatasetDefinition
	err := s.client.performRequest(ctx, "PATCH", fmt.Sprintf("/1/dataset_definitions/%s", urlEncodeDataset(dataset)), data, &definition)
	return &definition, err
}

// Custom Dataset Definitions Logic
var ValidDatasetDefinitions map[string]bool = map[string]bool{
	"duration_ms":     true,
	"error":           true,
	"name":            true,
	"parent_id":       true,
	"route":           true,
	"service_name":    true,
	"span_id":         true,
	"span_kind":       true,
	"annotation_type": true,
	"link_trace_id":   true,
	"link_span_id":    true,
	"status":          true,
	"trace_id":        true,
	"user":            true,
}

func ValidateDatasetDefinition(definition string) bool {
	if ValidDatasetDefinitions[definition] {
		return true
	} else {
		fmt.Printf("definition \"%s\" is not valid.", definition)
		return false
	}
}
