package honeycombio

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceHoneycombioQuery_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("query_json", expectedJSON),
				),
			},
		},
	})
}

func testAccQueryConfig() string {
	return `
data "honeycombio_query" "test" {
    calculation {
        op     = "AVG"
        column = "duration_ms"
    }

    filter {
        column = "trace.parent_id"
        op     = "does-not-exist"
    }

    filter {
        column = "app.tenant"
        op     = "="
        value  = "ThatSpecialTenant"
    }
}

output "query_json" {
    value = data.honeycombio_query.test.json
}`
}

const expectedJSON string = `{
  "calculations": [
    {
      "op": "AVG",
      "column": "duration_ms"
    }
  ],
  "filters": [
    {
      "column": "trace.parent_id",
      "op": "does-not-exist",
      "value": ""
    },
    {
      "column": "app.tenant",
      "op": "=",
      "value": "ThatSpecialTenant"
    }
  ],
  "filter_combination": "AND"
}`
