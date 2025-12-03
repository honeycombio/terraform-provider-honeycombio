variable "dataset" {
  type = string
}

variable "app_version" {
  type = string
}

resource "honeycombio_marker" "app_deploy" {
  message = "deploy ${var.app_version}"
  type    = "deploy"
  url     = "http://www.example.com/"

  dataset = var.dataset
}
