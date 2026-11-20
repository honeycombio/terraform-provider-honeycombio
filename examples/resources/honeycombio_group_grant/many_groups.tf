data "honeycombio_auth_metadata" "current" {}

locals {
  payments_datasets = toset(["checkout", "billing", "refunds"])
  payments_groups   = toset(["payments-oncall", "finance"])
}

data "honeycombio_dataset" "payments" {
  for_each = local.payments_datasets
  slug     = each.key
}

resource "honeycombio_dataset_restriction" "payments" {
  for_each       = data.honeycombio_dataset.payments
  environment_id = data.honeycombio_auth_metadata.current.environment.id
  dataset_id     = each.value.hcid
}

resource "honeycombio_group" "payments" {
  for_each = local.payments_groups
  name     = each.key
}

# One grant for each pair of a Group and a Dataset
resource "honeycombio_group_grant" "payments" {
  for_each = {
    for pair in setproduct(local.payments_groups, local.payments_datasets) :
    "${pair[0]}/${pair[1]}" => { group = pair[0], dataset = pair[1] }
  }
  group_id   = honeycombio_group.payments[each.value.group].id
  dataset_id = honeycombio_dataset_restriction.payments[each.value.dataset].dataset_id
}
