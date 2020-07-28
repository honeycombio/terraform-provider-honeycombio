provider "honeycombio" {
  # Set api_key or the environment variable HONEYCOMBIO_APIKEY
  #api_key = "<your API key>"

  # Set dataset or the environment variable HONEYCOMBIO_DATASET
  #dataset = "<your dataset>"
}

resource "honeycombio_marker" "marker" {
  message = "Hello world!"
  type    = "deploys"
  url     = "https://www.honeycomb.io/"
}
