# Resource: honeycombio_flexible_board

Creates a flexible board. For more information about boards, check out [Create Custom Boards](https://docs.honeycomb.io/observe/boards).

## Example Usage

### Simple Flexible Board

```hcl
variable "dataset" {
  type = string
}

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

resource "honeycombio_derived_column" "request_latency_sli" {
  alias       = "sli.request_latency"
  description = "SLI: request latency less than 300ms"
  dataset     = var.dataset

  # heredoc also works
  expression = file("../sli/sli.request_latency.honeycomb")

  lifecycle {
    # in order to avoid potential conflicts with renaming the derived column
    # while in use by the SLO, we set create_before_destroy to true
    create_before_destroy = true
  }
}

resource "honeycombio_slo" "slo" {
  name              = "Latency SLO"
  description       = "example SLO"
  dataset           = var.dataset
  sli               = honeycombio_derived_column.request_latency_sli.alias
  target_percentage = 99.9
  time_period       = 30

  tags = {
    team = "web"
  }
}

resource "honeycombio_flexible_board" "overview" {
  name        = "Service Overview"
  description = "My flexible baord description"

  panel {
    type = "query"

    position {
      x_coordinate = 0
      y_coordinate = 0
      width        = 6
      height       = 6
    }

    query_panel {
      query_id            = honeycombio_query.latency_by_userid.id
      query_annotation_id = honeycombio_query_annotation.latency_by_userid.id
      query_style         = "combo"
      visualization_settings {
        use_utc_xaxis = true
        chart {
          chart_type          = "line"
          chart_index         = 0
          omit_missing_values = true
          use_log_scale       = true
        }
      }
    }
  }

  panel {
    type = "slo"
    slo_panel {
      slo_id = honeycombio_slo.slo.id
    }
  }
}
```

## Argument Reference

The following arguments are supported for flexible boards:

- `name` - (Required) Name of the board.
- `description` - (Optional) Description of the board. Supports Markdown.
- `panel` - (Optional) zero or more configurations blocks

Each board configuration may have zero or more `panel` blocks which accept the following arguments:

- `position` - (Optional) Single configuration block to determine position of the panel.
- `slo_panel` - (Optional) This is only required for `type` slo panels. Single configuration block that contains board slo information.
- `query_panel` - (Optional) This is only required for `type` query panels. Single configuration block that contains board query information.

Each `position` configuration accepts the following arguments:

- `x_coordinate` - (Optional) The x-axis origin point for placing the panel within the layout.
- `y_coordinate` - (Optional) The y-axis origin point for placing the panel within the layout.
- `width` - (Optional) The width of the panel in honeycomb UI columns. Defaults to 6 for queries and 3 for slos. Maximum value is 12.
- `height` - (Optional) The height of the panel in rows. Defaults to 4.

Each `slo_panel` configuration accepts the following arguments:

- `slo_id` the ID of the SLO to add to the board.

Each `query_panel` configuration accepts the following arguments:

- `query_id` - (Required) The ID of the Query to show on the board.
- `query_style` - (Optional) How the query should be displayed within the board, either `graph` (the default), `table` or `combo`.
- `query_annotation_id` - (Required) The ID of the Query Annotation to associate with this query.
- `visualization_settings` - (Optional) A configuration block to manage the query visualization and charts.

Each `visualization_settings` configuration accepts the following arguments:

- `use_utc_xaxis` - (Optional) Display UTC Time X-Axis or Localtime X-Axis.
- `hide_markers` - (Optional) Hide [markers](https://docs.honeycomb.io/investigate/query/customize-results/#markers) from appearing on graph.
- `hide_hovers` - (Optional) Disable Graph tooltips in the results display when hovering over a graph.
- `overlaid_charts` - (Optional) Combine any visualized AVG, MIN, MAX, and PERCENTILE clauses into a single chart.
- `chart` - (Optional) a configuration block to manage the query's charts.

Each `chart` configuration accepts the following arguments:

- `chart_type` - (Optional) the type of chart to render. Some example values: `line`, `tsbar`, `stacked`, `stat`, `tsbar`, `cpie`, `cbar`. Default to `default`
- `chart_index` - (Optional) index of the charts this configuration corresponds to.
- `omit_missing_values` - (Optional) Interpolates between points when the intervening time buckets have no matching events. Use to display a continuous line graph with no drops to zero.
- `use_log_scale` - (Optional) Use logarithmic scale on Y axis. The y-axis of a Log Scale graph increases exponentially. Useful for data with an extremely large range of values.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- `id` - ID of the board.
- `board_url` - The URL to the board in the Honeycomb UI.

## Import

Boards can be imported using their ID, e.g.

```shell
terraform import honeycombio_flexible_board.my_board AobW9oAZX71
```

You can find the ID in the URL bar when visiting the board from the UI.
