# search for a Slack recipient with channel name "honeycomb-triggers"
data "honeycombio_recipient" "slack" {
  type    = "slack"

  detail_filter {
    name  = "channel"
    value = "#honeycomb-triggers"
  }
}

data "honeycombio_query_specification" "example" {
  calculation {
    op     = "AVG"
    column = "duration_ms"
  }
}

resource "honeycombio_trigger" "example" {
  name        = "Requests are slower than usual"

  query_json = data.honeycombio_query_specification.example.json
  dataset    = var.dataset

  threshold {
    op    = ">"
    value = 1000
  }

  recipient {
    type   = "email"
    target = "hello@example.com"
  }

  # add an already existing recipient
  recipient {
    id = data.honeycombio_recipient.slack.id
  }
}
