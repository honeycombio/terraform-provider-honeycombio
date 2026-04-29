# Flexible board surfacing the metrics queries side-by-side.
resource "honeycombio_flexible_board" "metrics_overview" {
  name        = "Metrics overview"
  description = "Datapoint and histogram volume for the metrics pipeline."

  panel {
    type = "query"
    position {
      x_coordinate = 0
      y_coordinate = 0
      width        = 6
      height       = 6
    }
    query_panel {
      query_id            = honeycombio_query.datapoints_total.id
      query_annotation_id = honeycombio_query_annotation.datapoints_total.id
      query_style         = "graph"
    }
  }

  panel {
    type = "query"
    position {
      x_coordinate = 6
      y_coordinate = 0
      width        = 6
      height       = 6
    }
    query_panel {
      query_id            = honeycombio_query.datapoints_for_metric.id
      query_annotation_id = honeycombio_query_annotation.datapoints_for_metric.id
      query_style         = "graph"
    }
  }

  panel {
    type = "query"
    position {
      x_coordinate = 0
      y_coordinate = 6
      width        = 12
      height       = 6
    }
    query_panel {
      query_id            = honeycombio_query.histogram_events.id
      query_annotation_id = honeycombio_query_annotation.histogram_events.id
      query_style         = "graph"
    }
  }
}
