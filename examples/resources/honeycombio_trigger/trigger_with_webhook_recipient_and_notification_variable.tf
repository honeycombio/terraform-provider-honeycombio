variable "dataset" {
    type = string
}

data "honeycombio_recipient" "custom_webhook" {
    type = "webhook"

    detail_filter {
        name  = "name"
        value = "My Custom Webhook"
    }
}

data "honeycombio_query_specification" "example" {
    calculation {
        op     = "AVG"
        column = "duration_ms"
    }

    filter {
        column = "trace.parent_id"
        op     = "does-not-exist"
    }
}

resource "honeycombio_trigger" "example" {
    name        = "Requests are slower than usual"
    description = "Average duration of all requests for the last 10 minutes."

    query_json = data.honeycombio_query_specification.example.json
    dataset    = var.dataset

    frequency = 600 // in seconds, 10 minutes

    threshold {
        op             = ">"
        value          = 1000
        exceeded_limit = 3
    }

    recipient {
        id = data.honeycombio_recipient.custom_webhook.id

        notification_details {
            variable {
                name = "severity"
                value = "info"
            }
        }
    }

    evaluation_schedule {
        start_time = "13:00"
        end_time   = "21:00"

        days_of_week = [
            "monday",
            "wednesday",
            "friday"
        ]
    }

  tags = {
      team = "backend"
      env  = "production"
  }
}
