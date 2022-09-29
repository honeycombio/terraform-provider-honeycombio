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
  type  = "deploy"
  color = "${var.color}"

  dataset = var.dataset
}
