package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/client"
)

func TestAccHoneycombioBoard_classic_basic(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: testClassicBoardConfig(dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_board.test"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "name", "Test board from terraform-provider-honeycombio"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "style", "visual"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "description", ""),
					resource.TestCheckResourceAttr("honeycombio_board.test", "type", "classic"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.#", "2"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.caption", "test query 0"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.dataset", dataset),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.graph_settings.0.utc_xaxis", "true"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.graph_settings.0.omit_missing_values", "false"),
					resource.TestCheckResourceAttrPair("honeycombio_board.test", "query.0.query_id", "honeycombio_query.test.0", "id"),
					resource.TestCheckResourceAttrPair("honeycombio_board.test", "query.0.query_annotation_id", "honeycombio_query_annotation.test.0", "id"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.1.caption", "test query 1"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.1.dataset", dataset),
					resource.TestCheckResourceAttrPair("honeycombio_board.test", "query.1.query_id", "honeycombio_query.test.1", "id"),
					resource.TestCheckResourceAttrPair("honeycombio_board.test", "query.1.query_annotation_id", "honeycombio_query_annotation.test.1", "id"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.1.graph_settings.0.omit_missing_values", "true"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.1.graph_settings.0.utc_xaxis", "false"),
				),
			},
			{
				ResourceName:      "honeycombio_board.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccHoneycombioBoard_updateGraphSettings(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			// setup a board with a single query with no graph settings
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

resource "honeycombio_board" "test" {
  name          = "simple board"

  query {
    query_id = honeycombio_query.test.id
  }
}`, dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_board.test"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.graph_settings.#", "0"),
				),
			},
			// now add a few graph settings to our query
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

resource "honeycombio_board" "test" {
  name          = "simple board"

  query {
    query_id = honeycombio_query.test.id

    graph_settings {
      utc_xaxis = false
      log_scale = true
    }
  }
}`, dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_board.test"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "name", "simple board"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.#", "1"),
					resource.TestCheckResourceAttrPair("honeycombio_board.test", "query.0.query_id", "honeycombio_query.test", "id"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.graph_settings.0.log_scale", "true"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.graph_settings.0.utc_xaxis", "false"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.graph_settings.0.overlaid_charts", "false"),
				),
			},
			// do a little shuffle of the settings and make sure we're all still updating
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

resource "honeycombio_board" "test" {
  name = "simple board"

  query {
    query_id = honeycombio_query.test.id

    graph_settings {
      utc_xaxis = true
    }
  }
}`, dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_board.test"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.graph_settings.0.log_scale", "false"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.graph_settings.0.utc_xaxis", "true"),
				),
			},
			// finally remove the graph settings and an ensure we're back to the defaults
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

resource "honeycombio_board" "test" {
  name          = "simple board"

  query {
    query_id = honeycombio_query.test.id
  }
}`, dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_board.test"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.graph_settings.#", "0"),
				),
			},
		},
	})
}

func TestAccBoard_withSLOs(t *testing.T) {
	ctx := context.Background()
	dataset := testAccDataset()
	c := testAccClient(t)

	sli1, err := c.DerivedColumns.Create(ctx, dataset, &client.DerivedColumn{
		Alias:      "sli." + acctest.RandString(8),
		Expression: "BOOL(1)",
	})
	require.NoError(t, err)
	slo1, err := c.SLOs.Create(ctx, dataset, &client.SLO{
		Name:             acctest.RandString(8) + " SLO",
		TimePeriodDays:   14,
		TargetPerMillion: 995000,
		SLI:              client.SLIRef{Alias: sli1.Alias},
	})
	require.NoError(t, err)

	sli2, err := c.DerivedColumns.Create(ctx, dataset, &client.DerivedColumn{
		Alias:      "sli." + acctest.RandString(8),
		Expression: "BOOL(1)",
	})
	require.NoError(t, err)
	slo2, err := c.SLOs.Create(ctx, dataset, &client.SLO{
		Name:             acctest.RandString(8) + " SLO",
		TimePeriodDays:   14,
		TargetPerMillion: 995000,
		SLI:              client.SLIRef{Alias: sli2.Alias},
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		// remove SLOs, and SLIs at end of test run
		c.SLOs.Delete(ctx, dataset, slo1.ID)
		c.SLOs.Delete(ctx, dataset, slo2.ID)
		c.DerivedColumns.Delete(ctx, dataset, sli1.ID)
		c.DerivedColumns.Delete(ctx, dataset, sli2.ID)
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			// create a board with one SLO on it
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

resource "honeycombio_board" "test" {
  name          = "board with some SLOs"

  query {
    query_id = honeycombio_query.test.id
  }

  slo {
    id = "%s"
  }
}
`, dataset, slo1.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_board.test"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "slo.#", "1"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "slo.0.id", slo1.ID),
				),
			},
			{
				// update that board to have two SLOs
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

resource "honeycombio_board" "test" {
  name          = "board with some SLOs"

  query {
    query_id = honeycombio_query.test.id
  }

  slo {
    id = "%s"
  }

  slo {
    id = "%s"
  }
}`, dataset, slo2.ID, slo1.ID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_board.test"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "slo.#", "2"),
				),
			},
			{
				// remove all SLOs from that board
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

resource "honeycombio_board" "test" {
  name          = "board with no SLOs"

  query {
    query_id = honeycombio_query.test.id
  }
}`, dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_board.test"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "slo.#", "0"),
				),
			},
		},
	})
}

// TestAcc_BoardResourceUpgradeFromVersion032 is intended to test the migration
// case from the last SDK-based version of the Board resource to the current Framework-based
// version.
//
// See: https://developer.hashicorp.com/terraform/plugin/framework/migrating/testing#testing-migration
func TestAcc_BoardResourceUpgradeFromVersion032(t *testing.T) {
	dataset := testAccDataset()
	config := testClassicBoardConfig(dataset)

	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"honeycombio": {
						VersionConstraint: "0.32.0",
						Source:            "honeycombio/honeycombio",
					},
				},
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_board.test"),
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

func testClassicBoardConfig(dataset string) string {
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

resource "honeycombio_board" "test" {
  name  = "Test board from terraform-provider-honeycombio"

  query {
    caption             = "test query 0"
    dataset             = "%[1]s"
    query_id            = honeycombio_query.test[0].id
    query_annotation_id = honeycombio_query_annotation.test[0].id

    graph_settings {
      utc_xaxis = true
    }
  }
  query {
    caption             = "test query 1"
    query_style         = "combo"
    query_id            = honeycombio_query.test[1].id
    query_annotation_id = honeycombio_query_annotation.test[1].id

    graph_settings {
      omit_missing_values = true
    }
  }
}`, dataset)
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
