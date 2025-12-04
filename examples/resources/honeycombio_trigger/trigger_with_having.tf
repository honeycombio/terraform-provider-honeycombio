variable "dataset" {
  type = string
}

data "honeycombio_query_specification" "query" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }

  // `having`'s must have a matching `calculation`. This won't be used as the
  // trigger's threshold since it matches the `having` - `AVG(duration_ms)` will
  // be used for the threshold. To use a `having` to restrict the trigger's
  // threshold, omit the second (different) `calculation`.
  calculation {
    op = "MAX"
    column = "retries"
  }

  filter {
    column = "error.type"
    op = "exists"
  }

  // Only returns results with at least one retry
  having {
    calculate_op = "MAX"
    column = "retries"
    op = ">"
    value = 0
  }

  time_range = 900 // in seconds, 15 minutes
}

resource "honeycombio_trigger" "trigger" {
  name        = "Retried errors are slower than usual"
  description = "Average duration of requests with errors and at least one retry is slower than expected for the last 15 minutes."

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
