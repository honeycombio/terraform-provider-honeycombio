// Package fixture provides Honeycomb resources for acceptance tests which
// need state of their own rather than the shared HONEYCOMB_DATASET.
package fixture

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

// NewDataset creates a throwaway dataset for tests which can't share
// HONEYCOMB_DATASET with others: those mutating per-dataset state, such as
// dataset definitions, and those asserting on everything in a dataset,
// such as unscoped list filters.
//
// Any provided columns are created, and NewDataset waits for them to become
// visible before returning, as schema changes propagate asynchronously.
// The dataset, and everything in it, is deleted when the test completes.
func NewDataset(ctx context.Context, t *testing.T, c *client.Client, columns ...client.Column) string {
	t.Helper()

	// named with the sweeper's prefix so it is cleaned up if a run is interrupted
	ds, err := c.Datasets.Create(ctx, &client.Dataset{
		Name: test.RandomStringWithPrefix("test.", 10),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := c.Datasets.Update(ctx, &client.Dataset{
			Slug:     ds.Slug,
			Settings: client.DatasetSettings{DeleteProtected: new(false)},
		})
		if err != nil {
			t.Errorf("could not disable deletion protection on dataset %q: %v", ds.Slug, err)
			return
		}
		if err := c.Datasets.Delete(ctx, ds.Slug); err != nil {
			t.Errorf("could not delete dataset %q: %v", ds.Slug, err)
		}
	})

	createColumns(ctx, t, c, ds.Slug, columns)

	return ds.Slug
}

// DefinitionDefaultColumns returns a column for every dataset definition's
// default, so resetting definitions on a new dataset has columns to fall back to.
func DefinitionDefaultColumns() []client.Column {
	return []client.Column{
		{KeyName: "duration_ms", Type: new(client.ColumnTypeFloat)},
		{KeyName: "error", Type: new(client.ColumnTypeBoolean)},
		{KeyName: "name", Type: new(client.ColumnTypeString)},
		{KeyName: "trace.parent_id", Type: new(client.ColumnTypeString)},
		{KeyName: "http.route", Type: new(client.ColumnTypeString)},
		{KeyName: "service.name", Type: new(client.ColumnTypeString)},
		{KeyName: "trace.span_id", Type: new(client.ColumnTypeString)},
		{KeyName: "meta.span_type", Type: new(client.ColumnTypeString)},
		{KeyName: "meta.annotation_type", Type: new(client.ColumnTypeString)},
		{KeyName: "http.status_code", Type: new(client.ColumnTypeInteger)},
		{KeyName: "trace.trace_id", Type: new(client.ColumnTypeString)},
		{KeyName: "request.user.id", Type: new(client.ColumnTypeString)},
		{KeyName: "request.user.username", Type: new(client.ColumnTypeString)},
		{KeyName: "trace.link.trace_id", Type: new(client.ColumnTypeString)},
		{KeyName: "trace.link.span_id", Type: new(client.ColumnTypeString)},
		{KeyName: "body", Type: new(client.ColumnTypeString)},
		{KeyName: "severity", Type: new(client.ColumnTypeString)},
	}
}

// createColumns creates the columns concurrently, then waits until all of
// them are visible in the dataset.
func createColumns(ctx context.Context, t *testing.T, c *client.Client, dataset string, columns []client.Column) {
	t.Helper()

	if len(columns) == 0 {
		return
	}

	var g errgroup.Group
	g.SetLimit(5)
	for _, col := range columns {
		g.Go(func() error {
			_, err := c.Columns.Create(ctx, dataset, &col)
			return err
		})
	}
	require.NoError(t, g.Wait())

	require.Eventually(t, func() bool {
		existing, err := c.Columns.List(ctx, dataset)
		if err != nil {
			return false
		}
		visible := make(map[string]bool, len(existing))
		for _, col := range existing {
			visible[col.KeyName] = true
		}
		for _, col := range columns {
			if !visible[col.KeyName] {
				return false
			}
		}
		return true
	}, 60*time.Second, 500*time.Millisecond, "columns never became visible in dataset %q", dataset)
}
