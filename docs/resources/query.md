# Resource: honeycombio_query

Creates a Query scoped to a Dataset or Environment.

Queries can be used by Triggers and Boards, or be executed via the [Query Data API](https://docs.honeycomb.io/api/query-results/).

-> **API Keys** Note that this requires a [v1 API Key](https://registry.terraform.io/providers/honeycombio/honeycombio/latest/docs#v1-apis)

-> Queries are immutable and can not be deleted -- only created or read.
  Any changes will result in a new query object being created.

## Example Usage

```hcl
variable "dataset" {
  type = string
}

resource "honeycombio_derived_column" "duration_ms_log10" {
  alias       = "duration_ms_log10"
  expression  = "LOG10($duration_ms)"
  description = "LOG10 of duration_ms"

  dataset = var.dataset
}

data "honeycombio_query_specification" "example" {
  calculation {
    op     = "P90"
    column = "duration_ms"
  }

  calculation {
    op     = "HEATMAP"
    column = $honeycombio_derived_column.duration_ms_log10.alias
  }

  filter {
    column = "duration_ms"
    op     = "exists"
  }

}

resource "honeycombio_query" "example" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.example.json

  lifecycle {
    replace_triggered_by = [
      // re-create the query if the derived column is changed
      // to ensure we're using the latest definition
      honeycombio_derived_column.duration_ms_log10
    ]
  }
}
```

-> **Note** If you are referencing a [Derived Column](derived_column.md) in your query and want to ensure you are always using the latest definition
  of the derived column you should use the [replace_triggered_by](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#replace_triggered_by)
  lifecycle argument as shown in the example above.

## Argument Reference

The following arguments are supported:

* `dataset` - (Optional) The dataset this query is scoped to.  If not set, an Environment-wide query will be created.
* `query_json` - (Required) A JSON object describing the query according to the [Query Specification](https://docs.honeycomb.io/api/query-specification/#fields-on-a-query-specification).
  While the JSON can be constructed manually, using the [`honeycombio_query_specification`](../data-sources/query_specification.md) data source provides deeper validation.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the query. Useful for adding it to a board and/or creating a query annotation.

## Import

Querys can be imported using a combination of the dataset name and their ID, e.g.

```
$ terraform import honeycombio_query.my_query my-dataset/bj8BwOa1uRz
```
