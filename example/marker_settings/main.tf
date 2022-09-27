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

variable "color" {
  type = string
}

resource "honeycombio_marker_setting" "markerSetting" {
  color = "deploy ${var.color}"
  type    = "deploy"

  dataset = var.dataset
}
