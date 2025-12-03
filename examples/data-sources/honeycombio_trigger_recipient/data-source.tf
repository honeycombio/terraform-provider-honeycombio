variable "dataset" {
  type = string
}

# search for a trigger recipient of type "slack" and target "honeycomb-triggers" in the given dataset
data "honeycombio_trigger_recipient" "slack" {
  dataset = var.dataset
  type    = "slack"
  target  = "honeycomb-triggers"
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
    id = data.honeycombio_trigger_recipient.slack.id
  }
}
