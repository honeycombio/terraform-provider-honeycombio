package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAccHoneycombioFlexibleBoard_basic(t *testing.T) {
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
		_ = c.SLOs.Delete(ctx, dataset, slo.ID)
		_ = c.DerivedColumns.Delete(ctx, dataset, sli.ID)
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
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "type", "flexible"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.#", "2"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.type", "slo"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.query_panel.#", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.slo_panel.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.slo_panel.0.slo_id", slo.ID),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.0.x_coordinate", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.0.y_coordinate", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.0.height", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.0.width", "1"),
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
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.query_panel.0.visualization_settings.0.chart.0.chart_type", "line"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.query_panel.0.visualization_settings.0.chart.0.chart_index", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.query_panel.0.visualization_settings.0.chart.0.omit_missing_values", "true"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.1.query_panel.0.visualization_settings.0.chart.0.use_log_scale", "false"),
				),
			},
			{
				ResourceName:      "honeycombio_flexible_board.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

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
  name  = "Test flexible board from terraform-provider-honeycombio"
  description = "Test flexible board description"
  type = "flexible"

  panel {
    type = "slo"
	slo_panel {
		slo_id = "%[2]s"
	}
	position {
		x_coordinate = 0
		y_coordinate = 0
		height      = 1
		width       = 1
	}
  }

  panel {
    type = "query"
	query_panel {
		query_id = honeycombio_query.test.id
		query_annotation_id = honeycombio_query_annotation.test.id
		query_style = "combo"
		visualization_settings {
            use_utc_xaxis = true
			chart {
                chart_type = "line"
                chart_index = 0
                omit_missing_values = true
            }
        }
	}
	position {
		x_coordinate = 0
		y_coordinate = 0
		height      = 6
		width       = 6
	}
 }
}`, dataset, sloID)
}
