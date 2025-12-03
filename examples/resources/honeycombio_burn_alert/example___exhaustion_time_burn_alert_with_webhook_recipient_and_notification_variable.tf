variable "dataset" {
  type = string
}

variable "slo_id" {
  type = string
}

data "honeycombio_recipient" "custom_webhook" {
    type = "webhook"

    detail_filter {
        name  = "name"
        value = "My Custom Webhook"
    }
}

resource "honeycombio_burn_alert" "example_alert" {
    exhaustion_minutes = 60
    description        = "Burn alert description"
    dataset            = var.dataset
    slo_id             = var.slo_id

    dataset = var.dataset
    slo_id  = var.slo_id

    recipient {
      id = data.honeycombio_recipient.custom_webhook.id

      notification_details {
          variable {
              name = "severity"
              value = "info"
          }
      }
    }
}
