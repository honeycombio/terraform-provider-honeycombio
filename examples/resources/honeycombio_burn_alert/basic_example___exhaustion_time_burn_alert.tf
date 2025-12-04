variable "dataset" {
  type = string
}

variable "slo_id" {
  type = string
}

resource "honeycombio_burn_alert" "example_alert" {
  alert_type         = "exhaustion_time"
  exhaustion_minutes = 480
  description        = "Exhaustion burn alert description"

  dataset = var.dataset
  slo_id  = var.slo_id

  # one or more recipients
  recipient {
    type   = "email"
    target = "hello@example.com"
  }

  recipient {
    type   = "slack"
    target = "#example-channel"
  }
}
