variable "dataset" {
  type = string
}

variable "slo_id" {
  type = string
}

data "honeycombio_recipient" "pd_prod" {
  type = "pagerduty"

  detail_filter {
    name  = "integration_name"
    value = "Prod On-Call"
  }
}

resource "honeycombio_burn_alert" "example_alert" {
  exhaustion_minutes = 60
  description        = "Burn alert description"
  dataset            = var.dataset
  slo_id             = var.slo_id

  recipient {
    id = data.honeycombio_recipient.pd_prod.id

    notification_details {
      pagerduty_severity = "critical"
    }
  }
}
