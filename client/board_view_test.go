package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestBoardViews(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var boardView *client.BoardView
	var err error

	c := newTestClient(t)
	dataset := testDataset(t)

	// Create a board first since views belong to boards
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
	require.NotEmpty(t, query.ID)

	qa := &client.QueryAnnotation{
		Name:        test.RandomStringWithPrefix("test.", 20),
		Description: "This query annotation is created by a test",
		QueryID:     *query.ID,
	}
	queryAnnotation, err := c.QueryAnnotations.Create(ctx, dataset, qa)
	require.NoError(t, err)
	require.NotNil(t, queryAnnotation.ID)

	board, err := c.Boards.Create(ctx, &client.Board{
		Name:      test.RandomStringWithPrefix("test.", 8),
		BoardType: "flexible",
		Panels: []client.BoardPanel{
			{
				PanelType: client.BoardPanelTypeQuery,
				QueryPanel: &client.BoardQueryPanel{
					QueryID:           *query.ID,
					QueryAnnotationID: queryAnnotation.ID,
					Style:             client.BoardQueryStyleGraph,
				},
			},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, board.ID)
	t.Cleanup(func() {
		c.Boards.Delete(ctx, board.ID)
	})

	t.Run("Create board view with unary filter", func(t *testing.T) {
		data := &client.BoardView{
			Name: test.RandomStringWithPrefix("test.", 8),
			Filters: []client.BoardViewFilter{
				{
					Column:    column.KeyName,
					Operation: string(client.FilterOpExists),
					// Value should be nil for unary operations
				},
			},
		}
		boardView, err = c.BoardViews.Create(ctx, board.ID, data)
		require.NoError(t, err)
		assert.NotNil(t, boardView.ID)
		assert.Equal(t, data.Name, boardView.Name)
		assert.Len(t, boardView.Filters, 1)
		assert.Equal(t, column.KeyName, boardView.Filters[0].Column)
		assert.Equal(t, string(client.FilterOpExists), boardView.Filters[0].Operation)
	})

	t.Run("Create board view with scalar filter", func(t *testing.T) {
		data := &client.BoardView{
			Name: test.RandomStringWithPrefix("test.", 8),
			Filters: []client.BoardViewFilter{
				{
					Column:    column.KeyName,
					Operation: string(client.FilterOpEquals),
					Value:     "test-value",
				},
			},
		}
		boardView, err = c.BoardViews.Create(ctx, board.ID, data)
		require.NoError(t, err)
		assert.NotNil(t, boardView.ID)
		assert.Equal(t, data.Name, boardView.Name)
		assert.Len(t, boardView.Filters, 1)
		assert.Equal(t, column.KeyName, boardView.Filters[0].Column)
		assert.Equal(t, string(client.FilterOpEquals), boardView.Filters[0].Operation)
		assert.Equal(t, "test-value", boardView.Filters[0].Value)
	})

	t.Run("Create board view with array filter", func(t *testing.T) {
		data := &client.BoardView{
			Name: test.RandomStringWithPrefix("test.", 8),
			Filters: []client.BoardViewFilter{
				{
					Column:    column.KeyName,
					Operation: string(client.FilterOpIn),
					Value:     []any{"value1", "value2", "value3"},
				},
			},
		}
		boardView, err = c.BoardViews.Create(ctx, board.ID, data)
		require.NoError(t, err)
		assert.NotNil(t, boardView.ID)
		assert.Equal(t, data.Name, boardView.Name)
		assert.Len(t, boardView.Filters, 1)
		assert.Equal(t, column.KeyName, boardView.Filters[0].Column)
		assert.Equal(t, string(client.FilterOpIn), boardView.Filters[0].Operation)
		// Verify the array value
		valueArray, ok := boardView.Filters[0].Value.([]any)
		require.True(t, ok, "value should be an array")
		assert.Len(t, valueArray, 3)
	})

	t.Run("Create board view with multiple filters", func(t *testing.T) {
		data := &client.BoardView{
			Name: test.RandomStringWithPrefix("test.", 8),
			Filters: []client.BoardViewFilter{
				{
					Column:    column.KeyName,
					Operation: string(client.FilterOpExists),
				},
				{
					Column:    column.KeyName,
					Operation: string(client.FilterOpGreaterThan),
					Value:     100.0,
				},
			},
		}
		boardView, err = c.BoardViews.Create(ctx, board.ID, data)
		require.NoError(t, err)
		assert.NotNil(t, boardView.ID)
		assert.Equal(t, data.Name, boardView.Name)
		assert.Len(t, boardView.Filters, 2)
	})

	t.Run("List", func(t *testing.T) {
		result, err := c.BoardViews.List(ctx, board.ID)
		require.NoError(t, err)

		found := false
		for _, view := range result {
			if view.ID == boardView.ID {
				found = true
				break
			}
		}

		assert.True(t, found, "could not find newly created board view with List")
	})

	t.Run("Get", func(t *testing.T) {
		view, err := c.BoardViews.Get(ctx, board.ID, boardView.ID)
		require.NoError(t, err)

		assert.Equal(t, boardView.ID, view.ID)
		assert.Equal(t, boardView.Name, view.Name)
		assert.Equal(t, boardView.Filters, view.Filters)
	})

	t.Run("Delete", func(t *testing.T) {
		// Create a view to delete
		deleteView, err := c.BoardViews.Create(ctx, board.ID, &client.BoardView{
			Name: test.RandomStringWithPrefix("test.", 8),
			Filters: []client.BoardViewFilter{
				{
					Column:    column.KeyName,
					Operation: string(client.FilterOpExists),
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, deleteView.ID)

		// Delete the view
		err = c.BoardViews.Delete(ctx, board.ID, deleteView.ID)
		require.NoError(t, err)

		// Verify it's deleted by trying to get it
		_, err = c.BoardViews.Get(ctx, board.ID, deleteView.ID)
		require.Error(t, err)
	})
}
