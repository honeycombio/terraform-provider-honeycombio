package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/honeycombio/terraform-provider-honeycombio/client"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"
)

func TestAccHoneycombioFlexibleBoard(t *testing.T) {
	ctx := context.Background()
	dataset := testAccDataset()
	c := testAccClient(t)

	sli, err := c.DerivedColumns.Create(ctx, dataset, &client.DerivedColumn{
		Alias:      "sli." + acctest.RandString(8),
		Expression: "BOOL(1)",
	})
	require.NoError(t, err)
	slo, err := c.SLOs.Create(ctx, dataset, &client.SLO{
		Name:             acctest.RandString(8) + " SLO",
		TimePeriodDays:   14,
		TargetPerMillion: 995000,
		SLI:              client.SLIRef{Alias: sli.Alias},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		// remove SLOs, and SLIs at end of test run
		c.SLOs.Delete(ctx, dataset, slo.ID)
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: testFlexibleBoardConfig(dataset, slo.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_flexible_board.test"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "name", "Test flexible board from terraform-provider-honeycombio"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "description", "Test flexible board description"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.#", "3"),

					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.type", "slo"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.query_panel.#", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.slo_panel.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.slo_panel.0.slo_id", slo.ID),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.height", "4"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.width", "3"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.x_coordinate", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.y_coordinate", "0"),

					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.type", "query"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.slo_panel.#", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.query_panel.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.query_panel.0.query_style", "combo"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.query_panel.0.visualization_settings.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.query_panel.0.visualization_settings.0.use_utc_xaxis", "true"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.query_panel.0.visualization_settings.0.hide_markers", "false"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.query_panel.0.visualization_settings.0.hide_hovers", "false"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.query_panel.0.visualization_settings.0.prefer_overlaid_charts", "false"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.query_panel.0.visualization_settings.0.chart.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.query_panel.0.visualization_settings.0.chart.0.chart_type", "default"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.query_panel.0.visualization_settings.0.chart.0.chart_index", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.query_panel.0.visualization_settings.0.chart.0.omit_missing_values", "true"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.query_panel.0.visualization_settings.0.chart.0.use_log_scale", "false"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.position.height", "6"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.position.width", "6"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.position.x_coordinate", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.position.y_coordinate", "3"),

					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.2.type", "text"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.2.text_panel.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.2.text_panel.0.content", "This is a text panel"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.2.position.height", "3"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.2.position.width", "4"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.2.position.x_coordinate", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.2.position.y_coordinate", "7"),
				),
			},
			// remove board's panels, add tags
			{
				Config: `
resource "honeycombio_flexible_board" "test" {
	name          = "simple flexible board"

	tags = {
	  team = "blue"
	  env  = "dev"
	}
}
			  `,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_flexible_board.test"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "name", "simple flexible board"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "description", ""),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.#", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "tags.%", "2"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "tags.team", "blue"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "tags.env", "dev"),
				),
			},
			// now add a query panel with no position (auto generated positions), update tags, and ensure the board is updated
			{
				Config: fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
    op = "COUNT"
  }
}

resource "honeycombio_query" "test" {
  dataset    = "%s"
  query_json = data.honeycombio_query_specification.test.json
}

resource "honeycombio_query_annotation" "test" {
  dataset     = "%[1]s"
  name        = "My annotated query"
  description = "My lovely description"
  query_id    = honeycombio_query.test.id
}

