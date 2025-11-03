package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestFlexibleBoards(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var flexibleBoard *client.Board
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
	require.NotEmpty(t, query.ID)

	qa := &client.QueryAnnotation{
		Name:        test.RandomStringWithPrefix("test.", 20),
		Description: "This derived column is created by a test",
		QueryID:     *query.ID,
	}
	queryAnnotation, err := c.QueryAnnotations.Create(ctx, dataset, qa)
	require.NoError(t, err)
	require.NotNil(t, queryAnnotation.ID)

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

	t.Run("Create flexible board", func(t *testing.T) {
		data := &client.Board{
			Name:        test.RandomStringWithPrefix("test.", 8),
			BoardType:   "flexible",
			Description: "A board with some panels",
			Panels: []client.BoardPanel{
				{
					PanelType: client.BoardPanelTypeQuery,
					PanelPosition: client.BoardPanelPosition{
						X:      0,
						Y:      0,
						Height: 3,
						Width:  4,
					},
					QueryPanel: &client.BoardQueryPanel{
						QueryID:           *query.ID,
						QueryAnnotationID: queryAnnotation.ID,
						Style:             client.BoardQueryStyleGraph,
					},
				},
				{
					PanelType: client.BoardPanelTypeSLO,
					PanelPosition: client.BoardPanelPosition{
						X:      6,
						Y:      0,
						Height: 3,
						Width:  4,
					},
					SLOPanel: &client.BoardSLOPanel{
						SLOID: slo.ID,
					},
				},
				{
					PanelType: client.BoardPanelTypeText,
					PanelPosition: client.BoardPanelPosition{
						X:      0,
						Y:      3,
						Height: 3,
						Width:  4,
					},
					TextPanel: &client.BoardTextPanel{
						Content: "This is a text panel",
					},
				},
			},
		}
		flexibleBoard, err = c.Boards.Create(ctx, data)
		require.NoError(t, err)
		assert.NotNil(t, flexibleBoard.ID)

		// copy ID before asserting equality
		data.ID = flexibleBoard.ID

		// ensure the board URL got populated
		assert.NotEmpty(t, flexibleBoard.Links.BoardURL)
		data.Links.BoardURL = flexibleBoard.Links.BoardURL

		// copy dataset name into query panel for comparison
		for i, panel := range flexibleBoard.Panels {
			if panel.PanelType == client.BoardPanelTypeQuery {
				assert.Equal(t, dataset, panel.QueryPanel.Dataset)
				data.Panels[i].QueryPanel.Dataset = dataset
			}
		}

		assert.Equal(t, data, flexibleBoard)
	})

	t.Run("Create flexible board with auto layout generation", func(t *testing.T) {
		data := &client.Board{
			Name:             test.RandomStringWithPrefix("test.", 8),
			BoardType:        "flexible",
			Description:      "A board with some panels",
			LayoutGeneration: client.LayoutGenerationAuto,
			Panels: []client.BoardPanel{
				{
					PanelType: client.BoardPanelTypeQuery,
					QueryPanel: &client.BoardQueryPanel{
						QueryID:           *query.ID,
						QueryAnnotationID: queryAnnotation.ID,
						Style:             client.BoardQueryStyleGraph,
					},
				},
				{
					PanelType: client.BoardPanelTypeSLO,
					SLOPanel: &client.BoardSLOPanel{
						SLOID: slo.ID,
					},
				},
				{
					PanelType: client.BoardPanelTypeText,
					TextPanel: &client.BoardTextPanel{
						Content: "This is a text panel",
					},
				},
			},
		}
		flexibleBoard, err = c.Boards.Create(ctx, data)
		require.NoError(t, err)
		assert.NotNil(t, flexibleBoard.ID)

		// copy ID before asserting equality
		data.ID = flexibleBoard.ID

		// ensure the board URL got populated
		assert.NotEmpty(t, flexibleBoard.Links.BoardURL)
		data.Links.BoardURL = flexibleBoard.Links.BoardURL

		// copy dataset name into query panel for comparison
		for i, panel := range flexibleBoard.Panels {
			if panel.PanelType == client.BoardPanelTypeQuery {
				assert.Equal(t, dataset, panel.QueryPanel.Dataset)
				data.Panels[i].QueryPanel.Dataset = dataset
			}
		}

		for i, panel := range flexibleBoard.Panels {
			assert.Equal(t, data.Panels[i].PanelType, panel.PanelType)
			assert.Equal(t, data.Panels[i].QueryPanel, panel.QueryPanel)
			assert.Equal(t, data.Panels[i].SLOPanel, panel.SLOPanel)

			// since positions are auto generated, we can't assert their exact values
			// but we can assert that they are not empty
			assert.NotEmpty(t, panel.PanelPosition)
			assert.GreaterOrEqual(t, panel.PanelPosition.X, 0)
			assert.GreaterOrEqual(t, panel.PanelPosition.Y, 0)
			assert.Positive(t, panel.PanelPosition.Height)
			assert.Positive(t, panel.PanelPosition.Width)
		}
	})

	t.Run("Create flexible board with tags", func(t *testing.T) {
		data := &client.Board{
			Name:        test.RandomStringWithPrefix("test.", 8),
			BoardType:   "flexible",
			Description: "A board with some tags",
			Panels: []client.BoardPanel{
				{
					PanelType: client.BoardPanelTypeQuery,
					PanelPosition: client.BoardPanelPosition{
						X:      0,
						Y:      0,
						Height: 3,
						Width:  4,
					},
					QueryPanel: &client.BoardQueryPanel{
						QueryID:           *query.ID,
						QueryAnnotationID: queryAnnotation.ID,
						Style:             client.BoardQueryStyleGraph,
					},
				},
				{
					PanelType: client.BoardPanelTypeSLO,
					PanelPosition: client.BoardPanelPosition{
						X:      6,
						Y:      0,
						Height: 3,
						Width:  4,
					},
					SLOPanel: &client.BoardSLOPanel{
						SLOID: slo.ID,
					},
				},
				{
					PanelType: client.BoardPanelTypeText,
					PanelPosition: client.BoardPanelPosition{
						X:      0,
						Y:      3,
						Height: 3,
						Width:  4,
					},
					TextPanel: &client.BoardTextPanel{
						Content: "This is a text panel",
					},
				},
			},
			Tags: []client.Tag{
				{Key: "color", Value: "blue"},
			},
		}
		flexibleBoard, err = c.Boards.Create(ctx, data)
		require.NoError(t, err)
		assert.NotNil(t, flexibleBoard.ID)

		// copy ID before asserting equality
		data.ID = flexibleBoard.ID

		// ensure the board URL got populated
		assert.NotEmpty(t, flexibleBoard.Links.BoardURL)
		data.Links.BoardURL = flexibleBoard.Links.BoardURL

		// copy dataset name into query panel for comparison
		for i, panel := range flexibleBoard.Panels {
			if panel.PanelType == client.BoardPanelTypeQuery {
				assert.Equal(t, dataset, panel.QueryPanel.Dataset)
				data.Panels[i].QueryPanel.Dataset = dataset
			}
		}

		// ensure the tags were added
		assert.NotEmpty(t, flexibleBoard.Tags)
		assert.ElementsMatch(t, flexibleBoard.Tags, data.Tags, "tags do not match")

		assert.Equal(t, data, flexibleBoard)
	})

	t.Run("List", func(t *testing.T) {
		result, err := c.Boards.List(ctx)
		require.NoError(t, err)

		assert.Contains(t, result, *flexibleBoard, "could not find newly created board with List")
	})

	t.Run("Get", func(t *testing.T) {
		board, err := c.Boards.Get(ctx, flexibleBoard.ID)
		require.NoError(t, err)

		assert.ElementsMatch(t, board.Tags, flexibleBoard.Tags, "tags do not match")

		assert.Equal(t, *board, *flexibleBoard)
	})
}
