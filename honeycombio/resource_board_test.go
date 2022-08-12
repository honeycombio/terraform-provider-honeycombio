package honeycombio

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccHoneycombioBoard_basic(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBoardConfig(dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_board.test"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "name", "Test board from terraform-provider-honeycombio"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "style", "list"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "description", ""),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.#", "2"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.caption", "test query 0"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.dataset", dataset),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.0.graph_settings.0.utc_xaxis", "true"),
					resource.TestCheckResourceAttrPair("honeycombio_board.test", "query.0.query_id", "honeycombio_query.test.0", "id"),
					resource.TestCheckResourceAttrPair("honeycombio_board.test", "query.0.query_annotation_id", "honeycombio_query_annotation.test.0", "id"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.1.caption", "test query 1"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.1.dataset", dataset),
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
	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
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
  style         = "visual"
  column_layout = "single"

  query {
    query_id = honeycombio_query.test.id

    graph_settings {
      log_scale = true
    }
  }
}
`, dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_board.test"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "name", "simple board"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "style", "visual"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "column_layout", "single"),
					resource.TestCheckResourceAttr("honeycombio_board.test", "query.#", "1"),
					resource.TestCheckResourceAttrPair("honeycombio_board.test", "query.0.query_id", "honeycombio_query.test", "id"),
				),
			},
		},
	})
	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "honeycombio_board" "test" {
  name          = "error board"
  style         = "list"
  column_layout = "multi"
}
`,
				ExpectError: regexp.MustCompile(`list style boards cannot specify a column layout`),
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

  dataset    = "%s"
  query_json = data.honeycombio_query_specification.test[count.index].json
}

resource "honeycombio_query_annotation" "test" {
  count = 2

  dataset     = "%s"
  name        = "My annotated query"
  description = "My lovely description"
  query_id    = honeycombio_query.test[count.index].id
}

resource "honeycombio_board" "test" {
  name  = "Test board from terraform-provider-honeycombio"
  style = "list"

  query {
    caption             = "test query 0"
    dataset             = "%s"
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
}`, dataset, dataset, dataset)
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
