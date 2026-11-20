data "honeycombio_auth_metadata" "current" {}

resource "honeycombio_dataset" "checkout" {
  name = "checkout"
}

resource "honeycombio_dataset_restriction" "checkout" {
  environment_id = data.honeycombio_auth_metadata.current.environment.id
  dataset_id     = honeycombio_dataset.checkout.hcid
}

resource "honeycombio_group" "payments" {
  name = "payments-oncall"
}

# The reference to the restriction makes Terraform restrict the Dataset
# before it grants it, and revoke the grant before it removes the restriction.
resource "honeycombio_group_grant" "checkout" {
  group_id   = honeycombio_group.payments.id
  dataset_id = honeycombio_dataset_restriction.checkout.dataset_id
}
