terraform {
  required_providers {
    honeycombio = {
      source = "honeycombio/honeycombio"
    }
  }
}

variable "dataset" {
  type = string
}

resource "honeycombio_dataset_definition" "test" {
  dataset    = var.dataset

  trace_id {
    name = "trace.trace_id"
    column_type = "column"
  }

  error {
    name = "err"
    column_type = "column"
  }
}