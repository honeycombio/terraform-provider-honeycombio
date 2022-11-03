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
	// Same behavior as Update - needs to be invoked once to specifiy resource ID
	Create(ctx context.Context, dataset string, data *DatasetDefinition) (*DatasetDefinition, error)

	// Get All Dataset Definitions for a Dataset
	Get(ctx context.Context, dataset string) (*DatasetDefinition, error)

	// Get All Dataset Definitions
	Update(ctx context.Context, dataset string, data *DatasetDefinition) (*DatasetDefinition, error)

	// Get All Dataset Definitions
	Delete(ctx context.Context, dataset string) error
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
	ColumnType string `json:"column_type,omitempty"` //this is regular vs. derived column instead of data type -> IsDerivedColumn boolean
	//Type as datatype? -> since we don't want to set an integer column field to a string definiton?
	//IsHidden? ->  since we don't want to allow an integer column field to a string definiton?
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

func (s *datasetDefinitions) Create(ctx context.Context, dataset string, data *DatasetDefinition) (*DatasetDefinition, error) {
	var definition DatasetDefinition
	err := s.client.performRequest(ctx, "PATCH", fmt.Sprintf("/1/dataset_definitions/%s", urlEncodeDataset(dataset)), data, &definition)
	return &definition, err
}

func (s *datasetDefinitions) Get(ctx context.Context, dataset string) (*DatasetDefinition, error) {
	var definition DatasetDefinition
	err := s.client.performRequest(ctx, "GET", fmt.Sprintf("/1/dataset_definitions/%s", urlEncodeDataset(dataset)), nil, &definition)
	return &definition, err
}

func (s *datasetDefinitions) Update(ctx context.Context, dataset string, data *DatasetDefinition) (*DatasetDefinition, error) {
	var definition DatasetDefinition
	err := s.client.performRequest(ctx, "PATCH", fmt.Sprintf("/1/dataset_definitions/%s", urlEncodeDataset(dataset)), data, &definition)
	return &definition, err
}

func (s *datasetDefinitions) Delete(ctx context.Context, dataset string) error {
	// "Deleting dataset definitions" means set all values to ""
	definition := DatasetDefinition{
		DurationMs:     DefinitionColumn{Name: ""},
		Error:          DefinitionColumn{Name: ""},
		Name:           DefinitionColumn{Name: ""},
		ParentID:       DefinitionColumn{Name: ""},
		Route:          DefinitionColumn{Name: ""},
		ServiceName:    DefinitionColumn{Name: ""},
		SpanID:         DefinitionColumn{Name: ""},
		SpanType:       DefinitionColumn{Name: ""},
		AnnotationType: DefinitionColumn{Name: ""},
		LinkTraceID:    DefinitionColumn{Name: ""},
		LinkSpanID:     DefinitionColumn{Name: ""},
		Status:         DefinitionColumn{Name: ""},
		TraceID:        DefinitionColumn{Name: ""},
		User:           DefinitionColumn{Name: ""},
	}

	var dd DatasetDefinition
	err := s.client.performRequest(ctx, "PATCH", fmt.Sprintf("/1/dataset_definitions/%s", urlEncodeDataset(dataset)), definition, &dd)

	return err
}
