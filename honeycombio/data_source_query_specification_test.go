package honeycombio

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceHoneycombioQuery_EmptyDefaults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "honeycombio_query_specification" "test" {}

output "query_json" {
  value = data.honeycombio_query_specification.test.json
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("query_json", `{
  "calculations": [
    {
      "op": "COUNT"
    }
  ],
  "time_range": 7200
}`),
				),
			},
		},
	})
}

func TestAccDataSourceHoneycombioQuery_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccQueryConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("query_json", expectedJSON),
				),
			},
		},
	})
}

const testAccQueryConfig = `
data "honeycombio_query_specification" "test" {
    calculation {
        op     = "AVG"
        column = "duration_ms"
    }
    calculation {
        op     = "P99"
        column = "duration_ms"
    }

    filter {
        column = "trace.parent_id"
        op     = "does-not-exist"
    }
    filter {
        column = "duration_ms"
        op     = ">"
        value  = 0
    }
    filter {
        column        = "duration_ms"
        op            = "<"
        value_integer = 100
    }
    filter {
        column = "app.tenant"
        op     = "="
        value  = "ThatSpecialTenant"
    }
    filter {
        column = "app.database.shard"
        op     = "not-in"
        value  = "347338,837359"
    }

    filter_combination = "OR"

    breakdowns = ["column_1"]

    order {
        op     = "AVG"
        column = "duration_ms"
    }
    order {
        column = "column_1"
        order  = "ascending"
    }

    having {
        calculate_op = "P99"
        column       = "duration_ms"
        op           = ">"
        value        = 1000
    }

    limit 	    = 250
    time_range  = 7200
    start_time  = 1577836800
    granularity = 30
}

output "query_json" {
    value = data.honeycombio_query_specification.test.json
}`

// Note: By default go encodes `<` and `>` for html, hence the `\u003e`
const expectedJSON string = `{
  "calculations": [
    {
      "op": "AVG",
      "column": "duration_ms"
    },
    {
      "op": "P99",
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
      "value": 0
    },
    {
      "column": "duration_ms",
      "op": "\u003c",
      "value": 100
    },
    {
      "column": "app.tenant",
      "op": "=",
      "value": "ThatSpecialTenant"
    },
    {
      "column": "app.database.shard",
      "op": "not-in",
      "value": [
        347338,
        837359
      ]
    }
  ],
  "filter_combination": "OR",
  "breakdowns": [
    "column_1"
  ],
  "orders": [
    {
      "op": "AVG",
      "column": "duration_ms"
    },
    {
      "column": "column_1"
    }
  ],
  "havings": [
    {
      "calculate_op": "P99",
      "column": "duration_ms",
      "op": "\u003e",
      "value": 1000
    }
  ],
  "limit": 250,
  "time_range": 7200,
  "start_time": 1577836800,
  "granularity": 30
}`

func TestAccDataSourceHoneycombioQuery_validationChecks(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: appendAllTestSteps(
			testStepsQueryValidationChecks_calculation,
			testStepsQueryValidationChecks_filter,
			testStepsQueryValidationChecks_limit(),
			testStepsQueryValidationChecks_time,
			testStepsQueryValidationChecks_having,
		),
	})
}

var testStepsQueryValidationChecks_calculation = []resource.TestStep{
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "COUNT"
    column = "we-should-not-specify-a-column-with-COUNT"
  }
}
`,
		ExpectError: regexp.MustCompile("calculation op COUNT should not have an accompanying column"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "AVG"
  }
}
`,
		ExpectError: regexp.MustCompile("calculation op AVG is missing an accompanying column"),
	},
}

