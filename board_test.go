package honeycombio

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBoards(t *testing.T) {
	var b *Board
	var err error

	c := newTestClient(t)

	t.Run("Create", func(t *testing.T) {

		data := &Board{
			Name:        "Test Board",
			Description: "A board with some queries",
			Style:       BoardStyleVisual,
			Queries: []BoardQuery{
				{
					Caption: "A sample dataset",
					Dataset: c.dataset,
					Query:   QuerySpec{},
				},
			},
		}
		b, err = c.Boards.Create(data)
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
		boards, err := c.Boards.List()
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

	t.Run("Get", func(t *testing.T) {
		board, err := c.Boards.Get(b.ID)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, *b, *board)
	})

	t.Run("Get_unexistingID", func(t *testing.T) {
		_, err := c.Boards.Get("0")
		assert.Equal(t, ErrNotFound, err)
	})
}
