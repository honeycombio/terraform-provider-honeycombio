package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestBoards(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var b *client.Board
	var err error

	c := newTestClient(t)
	dataset := testDataset(t)

	column, err := c.Columns.Create(ctx, dataset, &client.Column{
		KeyName: test.RandomStringWithPrefix("test.", 8),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		c.Columns.Delete(ctx, dataset, column.ID)
	})

	query, err := c.Queries.Create(ctx, dataset, &client.QuerySpec{
		Calculations: []client.CalculationSpec{
			{
				Op:     client.CalculationOpAvg,
				Column: &column.KeyName,
			},
		},
		TimeRange: client.ToPtr(3600), // 1 hour
	})
	require.NoError(t, err)

	sli, err := c.DerivedColumns.Create(ctx, dataset, &client.DerivedColumn{
		Alias:      test.RandomStringWithPrefix("test.", 8),
		Expression: "BOOL(1)",
	})
	require.NoError(t, err)
	slo, err := c.SLOs.Create(ctx, dataset, &client.SLO{
		Name:             test.RandomStringWithPrefix("test.", 8),
		TimePeriodDays:   7,
		TargetPerMillion: 990000,
		SLI:              client.SLIRef{Alias: sli.Alias},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		c.SLOs.Delete(ctx, dataset, slo.ID)
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)
	})

	t.Run("Create", func(t *testing.T) {
		data := &client.Board{
			Name:         test.RandomStringWithPrefix("test.", 8),
			Description:  "A board with some queries",
			Style:        client.BoardStyleVisual,
			ColumnLayout: client.BoardColumnStyleSingle,
			Queries: []client.BoardQuery{
				{
					Caption:       "A sample query",
					QueryStyle:    client.BoardQueryStyleCombo,
					Dataset:       dataset,
					QueryID:       *query.ID,
					GraphSettings: client.BoardGraphSettings{OmitMissingValues: true, UseUTCXAxis: true},
				},
			},
			SLOs: []string{slo.ID},
		}
		b, err = c.Boards.Create(ctx, data)
		require.NoError(t, err)
		assert.NotNil(t, b.ID)

		// copy ID before asserting equality
		data.ID = b.ID
		data.Queries[0].QueryID = b.Queries[0].QueryID

		// ensure the board URL got populated
		assert.NotEqual(t, "", b.Links.BoardURL)
		data.Links.BoardURL = b.Links.BoardURL

		assert.Equal(t, data, b)
	})

	t.Run("List", func(t *testing.T) {
		result, err := c.Boards.List(ctx)
		require.NoError(t, err)

		assert.Contains(t, result, *b, "could not find newly created board with List")
	})

	t.Run("Get", func(t *testing.T) {
		board, err := c.Boards.Get(ctx, b.ID)
		require.NoError(t, err)

		assert.Equal(t, *b, *board)
	})

	t.Run("Update", func(t *testing.T) {
		newQuery, err := c.Queries.Create(ctx, dataset, &client.QuerySpec{
			Calculations: []client.CalculationSpec{
				{
					Op: client.CalculationOpCount,
				},
			},
			TimeRange: client.ToPtr(client.DefaultQueryTimeRange),
		})
		require.NoError(t, err)
		b.ColumnLayout = client.BoardColumnStyleMulti
		b.Queries = append(b.Queries, client.BoardQuery{
			Caption:       "A second query",
			QueryStyle:    client.BoardQueryStyleGraph,
			QueryID:       *newQuery.ID,
			GraphSettings: client.BoardGraphSettings{UseUTCXAxis: true},
		})
		b.SLOs = []string{}

		result, err := c.Boards.Update(ctx, b)
		require.NoError(t, err)
		assert.Equal(t, b, result)
	})

	t.Run("Delete", func(t *testing.T) {
		err := c.Boards.Delete(ctx, b.ID)

		require.NoError(t, err)
	})

	t.Run("Fail to get deleted Board", func(t *testing.T) {
		_, err := c.Boards.Get(ctx, b.ID)

		var de client.DetailedError
		require.Error(t, err)
		require.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})
}
