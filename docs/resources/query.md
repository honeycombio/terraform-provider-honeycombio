# Resource: honeycombio_query

Creates a query in a dataset.

Queries can be used by triggers and boards, or be executed via the [Query Data API](https://docs.honeycomb.io/api/query-results/).

-> **Note** Queries can only be created or read. Any changes will result in a new query object being created, and destroying it does nothing.

## Example Usage

```hcl
variable "dataset" {
  type = string
}

data "honeycombio_query_specification" "test_query" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }

  filter {
    column = "duration_ms"
    op     = ">"
    value  = "200"
  }
}

resource "honeycombio_query" "test_query" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.test_query.json
}
```

## Argument Reference

The following arguments are supported:

* `dataset` - (Required) The dataset this query is added to. Use `__all__` for Environment-wide queries.
* `query_json` - (Required) A JSON object describing the query according to the [Query Specification](https://docs.honeycomb.io/api/query-specification/#fields-on-a-query-specification). While the JSON can be constructed manually, it is easiest to use the [`honeycombio_query_specification`](../data-sources/query_specification.md) data source.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the query. Useful for adding it to a board and/or creating a query annotation.

## Import

Querys can be imported using a combination of the dataset name and their ID, e.g.

```
$ terraform import honeycombio_query.my_query my-dataset/bj8BwOa1uRz
```
