variable "dataset" {
  type = string
}

data "honeycombio_recipient" "pd_checkout" {
  type = "pagerduty"

  detail_filter {
    name  = "integration_name"
    value = "Checkout On-Call"
  }
}

data "honeycombio_recipient" "slack_oncall" {
  type = "slack"

  detail_filter {
    name  = "channel"
    value = "#oncall"
  }
}

data "honeycombio_query_specification" "errors_by_service" {
  calculation {
    op = "COUNT"
  }

  filter {
    column = "error"
    op     = "="
    value  = "true"
  }

  # per-group routing requires the query to group by at least one column
  breakdowns = ["service.name"]

  time_range = 1800
}

resource "honeycombio_trigger" "errors_by_service" {
  name        = "Errors by service"
  description = "Routes each service's errors to the team which owns it."

  query_json = data.honeycombio_query_specification.errors_by_service.json
  dataset    = var.dataset

  # required for any per-group recipient routing: it is what makes Honeycomb
  # evaluate and resolve each group's state independently
  alert_type = "on_group_change"

  frequency = 1800

  threshold {
    op    = ">"
    value = 100
  }

  # the checkout team only hears about its own services, and gets a separate
  # PagerDuty incident per service so they can be resolved independently
  recipient {
    id = data.honeycombio_recipient.pd_checkout.id

    group_filter = {
      "service.name" = ["checkout", "cart"]
    }

    pagerduty_per_group_incidents = true
  }

  # a recipient with no group_filter is a catch-all: it is notified about every
  # group which crosses the threshold
  recipient {
    id = data.honeycombio_recipient.slack_oncall.id
  }
}
