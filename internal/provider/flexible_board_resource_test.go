package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/honeycombio/terraform-provider-honeycombio/client"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
					resource.TestCheckResourceAttrWith("honeycombio_flexible_board.test", "panel.2.text_panel.0.content", func(value string) error {
						if !strings.Contains(value, "# Text Panel Title") {
							return fmt.Errorf("expected content to contain '# Text Panel Title', got: %s", value)
						}
						if !strings.Contains(value, "multiline") {
							return fmt.Errorf("expected content to contain 'multiline', got: %s", value)
						}
						return nil
					}),
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
      content = <<EOF
# Dynamic Text Panel

This is a **multiline** text panel with:
- No fixed positions
- Auto-generated coordinates
- Flexible layout support
EOF
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
					resource.TestCheckResourceAttrWith("honeycombio_flexible_board.test", "panel.1.text_panel.0.content", func(value string) error {
						if !strings.Contains(value, "# Dynamic Text Panel") {
							return fmt.Errorf("expected content to contain '# Dynamic Text Panel', got: %s", value)
						}
						if !strings.Contains(value, "Auto-generated coordinates") {
							return fmt.Errorf("expected content to contain 'Auto-generated coordinates', got: %s", value)
						}
						return nil
					}),
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
      content = <<EOF
# Positioned Text Panel

This is a **multiline** text panel with:
- Fixed position and size
- Rich markdown formatting
- Multiple content sections

## Additional Info
Content positioned at specific coordinates.
EOF
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
					resource.TestCheckResourceAttrWith("honeycombio_flexible_board.test", "panel.1.text_panel.0.content", func(value string) error {
						if !strings.Contains(value, "# Positioned Text Panel") {
							return fmt.Errorf("expected content to contain '# Positioned Text Panel', got: %s", value)
						}
						if !strings.Contains(value, "Fixed position and size") {
							return fmt.Errorf("expected content to contain 'Fixed position and size', got: %s", value)
						}
						return nil
					}),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.position.height", "3"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.position.width", "4"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.position.x_coordinate", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.position.y_coordinate", "0"),
				),
			},
			// add preset filters
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
      content = <<EOF
# Positioned Text Panel

This is a **multiline** text panel with:
- Fixed position and size
- Rich markdown formatting
- Multiple content sections

## Additional Info
Content positioned at specific coordinates.
EOF
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

  preset_filter {
    column = "column1"
    alias  = "alias1"
  }
  preset_filter {
    column = "column2"
    alias  = "alias2"
  }
}
						  `, dataset, slo.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_flexible_board.test"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.#", "2"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.0.column", "column1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.0.alias", "alias1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.1.column", "column2"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.1.alias", "alias2"),
				),
			},
			// update preset filters - change one and remove one
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
      content = <<EOF
# Positioned Text Panel

This is a **multiline** text panel with:
- Fixed position and size
- Rich markdown formatting
- Multiple content sections

## Additional Info
Content positioned at specific coordinates.
EOF
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

  preset_filter {
    column = "column1"
    alias  = "updated_alias1"
  }
}
						  `, dataset, slo.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_flexible_board.test"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.0.column", "column1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.0.alias", "updated_alias1"),
				),
			},
			// remove preset filters
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
      content = <<EOF
# Positioned Text Panel

This is a **multiline** text panel with:
- Fixed position and size
- Rich markdown formatting
- Multiple content sections

## Additional Info
Content positioned at specific coordinates.
EOF
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
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.#", "0"),
				),
			},
		},
	})
}

// TestAccHoneycombioFlexibleBoard_presetFilters tests preset_filters functionality
func TestAccHoneycombioFlexibleBoard_presetFilters(t *testing.T) {
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
			// create board with preset filters
			{
				Config: `
resource "honeycombio_flexible_board" "test" {
  name        = "Test board with preset filters"
  description = "Testing preset filters"

  preset_filter {
    column = "service.name"
    alias  = "service"
  }
  preset_filter {
    column = "trace.trace_id"
    alias  = "trace"
  }
}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_flexible_board.test"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "name", "Test board with preset filters"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.#", "2"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.0.column", "service.name"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.0.alias", "service"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.1.column", "trace.trace_id"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.1.alias", "trace"),
				),
			},
			// update preset filters - modify values
			{
				Config: `
			resource "honeycombio_flexible_board" "test" {
			  name        = "Test board with preset filters"
			  description = "Testing preset filters"

			  preset_filter {
			    column = "service.name"
			    alias  = "updated_service"
			  }
			  preset_filter {
			    column = "environment"
			    alias  = "env"
			  }
			  preset_filter {
			    column = "deployment.id"
			    alias  = "deployment"
			  }
			}
							`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_flexible_board.test"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.#", "3"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.0.column", "service.name"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.0.alias", "updated_service"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.1.column", "environment"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.1.alias", "env"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.2.column", "deployment.id"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.2.alias", "deployment"),
				),
			},
			// remove all preset filters
			{
				Config: `
			resource "honeycombio_flexible_board" "test" {
			  name        = "Test board with preset filters"
			  description = "Testing preset filters"
			}
							`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_flexible_board.test"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.#", "0"),
				),
			},
			// add preset filters back with panels
			{
				Config: fmt.Sprintf(`
			resource "honeycombio_flexible_board" "test" {
			  name        = "Test board with preset filters"
			  description = "Testing preset filters"

			  panel {
			    type = "slo"
			    slo_panel {
			      slo_id = "%[2]s"
			    }
			  }

			  preset_filter {
			    column = "final.column"
			    alias  = "final_alias"
			  }
			}
							`, dataset, slo.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_flexible_board.test"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.0.column", "final.column"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.0.alias", "final_alias"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.#", "1"),
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

func TestAccHoneycombioFlexibleBoard_upgradeFromVersion043_0(t *testing.T) {
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
						VersionConstraint: "0.43.0",
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
      content = <<EOF
# Text Panel Title

This is a **multiline** text panel with:
- Support for markdown
- Multiple lines of content
- Rich formatting options

## Section 2
Additional content can be added here.
EOF
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

func testAccCheckBoardExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		client := testAccClient(t)
		_, err := client.Boards.Get(context.Background(), resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("could not find created board: %w", err)
		}
		return nil
	}
}

func TestImportHoneycombioFlexibleBoard(t *testing.T) {
	config := `
resource "honeycombio_flexible_board" "test" {
  name        = "Test board for import"
  description = "Testing import with preset filters"

  preset_filter {
    column = "service.name"
    alias  = "service"
  }
  preset_filter {
    column = "trace.trace_id"
    alias  = "trace"
  }
  preset_filter {
    column = "environment"
    alias  = "env"
  }
}
	`

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_flexible_board.test"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "name", "Test board for import"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "description", "Testing import with preset filters"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.#", "3"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.0.column", "service.name"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.0.alias", "service"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.1.column", "trace.trace_id"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.1.alias", "trace"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.2.column", "environment"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "preset_filter.2.alias", "env"),
				),
			},
			{
				ResourceName:      "honeycombio_flexible_board.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Don't ignore any import fields when verifying the imported state
				ImportStateVerifyIgnore: []string{},
			},
		},
	})
}
