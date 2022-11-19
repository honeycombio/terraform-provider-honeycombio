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

resource "honeycombio_dataset_definition" "example" {
  for_each = {
    "span_id"      = "trace.span_id",
    "trace_id"     = "trace.trace_id",
    "error"        = "error",
    "status"       = "http.status_code",
    "name"         = "name",
    "parent_id"    = "trace.parent_id",
    "route"        = "http.route",
    "service_name" = "service.name",
    "duration_ms"  = "duration_ms"
  }

  dataset = var.dataset
  name    = each.key
  column  = each.value
}
