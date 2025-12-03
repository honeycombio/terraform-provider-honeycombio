variable "dataset" {
  type = string
}

data "honeycombio_query_specification" "example" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }

  filter {
    column = "trace.parent_id"
    op     = "does-not-exist"
  }

  time_range = 1800
}

resource "honeycombio_trigger" "example" {
  name        = "Requests are slower than usual"
  description = "Average duration of all requests for the last 10 minutes."

  query_json = data.honeycombio_query_specification.example.json
  dataset    = var.dataset

  frequency = 600 // in seconds, 10 minutes

  alert_type = "on_change" // on_change is default, on_true can refers to the "Alert on True" checkbox in the UI

  threshold {
    op    = ">"
    value = 1000
  }

  # zero or more recipients
  recipient {
    type   = "email"
    target = "hello@example.com"
  }

  recipient {
    type   = "marker"
    target = "Trigger - requests are slow"
  }
}
