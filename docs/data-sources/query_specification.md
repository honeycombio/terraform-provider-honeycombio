# Data Source: honeycombio_query_specification

Generates a [Query Specification](https://docs.honeycomb.io/api/query-specification/) in JSON format for use with resources that expect a JSON-formatted Query Specification like [`honeycombio_query`](../resources/query.md).

Using this data source to generate query specifications is optional.
It is also valid to use literal JSON strings in your configuration or to use the file interpolation function to read a raw JSON query specification from a file.

## Example Usage

```hcl
data "honeycombio_query_specification" "example" {
  # zero or more calculation blocks
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }

  calculated_field {
    name       = "fast_enough"
    expression = "LTE($response.duration_ms, 200)"
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

  filter {
    column = "fast_enough"
    op     = "="
    value  = false
  }

  filter_combination = "AND"

  breakdowns = ["app.tenant"]
    
  time_range = 28800 // in seconds, 8 hours
  
  compare_time_offset = 86400 // in seconds, compare with data from 1 day ago
}

output "json_query" {
    value = data.honeycombio_query_specification.example.json
}
```

## Argument Reference

The following arguments are supported:

* `calculation` - (Optional) Zero or more configuration blocks (described below) with the calculations that should be displayed. If no calculations are specified, `COUNT` will be used.
* `calculated_field` - (Optional) Zero or more configuration blocks (described below) of inline Calculated Fields.
* `filter` - (Optional) Zero or more configuration blocks (described below) with the filters that should be applied.
* `filter_combination` - (Optional) How to combine multiple filters, either `AND` (default) or `OR`.
* `breakdowns` - (Optional) A list of fields to group by.
* `order` - (Optional) Zero or more configuration blocks (described below) describing how to order the query results. Each term must appear in either `calculation` or `breakdowns`.
* `having` - (Optional) Zero or more filters used to restrict returned groups in the query result.
* `limit` - (Optional)  The maximum number of query results, must be between 1 and 1000.
* `time_range` - (Optional) The time range of the query in seconds, defaults to `7200` (two hours).
* `start_time` - (Optional) The absolute start time of the query in Unix Time (= seconds since epoch).
* `end_time` - (Optional) The absolute end time of the query in Unix Time (= seconds since epoch).
* `granularity` - (Optional) The time resolution of the query’s graph, in seconds. Valid values must be in between the query’s time range /10 at maximum, and /1000 at minimum.
* `compare_time_offset` - (Optional) The time offset for comparison queries, in seconds. Used to compare current time range data with data from a previous time period. Valid values are the query time range, `1800`, `3600`, `7200`, `28800`, `86400`, `604800`, `2419200`, or `15724800`.

~> **NOTE** It is not allowed to specify all three of `time_range`, `start_time` and `end_time`. For more details about specifying time windows, check [Query specification: A caveat on time](https://docs.honeycomb.io/api/query-specification/#a-caveat-on-time).

Each query configuration may have zero or more `calculation` blocks, which each accept the following arguments:

* `op` - (Required) The operator to apply, see the supported list of calculation operators at [Calculation Operators](https://docs.honeycomb.io/api/query-specification/#calculation-operators).
* `column` - (Optional) The column to apply the operator to, not needed with `COUNT` or `CONCURRENCY`.

Each query configuration may have zero or more `calculated_field` blocks, which each accept the following arguments:

* `name` - (Required) The name of this Temporary Calculated Field.
* `expression` - (Required) The formula for your Calculated Field. To learn more about syntax and available functions, and to explore some example formulas, visit the [Calculated Field Formula Reference](https://docs.honeycomb.io/reference/derived-column-formula/).

Each query configuration may have zero or more `filter` blocks, which each accept the following arguments:

* `column` - (Required) The column to apply the filter to.
* `op` - (Required) The operator to apply, see the supported list of filter operators at [Filter Operators](https://docs.honeycomb.io/api/query-specification/#filter-operators). Not all operators require a value.
* `value` - (Optional) The value used for the filter. Not needed if op is `exists` or `not-exists`. Mutually exclusive with the other `value_*` options.

* -> **NOTE** Filter op `in` and `not-in` expect an array of strings as value. Use the `value` attribute and pass the values in single string separated by `,` without additional spaces (similar to the query builder in the UI). For example: the list `foo`, `bar` becomes `foo,bar`.

Each query configuration may have zero or more `order` blocks, which each accept the following arguments. An order term can refer to a `calculation` or a column set in `breakdowns`. When referring to a `calculation`, `op` and `column` must be the same as the calculation.

* `op` - (Optional) The calculation operator to apply, see the supported list of calculation operators at [Calculation Operators](https://docs.honeycomb.io/api/query-specification/#calculation-operators).
* `column` - (Optional) Either a column present in `breakdown` or a column that `op` applies to.
* `order` - (Optional) The sort direction, if set must be `ascending` or `descending`.

Each query configuration may have zero or more `having` blocks, which each accept the following arguments.

* `op` - The operator to apply to filter the query results. One of `=`, `!=`, `>`, `>=`, `<`, or `<=`.
* `calculate_op` - The calculation operator to apply, supports all of the [Calculation Operators](https://docs.honeycomb.io/api/query-specification/#calculation-operators) with the exception of `HEATMAP`.
* `column` - The column to apply the `calculate_op` to, not needed with `COUNT` or `CONCURRENCY`.
* `value` - The value used with `op`. Currently assumed to be a number.

~> **NOTE** A having term's `column`/`calculate_op` pair must have a corresponding `calculation`. There can be multiple `having` blocks for the same `column`/`calculate_op` pair.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the query specification.
* `json` - JSON representation of the query according to the [Query Specification](https://docs.honeycomb.io/api/query-specification/#fields-on-a-query-specification), can be used as input for other resources.
