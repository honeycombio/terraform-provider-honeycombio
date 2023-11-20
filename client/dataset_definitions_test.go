package client

import (
	"context"
	"testing"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatasetDefinitions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	c := newTestClient(t)
	dataset := testDataset(t)
	definitionDefaults := DatasetDefinitionDefaults()

	// ensure default definition columns exist -- create any which may be missing.
	// we leave these behind at the end of the test run
	// as they can't be deleted while being used as a definition column
	for _, col := range []Column{
		{KeyName: "duration_ms", Type: ToPtr(ColumnTypeFloat)},
		{KeyName: "error", Type: ToPtr(ColumnTypeBoolean)},
		{KeyName: "name", Type: ToPtr(ColumnTypeString)},
		{KeyName: "trace.parent_id", Type: ToPtr(ColumnTypeString)},
		{KeyName: "http.route", Type: ToPtr(ColumnTypeString)},
		{KeyName: "service.name", Type: ToPtr(ColumnTypeString)},
		{KeyName: "trace.span_id", Type: ToPtr(ColumnTypeString)},
		{KeyName: "meta.span_type", Type: ToPtr(ColumnTypeString)},
		{KeyName: "meta.annotation_type", Type: ToPtr(ColumnTypeString)},
		{KeyName: "http.status_code", Type: ToPtr(ColumnTypeInteger)},
		{KeyName: "trace.trace_id", Type: ToPtr(ColumnTypeString)},
		{KeyName: "request.user.id", Type: ToPtr(ColumnTypeString)},
		{KeyName: "request.user.username", Type: ToPtr(ColumnTypeString)},
		{KeyName: "trace.link.trace_id", Type: ToPtr(ColumnTypeString)},
		{KeyName: "trace.link.span_id", Type: ToPtr(ColumnTypeString)},
	} {
		//nolint:errcheck
		// ignore errors, we don't care if the column already exists
		c.Columns.Create(ctx, dataset, &col)
	}

	// create some new columns to assign as definitions -- we will clean these up at the end of the test run
	testCol, err := c.Columns.Create(ctx, dataset, &Column{KeyName: test.RandomStringWithPrefix("test.", 10)})
	require.NoError(t, err)
	testDC, err := c.DerivedColumns.Create(ctx, dataset, &DerivedColumn{
		Alias:      test.RandomStringWithPrefix("test.", 10),
		Expression: "BOOL(1)",
	})
	require.NoError(t, err)

	// reset all defs and remove test helpers at end of test run
	//nolint:errcheck
	t.Cleanup(func() {
		c.DatasetDefinitions.ResetAll(ctx, dataset)
		c.Columns.Delete(ctx, dataset, testCol.ID)
		c.DerivedColumns.Delete(ctx, dataset, testDC.ID)
	})

	t.Run("Reset and Assert Default state", func(t *testing.T) {
		err := c.DatasetDefinitions.ResetAll(ctx, dataset)
		assert.NoError(t, err)

		result, err := c.DatasetDefinitions.Get(ctx, dataset)
		assert.NoError(t, err)
		assert.Contains(t, definitionDefaults["duration_ms"], result.DurationMs.Name)
		assert.Equal(t, "error", result.Error.Name)
		assert.Equal(t, "name", result.Name.Name)
		assert.Contains(t, definitionDefaults["parent_id"], result.ParentID.Name)
		assert.Contains(t, definitionDefaults["route"], result.Route.Name)
		assert.Contains(t, definitionDefaults["service_name"], result.ServiceName.Name)
		assert.Contains(t, definitionDefaults["span_id"], result.SpanID.Name)
		assert.Contains(t, definitionDefaults["span_kind"], result.SpanKind.Name)
		assert.Contains(t, definitionDefaults["annotation_type"], result.AnnotationType.Name)
		assert.Contains(t, definitionDefaults["link_trace_id"], result.LinkTraceID.Name)
		assert.Contains(t, definitionDefaults["link_span_id"], result.LinkSpanID.Name)
		assert.Contains(t, definitionDefaults["status"], result.Status.Name)
		assert.Contains(t, definitionDefaults["trace_id"], result.TraceID.Name)
		assert.Contains(t, definitionDefaults["user"], result.User.Name)
	})

	t.Run("Update a pair of definitions", func(t *testing.T) {
		_, err := c.DatasetDefinitions.Update(ctx, dataset, &DatasetDefinition{
			Name:  &DefinitionColumn{Name: testCol.KeyName},
			Error: &DefinitionColumn{Name: testDC.Alias},
		})
		assert.NoError(t, err)
		// refetch to be extra sure that our update took effect
		datasetDef, err := c.DatasetDefinitions.Get(ctx, dataset)
		assert.NoError(t, err)
		assert.Equal(t, datasetDef.Name.Name, testCol.KeyName)
		assert.Equal(t, datasetDef.Name.ColumnType, "column")
		assert.Equal(t, datasetDef.Error.Name, testDC.Alias)
		assert.Equal(t, datasetDef.Error.ColumnType, "derived_column")
	})

	t.Run("Reset the fields: ensure reverted to default", func(t *testing.T) {
		result, err := c.DatasetDefinitions.Update(ctx, dataset, &DatasetDefinition{
			Name:  EmptyDatasetDefinition(),
			Error: EmptyDatasetDefinition(),
		})
		assert.NoError(t, err)
		assert.Equal(t, result.Name.Name, "name")
		assert.Equal(t, result.Error.Name, "error")
	})
}
