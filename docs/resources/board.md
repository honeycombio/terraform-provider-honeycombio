# Resource: honeycombio_board

Creates a board. For more information about boards, check out [Collaborate with Boards](https://docs.honeycomb.io/working-with-your-data/collaborating/boards/#docs-sidebar).

## Example Usage

### Simple Board

```hcl
data "honeycombio_query_specification" "query" {
  calculation {
    op     = "P99"
    column = "duration_ms"
  }

  filter {
    column = "trace.parent_id"
    op     = "does-not-exist"
  }

  breakdowns = ["app.tenant"]
}

resource "honeycombio_query" "query" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.query.json
}

resource "honeycombio_board" "board" {
  name        = "My Board"

  query {
    dataset  = var.dataset
    query_id = honeycombio_query.query.id
  }
}
```

### Annotated Board

```hcl
data "honeycombio_query_specification" "latency_by_userid" {
  time_range = 86400
  breakdowns = ["app.user_id"]

  calculation {
    op     = "HEATMAP"
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

  order {
    column = "duration_ms"
    op     = "P99"
    order  = "descending"
  }
}

resource "honeycombio_query" "latency_by_userid" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.latency_by_userid.json
}

resource "honeycombio_query_annotation" "latency_by_userid" {
  dataset     = var.dataset
  query_id    = honeycombio_query.latency_by_userid.id
  name        = "Latency by User"
  description = "A breakdown of trace latency by User over the last 24 hours"
}

resource "honeycombio_board" "overview" {
  name        = "Service Overview"
  style       = "visual"

  query {
    caption             = "Latency by User"
    query_id            = honeycombio_query.latency_by_userid.id
    query_annotation_id = honeycombio_query_annotation.latency_by_userid.id
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

* `query_id` - (Required) The ID of the Query to run.
* `query_annotation_id` - (Optional) The ID of the Query Annotation to associate with this query.
* `dataset` - (Deprecated) The dataset this query is associated with.
* `caption` - (Optional) A description of the query that will be displayed on the board. Supports markdown.
* `query_style` - (Optional) How the query should be displayed within the board, either `graph` (the default), `table` or `combo`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the board.

## Import

Boards can be imported using their ID, e.g.

```
$ terraform import honeycombio_board.my_board AobW9oAZX71
```

You can find the ID in the URL bar when visiting the board from the UI.
