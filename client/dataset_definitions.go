package client

import (
	"context"
	"fmt"
)

// Dataset Definitions define the fields in your Dataset that have special meaning.
//
// API docs: https://docs.honeycomb.io/api/dataset-definitions/
type DatasetDefinitions interface {
	// Get the Dataset Definitions for a Dataset
	Get(ctx context.Context, dataset string) (*DatasetDefinition, error)

	// Resets the Dataset Definitions for a Dataset to the default state
	ResetAll(ctx context.Context, dataset string) error

	// Update the Dataset Definitions in a Dataset
	Update(ctx context.Context, dataset string, d *DatasetDefinition) (*DatasetDefinition, error)
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
	DurationMs     *DefinitionColumn `json:"duration_ms,omitempty"`
	Error          *DefinitionColumn `json:"error,omitempty"`
	Name           *DefinitionColumn `json:"name,omitempty"`
	ParentID       *DefinitionColumn `json:"parent_id,omitempty"`
	Route          *DefinitionColumn `json:"route,omitempty"`
	ServiceName    *DefinitionColumn `json:"service_name,omitempty"`
	SpanID         *DefinitionColumn `json:"span_id,omitempty"`
	SpanKind       *DefinitionColumn `json:"span_kind,omitempty"`
	AnnotationType *DefinitionColumn `json:"annotation_type,omitempty"`
	LinkTraceID    *DefinitionColumn `json:"link_trace_id,omitempty"`
	LinkSpanID     *DefinitionColumn `json:"link_span_id,omitempty"`
	Status         *DefinitionColumn `json:"status,omitempty"`
	TraceID        *DefinitionColumn `json:"trace_id,omitempty"`
	User           *DefinitionColumn `json:"user,omitempty"`
}

// Resetting or Unsetting a Dataset Definition is done by setting the Name
// to the empty string.
func EmptyDatasetDefinition() *DefinitionColumn {
	return &DefinitionColumn{Name: ""}
}

// The names of all possible Dataset Definitions which can be set
func DatasetDefinitionFields() []string {
	return []string{
		"duration_ms",
		"error",
		"name",
		"parent_id",
		"route",
		"service_name",
		"span_id",
		"span_kind",
		"annotation_type",
		"link_trace_id",
		"link_span_id",
		"status",
		"trace_id",
		"user",
	}
}

// A mapping of Dataset Definition names to their possible default values (excluding 'nil')
func DatasetDefinitionDefaults() map[string][]string {
	return map[string][]string{
		"duration_ms":     {"duration_ms", "durationMs", "request_processing_time"},
		"error":           {"error"},
		"name":            {"name"},
		"parent_id":       {"trace.parent_id", "parentId"},
		"route":           {"route", "http.route", "request_path"},
		"service_name":    {"service_name", "service.name", "serviceName"},
		"span_id":         {"id", "trace.span_id"},
		"span_kind":       {"meta.span_type"},
		"annotation_type": {"meta.annotation_type"},
		"link_trace_id":   {"trace.link.trace_id"},
		"link_span_id":    {"trace.link.span_id", "trace.span_id"},
		"status":          {"response.status_code", "http.status_code", "elb_status_code"},
		"trace_id":        {"http.status_code", "trace.trace_id", "traceId"},
		"user":            {"user.id", "user.email", "request.user.id", "request.user.username"},
	}
}

func (s *datasetDefinitions) Get(ctx context.Context, dataset string) (*DatasetDefinition, error) {
	var result DatasetDefinition
	err := s.client.performRequest(ctx, "GET", fmt.Sprintf("/1/dataset_definitions/%s", urlEncodeDataset(dataset)), nil, &result)
	return &result, err
}

func (s *datasetDefinitions) ResetAll(ctx context.Context, dataset string) error {
	emptyDefinitions := DatasetDefinition{
		DurationMs:     EmptyDatasetDefinition(),
		Error:          EmptyDatasetDefinition(),
		Name:           EmptyDatasetDefinition(),
		ParentID:       EmptyDatasetDefinition(),
		Route:          EmptyDatasetDefinition(),
		ServiceName:    EmptyDatasetDefinition(),
		SpanID:         EmptyDatasetDefinition(),
		SpanKind:       EmptyDatasetDefinition(),
		AnnotationType: EmptyDatasetDefinition(),
		LinkTraceID:    EmptyDatasetDefinition(),
		LinkSpanID:     EmptyDatasetDefinition(),
		Status:         EmptyDatasetDefinition(),
		TraceID:        EmptyDatasetDefinition(),
		User:           EmptyDatasetDefinition(),
	}

	return s.client.performRequest(ctx, "PATCH", fmt.Sprintf("/1/dataset_definitions/%s", urlEncodeDataset(dataset)), emptyDefinitions, nil)
}

func (s *datasetDefinitions) Update(ctx context.Context, dataset string, d *DatasetDefinition) (*DatasetDefinition, error) {
	var result DatasetDefinition
	err := s.client.performRequest(ctx, "PATCH", fmt.Sprintf("/1/dataset_definitions/%s", urlEncodeDataset(dataset)), d, &result)
	return &result, err
}
