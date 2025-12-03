variable "dataset" {
  type = string
}

data "honeycombio_query_specification" "query" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }

  calculation {
    op = "COUNT"
  }

  // Only returns results where more than 100 events were received. Windows with
  // less than 100 events will yield zero, and this trigger will consider them
  // "ok"
  having {
    calculate_op = "COUNT"
    op = ">"
    value = 100
  }

  time_range = 900 // in seconds, 15 minutes
}

resource "honeycombio_trigger" "trigger" {
  name        = "Common requests are slower than usual"
  description = "Average duration of common requests is slower than expected for the last 15 minutes."

  disabled = false

  query_json = data.honeycombio_query_specification.query.json
  dataset    = var.dataset

  frequency = 900 // in seconds, 15 minutes

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
    target = "Trigger - slow requests" # name of the marker
  }
}
