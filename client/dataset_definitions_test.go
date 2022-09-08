package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// update, validate and clear a dataset definition
func TestDatasetDefinitions(t *testing.T) {
	ctx := context.Background()

	// Definition to create
	//  "trace_id": { "name": "trace.trace_id" }, "duration_ms": { "name": "request_processing_time"}

	traceIDValue := "trace.trace_id"
	// Exisiting Derived Column: GTE(INT($duration_ms), 50)
	durationMsValue := "gt50_duration_ms"

	traceIDDefinition := DefinitionColumn{
		Name: traceIDValue,
	}

	durationMsDefinition := DefinitionColumn{
		Name: durationMsValue,
	}

	// Create the parent that contains all of the dataset defs (TraceID = trace_id)
	datasetDefinition := DatasetDefinition{
		DurationMs: durationMsDefinition,
		TraceID:    traceIDDefinition,
	}

	c := newTestClient(t)
	dataset := testDataset(t)

	t.Run("Get", func(t *testing.T) {
		result, err := c.DatasetDefinitions.Get(ctx, dataset)
		assert.NoError(t, err)
		assert.Equal(t, "column", result.TraceID.ColumnType)
	})

	// set the Trace ID definition
	t.Run("Update", func(t *testing.T) {
		result, err := c.DatasetDefinitions.Update(ctx, dataset, &datasetDefinition)
		assert.NoError(t, err)

		// check duration ms derived column
		assert.Equal(t, "gt50_duration_ms", result.DurationMs.Name)
		assert.Equal(t, "derived_column", result.DurationMs.ColumnType)

		// check trace ID
		assert.Equal(t, "trace.trace_id", result.TraceID.Name)
		assert.Equal(t, "column", result.TraceID.ColumnType)

		// check Error unset field is still empty
		assert.Equal(t, "", result.Error.Name)
		assert.Equal(t, "", result.Error.ColumnType)
	})

	// remove trace ID
	t.Run("Delete", func(t *testing.T) {
		err := c.DatasetDefinitions.Delete(ctx, dataset)
		assert.NoError(t, err)
	})
}
