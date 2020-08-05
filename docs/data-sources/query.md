# Data Source: honeycombio_query

Construct a query that can be used in triggers. For more information about the query specification, check out [Query Specification](https://docs.honeycomb.io/api/query-specification/).

The `json` attribute contains a serialized JSON representation which can be passed to the `query_json` field of `honeycombio_trigger`.

## Example Usage

```hcl
data "honeycombio_query" "example" {
  # zero or more calculation blocks
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }

  # zero or more filter blocks
  filter {
    column = "trace.parent_id"
    op     = "does-not-exist"
  }

  filter {
    column = "app.tenant"
    op     = "="
    value  = "ThatSpecialTenant" 
  }

  filter_combination = "AND"

  breakdowns = ["app.tenant"]
}

output "json_query" {
    value = data.honeycombio_query.example.json
}
```

## Argument Reference

The following arguments are supported:

* `calculation` - (Optional) Zero or more configuration blocks (described below) with the calculations that should be displayed. If no calculations are specified, `COUNT` will be used.
* `filter` - (Optional) Zero or more configuration blocks (described below) with the filters that should be applied.
* `filter_combination` - (Optional) How to combine multiple filters, either `AND` (default) or `OR`.
* `breakdowns` - (Optional) A list of fields to group by.

Each query configuration may have zero or more `calculation` blocks, which each accept the following arguments:

* `op` - (Required) The operator to apply, see the supported list of calculation operators at [Calculation Operators](https://docs.honeycomb.io/api/query-specification/#calculation-operators).
* `column` - (Optional) The column to apply the operator to, not needed with `COUNT`.

Each query configuration may have zero or more `filter` blocks, which each accept the following arguments:

* `column` - (Required) The column to apply the filter to.
* `op` - (Required) The operator to apply, see the supported list of filter operators at [Filter Operators](https://docs.honeycomb.io/api/query-specification/#filter-operators).
* `value` - (Optional) The value to be used with the operator, not all operators require a value.


## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the trigger.
* `json` - JSON representation of the query according to the [Query Specification](https://docs.honeycomb.io/api/query-specification/#fields-on-a-query-specification), can be used as input for other resources.
