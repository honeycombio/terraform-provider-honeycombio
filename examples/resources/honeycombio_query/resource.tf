variable "dataset" {
  type = string
}

resource "honeycombio_derived_column" "duration_ms_log10" {
  alias       = "duration_ms_log10"
  expression  = "LOG10($duration_ms)"
  description = "LOG10 of duration_ms"

  dataset = var.dataset
}

data "honeycombio_query_specification" "example" {
  calculation {
    op     = "P90"
    column = "duration_ms"
  }

  calculation {
    op     = "HEATMAP"
    column = $honeycombio_derived_column.duration_ms_log10.alias
  }

  filter {
    column = "duration_ms"
    op     = "exists"
  }

}

resource "honeycombio_query" "example" {
  dataset    = var.dataset
  query_json = data.honeycombio_query_specification.example.json

  lifecycle {
    replace_triggered_by = [
      // re-create the query if the derived column is changed
      // to ensure we're using the latest definition
      honeycombio_derived_column.duration_ms_log10
    ]
  }
}
