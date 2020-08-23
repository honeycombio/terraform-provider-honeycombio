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
        column        = "duration_ms"
        op            = ">"
        value_integer = 100
    }
    filter {
        column = "app.tenant"
        op     = "="
        value  = "ThatSpecialTenant"
    }

    breakdowns = ["column_1"]

    order {
        op     = "AVG"
        column = "duration_ms"
    }
    order {
        column = "column_1"
        order  = "descending"
    }

    limit = 250
}

output "query_json" {
    value = data.honeycombio_query.test.json
}`
}

//Note: By default go encodes `<` and `>` for html, hence the `\u003e`
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
      "column": "duration_ms",
      "op": "\u003e",
      "value": 100
    },
    {
      "column": "app.tenant",
      "op": "=",
      "value": "ThatSpecialTenant"
    }
  ],
  "filter_combination": "AND",
  "breakdowns": [
    "column_1"
  ],
  "orders": [
    {
      "op": "AVG",
      "column": "duration_ms"
    },
    {
      "column": "column_1",
      "order": "descending"
    }
  ],
  "limit": 250
}`

func TestAccDataSourceHoneycombioQuery_validationChecks(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testBadExistsQuery(),
				ExpectError: regexp.MustCompile("filter operation exists must not"),
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
			{
				Config:      testConflictingValues(),
				ExpectError: regexp.MustCompile(multipleValuesError),
			},
			{
				Config:      testMissingFilterValue(),
				ExpectError: regexp.MustCompile("filter operation > requires a value"),
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

func testConflictingValues() string {
	return `
data "honeycombio_query" "test" {
  filter {
    column = "column"
    op     = ">"
    value  = "1"
    value_integer = 10
  }
}
`
}

func testMissingFilterValue() string {
	return `
data "honeycombio_query" "test" {
  filter {
    column = "column"
    op     = ">"
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
