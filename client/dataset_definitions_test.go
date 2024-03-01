package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestDatasetDefinitions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	c := newTestClient(t)
	dataset := testDataset(t)
	definitionDefaults := client.DatasetDefinitionDefaults()

	// ensure default definition columns exist -- create any which may be missing.
	// we leave these behind at the end of the test run
	// as they can't be deleted while being used as a definition column
	for _, col := range []client.Column{
		{KeyName: "duration_ms", Type: client.ToPtr(client.ColumnTypeFloat)},
		{KeyName: "error", Type: client.ToPtr(client.ColumnTypeBoolean)},
		{KeyName: "name", Type: client.ToPtr(client.ColumnTypeString)},
		{KeyName: "trace.parent_id", Type: client.ToPtr(client.ColumnTypeString)},
		{KeyName: "http.route", Type: client.ToPtr(client.ColumnTypeString)},
		{KeyName: "service.name", Type: client.ToPtr(client.ColumnTypeString)},
		{KeyName: "trace.span_id", Type: client.ToPtr(client.ColumnTypeString)},
		{KeyName: "meta.span_type", Type: client.ToPtr(client.ColumnTypeString)},
		{KeyName: "meta.annotation_type", Type: client.ToPtr(client.ColumnTypeString)},
		{KeyName: "http.status_code", Type: client.ToPtr(client.ColumnTypeInteger)},
		{KeyName: "trace.trace_id", Type: client.ToPtr(client.ColumnTypeString)},
		{KeyName: "request.user.id", Type: client.ToPtr(client.ColumnTypeString)},
		{KeyName: "request.user.username", Type: client.ToPtr(client.ColumnTypeString)},
		{KeyName: "trace.link.trace_id", Type: client.ToPtr(client.ColumnTypeString)},
		{KeyName: "trace.link.span_id", Type: client.ToPtr(client.ColumnTypeString)},
	} {
		//nolint:errcheck
		// ignore errors, we don't care if the column already exists
		c.Columns.Create(ctx, dataset, &col)
	}

	// create some new columns to assign as definitions -- we will clean these up at the end of the test run
	testCol, err := c.Columns.Create(ctx, dataset, &client.Column{KeyName: test.RandomStringWithPrefix("test.", 10)})
	require.NoError(t, err)
	testDC, err := c.DerivedColumns.Create(ctx, dataset, &client.DerivedColumn{
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
		require.NoError(t, err)

		result, err := c.DatasetDefinitions.Get(ctx, dataset)
		require.NoError(t, err)
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
		_, err := c.DatasetDefinitions.Update(ctx, dataset, &client.DatasetDefinition{
			Name:  &client.DefinitionColumn{Name: testCol.KeyName},
			Error: &client.DefinitionColumn{Name: testDC.Alias},
		})
		require.NoError(t, err)
		// refetch to be extra sure that our update took effect
		datasetDef, err := c.DatasetDefinitions.Get(ctx, dataset)
		require.NoError(t, err)
		assert.Equal(t, datasetDef.Name.Name, testCol.KeyName)
		assert.Equal(t, "column", datasetDef.Name.ColumnType)
		assert.Equal(t, datasetDef.Error.Name, testDC.Alias)
		assert.Equal(t, "derived_column", datasetDef.Error.ColumnType)
	})

	t.Run("Reset the fields: ensure reverted to default", func(t *testing.T) {
		result, err := c.DatasetDefinitions.Update(ctx, dataset, &client.DatasetDefinition{
			Name:  client.EmptyDatasetDefinition(),
			Error: client.EmptyDatasetDefinition(),
		})
		require.NoError(t, err)
		assert.Equal(t, "name", result.Name.Name)
		assert.Equal(t, "error", result.Error.Name)
	})
}
