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

resource "honeycombio_board" "overview" {
  name        = "Service Overview"
  style       = "visual"
  description = <<EOT
Helpful queries to get an overview of our service overall health and performance.
Useful as a jumping off point for BubbleUp or a quick investigation.

See the [wiki](https://wiki.company.internal.tld) for more information.
EOT

  query {
    caption             = "Latency by User"
    query_id            = honeycombio_query.latency_by_userid.id
    query_annotation_id = honeycombio_query_annotation.latency_by_userid.id
    query_style         = "graph"
  }
}
