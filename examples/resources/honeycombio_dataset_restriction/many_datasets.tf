# The Environment of the provider's API key
data "honeycombio_auth_metadata" "current" {}

locals {
  payments_datasets = toset(["checkout", "billing", "refunds"])
}

# Turns each slug into the Dataset's ID
data "honeycombio_dataset" "payments" {
  for_each = local.payments_datasets
  slug     = each.key
}

resource "honeycombio_dataset_restriction" "payments" {
  for_each       = data.honeycombio_dataset.payments
  environment_id = data.honeycombio_auth_metadata.current.environment.id
  dataset_id     = each.value.hcid
}