var testStepsQueryValidationChecks_filter = []resource.TestStep{
	{
		Config: `
data "honeycombio_query_specification" "test" {
  filter {
    column = "column"
    op     = "exists"
    value  = "this-value-should-not-be-here"
  }
}
`,
		ExpectError: regexp.MustCompile("filter operation exists must not contain a value"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  filter {
    column = "column"
    op     = ">"
  }
}
`,
		ExpectError: regexp.MustCompile("filter operation > requires a value"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  filter {
    column        = "column"
    op            = ">"
    value_string  = "1"
    value_integer = 10
  }
}
`,
		ExpectError: regexp.MustCompile(multipleValuesError),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  filter {
    column        = "column"
    op            = "in"
    value_integer = 10
  }
}
`,
		ExpectError: regexp.MustCompile("value must be a string if filter op is 'in' or 'not-in'"),
	},
}

func testStepsQueryValidationChecks_limit() []resource.TestStep {
	var queryLimitFmt = `
data "honeycombio_query_specification" "test" {
  limit = %d
}`
	return []resource.TestStep{
		{
			Config:      fmt.Sprintf(queryLimitFmt, 0),
			ExpectError: regexp.MustCompile(`expected limit to be in the range \(1 - 1000\)`),
		},
		{
			Config:      fmt.Sprintf(queryLimitFmt, -5),
			ExpectError: regexp.MustCompile(`expected limit to be in the range \(1 - 1000\)`),
		},
		{
			Config:      fmt.Sprintf(queryLimitFmt, 1200),
			ExpectError: regexp.MustCompile(`expected limit to be in the range \(1 - 1000\)`),
		},
	}
}

var testStepsQueryValidationChecks_time = []resource.TestStep{
	{
		Config: `
data "honeycombio_query_specification" "test" {
  time_range = 7200
  start_time = 1577836800
  end_time   = 1577844000
}
`,
		ExpectError: regexp.MustCompile("specify at most two of time_range, start_time and end_time"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  time_range  = 120
  granularity = 13
}
`,
		ExpectError: regexp.MustCompile("granularity can not be greater than time_range/10"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  time_range  = 60000
  granularity = 59
}
`,
		ExpectError: regexp.MustCompile("granularity can not be less than time_range/1000"),
	},
}

var testStepsQueryValidationChecks_having = []resource.TestStep{
	{
		Config: `
data "honeycombio_query_specification" "test" {
  having {
    calculate_op = "COUNT"
    column       = "we-should-not-specify-a-column-with-COUNT"
    op           = ">"
    value        = 1
  }
}
`,
		ExpectError: regexp.MustCompile("calculate_op COUNT should not have an accompanying column"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  having {
    calculate_op = "CONCURRENCY"
    column       = "we-should-not-specify-a-column-with-CONCURRENCY"
    op           = ">"
    value        = 1
  }
}
`,
		ExpectError: regexp.MustCompile("calculate_op CONCURRENCY should not have an accompanying column"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  having {
    calculate_op = "P99"
    op           = ">="
    value        = 1000
  }
}
`,
		ExpectError: regexp.MustCompile("calculate_op P99 requires a column"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  having {
    calculate_op = "P95"
    op           = ">"
    column       = "duration_ms"
    value        = "850"
  }
}
`,
		ExpectError: regexp.MustCompile("having 0 without matching column in query"),
	},
}

func appendAllTestSteps(steps ...[]resource.TestStep) []resource.TestStep {
	var allSteps []resource.TestStep
	for _, s := range steps {
		allSteps = append(allSteps, s...)
	}
	return allSteps
}

func TestAccDataSourceHoneycombioQuery_filterOpInAndNotIn(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op = "COUNT"
  }

  filter {
    column = "app.tenant"
    op     = "in"
    value  = "foo,bar"
  }
  filter {
    column = "app.tenant"
    op     = "not-in"
    value  = "fzz,bzz"
  }
}
`,
			},
		},
	})
}

func TestAccDataSourceHoneycombioQuery_zerovalue(t *testing.T) {
	properZeroValueJSON := `{
  "calculations": [
    {
      "op": "COUNT"
    }
  ],
  "filters": [
    {
      "column": "duration_ms",
      "op": "\u003e",
      "value": 0
    }
  ],
  "time_range": 7200
}`

	resource.Test(t, resource.TestCase{
		PreCheck:          testAccPreCheck(t),
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op = "COUNT"
  }

  filter {
    column = "duration_ms"
    op     = ">"
    value  = 0
  }
}

output "query_json" {
  value = data.honeycombio_query_specification.test.json
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("query_json", properZeroValueJSON),
				),
			},
		},
	})
}
