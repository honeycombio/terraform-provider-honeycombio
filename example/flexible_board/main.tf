terraform {
  required_providers {
    honeycombio = {
      source = "honeycombio/honeycombio"
    }
  }
}

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