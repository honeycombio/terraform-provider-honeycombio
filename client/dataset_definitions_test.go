package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// update, validate and clear a dataset definition
func TestDatasetDefinitions(t *testing.T) {
	ctx := context.Background()

	var datasetDef *DatasetDefinition
	var err error

	c := newTestClient(t)
	dataset := testDataset(t)

	dcs, err := c.DerivedColumns.Create(ctx, dataset, &DerivedColumn{
		Alias:      "test123",
		Expression: "GTE(INT($duration_ms), 50)",
	})
	if err != nil {
		t.Fatal(err)
	}

	// remove dataset definition DC at end of test run
	t.Cleanup(func() {
		c.DerivedColumns.Delete(ctx, dataset, dcs.ID)
	})

	t.Run("Create", func(t *testing.T) {
		// set the Trace ID definition
		traceIDDefinition := DefinitionColumn{
			Name:       "trace.trace_id",
			ColumnType: "column",
		}

		//set the durationMs definition
		durationMsDerivedDefinition := DefinitionColumn{
			Name:       dcs.Alias,
			ColumnType: "derived_column",
		}

		data := &DatasetDefinition{
			TraceID:    traceIDDefinition,
			DurationMs: durationMsDerivedDefinition,
		}

		datasetDef, err := c.DatasetDefinitions.Create(ctx, dataset, data)
		assert.NoError(t, err)
		assert.Equal(t, "column", datasetDef.TraceID.ColumnType)
		assert.Equal(t, "derived_column", datasetDef.DurationMs.ColumnType)
	})

	t.Run("Get", func(t *testing.T) {
		result, err := c.DatasetDefinitions.Get(ctx, dataset)
		assert.NoError(t, err)
		assert.Equal(t, "column", result.TraceID.ColumnType)
		assert.Equal(t, "derived_column", result.DurationMs.ColumnType)
	})

	t.Run("Update", func(t *testing.T) {

		data := &Column{
			KeyName:     "error_test1",
			Hidden:      BoolPtr(false),
			Description: "This column is created by a test",
			Type:        ColumnTypePtr(ColumnTypeFloat),
		}

		_, err = c.Columns.Create(ctx, dataset, data)

		errorDefinition := DefinitionColumn{
			Name:       data.KeyName,
			ColumnType: "column",
		}

		//Josslyn - set field name & columnType to empty should work, currently doesn't
		// durationMsDefinition := DefinitionColumn{
		// 	Name:       "",
		// 	ColumnType: "",
		// }

		datasetDef = &DatasetDefinition{
			Error: errorDefinition,
			// DurationMs: durationMsDefinition,
		}

		result, err := c.DatasetDefinitions.Update(ctx, dataset, datasetDef)
		assert.NoError(t, err)

		// check Error field set to new column created
		assert.Equal(t, "error_test1", result.Error.Name)
		assert.Equal(t, "column", result.Error.ColumnType)

		// check trace ID unchanged
		assert.Equal(t, "trace.trace_id", result.TraceID.Name)
		assert.Equal(t, "column", result.TraceID.ColumnType)

		// check duration_ms update to null
		//assert.Equal(t, "", result.DurationMs.Name)
		// assert.Equal(t, "", result.DurationMs.ColumnType)
	})

	//Provider cleanup
	t.Run("Delete", func(t *testing.T) {
		err := c.DatasetDefinitions.Delete(ctx, dataset)
		assert.NoError(t, err)
	})
}
