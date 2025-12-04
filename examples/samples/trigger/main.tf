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
    op     = "AVG"
    column = "duration_ms"
  }

  filter {
    column = "trace.parent_id"
    op     = "does-not-exist"
  }
  filter {
    column = "app.tenant"
    op     = "="
    value  = "ThatSpecialTenant"
  }

  time_range = 900 // in seconds, 15 minutes
}

resource "honeycombio_trigger" "trigger" {
  name        = "Requests are slower than usual"
  description = "Average duration of all requests for ThatSpecialTenant for the last 15 minutes."

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
