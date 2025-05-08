# Resource: honeycombio_flexible_board

Creates a flexible board. For more information about boards, check out [Create Custom Boards](https://docs.honeycomb.io/observe/boards).

## Example Usage

### Simple Flexible Board

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
  description = "overview of my service's performance"

  panel {
    type = "query"
    query_panel {
      query_id            = honeycombio_query.latency_by_userid.id
      query_annotation_id = honeycombio_query_annotation.latency_by_userid.id
    }
    position {
        x_coordinate = 0
        y_coordinate = 0
        width = 3
        height = 3
    }
  }

  panel {
    type = "slo"
    slo_panel {
      slo_id = var.slo_id
    }
    position {
        x_coordinate = 3
        y_coordinate = 3
        width = 3
        height = 3
    }
  }
}
```

## Argument Reference

The following arguments are supported for flexible boards:

* `name` - (Required) Name of the board.
* `description` - (Optional) Description of the board. Supports Markdown.
* `panel` - (Optional) zero or more configurations blocks

Ecach board configuration may have zero or more `panel` blocks which accept the following arguments:

* `type` - (Required) Type of the panel.
* `position` - (Optional) Single configuration block to determine position of the panel.
* `slo_panel` - (Optional) This is only required for `type` slo panels. Single configuration block that contains board slo information.
* `query_panel` - (Optional) This is only required for `type` query panels. Single configuration block that contains board query information.

Each `position` configuration accepts the following arguments:

* `x_coordinate` - (Optional) The x-axis origin point for placing the panel within the layout.
* `y_coordinate` - (Optional) The y-axis origin point for placing the panel within the layout.
* `width` - (Optional) The width of the panel in honeycomb UI columns. Defaults to 6 for queries and 3 for slos. Maximum value is 12.
* `height` - (Optional) The height of the panel in rows. Defaults to 4.

Each `slo_panel` configuration accepts the following arguments:

* `slo_id` the ID of the SLO to add to the board.

Each `query_panel` configuration accepts the following arguments:

* `query_id` - (Required) The ID of the Query to show on the board.
* `query_style` - (Optional) How the query should be displayed within the board, either `graph` (the default), `table` or `combo`.
* `query_annotation_id` - (Required) The ID of the Query Annotation to associate with this query.
* `visualization_settings` - (Optional) A configuration block to manage the query visualization and charts

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - ID of the board.
* `board_url` - The URL to the board in the Honeycomb UI.

## Import

Boards can be imported using their ID, e.g.

```shell
terraform import honeycombio_board.my_board AobW9oAZX71
```

You can find the ID in the URL bar when visiting the board from the UI.
