package honeycombio

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
      "op": "does-not-exist"
    },
    {
      "column": "app.tenant",
      "op": "=",
      "value": "ThatSpecialTenant"
    }
  ],
  "filter_combination": "AND"
}`

//Honeycomb API blows up if you send a Value with a FilterOp of `exists` or `does-not-exist`
func TestAccDataSourceHoneycombioQuery_noValueForExists(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testBadExistsQuery(),
				ExpectError: regexp.MustCompile("Filter operation exists must not"),
			},
		},
	})
}

func testBadExistsQuery() string {
	return `
data "honeycombio_query" "test" {
    filter {
        column = "column"
        op     = "exists"
        value  = "this-value-should-not-be-here"
    }
}
`
}
