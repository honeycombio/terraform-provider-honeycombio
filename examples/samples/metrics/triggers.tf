# Fires when no datapoints are reported, typically a sign that a metrics
# emitter has stopped.
resource "honeycombio_trigger" "datapoint_volume_drop" {
  name        = "Datapoint volume drop"
  description = "Fires when no datapoints are reported."
  dataset     = var.dataset

  query_json = data.honeycombio_query_specification.datapoints_total.json

  frequency = 900

  threshold {
    op    = "<"
    value = 1
  }

  recipient {
    type   = "email"
    target = "hello@example.com"
  }
}

# Fires when histogram event volume falls below an expected floor.
resource "honeycombio_trigger" "histogram_event_drop" {
  name        = "Histogram event volume drop"
  description = "Fires when histogram event volume falls below an expected floor."
  dataset     = var.dataset

  query_json = data.honeycombio_query_specification.histogram_events.json

  frequency = 900

  threshold {
    op    = "<"
    value = 1
  }

  recipient {
    type   = "email"
    target = "hello@example.com"
  }
}
