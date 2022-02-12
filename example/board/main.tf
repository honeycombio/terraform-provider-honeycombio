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

data "honeycombio_query_specification" "query" {
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

  breakdowns = ["app.tenant"]
}

resource "honeycombio_query" "query" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.query.json
}

resource "honeycombio_board" "board" {
  name        = "Request Latency"
  description = "Latencies of all requests by Tenant for the last 15 minutes."
  style       = "list"

  query {
    caption  = "Latency"
    dataset  = var.dataset
    query_id = honeycombio_query.query.id
  }
}
