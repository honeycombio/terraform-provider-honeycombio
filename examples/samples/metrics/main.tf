terraform {
  required_providers {
    honeycombio = {
      source = "honeycombio/honeycombio"
    }
  }
}

variable "dataset" {
  description = "A Honeycomb metrics dataset. Calculations like COUNT_DATAPOINTS and HISTOGRAM_COUNT are only valid against metrics datasets."
  type        = string
}
