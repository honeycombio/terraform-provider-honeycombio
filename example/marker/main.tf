provider "honeycombio" {
  # You can also set the environment variable HONEYCOMBIO_APIKEY
  api_key = "<your API key>"

  # You can also set the environment variable HONEYCOMBIO_DATASET
  dataset = "<your dataset>"
}

variable "app_version" {
    type = string
}

resource "honeycombio_marker" "marker" {
  message = "deploy ${var.app_version}"
  type    = "deploys"
  url     = "https://www.honeycomb.io/"
}
