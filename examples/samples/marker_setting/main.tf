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

variable "type" {
  type = string
}

variable "color" {
  type = string
}

resource "honeycombio_marker_setting" "markerSetting" {
  type  = var.type
  color = var.color
  dataset = var.dataset
}
