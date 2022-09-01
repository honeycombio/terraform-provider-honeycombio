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
	//  "trace_id": { "name": "trace.trace_id" }
	definitionValue := "trace.trace_id"

	// Get the name/type of the Dataset Definition
	traceIDValue := DefinitionColumn{
		Name: definitionValue,
	}

	// Create the parent that contains all of the dataset defs (TraceID = trace_id)
	datasetDefinition := DatasetDefinition{
		TraceID: traceIDValue,
	}

	c := newTestClient(t)
	dataset := testDataset(t)

	t.Run("List", func(t *testing.T) {
		result, err := c.DatasetDefinitions.List(ctx, dataset)
		assert.NoError(t, err)
		assert.Equal(t, "column", result.TraceID.ColumnType)
	})

	// set the Trace ID definition
	t.Run("Update", func(t *testing.T) {
		result, err := c.DatasetDefinitions.Update(ctx, dataset, &datasetDefinition)
		assert.NoError(t, err)
		assert.Equal(t, "trace.trace_id", result.TraceID.Name)
		assert.Equal(t, "column", result.TraceID.ColumnType)

		// check Error unset field is still empty
		assert.Equal(t, "", result.Error.Name)
		assert.Equal(t, "", result.Error.ColumnType)

		// reset to empty - currently WIP
		//datasetDefinition.TraceID.Name = ""
		//result, err = c.DatasetDefinitions.Update(ctx, dataset, &datasetDefinition)
		//assert.Equal(t, "", result.TraceID.Name)
		//assert.Equal(t, "column", result.TraceID.ColumnType)
	})
}
