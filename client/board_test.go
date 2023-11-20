package client

import (
	"context"
	"testing"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoards(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var b *Board
	var err error

	c := newTestClient(t)
	dataset := testDataset(t)

	column, err := c.Columns.Create(ctx, dataset, &Column{
		KeyName: test.RandomStringWithPrefix("test.", 8),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		//nolint:errcheck
		c.Columns.Delete(ctx, dataset, column.ID)
	})

	query, err := c.Queries.Create(ctx, dataset, &QuerySpec{
		Calculations: []CalculationSpec{
			{
				Op:     CalculationOpAvg,
				Column: &column.KeyName,
			},
		},
		TimeRange: ToPtr(3600), // 1 hour
	})
	require.NoError(t, err)

	t.Run("Create", func(t *testing.T) {
		data := &Board{
			Name:         test.RandomStringWithPrefix("test.", 8),
			Description:  "A board with some queries",
			Style:        BoardStyleVisual,
			ColumnLayout: BoardColumnStyleSingle,
			Queries: []BoardQuery{
				{
					Caption:       "A sample query",
					QueryStyle:    BoardQueryStyleCombo,
					Dataset:       dataset,
					QueryID:       *query.ID,
					GraphSettings: BoardGraphSettings{OmitMissingValues: true, UseUTCXAxis: true},
				},
			},
		}
		b, err = c.Boards.Create(ctx, data)

		assert.NoError(t, err)
		assert.NotNil(t, b.ID)

		// copy ID before asserting equality
		data.ID = b.ID
		data.Queries[0].QueryID = b.Queries[0].QueryID

		// ensure the board URL got populated
		assert.NotEqual(t, b.Links.BoardURL, "")
		data.Links.BoardURL = b.Links.BoardURL

		assert.Equal(t, data, b)
	})

	t.Run("List", func(t *testing.T) {
		result, err := c.Boards.List(ctx)

		assert.NoError(t, err)
		assert.Contains(t, result, *b, "could not find newly created board with List")
	})

	t.Run("Get", func(t *testing.T) {
		board, err := c.Boards.Get(ctx, b.ID)
		assert.NoError(t, err)

		assert.Equal(t, *b, *board)
	})

	t.Run("Update", func(t *testing.T) {
		newQuery, err := c.Queries.Create(ctx, dataset, &QuerySpec{
			Calculations: []CalculationSpec{
				{
					Op: CalculationOpCount,
				},
			},
			TimeRange: ToPtr(DefaultQueryTimeRange),
		})
		assert.NoError(t, err)
		b.ColumnLayout = BoardColumnStyleMulti
		b.Queries = append(b.Queries, BoardQuery{
			Caption:       "A second query",
			QueryStyle:    BoardQueryStyleGraph,
			QueryID:       *newQuery.ID,
			GraphSettings: BoardGraphSettings{UseUTCXAxis: true},
		})

		result, err := c.Boards.Update(ctx, b)
		assert.NoError(t, err)
		assert.Equal(t, b, result)
	})

	t.Run("Delete", func(t *testing.T) {
		err := c.Boards.Delete(ctx, b.ID)

		assert.NoError(t, err)
	})

	t.Run("Fail to get deleted Board", func(t *testing.T) {
		_, err := c.Boards.Get(ctx, b.ID)

		var de DetailedError
		assert.Error(t, err)
		assert.ErrorAs(t, err, &de)
		assert.True(t, de.IsNotFound())
	})
}
