package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/honeycombio/terraform-provider-honeycombio/internal/helper/test"
)

func TestAcc_QuerySpecificationDataSource_EmptyDefaults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: `
data "honeycombio_query_specification" "test" {}

output "query_json" {
  value = data.honeycombio_query_specification.test.json
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("query_json", `{"calculations":[{"op":"COUNT"}],"time_range":7200}`),
				),
			},
		},
	})
}

func TestAcc_QuerySpecificationDataSource_basic(t *testing.T) {
	// Note: By default go encodes `<` and `>` for html, hence the `\u003e`
	expected, err := test.MinifyJSON(`
{
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
  "calculated_fields": [
    {
      "name": "adhoc_test",
      "expression": "BOOL(1)"
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
      "value": [347338,837359]
    },
    {
      "column": "app.region.name",
      "op": "in",
      "value": [
        "us-west-1",
        "us-west-2"
      ]
    }
  ],
  "filter_combination": "OR",
  "breakdowns": ["column_1"],
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
}`)
	require.NoError(t, err)
	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: `
data "honeycombio_query_specification" "test" {
    calculation {
        op     = "AVG"
        column = "duration_ms"
    }
    calculation {
        op     = "P99"
        column = "duration_ms"
    }

    calculated_field {
      name       = "adhoc_test"
      expression = "BOOL(1)"
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
        column = "duration_ms"
        op     = "<"
        value  = 100
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
    filter {
        column = "app.region.name"
        op     = "in"
        value  = "us-west-1,us-west-2"
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

    limit                       = 250
    time_range                  = 7200
    start_time                  = 1577836800
    granularity                 = 30
}

output "query_json" {
    value = data.honeycombio_query_specification.test.json
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("query_json", expected),
				),
			},
		},
	})
}

func TestAcc_QuerySpecificationDataSource_validationChecks(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: appendAllTestSteps(
			testStepsQueryValidationChecks_calculation,
			testStepsQueryValidationChecks_calculationFilters,
			testStepsQueryValidationChecks_formulas,
			testStepsQueryValidationChecks_filter,
			testStepsQueryValidationChecks_limit(),
			testStepsQueryValidationChecks_time,
			testStepsQueryValidationChecks_having,
			testStepsQueryValidationChecks_order,
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
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("column is not allowed with operator COUNT"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "AVG"
  }
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("AVG requires a colum"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "HEATMAP"
    column = "duration_ms"
  }
  calculation {
    op   = "COUNT"
    name = "total"
  }
  formula {
    name       = "rate"
    expression = "DIV($total, 100)"
  }
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("HEATMAP calculations cannot be used with formulas"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op   = "COUNT"
    name = "my_count"
  }
  calculation {
    op     = "AVG"
    column = "duration_ms"
    name   = "my_count"
  }
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile(`duplicate name.*already used by calculation`),
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
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("exists does not take a value"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  filter {
    column = "column"
    op     = ">"
  }
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("operator > requires a value"),
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
			PlanOnly:    true,
			ExpectError: regexp.MustCompile("limit value must be between 1 and 1000"),
		},
		{
			Config:      fmt.Sprintf(queryLimitFmt, -5),
			PlanOnly:    true,
			ExpectError: regexp.MustCompile("limit value must be between 1 and 1000"),
		},
		{
			Config:      fmt.Sprintf(queryLimitFmt, 1200),
			PlanOnly:    true,
			ExpectError: regexp.MustCompile("limit value must be between 1 and 1000"),
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
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("specify at most two of time_range, start_time and end_time"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  time_range  = 120
  granularity = 121
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("granularity can not be greater than time_range"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  time_range  = 60000
  granularity = 59
}`,
		PlanOnly:    true,
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
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("COUNT should not have an accompanying column"),
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
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("CONCURRENCY should not have an accompanying column"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  having {
    calculate_op = "P99"
    op           = ">="
    value        = 1000
  }
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("P99 requires a column"),
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
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("P95 missing matching calculation"),
	},
}

var testStepsQueryValidationChecks_order = []resource.TestStep{
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op = "COUNT"
  }

  order {
    op    = "COUNT"
    order = "descending"
  }
}`,
		PlanOnly: true,
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  order {
    op    = "COUNT"
    order = "descending"
  }
}`,
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }
}`,
		PlanOnly: true,
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }

  breakdowns = ["app.tenant"]

  order {
    op     = "AVG"
    column = "duration_ms"
    order  = "ascending"
  }

  order {
    column = "app.tenant"
    order  = "descending"
  }
}`,
		PlanOnly: true,
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "MAX"
    column = "duration_ms"
  }

  order {
    op    = "COUNT"
    order = "ascending"
  }
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("missing matching calculation, formula, or breakdown"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "MAX"
    column = "duration_ms"
  }

  breakdowns = ["app.tenant"]

  order {
    column = "column_1"
    order  = "ascending"
  }
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("missing matching calculation, formula, or breakdown"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "HEATMAP"
    column = "duration_ms"
  }

  breakdowns = ["app.tenant"]

  order {
    op     = "HEATMAP"
    column = "duration_ms"
    order  = "ascending"
  }
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("cannot order by HEATMAP"),
	},
}

func appendAllTestSteps(steps ...[]resource.TestStep) []resource.TestStep {
	var allSteps []resource.TestStep
	for _, s := range steps {
		allSteps = append(allSteps, s...)
	}
	return allSteps
}

func TestAcc_QuerySpecificationDataSource_filterOpInAndNotIn(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
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
}`,
				PlanOnly: true,
			},
		},
	})
}

