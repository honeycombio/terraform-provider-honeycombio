provider "honeycombio" {
  # You can also set the environment variable HONEYCOMBIO_APIKEY
  api_key = "<your API key>"
}

resource "honeycombio_trigger" "trigger" {
  name        = "Requests are slower than usuals"
  description = "Average duration of all requests for ThatSpecialTenant for the last 15 minutes."
  dataset     = "<your dataset>"

  disabled = false

  query {
    # exactly one calculation is required
    calculation {
      op     = "AVG"
      column = "duration_ms"
    }

    # zero or more filter blocks
    filter {
      column = "trace.parent_id"
      op     = "does-not-exist"
    }

    filter {
      column = "app.tenant"
      op     = "="
      value  = "ThatSpecialTenant"
    }

    # this can be ommited, AND is the default
    filter_combination = "AND"

    # also supported: breakdowns
  }

  frequency = 900 // in seconds, 15 minutes

  threshold {
    op    = ">"
    value = 100
  }

  # zero or more recipients
  recipient {
    type   = "email"
    target = "hello@example.com"
  }

  recipient {
    type   = "email"
    target = "bye@example.com"
  }
}
