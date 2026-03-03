variable "dataset" {
  type = string
}

data "honeycombio_query_specification" "query" {
  calculation {
    op   = "COUNT"
    name = "total"
  }

  calculation {
    op   = "COUNT"
    name = "errors"

    filter {
      column = "error"
      op     = "exists"
    }
  }

  formula {
    name       = "error_rate"
    expression = "DIV($errors, $total)"
  }

  time_range = 900 // in seconds, 15 minutes
}

resource "honeycombio_trigger" "trigger" {
  name        = "Error rate is too high"
  description = "The error rate has exceeded the threshold for the last 15 minutes."

  query_json = data.honeycombio_query_specification.query.json
  dataset    = var.dataset

  frequency = 900 // in seconds, 15 minutes

  threshold {
    op    = ">"
    value = 0.1
  }

  recipient {
    type   = "email"
    target = "hello@example.com"
  }
}