func TestAcc_QuerySpecificationDataSource_zerovalue(t *testing.T) {
	// Note: By default go encodes `<` and `>` for html, hence the `\u003e`
	expected, err := test.MinifyJSON(`
{
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
  "time_range": 7200,
  "granularity": 0
}`)
	require.NoError(t, err)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
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

  granularity = 0
}

output "query_json" {
  value = data.honeycombio_query_specification.test.json
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("query_json", expected),
				),
			},
		},
	})
}

func TestAcc_QuerySpecificationDataSource_CompareTimeOffsetSecondsValid(t *testing.T) {
	expected, err := test.MinifyJSON(`
{
  "calculations": [
    {
      "op": "COUNT"
    }
  ],
  "filters": [
    {
      "column": "message",
      "op": "contains",
      "value": "a"
    },
    {
      "column": "message",
      "op": "contains",
      "value": "b"
    }
  ],
  "time_range": 7200,
  "granularity": 0,
  "compare_time_offset_seconds": 7200
}`)
	require.NoError(t, err)

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op = "COUNT"
  }

  dynamic "filter" {
    for_each = toset(["a", "b"])
    content {
      column = "message"
      op     = "contains"
      value  = filter.value
    }
  }

  granularity = 0
  compare_time_offset = 7200
}

output "query_json" {
  value = data.honeycombio_query_specification.test.json
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("query_json", expected),
				),
			},
		},
	})
}

func TestAcc_QuerySpecificationDataSource_formulas(t *testing.T) {
	expected, err := test.MinifyJSON(`
{
  "calculations": [
    {"op": "COUNT", "name": "total"},
    {"op": "COUNT", "name": "errors", "filters": [{"column": "status_code", "op": "\u003e=", "value": 500}]}
  ],
  "formulas": [
    {"name": "error_rate", "expression": "DIV($errors, $total)"}
  ],
  "time_range": 7200
}`)
	require.NoError(t, err)
	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op   = "COUNT"
    name = "total"
  }
  calculation {
    op   = "COUNT"
    name = "errors"
    filter {
      column = "status_code"
      op     = ">="
      value  = 500
    }
  }
  formula {
    name       = "error_rate"
    expression = "DIV($errors, $total)"
  }
}

output "query_json" {
  value = data.honeycombio_query_specification.test.json
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("query_json", expected),
				),
			},
		},
	})
}

func TestAcc_QuerySpecificationDataSource_calculationFilters(t *testing.T) {
	expected, err := test.MinifyJSON(`
{
  "calculations": [
    {
      "op": "AVG",
      "column": "duration_ms",
      "name": "avg_duration",
      "filters": [
        {"column": "status", "op": "=", "value": "success"},
        {"column": "region", "op": "in", "value": ["us-west-1", "us-east-1"]}
      ],
      "filter_combination": "OR"
    }
  ],
  "time_range": 7200
}`)
	require.NoError(t, err)
	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
    name   = "avg_duration"
    filter_combination = "OR"
    filter {
      column = "status"
      op     = "="
      value  = "success"
    }
    filter {
      column = "region"
      op     = "in"
      value  = "us-west-1,us-east-1"
    }
  }
}

output "query_json" {
  value = data.honeycombio_query_specification.test.json
}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("query_json", expected),
				),
			},
		},
	})
}

