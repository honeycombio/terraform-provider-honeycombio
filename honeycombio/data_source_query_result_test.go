package honeycombio

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceHoneycombioQueryResult_basic(t *testing.T) {
	dataset := testAccDataset()

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceQueryResultConfig(dataset),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("results", "COUNT"),
				),
			},
		},
	})
}

func testAccDataSourceQueryResultConfig(dataset string) string {
	return fmt.Sprintf(`
data "honeycombio_query_specification" "test" {
  time_range = 86400

  calculation {
    op = "COUNT"
  }
}

resource "honeycombio_query" "test" {
  dataset    = "%s"
  query_json = data.honeycombio_query_specification.test.json
}

data "honeycombio_query_result" "test" {
  dataset  = "%s"
  query_id = honeycombio_query.test.id
}

output "results" {
  # do some type silly iteration over the result to output the calculation ('COUNT')
  # as the resulting value is variable
  value = join(", ",
    flatten(
      [
        for result in data.honeycombio_query_result.test.results : [for k, v in result : "${k}"]
      ]
    )
  )
}`, dataset, dataset)
}
