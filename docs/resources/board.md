# Resource: honeycombio_board

Creates a board. For more information about boards, check out [Collaborate with Boards](https://docs.honeycomb.io/working-with-your-data/collaborating/boards/#docs-sidebar).

## Example Usage

```hcl
variable "dataset" {
  type = string
}

locals {
  percentiles = ["P50", "P75", "P90", "P95"]
}

data "honeycombio_query" "query" {
  count = length(local.percentiles)

  calculation {
    op     = local.percentiles[count.index]
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

resource "honeycombio_board" "board" {
  name        = "Request percentiles"
  description = "${join(", ", local.percentiles)} of all requests for ThatSpecialTenant for the last 15 minutes."
  style       = "list"

  //Use dynamic config blocks to generate a query for each of the percentiles we're interested in
  dynamic "query" {
    for_each = local.percentiles

    content {
      caption     = query.value
      query_style = "combo"
      dataset     = var.dataset
      query_json  = data.honeycombio_query.query[query.key].json
    }
  }
}
```

## Example with Query IDs and Annotations

```hcl
resource "honeycombio_query" "example" {
  dataset    = "my-traces"
  query_json = file("${path.cwd}/board-queries/example.hny")
}

resource "honeycombio_query_annotation" "example" {
  dataset     = "my-traces"
  query_id    = honeycombio_query.example.id
  name        = "My Example Query"
  description = "My Helpful Description"
}

resource "honeycombio_board" "example" {
  name  = "My example board"
  style = "list"

  query {
    query_id            = honeycombio_query.example.id
    query_annotation_id = honeycombio_query_annotation.example.id
    query_style         = "combo"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the board.
* `description` - (Optional) Description of the board. Supports markdown.
* `style` - (Optional) How the board should be displayed in the UI, either `list` (the default) or `visual`.
* `query` - (Optional) Zero or more configurations blocks (described below) with the queries of the board.

Each board configuration may have zero or more `query` blocks, which accepts the following arguments:

* `query_json` - (Optional) A JSON object describing the query according to the [Query Specification](https://docs.honeycomb.io/api/query-specification/#fields-on-a-query-specification). While the JSON can be constructed manually, it is easiest to use the [`honeycombio_query`](../data-sources/query.md) data source.
* `query_id` - (Optional) The ID of the Query to run.
* `query_annotation_id` - (Optional) The ID of the Query Annotation to associate with this query.
* `dataset` - (Required) The dataset this query is associated with.
* `caption` - (Optional) A description of the query that will be displayed on the board. Supports markdown.
* `query_style` - (Optional) How the query should be displayed within the board, either `graph` (the default), `table` or `combo`.

~> **NOTE** One of `query_id` or `query_json` is required.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the board.

## Import

Boards can be imported using their ID, e.g.

```
$ terraform import honeycombio_board.my_board AobW9oAZX71
```

You can find the ID in the URL bar when visiting the board from the UI.
