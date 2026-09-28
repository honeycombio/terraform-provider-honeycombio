package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test/fixture"
)

func TestDatasetDefinitions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	c := newTestClient(t)
	// definitions are per-dataset state, so don't clobber the shared dataset's
	dataset := fixture.NewDataset(ctx, t, c, fixture.DefinitionDefaultColumns()...)
	definitionDefaults := client.DatasetDefinitionDefaults()

	// create some new columns to assign as definitions -- these are removed with the dataset
	testCol, err := c.Columns.Create(ctx, dataset, &client.Column{KeyName: test.RandomStringWithPrefix("test.", 10)})
	require.NoError(t, err)
	testDC, err := c.DerivedColumns.Create(ctx, dataset, &client.DerivedColumn{
		Alias:      test.RandomStringWithPrefix("test.", 10),
		Expression: "BOOL(1)",
	})
	require.NoError(t, err)

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
		assert.Contains(t, definitionDefaults["log_message"], result.LogMessage.Name)
		assert.Contains(t, definitionDefaults["log_severity"], result.LogSeverity.Name)
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
