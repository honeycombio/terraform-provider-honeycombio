variable "dataset" {
  type = string
}

resource "honeycombio_derived_column" "any_error" {
  alias       = "dc.any_error"
  expression  = "COALESCE($error.message, $app.legacy_error)"

  description = "Collapse OTEL semantic convention and legacy error messages into one field"
  dataset = var.dataset
}
