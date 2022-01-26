package honeycombio

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	honeycombio "github.com/honeycombio/terraform-provider-honeycombio/client"
	"github.com/stretchr/testify/assert"
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

func testAccBoardConfig(dataset string) string {
	return fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  count = 2

  calculation {
    op     = "AVG"
    column = "duration_ms"
  }

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
  }
  query {
    caption             = "test query 1"
    query_style         = "combo"
    dataset             = "%s"
    query_id            = honeycombio_query.test[1].id
    query_annotation_id = honeycombio_query_annotation.test[1].id
  }
}`, dataset, dataset, dataset, dataset)
}

func testAccCheckBoardExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		client := testAccClient(t)
		createdBoard, err := client.Boards.Get(context.Background(), resourceState.Primary.ID)
		if err != nil {
			return fmt.Errorf("could not find created board: %w", err)
		}

		expectedBoard := &honeycombio.Board{
			ID:          createdBoard.ID,
			Name:        "Test board from terraform-provider-honeycombio",
			Description: "",
			Style:       honeycombio.BoardStyleList,
			Queries: []honeycombio.BoardQuery{
				{
					Caption:           "test query 0",
					QueryStyle:        honeycombio.BoardQueryStyleGraph,
					Dataset:           testAccDataset(),
					QueryID:           createdBoard.Queries[0].QueryID,
					QueryAnnotationID: createdBoard.Queries[0].QueryAnnotationID,
					Query: &honeycombio.QuerySpec{
						Calculations: []honeycombio.CalculationSpec{
							{
								Op:     honeycombio.CalculationOpAvg,
								Column: honeycombio.StringPtr("duration_ms"),
							},
						},
						Filters: []honeycombio.FilterSpec{
							{
								Column: "duration_ms",
								Op:     ">",
								Value:  "0",
							},
						},
						TimeRange: honeycombio.IntPtr(7200),
					},
				},
				{
					Caption:           "test query 1",
					QueryStyle:        honeycombio.BoardQueryStyleCombo,
					Dataset:           testAccDataset(),
					QueryID:           createdBoard.Queries[1].QueryID,
					QueryAnnotationID: createdBoard.Queries[1].QueryAnnotationID,
					Query: &honeycombio.QuerySpec{
						Calculations: []honeycombio.CalculationSpec{
							{
								Op:     honeycombio.CalculationOpAvg,
								Column: honeycombio.StringPtr("duration_ms"),
							},
						},
						Filters: []honeycombio.FilterSpec{
							{
								Column: "duration_ms",
								Op:     ">",
								Value:  "1",
							},
						},
						TimeRange: honeycombio.IntPtr(7200),
					},
				},
			},
		}

		ok = assert.Equal(t, expectedBoard, createdBoard)
		if !ok {
			return errors.New("created board did not match expected board")
		}
		return nil
	}
}
