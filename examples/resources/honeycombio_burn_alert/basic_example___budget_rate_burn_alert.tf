variable "dataset" {
  type = string
}

variable "slo_id" {
  type = string
}

resource "honeycombio_burn_alert" "example_alert" {
  alert_type                   = "budget_rate"
  budget_rate_window_minutes   = 480
  budget_rate_decrease_percent = 1
  description                  =  "my example description"

  dataset = var.dataset
  slo_id  = var.slo_id

  # one or more recipients
  recipient {
    type   = "webhook"
    target = "name of the webhook"
  }
}
