# The Environment of the provider's API key
data "honeycombio_auth_metadata" "current" {}

resource "honeycombio_dataset" "checkout" {
  name = "checkout"
}

resource "honeycombio_dataset_restriction" "checkout" {
  environment_id = data.honeycombio_auth_metadata.current.environment.id
  dataset_id     = honeycombio_dataset.checkout.hcid
}