func TestAcc_QuerySpecificationDataSource_orderByFormula(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op   = "COUNT"
    name = "total"
  }
  calculation {
    op   = "COUNT"
    name = "errors"
  }
  formula {
    name       = "error_rate"
    expression = "DIV($errors, $total)"
  }
  order {
    column = "error_rate"
    order  = "descending"
  }
}`,
				PlanOnly: true,
			},
		},
	})
}

func TestAcc_QuerySpecificationDataSource_orderByNamedCalculation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op   = "COUNT"
    name = "my_count"
  }
  order {
    column = "my_count"
    order  = "descending"
  }
}`,
				PlanOnly: true,
			},
		},
	})
}

var testStepsQueryValidationChecks_calculationFilters = []resource.TestStep{
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op = "COUNT"
    filter {
      column = "status"
      op     = "="
      value  = "error"
    }
  }
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("name is required when using calculation filters"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op   = "COUNT"
    name = "filtered_count"
    filter {
      column = "status"
      op     = "exists"
      value  = "should-not-be-here"
    }
  }
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("exists does not take a value"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op   = "COUNT"
    name = "filtered_count"
    filter {
      column = "status"
      op     = "="
    }
  }
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("operator = requires a value"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op   = "COUNT"
    name = "filtered_count"
    filter {
      column = "root.status_code"
      op     = "="
      value  = "500"
    }
  }
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("relational fields are not supported in calculation filters"),
	},
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op   = "COUNT"
    name = "filtered_count"
    filter {
      column = "child.duration_ms"
      op     = ">"
      value  = "100"
    }
  }
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("relational fields are not supported in calculation filters"),
	},
}

var testStepsQueryValidationChecks_formulas = []resource.TestStep{
	// Duplicate formula names not allowed
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op   = "COUNT"
    name = "total"
  }
  calculation {
    op   = "COUNT"
    name = "errors"
  }
  formula {
    name       = "rate"
    expression = "DIV($total, 100)"
  }
  formula {
    name       = "rate"
    expression = "DIV($errors, 100)"
  }
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile(`duplicate name.*already used by formula`),
	},
	// Formula name cannot match calculation name
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op   = "COUNT"
    name = "total"
  }
  formula {
    name       = "total"
    expression = "DIV($total, 100)"
  }
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile(`duplicate name.*already used by calculation`),
	},
	// Relational fields in filters not allowed when using formulas
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op   = "COUNT"
    name = "total"
  }
  formula {
    name       = "rate"
    expression = "DIV($total, 100)"
  }
  filter {
    column = "root.status_code"
    op     = "="
    value  = "200"
  }
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("relational fields are not supported when using formulas or calculation filters"),
	},
	// Relational fields in breakdowns not allowed when using formulas
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op   = "COUNT"
    name = "total"
  }
  formula {
    name       = "rate"
    expression = "DIV($total, 100)"
  }
  breakdowns = ["child.service_name"]
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("relational fields are not supported when using formulas or calculation filters"),
	},
	// Relational fields in filters not allowed when using calculation filters
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op   = "COUNT"
    name = "filtered"
    filter {
      column = "status"
      op     = "="
      value  = "error"
    }
  }
  filter {
    column = "parent.trace_id"
    op     = "exists"
  }
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("relational fields are not supported when using formulas or calculation filters"),
	},
	// Relational fields in breakdowns not allowed when using calculation filters
	{
		Config: `
data "honeycombio_query_specification" "test" {
  calculation {
    op   = "COUNT"
    name = "filtered"
    filter {
      column = "status"
      op     = "="
      value  = "error"
    }
  }
  breakdowns = ["root.service_name"]
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("relational fields are not supported when using formulas or calculation filters"),
	},
}

func TestAcc_QuerySpecificationDataSource_TimeRange_CompareTimeOffsetInvalid(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: `
data "honeycombio_query_specification" "less_than_time_range" {
  calculation {
    op = "COUNT"
  }

  filter {
    column = "duration_ms"
    op     = "="
    value  = 0
  }

  start_time = 1714738800
  compare_time_offset = 1800
}

output "query_json" {
  value = data.honeycombio_query_specification.less_than_time_range.json
}`,
				ExpectError: regexp.MustCompile("compare_time_offset must be greater than the queries time range"),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		PreCheck:                 testAccPreCheck(t),
		ProtoV5ProviderFactories: testAccProtoV5MuxServerFactory,
		Steps: []resource.TestStep{
			{
				Config: `
data "honeycombio_query_specification" "invalid_compare_time_offset" {
  calculation {
    op = "COUNT"
  }

  filter {
    column = "duration_ms"
    op     = "="
    value  = 0
  }

  compare_time_offset = 1
}

output "query_json" {
  value = data.honeycombio_query_specification.invalid_compare_time_offset.json
}`,
				ExpectError: regexp.MustCompile("compare_time_offset is an invalid value"),
			},
		},
	})
}
