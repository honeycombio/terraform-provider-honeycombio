terraform {
  required_providers {
    honeycombio = {
      source = "honeycombio/honeycombio"
    }
  }
}

locals {
  # The people who can read the payments data
  member_emails = toset(["alice@example.com", "bob@example.com"])

  # Payments Datasets in a shared Environment: the Environment of the
  # provider's Configuration Key. The other Datasets in it stay open.
  dataset_slugs = toset(["checkout", "billing", "refunds"])

  # An Environment that holds only payments data. Every Dataset in it is
  # restricted, including Datasets added later.
  environment_name = "payments-production"
}

# 1. The Group and its members

resource "honeycombio_group" "payments" {
  name        = "payments"
  description = "Engineers who can read the payments data."
}

# the plan fails if an email matches no user on the Team
data "honeycombio_user" "payments" {
  for_each = local.member_emails
  email    = each.key
}

resource "honeycombio_group_members" "payments" {
  group_id = honeycombio_group.payments.id
  user_ids = [for u in data.honeycombio_user.payments : u.id]
}

# 2. Restrict single Datasets in the shared Environment

data "honeycombio_auth_metadata" "current" {}

data "honeycombio_dataset" "payments" {
  for_each = local.dataset_slugs
  slug     = each.key
}

resource "honeycombio_dataset_restriction" "payments" {
  for_each       = data.honeycombio_dataset.payments
  environment_id = data.honeycombio_auth_metadata.current.environment.id
  dataset_id     = each.value.hcid
}

# 3. Give the Group access to the Datasets and to the payments Environment

data "honeycombio_environment" "payments" {
  detail_filter {
    name  = "name"
    value = local.environment_name
  }
}

# The references to the restrictions make Terraform restrict each Dataset
# before it grants it, and revoke the grant before it removes the restriction.
resource "honeycombio_group_grants" "payments" {
  group_id        = honeycombio_group.payments.id
  dataset_ids     = [for r in honeycombio_dataset_restriction.payments : r.dataset_id]
  environment_ids = [data.honeycombio_environment.payments.id]
}

# 4. Restrict the payments Environment

# depends_on makes Terraform grant before it restricts, so the Group never
# loses access. On destroy, Terraform unrestricts before it revokes.
resource "honeycombio_environment_restriction" "payments" {
  environment_id = data.honeycombio_environment.payments.id

  depends_on = [honeycombio_group_grants.payments]
}
