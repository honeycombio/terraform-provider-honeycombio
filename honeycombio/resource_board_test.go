package honeycombio

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccHoneycombioBoard_basic(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
		Steps: []resource.TestStep{
			{
				Config: testAccBoardConfig(dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_board.test"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "name", "Test board from terraform-provider-honeycombio"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "style", "visual"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "description", ""),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.#", "2"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.caption", "test query 0"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.dataset", dataset),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.graph_settings.0.utc_xaxis", "true"),
					resource.TestCheckResourceAttrPair("honeycombio_board.test", "query.0.query_id", "honeycombio_query.test.0", "id"),
					resource.TestCheckResourceAttrPair("honeycombio_board.test", "query.0.query_annotation_id", "honeycombio_query_annotation.test.0", "id"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.1.caption", "test query 1"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.1.dataset", dataset),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.1.graph_settings.0.utc_xaxis", "false"),
					resource.TestCheckResourceAttrPair("honeycombio_board.test", "query.1.query_id", "honeycombio_query.test.1", "id"),
					resource.TestCheckResourceAttrPair("honeycombio_board.test", "query.1.query_annotation_id", "honeycombio_query_annotation.test.1", "id"),
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
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactory,
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
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.graph_settings.0.log_scale", "false"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.graph_settings.0.utc_xaxis", "false"),
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
  name          = "simple board"

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
				// skipped due to bug: https://github.com/honeycombio/terraform-provider-honeycombio/issues/399
				SkipFunc: func() (bool, error) { return true, nil },
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

    // below would have the test pass and is a workaround for the bug
    // graph_settings {}
  }
}`, dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_board.test"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.graph_settings.0.log_scale", "false"),
					// this check currently fails due to the bug
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.graph_settings.0.utc_xaxis", "false"),
				),
			},
		},
	})
}

func testAccBoardConfig(dataset string) string {
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
  }
}`, dataset)
}

//nolint:unparam
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
