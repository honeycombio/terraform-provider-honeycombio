data "honeycombio_auth_metadata" "current" {}

locals {
  payments_datasets = toset(["checkout", "billing", "refunds"])
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
  name = "payments-oncall"
}

resource "honeycombio_group_grant" "payments" {
  for_each   = honeycombio_dataset_restriction.payments
  group_id   = honeycombio_group.payments.id
  dataset_id = each.value.dataset_id
}

# Warns when the Group has grants that this configuration does not declare,
# such as grants made in the UI. It never revokes them.
check "payments_grants" {
  data "honeycombio_group_grants" "payments" {
    group_id = honeycombio_group.payments.id

    # read the grants after Terraform changes them
    depends_on = [honeycombio_group_grant.payments]
  }

  assert {
    condition = length(setsubtract(
      data.honeycombio_group_grants.payments.dataset_ids,
      [for g in honeycombio_group_grant.payments : g.dataset_id],
    )) == 0
    error_message = "payments-oncall has grants that Terraform does not declare: ${join(", ", setsubtract(
      data.honeycombio_group_grants.payments.dataset_ids,
      [for g in honeycombio_group_grant.payments : g.dataset_id],
    ))}"
  }
}
