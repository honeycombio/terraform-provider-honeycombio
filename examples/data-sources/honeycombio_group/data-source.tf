# a Group made in the UI, or by another Terraform configuration
data "honeycombio_group" "payments" {
  name = "payments-oncall"
}

# a Dataset that is already restricted
data "honeycombio_dataset" "checkout" {
  slug = "checkout"
}

resource "honeycombio_group_grant" "checkout" {
  group_id   = data.honeycombio_group.payments.id
  dataset_id = data.honeycombio_dataset.checkout.hcid
}
