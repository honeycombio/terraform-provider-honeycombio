resource "honeycombio_dataset_definition" "trace_id" {
  dataset = var.dataset

  name   = "trace_id"
  column = "trace.trace_id"
}
