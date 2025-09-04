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
  granularity = 13
}`,
		PlanOnly:    true,
		ExpectError: regexp.MustCompile("granularity can not be greater than time_range/10"),
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
		ExpectError: regexp.MustCompile("missing matching calculation or breakdown"),
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
		ExpectError: regexp.MustCompile("missing matching calculation or breakdown"),
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
      "column": "duration_ms",
      "op": "=",
      "value": 0
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

  filter {
    column = "duration_ms"
    op     = "="
    value  = 0
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
