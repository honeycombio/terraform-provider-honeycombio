provider "honeycombio" {
  # You can also set the environment variable HONEYCOMBIO_APIKEY
  api_key = "<your API key>"
}

data "honeycombio_query" "query" {
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

  filter_combination = "AND"

  # also supported: breakdowns
}

resource "honeycombio_trigger" "trigger" {
  name        = "Requests are slower than usuals"
  description = "Average duration of all requests for ThatSpecialTenant for the last 15 minutes."
  dataset     = "<your dataset>"

  disabled = false

  query_json = data.honeycombio_query.query.json

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
