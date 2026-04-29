data "honeycombio_query_specification" "datapoints_total" {
  calculation {
    op = "COUNT_DATAPOINTS"
  }

  time_range  = 1800
  granularity = 60
}

data "honeycombio_query_specification" "datapoints_for_metric" {
  calculation {
    op     = "COUNT_DATAPOINTS"
    column = "app.cumulative"
  }

  breakdowns = ["node.name"]

  time_range  = 1800
  granularity = 60
}

data "honeycombio_query_specification" "histogram_events" {
  calculation {
    op     = "HISTOGRAM_COUNT"
    column = "app.histogram"
  }

  time_range  = 1800
  granularity = 60
}

# Saved queries + annotations so the queries can be referenced by boards and triggers
resource "honeycombio_query" "datapoints_total" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.datapoints_total.json
}

resource "honeycombio_query_annotation" "datapoints_total" {
  dataset     = var.dataset
  query_id    = honeycombio_query.datapoints_total.id
  name        = "Total datapoints"
  description = "Total datapoints reported across all metrics"
}

resource "honeycombio_query" "datapoints_for_metric" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.datapoints_for_metric.json
}

resource "honeycombio_query_annotation" "datapoints_for_metric" {
  dataset     = var.dataset
  query_id    = honeycombio_query.datapoints_for_metric.id
  name        = "Datapoint volume by node"
  description = "Per-node datapoint volume for app.cumulative over the last 30 minutes"
}

resource "honeycombio_query" "histogram_events" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.histogram_events.json
}

resource "honeycombio_query_annotation" "histogram_events" {
  dataset     = var.dataset
  query_id    = honeycombio_query.histogram_events.id
  name        = "Histogram event count"
  description = "Number of events recorded in app.histogram"
}