resource "honeycombio_flexible_board" "test" {
  name        = "simple flexible board updated"
  description = "simple flexible board description"
  tags = {
    team = "green"
  }
  panel {
    type = "query"
    query_panel {
      query_id            = honeycombio_query.test.id
      query_annotation_id = honeycombio_query_annotation.test.id
      query_style         = "combo"
      visualization_settings {
        use_utc_xaxis = true
        chart {
          chart_index         = 0
          omit_missing_values = true
          chart_type          = "line"
        }
      }
    }
  }
  panel {
    type = "text"
    text_panel {
      content = "This is a text panel with no positions."
    }
  }
}
						  `, dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_flexible_board.test"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "name", "simple flexible board updated"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "description", "simple flexible board description"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.#", "2"),
					resource.TestCheckResourceAttrPair("honeycombio_flexible_board.test", "panel.0.query_panel.0.query_id", "honeycombio_query.test", "id"),
					resource.TestCheckResourceAttrPair("honeycombio_flexible_board.test", "panel.0.query_panel.0.query_annotation_id", "honeycombio_query_annotation.test", "id"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "tags.%", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "tags.team", "green"),
					resource.TestCheckNoResourceAttr("honeycombio_flexible_board.test", "tags.env"),
					// when position not provided, the position should not be set in state
					resource.TestCheckNoResourceAttr("honeycombio_flexible_board.test", "panel.0.position.height"),
					resource.TestCheckNoResourceAttr("honeycombio_flexible_board.test", "panel.0.position.width"),
					resource.TestCheckNoResourceAttr("honeycombio_flexible_board.test", "panel.0.position.x_coordinate"),
					resource.TestCheckNoResourceAttr("honeycombio_flexible_board.test", "panel.0.position.y_coordinate"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.type", "text"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.text_panel.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.text_panel.0.content", "This is a text panel with no positions."),
					resource.TestCheckNoResourceAttr("honeycombio_flexible_board.test", "panel.1.position.height"),
					resource.TestCheckNoResourceAttr("honeycombio_flexible_board.test", "panel.1.position.width"),
					resource.TestCheckNoResourceAttr("honeycombio_flexible_board.test", "panel.1.position.x_coordinate"),
					resource.TestCheckNoResourceAttr("honeycombio_flexible_board.test", "panel.1.position.y_coordinate"),
				),
			},
			// now add an SLO panel, remove text panel, remove chart settings from the query panel, remove tags and update from generated positions to provided positions
			{
				Config: fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
    op = "COUNT"
  }
}

resource "honeycombio_query" "test" {
  dataset    = "%s"
  query_json = data.honeycombio_query_specification.test.json
}

resource "honeycombio_query_annotation" "test" {
  dataset     = "%[1]s"
  name        = "My annotated query"
  description = "My lovely description"
  query_id    = honeycombio_query.test.id
}

resource "honeycombio_flexible_board" "test" {
  name        = "simple flexible board updated"
  description = "simple flexible board description"
  panel {
    type = "query"
    query_panel {
      query_id            = honeycombio_query.test.id
      query_annotation_id = honeycombio_query_annotation.test.id
      query_style         = "combo"
      visualization_settings {
        use_utc_xaxis = false
      }
    }
	position {
      height       = 4
      width        = 3
      x_coordinate = 0
      y_coordinate = 0
    }
  }

  panel {
    type = "slo"
    slo_panel {
      slo_id = "%[2]s"
    }
	position {
      height       = 4
      width        = 3
      x_coordinate = 3
      y_coordinate = 0
    }
  }
}
						  `, dataset, slo.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_flexible_board.test"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "name", "simple flexible board updated"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "description", "simple flexible board description"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.#", "2"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.query_panel.#", "1"),
					resource.TestCheckResourceAttrPair("honeycombio_flexible_board.test", "panel.0.query_panel.0.query_id", "honeycombio_query.test", "id"),
					resource.TestCheckResourceAttrPair("honeycombio_flexible_board.test", "panel.0.query_panel.0.query_annotation_id", "honeycombio_query_annotation.test", "id"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.query_panel.0.query_style", "combo"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.query_panel.0.visualization_settings.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.query_panel.0.visualization_settings.0.use_utc_xaxis", "false"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.query_panel.0.visualization_settings.0.chart.#", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.height", "4"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.width", "3"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.x_coordinate", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.y_coordinate", "0"),

					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.position.height", "4"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.position.width", "3"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.position.x_coordinate", "3"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.position.y_coordinate", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "tags.%", "0"),
				),
			},
			// re-order the panels and remove viz settings for the query panel
			{
				Config: fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
    op = "COUNT"
  }
}

resource "honeycombio_query" "test" {
  dataset    = "%s"
  query_json = data.honeycombio_query_specification.test.json
}

resource "honeycombio_query_annotation" "test" {
  dataset     = "%[1]s"
  name        = "My annotated query"
  description = "My lovely description"
  query_id    = honeycombio_query.test.id
}

