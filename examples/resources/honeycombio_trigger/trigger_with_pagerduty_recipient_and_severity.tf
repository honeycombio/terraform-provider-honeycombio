variable "dataset" {
  type = string
}

data "honeycombio_recipient" "pd_prod" {
  type = "pagerduty"

  detail_filter {
    name  = "integration_name"
    value = "Prod On-Call"
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
    id = data.honeycombio_recipient.pd_prod.id

    notification_details {
      pagerduty_severity = "warning"
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
