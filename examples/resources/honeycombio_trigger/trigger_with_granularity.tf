variable "dataset" {
  type = string
}

data "honeycombio_query_specification" "metrics" {
  calculation {
    op     = "P95"
    column = "http.server.request.duration"
  }

  time_range  = 1800
  granularity = 300 # Custom granularity only available with Metrics
}

resource "honeycombio_trigger" "metrics" {
  name    = "High request duration"
  dataset = var.dataset

  query_json = data.honeycombio_query_specification.metrics.json

  frequency = 900

  threshold {
    op    = ">"
    value = 1000
  }

  recipient {
    type   = "email"
    target = "hello@example.com"
  }
}
