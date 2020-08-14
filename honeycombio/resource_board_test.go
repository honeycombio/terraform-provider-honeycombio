package honeycombio

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	honeycombio "github.com/kvrhdn/go-honeycombio"
	"github.com/stretchr/testify/assert"
)

func TestAccHoneycombioBoard_basic(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccBoardConfig(dataset),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBoardExists(t, "honeycombio_board.test"),
				),
			},
		},
	})
}

func testAccBoardConfig(dataset string) string {
	return fmt.Sprintf(`
data "honeycombio_query" "test" {
    count = 2
    calculation {
        op     = "AVG"
        column = "duration_ms"
    }
    filter {
        op     = ">"
        column = "duration_ms"
        value  = count.index 
    }
}

resource "honeycombio_board" "test" {
    name  = "Test board from terraform-provider-honeycombio"
    style = "list"
     
    query {
        caption = "test query 0"
        dataset = "%s"
        query_json = data.honeycombio_query.test[0].json
    }
      
    query {
        caption = "test query 1"
        dataset = "%s"
        query_json = data.honeycombio_query.test[1].json
    }

}`, dataset, dataset)
}

func testAccCheckBoardExists(t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		client := testAccProvider.Meta().(*honeycombio.Client)
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
					Caption: "test query 0",
					Dataset: testAccDataset(),
					Query: honeycombio.QuerySpec{
						Calculations: []honeycombio.CalculationSpec{
							{
								Op:     honeycombio.CalculateOpAvg,
								Column: &[]string{"duration_ms"}[0],
							},
						},
						Filters: []honeycombio.FilterSpec{
							{
								Column: "duration_ms",
								Op:     ">",
								Value:  "0",
							},
						},
					},
				},
				{
					Caption: "test query 1",
					Dataset: testAccDataset(),
					Query: honeycombio.QuerySpec{
						Calculations: []honeycombio.CalculationSpec{
							{
								Op:     honeycombio.CalculateOpAvg,
								Column: &[]string{"duration_ms"}[0],
							},
						},
						Filters: []honeycombio.FilterSpec{
							{
								Column: "duration_ms",
								Op:     ">",
								Value:  "1",
							},
						},
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