resource "honeycombio_flexible_board" "test" {
  name        = "simple flexible board updated"
  description = "simple flexible board description"

  panel {
    type = "slo"
    position {
      height = 4
      width  = 3
    }
    slo_panel {
      slo_id = "%[2]s"
    }
  }
  panel {
    type = "text"
	position {
      height = 3
      width  = 4
    }
    text_panel {
      content = "This is a text panel with no positions."
    }
  }

  panel {
    type = "query"
    query_panel {
      query_id            = honeycombio_query.test.id
      query_annotation_id = honeycombio_query_annotation.test.id
      query_style         = "table"
    }
    position {
      height = 5
      width  = 6
    }
  }
}
						  `, dataset, slo.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_flexible_board.test"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "name", "simple flexible board updated"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "description", "simple flexible board description"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.#", "3"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.type", "slo"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.slo_panel.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.slo_panel.0.slo_id", slo.ID),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.height", "4"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.width", "3"),
					resource.TestCheckResourceAttrPair("honeycombio_flexible_board.test", "panel.2.query_panel.0.query_id", "honeycombio_query.test", "id"),
					resource.TestCheckResourceAttrPair("honeycombio_flexible_board.test", "panel.2.query_panel.0.query_annotation_id", "honeycombio_query_annotation.test", "id"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.2.query_panel.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.2.query_panel.0.query_style", "table"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.2.query_panel.0.visualization_settings.#", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.2.type", "query"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.2.position.height", "5"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.2.position.width", "6"),
					// ensure x and y coordinates are dynamically generated when only height and width are provided
					resource.TestCheckResourceAttrSet("honeycombio_flexible_board.test", "panel.2.position.x_coordinate"),
					resource.TestCheckResourceAttrSet("honeycombio_flexible_board.test", "panel.2.position.y_coordinate"),

					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.type", "text"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.text_panel.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.text_panel.0.content", "This is a text panel with no positions."),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.position.height", "3"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.position.width", "4"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.position.x_coordinate", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.position.y_coordinate", "0"),
				),
			},
		},
	})
}

// TestAccHoneycombioFlexibleBoard_upgradeFromVersion036_2 tests that the flexible board resource can be upgraded from version 0.36.2
// This is needed because the position field was changed from a list to a single object causing a schema drift.
// This test ensures the conversion works as expected.
func TestAccHoneycombioFlexibleBoard_upgradeFromVersion036_2(t *testing.T) {
	ctx := context.Background()
	dataset := testAccDataset()
	c := testAccClient(t)

	sli, err := c.DerivedColumns.Create(ctx, dataset, &client.DerivedColumn{
		Alias:      "sli." + acctest.RandString(8),
		Expression: "BOOL(1)",
	})
	require.NoError(t, err)
	slo, err := c.SLOs.Create(ctx, dataset, &client.SLO{
		Name:             acctest.RandString(8) + " SLO",
		TimePeriodDays:   14,
		TargetPerMillion: 995000,
		SLI:              client.SLIRef{Alias: sli.Alias},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		// remove SLOs, and SLIs at end of test run
		c.SLOs.Delete(ctx, dataset, slo.ID)
		c.DerivedColumns.Delete(ctx, dataset, sli.ID)
	})

	config := testFlexibleBoardConfigNoTextPanel(dataset, slo.ID)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"honeycombio": {
						VersionConstraint: "0.36.2",
						Source:            "honeycombio/honeycombio",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_flexible_board.test"),
				),
			},
			{
				ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
				Config:                   config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// testFlexibleBoardConfig returns a configuration string for a flexible board
// with a query panel and an SLO panel.
func testFlexibleBoardConfig(dataset, sloID string) string {
	return fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
    op = "COUNT"
  }
}
resource "honeycombio_query" "test" {
  dataset    = "%s"
  query_json = data.honeycombio_query_specification.test.json
}
resource "honeycombio_query_annotation" "test" {
  dataset     = "%[1]s"
  name        = "My annotated query"
  description = "My lovely description"
  query_id    = honeycombio_query.test.id
}
resource "honeycombio_flexible_board" "test" {
  name        = "Test flexible board from terraform-provider-honeycombio"
  description = "Test flexible board description"
  panel {
    type = "slo"
    slo_panel {
      slo_id = "%[2]s"
    }
    position {
      height       = 4
      width        = 3
      x_coordinate = 0
      y_coordinate = 0
    }
  }
  panel {
    type = "query"
    query_panel {
      query_id            = honeycombio_query.test.id
      query_annotation_id = honeycombio_query_annotation.test.id
      query_style         = "combo"
      visualization_settings {
        use_utc_xaxis = true
        chart {
          chart_index         = 0
          omit_missing_values = true
        }
      }
    }
    position {
      height       = 6
      width        = 6
      x_coordinate = 0
      y_coordinate = 3
    }
  }
  panel {
    type = "text"
    text_panel {
      content = "This is a text panel"
    }
    position {
      height       = 3
      width        = 4
      x_coordinate = 0
      y_coordinate = 7
    }
  }
}
	`, dataset, sloID)
}

func testFlexibleBoardConfigNoTextPanel(dataset, sloID string) string {
	return fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  calculation {
    op = "COUNT"
  }
}
resource "honeycombio_query" "test" {
  dataset    = "%s"
  query_json = data.honeycombio_query_specification.test.json
}
resource "honeycombio_query_annotation" "test" {
  dataset     = "%[1]s"
  name        = "My annotated query"
  description = "My lovely description"
  query_id    = honeycombio_query.test.id
}
resource "honeycombio_flexible_board" "test" {
  name        = "Test flexible board from terraform-provider-honeycombio"
  description = "Test flexible board description"
  panel {
    type = "slo"
    slo_panel {
      slo_id = "%[2]s"
    }
    position {
      height       = 4
      width        = 3
      x_coordinate = 0
      y_coordinate = 0
    }
  }
  panel {
    type = "query"
    query_panel {
      query_id            = honeycombio_query.test.id
      query_annotation_id = honeycombio_query_annotation.test.id
      query_style         = "combo"
      visualization_settings {
        use_utc_xaxis = true
        chart {
          chart_index         = 0
          omit_missing_values = true
        }
      }
    }
    position {
      height       = 6
      width        = 6
      x_coordinate = 0
      y_coordinate = 3
    }
  }
}
	`, dataset, sloID)
}
