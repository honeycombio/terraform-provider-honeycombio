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

# A time-range marker spanning a maintenance window.
resource "honeycombio_marker" "maintenance" {
  message    = "scheduled maintenance"
  type       = "maintenance"
  start_time = 1700000000
  end_time   = 1700003600

  dataset = var.dataset
}
