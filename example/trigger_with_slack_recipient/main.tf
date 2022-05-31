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
  dataset = var.dataset
  type    = "slack"
  target  = "#honeycombio"
}

resource "honeycombio_query" "trigger-query" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.query.json
}

resource "honeycombio_trigger" "trigger" {
  name = "Requests are slower than usual"

  query_id = honeycombio_query.trigger-query.id
  dataset  = var.dataset

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
}
