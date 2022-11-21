package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatasetDefinitions(t *testing.T) {
	ctx := context.Background()

	c := newTestClient(t)
	dataset := testDataset(t)

	col, err := c.Columns.Create(ctx, dataset, &Column{KeyName: "happy.duck"})
	assert.NoError(t, err)
	dc, err := c.DerivedColumns.Create(ctx, dataset, &DerivedColumn{
		Alias:      "happy.error",
		Expression: `NOT(CONTAINS($app.error, "fatal"))`,
	})
	assert.NoError(t, err)
	// reset all defs and remove test helpers at end of test run
	t.Cleanup(func() {
		c.DatasetDefinitions.ResetAll(ctx, dataset)
		c.Columns.Delete(ctx, dataset, col.ID)
		c.DerivedColumns.Delete(ctx, dataset, dc.ID)
	})

	t.Run("Reset and Assert Default state", func(t *testing.T) {
		err := c.DatasetDefinitions.ResetAll(ctx, dataset)
		assert.NoError(t, err)

		result, err := c.DatasetDefinitions.Get(ctx, dataset)
		assert.NoError(t, err)
		assert.Equal(t, "duration_ms", result.DurationMs.Name)
		assert.Nil(t, result.Error)
		assert.Equal(t, "name", result.Name.Name)
		assert.Equal(t, "trace.parent_id", result.ParentID.Name)
		assert.Nil(t, result.Route)
		assert.Equal(t, "service_name", result.ServiceName.Name)
		assert.Equal(t, "trace.span_id", result.SpanID.Name)
		assert.Nil(t, result.SpanKind)
		assert.Nil(t, result.AnnotationType)
		assert.Nil(t, result.LinkTraceID)
		assert.Nil(t, result.LinkSpanID)
		assert.Nil(t, result.Status)
		assert.Equal(t, "trace.trace_id", result.TraceID.Name)
		assert.Nil(t, result.User)
	})

	t.Run("Update a pair of definitions", func(t *testing.T) {
		_, err := c.DatasetDefinitions.Update(ctx, dataset, &DatasetDefinition{
			Name:  &DefinitionColumn{Name: col.KeyName},
			Error: &DefinitionColumn{Name: dc.Alias},
		})
		assert.NoError(t, err)
		// refetch to be extra sure that our update took effect
		datasetDef, err := c.DatasetDefinitions.Get(ctx, dataset)
		assert.NoError(t, err)
		assert.Equal(t, datasetDef.Name.Name, col.KeyName)
		assert.Equal(t, datasetDef.Name.ColumnType, "column")
		assert.Equal(t, datasetDef.Error.Name, dc.Alias)
		assert.Equal(t, datasetDef.Error.ColumnType, "derived_column")
		// spot check a few of the original fields from above to ensure they were unchanged
		assert.Nil(t, datasetDef.Route)
		assert.Equal(t, "service_name", datasetDef.ServiceName.Name)
		assert.Equal(t, "trace.span_id", datasetDef.SpanID.Name)
	})

	t.Run("Reset two fields: one with a default", func(t *testing.T) {
		result, err := c.DatasetDefinitions.Update(ctx, dataset, &DatasetDefinition{
			Name:  EmptyDatasetDefinition(),
			Error: EmptyDatasetDefinition(),
		})
		assert.NoError(t, err)
		assert.Equal(t, result.Name.Name, "name")
		assert.Nil(t, result.Error)
	})
}
