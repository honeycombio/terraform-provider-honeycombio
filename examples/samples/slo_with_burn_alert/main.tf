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

resource "honeycombio_derived_column" "request_latency_sli" {
  alias       = "sli.request_latency"
  description = "SLI: request latency less than 300ms"
  dataset     = var.dataset

  # heredoc also works
  expression = file("../sli/sli.request_latency.honeycomb")

  lifecycle {
    # in order to avoid potential conflicts with renaming the derived column
    # while in use by the SLO, we set create_before_destroy to true
    create_before_destroy = true
  }
}

resource "honeycombio_slo" "slo" {
  name              = "Latency SLO"
  description       = "example SLO"
  dataset           = var.dataset
  sli               = honeycombio_derived_column.request_latency_sli.alias
  target_percentage = 99.9
  time_period       = 30

  tags = {
    team = "web"
  }
}

resource "honeycombio_burn_alert" "burn_alert" {
  dataset            = var.dataset
  slo_id             = honeycombio_slo.slo.id
  exhaustion_minutes = 90

  recipient {
    type   = "slack"
    target = "#example-channel"
  }
}
