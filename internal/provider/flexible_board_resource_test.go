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
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.type", "slo"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.query_panel.#", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.slo_panel.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.slo_panel.0.slo_id", slo.ID),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.0.x_coordinate", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.0.y_coordinate", "0"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.0.height", "1"),
					resource.TestCheckResourceAttr("honeycombio_flexible_board.test", "panel.0.position.0.width", "1"),
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
  count = 2

  calculation {
    op     = "AVG"
    column = "duration_ms"
  }

  filter_combination = "AND"

  filter {
    column = "duration_ms"
    op     = ">"
    value  = count.index
  }
}

resource "honeycombio_query" "test" {
  count = 2

  dataset    = "%[1]s"
  query_json = data.honeycombio_query_specification.test[count.index].json
}

resource "honeycombio_query_annotation" "test" {
  count = 2

  dataset     = "%[1]s"
  name        = "My annotated query"
  description = "My lovely description"
  query_id    = honeycombio_query.test[count.index].id
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
}`, dataset, sloID)
}
