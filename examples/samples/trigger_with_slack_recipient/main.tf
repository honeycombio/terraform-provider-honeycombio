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
}

// Since it is not possible to add Slack recipient using the API, we first
// search for a trigger that already has a Slack recipient. If there is none,
// this will fail during the plan phase.
data "honeycombio_recipient" "slack" {
  type = "slack"

  detail_filter {
    name  = "channel"
    value = "#honeycombio"
  }
}

resource "honeycombio_trigger" "trigger" {
  name = "Requests are slower than usual"

  query_json = data.honeycombio_query_specification.query.json
  dataset    = var.dataset

  threshold {
    op    = ">"
    value = 1000
  }

  // Add a recipient by ID, this will not create a new recipient.
  recipient {
    id = data.honeycombio_recipient.slack.id
  }
  recipient {
    type   = "email"
    target = "hello@example.com"
  }

  frequency = 1800

  tags = {
    team = "backend"
    env  = "production"
  }
}
