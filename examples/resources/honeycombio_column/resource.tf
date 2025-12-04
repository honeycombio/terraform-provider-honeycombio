variable "dataset" {
  type = string
}

resource "honeycombio_column" "duration_ms" {
  name        = "duration_ms_log10"
  type        = "float"
  description = "Duration of the trace"

  dataset = var.dataset
}
