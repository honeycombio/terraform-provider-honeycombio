package honeycombio

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBoards(t *testing.T) {
	ctx := context.Background()

	var b *Board
	var err error

	c := newTestClient(t)
	dataset := testDataset(t)

	t.Run("Create", func(t *testing.T) {
		data := &Board{
			Name:        fmt.Sprintf("Test Board, created at %v", time.Now()),
			Description: "A board with some queries",
			Style:       BoardStyleVisual,
			Queries: []BoardQuery{
				{
					Caption:    "A sample query",
					QueryStyle: BoardQueryStyleCombo,
					Dataset:    dataset,
					Query: &QuerySpec{
						Calculations: []CalculationSpec{
							{
								Op:     CalculationOpAvg,
								Column: StringPtr("duration_ms"),
							},
						},
						TimeRange: IntPtr(3600), // 1 hour
					},
				},
			},
		}
		b, err = c.Boards.Create(ctx, data)

		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(t, b.ID)

		// copy ID before asserting equality
		data.ID = b.ID
		data.Queries[0].QueryID = b.Queries[0].QueryID

		assert.Equal(t, data, b)
	})

	t.Run("List", func(t *testing.T) {
		result, err := c.Boards.List(ctx)

		assert.NoError(t, err)
		assert.Contains(t, result, *b, "could not find newly created board with List")
	})

	t.Run("Get", func(t *testing.T) {
		board, err := c.Boards.Get(ctx, b.ID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, *b, *board)
	})

	t.Run("Update", func(t *testing.T) {
		b.Queries[0].QueryID = ""
		b.Queries = append(b.Queries, BoardQuery{
			Caption:    "A second query",
			QueryStyle: BoardQueryStyleGraph,
			Dataset:    dataset,
			Query: &QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op: CalculationOpCount,
					},
				},
				TimeRange: IntPtr(7200), // 2 hours
			},
		})

		result, err := c.Boards.Update(ctx, b)
		b.Queries[0].QueryID = result.Queries[0].QueryID
		b.Queries[1].QueryID = result.Queries[1].QueryID

		assert.NoError(t, err)
		assert.Equal(t, b, result)
	})

	t.Run("Delete", func(t *testing.T) {
		err := c.Boards.Delete(ctx, b.ID)

		assert.NoError(t, err)
	})

	t.Run("Get_deletedBoard", func(t *testing.T) {
		_, err := c.Boards.Get(ctx, b.ID)

		assert.Equal(t, ErrNotFound, err)
	})
}
