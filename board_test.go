package honeycombio

import (
	"context"
	"testing"

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
			Name:        "Test Board",
			Description: "A board with some queries",
			Style:       BoardStyleVisual,
			Queries: []BoardQuery{
				{
					Caption: "A sample query",
					Dataset: dataset,
					Query: QuerySpec{
						Calculations: []CalculationSpec{
							{
								Op:     CalculateOpAvg,
								Column: &[]string{"duration_ms"}[0],
							},
						},
					},
				},
			},
		}
		b, err = c.Boards.Create(ctx, data)
		if err != nil {
			t.Fatal(err)
		}

		assert.NotNil(t, b.ID)
		assert.Equal(t, data.Name, b.Name)
		assert.Equal(t, data.Description, b.Description)
		assert.Equal(t, data.Style, b.Style)
		assert.Equal(t, data.Queries, b.Queries)
	})

	t.Run("List", func(t *testing.T) {
		boards, err := c.Boards.List(ctx)
		if err != nil {
			t.Fatal(err)
		}

		var createdBoard *Board

		for _, board := range boards {
			if board.ID == b.ID {
				createdBoard = &board
				break
			}
		}
		if createdBoard == nil {
			t.Fatalf("could not find newly created board with ID = %s", b.ID)
		}

		assert.Equal(t, *b, *createdBoard)
	})

	t.Run("Update", func(t *testing.T) {
		newBoard := *b
		newBoard.Queries = append(newBoard.Queries, BoardQuery{
			Caption: "A second query",
			Dataset: dataset,
			Query: QuerySpec{
				Calculations: []CalculationSpec{
					{
						Op: CalculateOpCount,
					},
				},
			},
		})

		updatedBoard, err := c.Boards.Update(ctx, &newBoard)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, newBoard, *updatedBoard)

		b = updatedBoard
	})

	t.Run("Get", func(t *testing.T) {
		board, err := c.Boards.Get(ctx, b.ID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, *b, *board)
	})

	t.Run("Delete", func(t *testing.T) {
		err := c.Boards.Delete(ctx, b.ID)
		assert.NoError(t, err)
	})

	t.Run("Get_unexistingID", func(t *testing.T) {
		_, err := c.Boards.Get(ctx, b.ID)
		assert.Equal(t, ErrNotFound, err)
	})
}
