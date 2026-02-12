variable "dataset" {
  type = string
}

data "honeycombio_query_specification" "metrics" {
  calculated_field {
    name       = "request_rate_5m"
    expression = "INCREASE($http.server.requests, 300)" # 5-minute range interval
  }

  calculation {
    op     = "AVG"
    column = "request_rate_5m"
  }

  time_range  = 1800
  granularity = 60 # 1-minute time step, but rate is calculated over 5 minutes
}

resource "honeycombio_trigger" "metrics" {
  name    = "High request rate (custom temporal aggregation)"
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
