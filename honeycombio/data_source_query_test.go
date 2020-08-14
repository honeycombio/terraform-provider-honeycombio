package honeycombio

import (
	"fmt"
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

    limit = 250
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
  "filter_combination": "AND",
  "limit": 250
}`

//Honeycomb API blows up if you send a Value with a FilterOp of `exists` or `does-not-exist`
func TestAccDataSourceHoneycombioQuery_noValueForExists(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testBadExistsQuery(),
				ExpectError: regexp.MustCompile("Filter operation exists must not"),
			},
			{
				Config:      testBadCountQuery(),
				ExpectError: regexp.MustCompile("calculation op COUNT should not have an accompanying column"),
			},
			{
				Config:      testQueryWithLimit(0),
				ExpectError: regexp.MustCompile("expected limit to be in the range \\(1 - 1000\\)"),
			},
			{
				Config:      testQueryWithLimit(-5),
				ExpectError: regexp.MustCompile("expected limit to be in the range \\(1 - 1000\\)"),
			},
			{
				Config:      testQueryWithLimit(1200),
				ExpectError: regexp.MustCompile("expected limit to be in the range \\(1 - 1000\\)"),
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

func testBadCountQuery() string {
	return `
data "honeycombio_query" "test" {
  calculation {
    op     = "COUNT"
    column = "we-should-not-specify-a-column-with-COUNT"
  }
}
`
}

func testQueryWithLimit(limit int) string {
	return fmt.Sprintf(`
data "honeycombio_query" "test" {
  limit = %d
}
`, limit)
}
